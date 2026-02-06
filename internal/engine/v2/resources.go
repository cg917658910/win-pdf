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
	res := getResourceDict(ctx, pageDict)
	xobj := res["XObject"].(types.Dict)
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

func getResourceDict(ctx *model.Context, pageDict types.Dict) types.Dict {
	// 确保 Resources 存在且为 types.Dict，容错处理
	var res types.Dict
	if r, ok := pageDict["Resources"]; ok && r != nil {
		switch v := r.(type) {
		case types.Dict:
			res = v
		case types.IndirectRef:
			rd, err := ctx.DereferenceDict(v)
			if err == nil && rd != nil {
				res = rd
			} else {
				res = types.Dict{}
				pageDict["Resources"] = res
			}
		default:
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
	return res
}

func injectExpiredOCGResources(
	ctx *model.Context,
	pageDict types.Dict,
	pageNr int,
	expiredText *types.IndirectRef,
	expiredOCG *types.IndirectRef,
) {
	res := getResourceDict(ctx, pageDict)
	xobj := res["XObject"].(types.Dict)
	xobj[fmt.Sprintf("expired_%02d", pageNr)] = *expiredText
	// Properties

	props, ok := res["Properties"].(types.Dict)
	if !ok {
		props = types.Dict{}
		res["Properties"] = props
	}
	props[fmt.Sprintf("expired_%02d", pageNr)] = *expiredOCG
	return
}

func injectExpiredMaskOCGResources(ctx *model.Context, pageDict types.Dict, pageNr int, expiredMaskXObj *types.IndirectRef, expiredMaskOCG *types.IndirectRef) {
	res := getResourceDict(ctx, pageDict)
	xobj := res["XObject"].(types.Dict)
	xobj[fmt.Sprintf("expired_mask_%02d", pageNr)] = *expiredMaskXObj
	// Properties
	props, ok := res["Properties"].(types.Dict)
	if !ok {
		props = types.Dict{}
		res["Properties"] = props
	}
	props[fmt.Sprintf("expired_mask_%02d", pageNr)] = *expiredMaskOCG
	return
}
