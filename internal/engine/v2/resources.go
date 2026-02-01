package engine

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// InjectOCGResources 在页面 Resources 中注入 XObject 和 Properties 键，注册 normal/mask/text 对象与 OCG 引用。
func injectOCGResources(
	ctx *model.Context,
	pageDict types.Dict,
	pageNr int,
	normalXObj *types.IndirectRef,
	masks []*types.IndirectRef,
	text *types.IndirectRef,
	maskOCGs []*types.IndirectRef,
	textOCG *types.IndirectRef,
) {
	// 确保 Resources 存在且为 types.Dict，容错处理
	var res types.Dict
	if r, ok := pageDict["Resources"]; ok && r != nil {
		if rd, ok := r.(types.Dict); ok {
			res = rd
		} else {
			// 非预期类型，创建新的 Resources
			res = types.Dict{}
			pageDict["Resources"] = res
		}
	} else {
		res = types.Dict{}
		pageDict["Resources"] = res
	}

	xobj, ok := res["XObject"].(types.Dict)
	if !ok {
		xobj = types.Dict{}
		res["XObject"] = xobj
	}

	// 注册 NormalContent（如果有）
	if normalXObj != nil {
		xobj["NormalContent"] = *normalXObj
	}
	for i, m := range masks {
		xobj[fmt.Sprintf("mask_%02d_%02d", pageNr, i)] = *m
	}

	xobj[fmt.Sprintf("text_%02d", pageNr)] = *text

	// Properties
	props, ok := res["Properties"].(types.Dict)
	if !ok {
		props = types.Dict{}
		res["Properties"] = props
	}
	for i, ocg := range maskOCGs {
		props[fmt.Sprintf("mask_%02d_%02d", pageNr, i)] = *ocg
	}
	if textOCG != nil {
		props[fmt.Sprintf("text_%02d", pageNr)] = *textOCG
	}
	return
}
