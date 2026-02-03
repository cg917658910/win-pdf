package main

import (
	"log"
	"time"

	eng "github.com/cg917658910/win-pdf/internal/engine/v2"
)

func main() {
	in := "./wzy.pdf"
	out := "./v6_cg.pdf"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 2, 23, 59, 59, 0, time.UTC)
	opts := eng.Options{
		Input:           in,
		Output:          out,
		StartTime:       start,
		EndTime:         end,
		UnsupportedText: "文件显示错误！请使用Adobe Reader、PDF-Xchange或福昕PDF阅读器打开当前文档！",
		UserPassword:    "",
		//PwdEnabled:      true,
		ExperiredText:  "已过期",
		AllowedPrint:   true,
		AllowedCopy:    true,
		AllowedEdit:    true,
		AllowedConvert: true,
	}
	if err := eng.Run(opts); err != nil {
		log.Fatalf("protect: %v", err)
	}

	log.Printf("wrote %s", out)
}
