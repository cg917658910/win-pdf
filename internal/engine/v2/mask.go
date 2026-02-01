package engine

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// BuildMaskXObject 创建一个用于遮罩/覆盖的 Form XObject。
func buildMaskXObject(ctx *model.Context, mediaBox types.Array) (*types.IndirectRef, error) {
	w := mediaBox[2]
	h := mediaBox[3]

	content := fmt.Sprintf(`
q
1 1 1 rg
0 0 %v %v re
f
Q
`, w, h)

	sd, err := ctx.NewStreamDictForBuf([]byte(content))
	if err != nil {
		return nil, err
	}
	if err := sd.Encode(); err != nil {
		return nil, err
	}
	sd.Dict["Type"] = types.Name("XObject")
	sd.Dict["Subtype"] = types.Name("Form")
	sd.Dict["BBox"] = mediaBox

	return ctx.IndRefForNewObject(*sd)
}

// 批量创建Mask XObjects
func buildMaskXObjects(ctx *model.Context, mediaBox types.Array, count int) ([]*types.IndirectRef, error) {
	var refs []*types.IndirectRef
	for i := 0; i < count; i++ {
		ref, err := buildMaskXObject(ctx, mediaBox)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// BuildMaskOCG 创建指定数量的遮罩 OCG。
func buildMaskOCGs(ctx *model.Context, pageNr, count int) ([]*types.IndirectRef, error) {
	var refs []*types.IndirectRef
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("mask_%02d_%02d", pageNr, i+1)
		ref, err := createOCG(ctx, name)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// BuildMaskOCGsAndXObjects 创建指定数量的遮罩 OCG 和对应的遮罩 XObject。
func buildMaskOCGsAndXObjectsForPage(ctx *model.Context, page types.Dict, pageNr int, count int) ([]*types.IndirectRef, []*types.IndirectRef, error) {
	mediaBox := getPageMediaBox(page)
	ocgs, err := buildMaskOCGs(ctx, pageNr, count)
	if err != nil {
		return nil, nil, fmt.Errorf("build mask ocgs: %w", err)
	}
	objs, err := buildMaskXObjects(ctx, mediaBox, count)
	if err != nil {
		return nil, nil, fmt.Errorf("build mask xobjects: %w", objs)
	}
	return ocgs, objs, nil
}
