package engine

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func extractPageContentAsXObject(
	ctx *model.Context,
	page types.Dict,
	pageNr int,
	ocg types.IndirectRef,
) (types.IndirectRef, error) {

	contents := page["Contents"]
	if contents == nil {
		return types.IndirectRef{}, fmt.Errorf("page has no contents")
	}

	// 内容流可能是一个或多个
	var streams types.Array
	switch c := contents.(type) {
	case types.IndirectRef:
		streams = types.Array{c}
	case types.Array:
		streams = c
	default:
		return types.IndirectRef{}, fmt.Errorf("unsupported contents type")
	}
	_, _, inhPAttrs, err := ctx.PageDict(pageNr, true)
	if err != nil {
		return types.IndirectRef{}, err
	}
	// 创建 Form XObject
	form := types.Dict{
		"Type":      types.Name("XObject"),
		"Subtype":   types.Name("Form"),
		"OC":        ocg, // ⭐ 关键：绑定 OCG_Normal
		"Resources": page["Resources"],
		"BBox":      inhPAttrs.MediaBox.Array(),
	}

	// 合并内容流
	form["Contents"] = streams

	formRef, err := ctx.IndRefForNewObject(form)

	// 将 Page 内容替换为仅绘制 fallback
	// 并在 Page Contents 末尾追加一个流：
	// q
	// /OC /OCG_Normal BDC
	// /NormalContent Do
	// EMC
	// Q
	// 这样 OCG 控制是否执行 Do，页面内容保持被 OCG 包裹但不被 JS 直接修改。

	return *formRef, err
}

func attachXObjectToPage(
	page types.Dict,
	xobj types.IndirectRef,
	ocg types.IndirectRef,
) {

	res := page["Resources"].(types.Dict)

	xobjs, ok := res["XObject"].(types.Dict)
	if !ok {
		xobjs = types.Dict{}
		res["XObject"] = xobjs
	}

	xobjs["NormalContent"] = xobj

	// Ensure OCG reference is available in page resource Properties mapping
	if props, ok := res["Properties"].(types.Dict); ok {
		props["OCG_Normal"] = ocg
		res["Properties"] = props
	} else {
		res["Properties"] = types.Dict{"OCG_Normal": ocg}
	}
}

func setFallbackContent(ctx *model.Context, page types.Dict, fallbackText string) error {
	// Ensure Resources exists
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
	cidRef, err := ctx.IndRefForNewObject(cidDict)
	if err != nil {
		return err
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
		return err
	}

	fonts["F1"] = *type0Ref
	res["Font"] = fonts
	page["Resources"] = res

	// Compute media box to position text at top-left
	var xmin, _, _, ymax float64 = 0, 0, 595, 842 // default A4-like
	if mb, ok := page["MediaBox"].(types.Array); ok && len(mb) >= 4 {
		if v, ok := mb[0].(types.Float); ok {
			xmin = float64(v)
		}
		/* if v, ok := mb[1].(types.Float); ok {
			ymin = float64(v)
		}
		if v, ok := mb[2].(types.Float); ok {
			xmax = float64(v)
		} */
		if v, ok := mb[3].(types.Float); ok {
			ymax = float64(v)
		}
	}
	// Position: left margin 40, top margin 40
	leftMargin := 40.0
	topMargin := 40.0
	x := xmin + leftMargin
	y := ymax - topMargin

	// Font size smaller
	fontSize := 12.0

	// Encode text to UTF-16BE with BOM
	utf16codes := utf16.Encode([]rune(fallbackText))
	buf := make([]byte, 0, 2*(len(utf16codes)+1))
	// BOM FE FF
	buf = append(buf, 0xFE, 0xFF)
	for _, cp := range utf16codes {
		buf = append(buf, byte(cp>>8), byte(cp&0xFF))
	}
	hexstr := strings.ToUpper(hex.EncodeToString(buf))

	// Build content stream: gray color 0.5, set font /F1 fontSize, move to computed position, show hex string (left-aligned on first line)
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

	sd, err := ctx.NewStreamDictForBuf([]byte(content))
	if err != nil {
		return err
	}
	if err := sd.Encode(); err != nil {
		return err
	}
	ref, err := ctx.IndRefForNewObject(*sd)
	if err != nil {
		return err
	}

	page["Contents"] = *ref
	return nil
}

func createUnlockWidget(
	ctx *model.Context,
	page types.Dict,
	pageNr int,
	normalXObj types.IndirectRef,
) error {

	// --- AP stream ---
	apContent := `
q
/NormalContent Do
Q
`
	// Compute inherited MediaBox early to set proper BBox in AP
	_, _, inhPAttrs, err := ctx.PageDict(pageNr, true)
	if err != nil {
		return err
	}

	apStream, err := ctx.NewStreamDictForBuf([]byte(apContent))
	if err != nil {
		return err
	}

	apStream.Dict = types.Dict{
		"Type":    types.Name("XObject"),
		"Subtype": types.Name("Form"),
		"BBox":    inhPAttrs.MediaBox.Array(),
		"Resources": types.Dict{
			"XObject": types.Dict{
				"NormalContent": normalXObj,
			},
		},
	}

	if err := apStream.Encode(); err != nil {
		return err
	}

	apRef, err := ctx.IndRefForNewObject(*apStream)
	if err != nil {
		return err
	}
	// --- Widget ---
	widget := types.Dict{
		"Type":    types.Name("Annot"),
		"Subtype": types.Name("Widget"),

		"FT": types.Name("Btn"),
		"T":  types.StringLiteral("tag_unlock"),

		// 覆盖整页，但默认隐藏
		"Rect": inhPAttrs.MediaBox.Array(),

		"F": types.Integer(
			//(1 << 1) | // Hidden
			(1 << 2) | // Print
				(1 << 7), // Locked
		),

		"AP": types.Dict{
			"N": *apRef,
		},
	}

	ref, err := ctx.IndRefForNewObject(widget)

	if page["Annots"] == nil {
		page["Annots"] = types.Array{*ref}
	} else {
		page["Annots"] = append(page["Annots"].(types.Array), *ref)
	}

	return nil
}
