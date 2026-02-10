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
	xobj := ensureXObjectDict(ctx, res)
	// 注册 NormalContent（如果有）
	if normalXObj != nil {
		xobj["NormalContent"] = *normalXObj
	}
	for i, m := range masks {
		xobj[fmt.Sprintf("mask_%02d_%02d", pageNr, i)] = *m
	}

	if text != nil {
		xobj[fmt.Sprintf("text_%02d", pageNr)] = *text
	}

	// Properties
	props := ensurePropertiesDict(ctx, res)
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
			}
		default:
			// 非预期类型，创建新的 Resources
			res = types.Dict{}
		}
	} else {
		res = types.Dict{}
	}
	pageDict["Resources"] = res
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
	xobj := ensureXObjectDict(ctx, res)
	if expiredText != nil {
		xobj[fmt.Sprintf("expired_%02d", pageNr)] = *expiredText
	}
	// Properties

	props := ensurePropertiesDict(ctx, res)
	props[fmt.Sprintf("expired_%02d", pageNr)] = *expiredOCG
	return
}

func injectExpiredMaskOCGResources(ctx *model.Context, pageDict types.Dict, pageNr int, expiredMaskXObj *types.IndirectRef, expiredMaskOCG *types.IndirectRef) {
	res := getResourceDict(ctx, pageDict)
	xobj := ensureXObjectDict(ctx, res)
	if expiredMaskXObj != nil {
		xobj[fmt.Sprintf("expired_mask_%02d", pageNr)] = *expiredMaskXObj
	}
	// Properties
	props := ensurePropertiesDict(ctx, res)
	props[fmt.Sprintf("expired_mask_%02d", pageNr)] = *expiredMaskOCG
	return
}

func ensureXObjectDict(ctx *model.Context, res types.Dict) types.Dict {
	if xo, ok := res["XObject"]; ok && xo != nil {
		switch v := xo.(type) {
		case types.Dict:
			return v
		case types.IndirectRef:
			rd, err := ctx.DereferenceDict(v)
			if err == nil && rd != nil {
				res["XObject"] = rd
				return rd
			}
		}
	}
	xobj := types.Dict{}
	res["XObject"] = xobj
	return xobj
}

func ensurePropertiesDict(ctx *model.Context, res types.Dict) types.Dict {
	if p, ok := res["Properties"]; ok && p != nil {
		switch v := p.(type) {
		case types.Dict:
			return v
		case types.IndirectRef:
			rd, err := ctx.DereferenceDict(v)
			if err == nil && rd != nil {
				res["Properties"] = rd
				return rd
			}
		}
	}
	props := types.Dict{}
	res["Properties"] = props
	return props
}
