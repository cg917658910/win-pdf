package engine

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdffont "github.com/pdfcpu/pdfcpu/pkg/font"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Options governs the processing behavior.
type Options struct {
	Input            string
	Output           string
	Files            string
	OutputDir        string
	StartTime        string
	EndTime          string
	ExperiredText    string
	UnsupportedText  string
	PwdEnabled       bool
	UserPassword     string
	OwnerPassword    string
	WatermarkEnabled bool
	WatermarkText    string
	WatermarkDesc    string
	WatermarkTiled   bool
	WatermarkSpacing float64

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
	maskNum     = 5
)

// 批量处理入口
func RunBatch(opt Options) (successCount int, err error) {
	if opt.Files == "" {
		return successCount, fmt.Errorf("no files provided for batch run")
	}

	raw := strings.ReplaceAll(opt.Files, ",", ";")
	parts := strings.Split(raw, ";")
	files := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			files = append(files, p)
		}
	}
	if len(files) == 0 {
		return successCount, fmt.Errorf("no valid files provided for batch run")
	}

	startedAt := time.Now()
	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(files) {
		workerCount = len(files)
	}

	var (
		firstErr error
		mu       sync.Mutex
		wg       sync.WaitGroup
	)
	jobs := make(chan string, len(files))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				cur := opt
				cur.Input = p
				cur.Output = filepath.Join(cur.OutputDir, filepath.Base(p))
				fmt.Printf("Running batch for %s -> %s\n", cur.Input, cur.Output)

				if runErr := Run(cur); runErr != nil {
					fmt.Printf("RunBatch error for %s: %v\n", p, runErr)
					mu.Lock()
					if firstErr == nil {
						firstErr = runErr
					}
					mu.Unlock()
					continue
				}

				fmt.Printf("RunBatch completed for %s\n", p)
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	for _, p := range files {
		jobs <- p
	}
	close(jobs)
	wg.Wait()

	elapsed := time.Since(startedAt)
	fmt.Printf("RunBatch finished in %s: %d/%d files processed successfully.\n", elapsed.Round(time.Second), successCount, len(files))
	return successCount, firstErr
}

// Run executes the full pipeline: read -> process -> write.
func Run(opt Options) error {
	ctx, err := readPDF(opt.Input)
	if err != nil {
		//  pdfcpu:
		if strings.Contains(err.Error(), "pdfcpu:") {
			return fmt.Errorf("此文档已加密过，不能再次加密！")
		}
		return err
	}

	processPDF(ctx, opt)

	return writePDF(ctx, opt.Output)
}

func writePDF(ctx *model.Context, output string) error {
	if err := api.OptimizeContext(ctx); err != nil {
		return fmt.Errorf("optimize context: %w", err)
	}
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
	if err := applyWatermarkToOriginalContent(ctx, opt); err != nil {
		return fmt.Errorf("apply watermark: %w", err)
	}
	// 规范化时间：如果为零，设置为很早或很晚的时间，避免 JS 逻辑出错
	startTime, endTime, err := parseOptionTimes(opt.StartTime, opt.EndTime)
	if err != nil {
		return err
	}
	start := normalizeTime(startTime, true)
	end := normalizeTime(endTime, false)
	// 处理页面
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

func parseOptionTimes(startStr, endStr string) (time.Time, time.Time, error) {
	startStr = strings.TrimSpace(startStr)
	endStr = strings.TrimSpace(endStr)
	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start or end time is empty")
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse start time: %w", err)
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse end time: %w", err)
	}
	return start, end, nil
}

// applyWatermarkToOriginalContent adds watermark into the original page content stream only.
// This runs before we extract NormalContent into a Form XObject, so watermark becomes part of NormalContent.
func applyWatermarkToOriginalContent(ctx *model.Context, opt Options) error {
	if !opt.WatermarkEnabled || strings.TrimSpace(opt.WatermarkText) == "" {
		return nil
	}
	desc := strings.TrimSpace(opt.WatermarkDesc)
	desc = ensureCJKFontForWatermark(opt.WatermarkText, desc)
	fmt.Printf("Applying watermark to original content with desc: %s\n", desc)
	wm, err := api.TextWatermark(opt.WatermarkText, desc, true, false, types.POINTS)
	if err != nil {
		return err
	}
	if !opt.WatermarkTiled {
		return pdfcpu.AddWatermarks(ctx, nil, wm)
	}
	return addTiledWatermarks(ctx, wm, opt.WatermarkSpacing)
}

func addTiledWatermarks(ctx *model.Context, base *model.Watermark, spacing float64) error {
	step := spacing
	if step <= 0 {
		step = 200
	}
	const maxPerPage = 400
	m := map[int][]*model.Watermark{}
	for p := 1; p <= ctx.PageCount; p++ {
		pageDict, _, _, err := ctx.PageDict(p, true)
		if err != nil || pageDict == nil {
			return fmt.Errorf("get page dict: %w", err)
		}
		mb := getPageMediaBox(pageDict)
		x0 := numToFloat(mb[0])
		y0 := numToFloat(mb[1])
		x1 := numToFloat(mb[2])
		y1 := numToFloat(mb[3])
		pageW := x1 - x0
		pageH := y1 - y0
		if pageW <= 0 || pageH <= 0 {
			continue
		}
		count := 0
		startX := x0 + step/2
		startY := y0 + step/2
		endX := x1 - step/2
		endY := y1 - step/2

		// If spacing is larger than page size, tiling loops below won't produce any
		// watermark. Ensure we still have one watermark centered on the page.
		if startX > endX || startY > endY {
			wm := new(model.Watermark)
			*wm = *base
			wm.Objs = types.IntSet{}
			wm.Pos = types.BottomLeft
			wm.Dx = types.ToUserSpace(pageW/2, wm.InpUnit)
			wm.Dy = types.ToUserSpace(pageH/2, wm.InpUnit)
			m[p] = append(m[p], wm)
			continue
		}

		for x := startX; x <= endX; x += step {
			for y := startY; y <= endY; y += step {
				wm := new(model.Watermark)
				*wm = *base
				// ensure each watermark has its own object/cache bookkeeping
				wm.Objs = types.IntSet{}
				wm.Pos = types.BottomLeft
				wm.Dx = types.ToUserSpace(x-x0, wm.InpUnit)
				wm.Dy = types.ToUserSpace(y-y0, wm.InpUnit)
				m[p] = append(m[p], wm)
				count++
				if count >= maxPerPage {
					break
				}
			}
			if count >= maxPerPage {
				break
			}
		}
	}

	// pdfcpu's slice-map path resolves text fonts without script info.
	// For CJK no-embed watermarks (ScriptName set), add per page/per watermark
	// to preserve script-based font encoding.
	if strings.TrimSpace(base.ScriptName) != "" {
		for p, wms := range m {
			sel := types.IntSet{p: true}
			for _, wm := range wms {
				if err := pdfcpu.AddWatermarks(ctx, sel, wm); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return pdfcpu.AddWatermarksSliceMap(ctx, m)
}

func numToFloat(o types.Object) float64 {
	switch v := o.(type) {
	case types.Float:
		return float64(v)
	case types.Integer:
		return float64(v)
	default:
		return 0
	}
}

func ensureCJKFontForWatermark(text, desc string) string {
	if !hasNonASCII(text) {
		return desc
	}
	items := parseWatermarkDesc(desc)
	fontKey, fontVal := findFontParam(items)
	if fontVal != "" && pdffont.IsUserFont(fontVal) {
		return joinWatermarkDesc(items)
	}
	// If no user font set or core font is used, pick a CJK-capable user font.
	cjkFont := pickCJKUserFont()
	if cjkFont == "" {
		fmt.Printf("No user CJK font installed; watermark text may not render.\n")
		return joinWatermarkDesc(items)
	}
	if fontKey == "" {
		items = append(items, wmItem{key: "fontname", raw: "fontname:" + cjkFont})
	} else {
		for i := range items {
			if items[i].key == fontKey {
				items[i].raw = fontKey + ":" + cjkFont
				break
			}
		}
	}
	return joinWatermarkDesc(items)
}

func hasNonASCII(s string) bool {
	for _, r := range s {
		if r > 0x7F {
			return true
		}
	}
	return false
}

type wmItem struct {
	key string
	raw string
}

func parseWatermarkDesc(desc string) []wmItem {
	parts := strings.Split(desc, ",")
	items := make([]wmItem, 0, len(parts))
	for _, p := range parts {
		raw := strings.TrimSpace(p)
		if raw == "" {
			continue
		}
		key := strings.ToLower(raw)
		if i := strings.IndexAny(raw, ":="); i >= 0 {
			key = strings.ToLower(strings.TrimSpace(raw[:i]))
		}
		items = append(items, wmItem{key: key, raw: raw})
	}
	return items
}

func joinWatermarkDesc(items []wmItem) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		if it.raw != "" {
			parts = append(parts, it.raw)
		}
	}
	return strings.Join(parts, ", ")
}

func findFontParam(items []wmItem) (string, string) {
	for _, it := range items {
		switch it.key {
		case "font", "fontname", "fo":
			if i := strings.IndexAny(it.raw, ":="); i >= 0 {
				return it.key, strings.TrimSpace(it.raw[i+1:])
			}
			return it.key, ""
		}
	}
	return "", ""
}

func pickCJKUserFont() string {
	preferred := []string{
		"MicrosoftYaHei",
		"MicrosoftYaHeiUI",
		"DengXian",
		"DengXian-Regular",
		"SimSun",
		"SimHei",
		"FangSong",
		"KaiTi",
	}
	userFonts := pdffont.UserFontNames()
	if len(userFonts) == 0 {
		return ""
	}
	for _, p := range preferred {
		for _, f := range userFonts {
			if strings.EqualFold(p, f) {
				return f
			}
		}
	}
	return userFonts[0]
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
	// 以“默认禁止全部操作”为基础，再按选项开放权限
	var permissions model.PermissionFlags = 0xF0C3 // PermissionsNone

	if opt.AllowedPrint {
		permissions |= model.PermissionsPrint
	}
	if opt.AllowedCopy || opt.AllowedConvert {
		permissions |= (model.PermissionExtract + model.PermissionExtractRev3)
	}
	if opt.AllowedEdit {
		permissions |= (model.PermissionModify + model.PermissionAssembleRev3 + model.PermissionFillRev3 + model.PermissionModAnnFillForm)
	}
	/* if opt.AllowedConvert {
		permissions |= model.PermissionExtract
	} */
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
