package engine

import (
	"bytes"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ExtractPageContentAsXObject 将页面的 Contents（可能是多个流）合并为一个 Form XObject，返回其间接引用。

func extractPageContentAsXObject(
	ctx *model.Context,
	page types.Dict,
	pageNr int,
) (*types.IndirectRef, error) {

	contents := page["Contents"]
	if contents == nil {
		return nil, fmt.Errorf("page has no contents")
	}

	// 内容流可能是一个或多个
	var streams types.Array
	switch c := contents.(type) {
	case types.IndirectRef:
		streams = types.Array{c}
	case types.Array:
		streams = c
	default:
		return nil, fmt.Errorf("unsupported contents type")
	}
	_, _, inhPAttrs, err := ctx.PageDict(pageNr, true)
	if err != nil {
		return nil, err
	}
	// 创建 Form XObject: 将页面的内容流合并为一个 StreamDict
	var buf bytes.Buffer
	for _, s := range streams {
		ir, ok := s.(types.IndirectRef)
		if !ok {
			continue
		}
		sd, _, err := ctx.DereferenceStreamDict(ir)
		if err != nil {
			return nil, err
		}
		// 确保内容已解码
		if err := sd.Decode(); err != nil {
			return nil, err
		}
		if sd.Content != nil {
			buf.Write(sd.Content)
		}
	}

	newSD, err := ctx.NewStreamDictForBuf(buf.Bytes())
	if err != nil {
		return nil, err
	}
	if err := newSD.Encode(); err != nil {
		return nil, err
	}
	newSD.Dict["Type"] = types.Name("XObject")
	newSD.Dict["Subtype"] = types.Name("Form")
	newSD.Dict["Resources"] = page["Resources"]
	newSD.Dict["BBox"] = inhPAttrs.MediaBox.Array()

	return ctx.IndRefForNewObject(*newSD)
}
func rewritePageWithMasksAndFallback(
	ctx *model.Context,
	pageDict types.Dict,
	pageNr int,
	masks []*types.IndirectRef,
	text *types.IndirectRef,
	maskOCGs []*types.IndirectRef,
	textOCG *types.IndirectRef,
) error {

	var buf bytes.Buffer

	buf.WriteString("q\nQ\n") // 🔥 清空历史 GS
	buf.WriteString("q\n/NormalContent Do\nQ\n")

	for i := range masks {
		idx := i
		buf.WriteString(fmt.Sprintf(
			"/OC /mask_%02d_%02d BDC\n/mask_%02d_%02d Do\nEMC\n",
			pageNr,
			idx,
			pageNr,
			idx,
		))
	}
	buf.WriteString(fmt.Sprintf(
		"/OC /text_%02d BDC\n/text_%02d Do\nEMC\n", pageNr, pageNr,
	))

	sd, err := ctx.NewStreamDictForBuf(
		buf.Bytes(),
	)
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
	pageDict["Contents"] = *ref

	return nil
}

// RewritePageWithMasks 生成一个包装内容流，引用 normal/mask/text XObjects 并追加到 pageDict.Contents。
