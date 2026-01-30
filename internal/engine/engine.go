package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Options struct {
	Input           string
	Output          string
	Files           string
	OutputDir       string
	StartTime       time.Time
	EndTime         time.Time
	Watermark       string
	ExperiredText   string
	UnsupportedText string
	PwdEnabled      bool
	UserPassword    string
	OwnerPassword   string

	// 打印/复制
	AllowedPrint bool
	AllowedCopy  bool
	// 编辑
	AllowedEdit bool
	// 转换
	AllowedConvert bool
}

const (
	ownerPWMask = "cg"
)

// 批量处理入口
func RunBatch(opt Options) error {
	// 解析 opt.Files，按;或,分割
	if opt.Files == "" {
		return fmt.Errorf("no files provided for batch run")
	}
	seps := []string{";", ","}
	raw := opt.Files
	for _, s := range seps {
		raw = strings.ReplaceAll(raw, s, ";")
	}
	parts := strings.Split(raw, ";")
	var firstErr error
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// 为当前文件构造专属 Options
		cur := opt
		cur.Input = p
		// 构造输出路径：优先 OutputDir，其次与输入同目录
		cur.Output = filepath.Join(cur.OutputDir, filepath.Base(p))
		fmt.Printf("Running batch for %s -> %s\n", cur.Input, cur.Output)
		if err := exec(cur); err != nil {
			fmt.Printf("RunBatch error for %s: %v\n", p, err)
			if firstErr == nil {
				firstErr = err
			}
			// continue processing remaining files
		} else {
			fmt.Printf("RunBatch completed for %s\n", p)
		}
	}
	return firstErr
}

func Run(opt Options) error {
	return exec(opt)
}

func exec(opt Options) error {
	//conf := model.NewDefaultConfiguration()
	// 规范化时间：如果为零，设置为很早或很晚的时间，避免 JS 逻辑出错
	start := normalizeTime(opt.StartTime, true)
	end := normalizeTime(opt.EndTime, false)

	fmt.Printf("Applying time-limited two-layer protection from %s to %s\n", start.Format(time.RFC3339), end.Format(time.RFC3339))
	ctx, err := api.ReadContextFile(opt.Input)
	if err != nil {
		fmt.Printf("read context file: %v\n", err)
		return err
	}
	conf := model.NewDefaultConfiguration()
	ctx.Configuration = conf
	// 1. 创建 OCG
	normalOCG, _ := ensureOCGs(ctx)

	// 2. 每一页加 Fallback Widget（遮罩）
	// 2. Pages
	for p := 1; p <= ctx.PageCount; p++ {
		pageDict, _, _, err := ctx.PageDict(p, true)
		if err != nil {
			fmt.Printf("get page dict for page %d: %v\n", p, err)
			return err
		}
		if pageDict == nil {
			continue
		}
		// 1.把原页面内容封装成 Form XObject
		pxd, err := extractPageContentAsXObject(ctx, pageDict, p, normalOCG)
		if err != nil {
			fmt.Printf("extract page content as XObject for page %d: %v\n", p, err)
			return err
		}
		// 2.重写页面 Contents（只画 fallback），然后在末尾追加 OCG 包裹的 Do NormalContent
		err = setFallbackContent(ctx, pageDict, opt.UnsupportedText)
		if err != nil {
			fmt.Printf("set fallback content for page %d: %v\n", p, err)
			return err
		}
		// 3.把 NormalContent XObject 加回页面 Resources
		attachXObjectToPage(pageDict, pxd, normalOCG)
		// 4.在 Page Contents 末尾追加一个流，执行 /NormalContent Do 并由 OCG 控制
		if err := appendDoNormalContent(ctx, pageDict); err != nil {
			fmt.Printf("append Do NormalContent for page %d: %v\n", p, err)
			return err
		}
		// 5.创建隐藏 Widget，仅用于兼容或备用（不作显示切换）
		//createUnlockWidget(ctx, pageDict, p, pxd)
		/* if err := addFallbackWidget(ctx, pageDict, p, fallbackOCG, opt.Watermark); err != nil {
			fmt.Printf("add fallback widget to page %d: %v\n", p, err)
			return err
		} */
	}

	// 3. 注入 JS（只隐藏 Widget），并传递提示文本
	injectOpenActionJS(ctx, start, end, opt.ExperiredText, opt.UnsupportedText)
	// 处理密码
	if opt.UserPassword != "" {
		ctx.Cmd = model.ENCRYPT
		ctx.UserPW = strings.TrimSpace(opt.UserPassword)
		ctx.OwnerPW = fmt.Sprintf("%s%s", strings.TrimSpace(opt.OwnerPassword), ownerPWMask)
		ctx.EncryptUsingAES = true
		ctx.EncryptKeyLength = 256
	}
	// 处理权限
	var permissions model.PermissionFlags = 0xF0C3 // PermissionsNone - 禁止所有操作

	if opt.AllowedPrint {
		permissions = model.PermissionsPrint
	}
	if opt.AllowedCopy {
		permissions = model.PermissionExtract
	}
	if opt.AllowedEdit {
		permissions = model.PermissionModify
	}
	if opt.AllowedConvert {
		permissions = model.PermissionAssembleRev3
	}
	ctx.Permissions = permissions
	fmt.Println("Applying time-limited two-layer protection completed.")
	return api.WriteContextFile(ctx, opt.Output)
}

func normalizeTime(t time.Time, isStart bool) time.Time {
	if !t.IsZero() {
		return t
	}
	if isStart {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}
