package engine

import (
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// CreateOCG 创建一个 Optional Content Group (OCG) 对象并返回其间接引用。
func createOCG(ctx *model.Context, name string) (*types.IndirectRef, error) {
	d := types.Dict{
		"Type": types.Name("OCG"),
		"Name": types.StringLiteral(name),
	}
	ref, err := ctx.IndRefForNewObject(d)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// ApplyOCProperties 在文档根字典上安装 /OCProperties，使阅读器识别并显示图层。
func applyOCProperties(ctx *model.Context, ocgs []*types.IndirectRef) {
	arr := types.Array{}
	for _, r := range ocgs {
		arr = append(arr, *r)
	}
	ctx.RootDict["OCProperties"] = types.Dict{
		"OCGs": arr,
		"D": types.Dict{
			"ON":       arr,
			"Order":    arr,
			"ListMode": types.Name("AllOn"),
		},
	}
}

func insertOCPropertiesOCGs(ctx *model.Context, ocgs []*types.IndirectRef) {
	arr := types.Array{}

	// 判断是否已经存在 OCProperties
	if _, found := ctx.RootDict["OCProperties"]; found {
		// 如果存在，更新 OCGs、ON 和 Order 数组
		ocProps := ctx.RootDict["OCProperties"].(types.Dict)
		// 判断是否存在 OCGs 键
		_, ocgsFound := ocProps["OCGs"]
		if ocgsFound {
			arr = ocProps["OCGs"].(types.Array)
		}
	}
	for _, ref := range ocgs {
		arr = append(arr, *ref)
	}
	ctx.RootDict["OCProperties"] = types.Dict{
		"OCGs": arr,
		"D": types.Dict{
			"ON":       arr,
			"Order":    arr,
			"ListMode": types.Name("AllOn"),
		},
	}
}
