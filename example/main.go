package main

import (
	"log"
	"time"

	eng "github.com/cg917658910/win-pdf/internal/engine/v2"
)

func main() {
	in := "./陈果-PHP.pdf"
	out := "./v6_cg.pdf"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 23, 59, 59, 0, time.UTC)
	opts := eng.Options{
		Input:           in,
		Output:          out,
		StartTime:       start.Format(time.RFC3339),
		EndTime:         end.Format(time.RFC3339),
		UnsupportedText: "文件显示错误！请使用Adobe Reader、PDF-Xchange或福昕PDF阅读器打开当前文档！",
		UserPassword:    "",
		//PwdEnabled:      true,
		ExperiredText:    "请使用正版授权的文档！当前文档已过期，无法查看内容！",
		AllowedPrint:     true,
		AllowedCopy:      true,
		AllowedEdit:      true,
		AllowedConvert:   true,
		WatermarkEnabled: true,
		WatermarkText:    "sf",
		WatermarkDesc:    "fontname:Helvetica-Bold, points:36, fillcolor:#ff0000, opacity:0.3, rot:45, pos:c",
	}
	if err := eng.Run(opts); err != nil {
		log.Fatalf("protect: %v", err)
	}

	log.Printf("wrote %s", out)
}
