package engine

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// BuildTextXObject 创建一个用于绘制文本的 Form XObject。
// 中文及所有非 ASCII 字符使用 CJK Type0 字体 F1 以 UTF-16BE 输出；
// 纯 ASCII 文本使用 F2 + WinAnsi，一次性输出对应段，兼顾中英文混排与字距效果。
func buildTextXObject(
	ctx *model.Context,
	page types.Dict,
	text string,
) (*types.IndirectRef, error) {
	// Compute media box to position text at top-left
	var xmin, _, _, ymax float64 = 0, 0, 595, 842
	mb, ok := page["MediaBox"].(types.Array)
	if ok && len(mb) >= 4 {
		if v, ok := mb[0].(types.Float); ok {
			xmin = float64(v)
		}
		if v, ok := mb[3].(types.Float); ok {
			ymax = float64(v)
		}
	}
	// Position: left margin 40, top margin 40
	leftMargin := 10.0
	topMargin := 40.0
	x := xmin + leftMargin
	y := ymax - topMargin
	// Font size smaller
	fontSize := 10.0

	var sb strings.Builder
	sb.WriteString("q\n")
	sb.WriteString("0.5 0.5 0.5 rg\n")
	sb.WriteString("BT\n")
	sb.WriteString(fmt.Sprintf("1 0 0 1 %s %s Tm\n", fmtFloat(x), fmtFloat(y)))

	// === 核心逻辑：ASCII 段 -> F2/WinAnsi；非 ASCII 段 -> F1/UTF-16BE(不带 BOM) ===
	runes := []rune(text)
	var asciiBuf []byte
	var cjkBuf []byte

	flushAscii := func() {
		if len(asciiBuf) == 0 {
			return
		}
		hexStr := strings.ToUpper(hex.EncodeToString(asciiBuf))
		sb.WriteString(fmt.Sprintf("/F2 %d Tf\n<%s> Tj\n", int(fontSize), hexStr))
		asciiBuf = asciiBuf[:0]
	}
	flushCJK := func() {
		if len(cjkBuf) == 0 {
			return
		}
		hexStr := strings.ToUpper(hex.EncodeToString(cjkBuf))
		sb.WriteString(fmt.Sprintf("/F1 %d Tf\n<%s> Tj\n", int(fontSize), hexStr))
		cjkBuf = cjkBuf[:0]
	}

	for _, r := range runes {
		// 可打印 ASCII：交给 F2/WinAnsi
		if r >= 0x20 && r <= 0x7E {
			flushCJK()
			asciiBuf = append(asciiBuf, byte(r))
			continue
		}

		// 非 ASCII：使用 F1/UTF-16BE（UniGB-UCS2-H 不要 BOM）
		flushAscii()
		for _, cp := range utf16.Encode([]rune{r}) {
			cjkBuf = append(cjkBuf, byte(cp>>8), byte(cp&0xFF))
		}
	}

	// 处理末尾残留
	flushAscii()
	flushCJK()

	sb.WriteString("ET\nQ\n")
	content := sb.String()

	sd, err := ctx.NewStreamDictForBuf([]byte(content))
	if err != nil {
		return nil, err
	}
	if err := sd.Encode(); err != nil {
		return nil, err
	}
	sd.Dict["Type"] = types.Name("XObject")
	sd.Dict["Subtype"] = types.Name("Form")
	sd.Dict["BBox"] = mb

	// 构建资源字典：F1 = CJK Type0；F2 = Helvetica/WinAnsi
	res := types.Dict{}
	fonts := types.Dict{}

	// F1: CJK Type0
	cidDict := types.Dict{
		"Type":     types.Name("Font"),
		"Subtype":  types.Name("CIDFontType2"),
		"BaseFont": types.Name("STSong-Light"),
		"CIDSystemInfo": types.Dict{
			"Registry":   types.StringLiteral("Adobe"),
			"Ordering":   types.StringLiteral("GB1"),
			"Supplement": types.Integer(2),
		},
	}
	fd := types.Dict{
		"Type":        types.Name("FontDescriptor"),
		"FontName":    types.Name("STSong-Light"),
		"Flags":       types.Integer(4),
		"FontBBox":    types.Array{types.Integer(-500), types.Integer(-500), types.Integer(1500), types.Integer(1500)},
		"ItalicAngle": types.Integer(0),
		"Ascent":      types.Integer(880),
		"Descent":     types.Integer(-120),
		"CapHeight":   types.Integer(880),
		"StemV":       types.Integer(80),
	}

	fdRef, err := ctx.IndRefForNewObject(fd)
	if err != nil {
		return nil, err
	}
	cidDict["FontDescriptor"] = *fdRef
	cidRef, err := ctx.IndRefForNewObject(cidDict)
	if err != nil {
		return nil, err
	}
	type0 := types.Dict{
		"Type":            types.Name("Font"),
		"Subtype":         types.Name("Type0"),
		"BaseFont":        types.Name("STSong-Light"),
		"Encoding":        types.Name("UniGB-UCS2-H"),
		"DescendantFonts": types.Array{*cidRef},
	}
	type0Ref, err := ctx.IndRefForNewObject(type0)
	if err != nil {
		return nil, err
	}
	fonts["F1"] = *type0Ref

	// F2: Latin Type1 Helvetica + WinAnsiEncoding
	latinFont := types.Dict{
		"Type":     types.Name("Font"),
		"Subtype":  types.Name("Type1"),
		"BaseFont": types.Name("Helvetica"),
		"Encoding": types.Name("WinAnsiEncoding"),
	}
	latinRef, err := ctx.IndRefForNewObject(latinFont)
	if err != nil {
		return nil, err
	}
	fonts["F2"] = *latinRef

	res["Font"] = fonts
	sd.Dict["Resources"] = res

	return ctx.IndRefForNewObject(*sd)
}

// 保守的浮点格式化，去掉多余小数
func fmtFloat(v interface{}) string {
	switch t := v.(type) {
	case int:
		return fmt.Sprintf("%d", t)
	case float64:
		// 避免科学计数法
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", t), "0"), ".")
	default:
		return "0"
	}
}

// ensurePageResources 确保 page 字典里存在 Resources 并返回它。
func ensurePageResources(page types.Dict) types.Dict {
	if r, ok := page["Resources"].(types.Dict); ok {
		return r
	}
	r := types.Dict{}
	page["Resources"] = r
	return r
}

// ensureFontDict 确保 Resources 包含 Font 字典并返回它。
func ensureFontDict(res types.Dict) types.Dict {
	if f, ok := res["Font"].(types.Dict); ok {
		return f
	}
	f := types.Dict{}
	res["Font"] = f
	return f
}

func buildFallbackOCGAndXObject(ctx *model.Context, pageDict types.Dict, pageNr int, text string) (*types.IndirectRef, *types.IndirectRef, error) {
	// 4. 处理 fallback：创建 fallback OCG，创建 fallback XObject（例如水印/提示）
	fallbackName := fmt.Sprintf("text_%02d", pageNr)
	ocg, err := createOCG(ctx, fallbackName)
	if err != nil {
		return nil, nil, fmt.Errorf("create fallback ocg: %w", err)
	}
	xobj, err := buildTextXObject(ctx, pageDict, text)
	if err != nil {
		return nil, nil, fmt.Errorf("build fallback text xobject: %w", err)
	}
	return ocg, xobj, nil
}
