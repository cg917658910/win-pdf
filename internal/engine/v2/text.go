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
// 它会保证 page 的 Resources 包含一个名为 F1 的 Type0 字体（如果不存在则创建 STSong-Light）。
func buildTextXObject(
	ctx *model.Context,
	page types.Dict,
	text string,
) (*types.IndirectRef, error) {

	// Compute media box to position text at top-left
	var xmin, _, _, ymax float64 = 0, 0, 595, 842 // default A4-like
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

	// Encode text to UTF-16BE with BOM
	utf16codes := utf16.Encode([]rune(text))
	buf := make([]byte, 0, 2*(len(utf16codes)+1))
	// BOM FE FF
	buf = append(buf, 0xFE, 0xFF)
	for _, cp := range utf16codes {
		buf = append(buf, byte(cp>>8), byte(cp&0xFF))
	}
	hexstr := strings.ToUpper(hex.EncodeToString(buf))
	content := fmt.Sprintf(`
q
0.5 0.5 0.5 rg
BT
/F1 %d Tf
%f %f Td
<%s> Tj
ET
Q
`, int(fontSize), x, y, hexstr)
	sd, err := ctx.NewStreamDictForBuf(
		[]byte(content),
	)
	if err != nil {
		return nil, err
	}
	if err := sd.Encode(); err != nil {
		return nil, err
	}
	sd.Dict["Type"] = types.Name("XObject")
	sd.Dict["Subtype"] = types.Name("Form")
	sd.Dict["BBox"] = mb

	// 如果提供了页面 Resources，则把它设置到 XObject 里，保证文本字体等资源可以被解析
	var res types.Dict
	if r, ok := page["Resources"].(types.Dict); ok {
		res = r
	} else {
		res = types.Dict{}
		page["Resources"] = res
	}

	// Ensure Font dictionary
	var fonts types.Dict
	if f, ok := res["Font"].(types.Dict); ok {
		fonts = f
	} else {
		fonts = types.Dict{}
		res["Font"] = fonts
	}

	// Create a CJK Type0 font referencing STSong-Light with UniGB-UCS2-H encoding
	// Many PDF viewers will map this to a local CJK font to render Chinese.
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
