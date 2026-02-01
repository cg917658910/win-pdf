package engine

import (
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// createCIDType0Font 创建一个 Type0 (CID) 字体并附带最小 FontDescriptor。
// baseFontName 例如 "STSong-Light"。返回 Type0 字体的间接引用。
func createCIDType0Font(ctx *model.Context, baseFontName string) (*types.IndirectRef, error) {
	cidDict := types.Dict{
		"Type":     types.Name("Font"),
		"Subtype":  types.Name("CIDFontType2"),
		"BaseFont": types.Name(baseFontName),
		"CIDSystemInfo": types.Dict{
			"Registry":   types.StringLiteral("Adobe"),
			"Ordering":   types.StringLiteral("GB1"),
			"Supplement": types.Integer(2),
		},
	}

	// 最小 FontDescriptor，防止 Adobe Reader 报 /FontBBox 问题
	fd := types.Dict{
		"Type":        types.Name("FontDescriptor"),
		"FontName":    types.Name(baseFontName),
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
		"BaseFont":        types.Name(baseFontName),
		"Encoding":        types.Name("UniGB-UCS2-H"),
		"DescendantFonts": types.Array{*cidRef},
	}

	type0Ref, err := ctx.IndRefForNewObject(type0)
	if err != nil {
		return nil, err
	}
	return type0Ref, nil
}
