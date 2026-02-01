package engine

import (
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func getPageMediaBox(pageDict types.Dict) types.Array {

	if mb, ok := pageDict["MediaBox"].(types.Array); ok && len(mb) >= 4 {
		return mb
	}

	defaultMediaBox := types.Array{types.Float(0), types.Float(0), types.Float(612), types.Float(792)} // 8.5 x 11 inches

	return defaultMediaBox
}
