package engine

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Options governs the processing behavior.
type Options struct {
	Input           string
	Output          string
	Files           string
	OutputDir       string
	StartTime       time.Time
	EndTime         time.Time
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
	maskNum     = 10
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
		if err := Run(cur); err != nil {
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

// Run executes the full pipeline: read -> process -> write.
func Run(opt Options) error {
	ctx, err := readPDF(opt.Input)
	if err != nil {
		return err
	}

	processPDF(ctx, opt)

	return writePDF(ctx, opt.Output)
}

func writePDF(ctx *model.Context, output string) error {
	out := uniqueOutputName(output)
	return api.WriteContextFile(ctx, out)
}

// readPDF reads the PDF into a pdfcpu Context.
func readPDF(input string) (*model.Context, error) {
	ctx, err := api.ReadContextFile(input)
	if err != nil {
		return nil, fmt.Errorf("read context file: %w", err)
	}
	return ctx, nil
}

// processPDF processes every page and returns all created OCG refs for the document.
func processPDF(ctx *model.Context, opt Options) error {

	conf := model.NewDefaultConfiguration()
	ctx.Configuration = conf
	// 规范化时间：如果为零，设置为很早或很晚的时间，避免 JS 逻辑出错
	start := normalizeTime(opt.StartTime, true)
	end := normalizeTime(opt.EndTime, false)

	for p := 1; p <= ctx.PageCount; p++ {
		err := processPageStructured(ctx, p, opt, maskNum)
		if err != nil {
			return fmt.Errorf("process page %d: %w", p, err)
		}
	}
	// 处理加密
	processEncryption(ctx, opt)
	// 处理权限
	processPermissions(ctx, opt)
	// 注入 OpenAction JS
	injectOpenActionJS(ctx, start, end, opt.ExperiredText, opt.UnsupportedText)
	return nil
}

// 处理加密
func processEncryption(ctx *model.Context, opt Options) {
	ctx.Cmd = model.ENCRYPT
	ctx.UserPW = strings.TrimSpace(opt.UserPassword)
	ctx.OwnerPW = fmt.Sprintf("%s%s", strings.TrimSpace(opt.OwnerPassword), ownerPWMask)
	ctx.EncryptUsingAES = true
	ctx.EncryptKeyLength = 256
}

// 处理权限
func processPermissions(ctx *model.Context, opt Options) {
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
}

// processPageStructured implements the per-page workflow with single-responsibility steps.
func processPageStructured(ctx *model.Context, pageNum int, opt Options, maskNum int) error {
	// 1. 获取 pageDict
	pageDict, _, _, err := ctx.PageDict(pageNum, true)
	if err != nil {
		return fmt.Errorf("get page dict: %w", err)
	}
	if pageDict == nil {
		return fmt.Errorf("page dict is nil")
	}
	// 2. 提取 pageContent XObject (原始内容合并为 Form XObject)
	normalXObj, err := extractPageContentAsXObject(ctx, pageDict, pageNum)
	if err != nil {
		return fmt.Errorf("extract page content as xobject: %w", err)
	}
	// 3. 处理 mask：创建 mask OCG，创建 mask XObject
	maskOCGs, maskXObjs, err := buildMaskOCGsAndXObjectsForPage(ctx, pageDict, pageNum, maskNum)
	if err != nil {
		return fmt.Errorf("build mask ocgs and xobjects for page: %w", err)
	}
	insertOCPropertiesOCGs(ctx, maskOCGs)
	// expired OCG and XObject
	expiredOCG, expiredXObj, err := buildExpiredOCGAndXObject(ctx, pageDict, pageNum, opt.ExperiredText)
	if err != nil {
		return fmt.Errorf("build expired ocg and xobject: %w", err)
	}
	insertOCPropertiesOCGs(ctx, []*types.IndirectRef{expiredOCG})
	// expired_mask OCG and XObject
	expiredMaskOCG, expiredMaskXObj, err := buildExpiredMaskOCGAndXObject(ctx, pageDict, pageNum)
	if err != nil {
		return fmt.Errorf("build expired mask ocg and xobject: %w", err)
	}
	insertOCPropertiesOCGs(ctx, []*types.IndirectRef{expiredMaskOCG})
	// 4. 处理 fallback：创建 fallback OCG，创建 fallback XObject
	fallbackOCG, fallbackXObj, err := buildFallbackOCGAndXObject(ctx, pageDict, pageNum, opt.UnsupportedText)
	if err != nil {
		return fmt.Errorf("build fallback ocg and xobject: %w", err)
	}
	insertOCPropertiesOCGs(ctx, []*types.IndirectRef{fallbackOCG})
	// 5. 绑定 mask OCGResources, 绑定 fallback OCGResources
	injectOCGResources(ctx, pageDict, pageNum, normalXObj, maskXObjs, fallbackXObj, maskOCGs, fallbackOCG)
	// expired_mask OCGResources
	injectExpiredMaskOCGResources(ctx, pageDict, pageNum, expiredMaskXObj, expiredMaskOCG)
	// expired OCGResources
	injectExpiredOCGResources(ctx, pageDict, pageNum, expiredXObj, expiredOCG)
	// 6. Rewrite 页面：插入引用 mask & fallback
	if err := rewritePageWithMasksAndFallback(ctx, pageDict, pageNum, maskXObjs); err != nil {
		return fmt.Errorf("rewrite page with masks: %w", err)
	}
	return nil
}

// uniqueOutputName returns a non-colliding filename. If target exists, append _副本XXXX
func uniqueOutputName(target string) string {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return target
	}
	ext := filepath.Ext(target)
	base := target[:len(target)-len(ext)]
	rand.Seed(time.Now().UnixNano())
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s_副本%04d%s", base, suffix, ext)
}
