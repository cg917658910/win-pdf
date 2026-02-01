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
	end := time.Date(2026, 1, 1, 23, 59, 59, 0, time.UTC)
	opts := eng.Options{
		Input:           in,
		Output:          out,
		StartTime:       start,
		EndTime:         end,
		UnsupportedText: "不支持的查看器",
		ExperiredText:   "文件已过期",
	}
	if err := eng.Run(opts); err != nil {
		log.Fatalf("protect: %v", err)
	}

	log.Printf("wrote %s", out)
}
