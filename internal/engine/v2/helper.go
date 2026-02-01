package engine

import (
	"bytes"
	"strings"
	"time"
)

func normalizeTime(t time.Time, isStart bool) time.Time {
	if !t.IsZero() {
		return t
	}
	if isStart {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}
func escapeJSString(s string) string {
	replacer := strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
		"\n", `\n`,
		"\r", ``,
	)
	return replacer.Replace(s)
}

func encodeJSUTF16BE(s string) []byte {
	runes := []rune(s)

	buf := bytes.NewBuffer(nil)

	// UTF-16BE BOM
	buf.Write([]byte{0xFE, 0xFF})

	for _, r := range runes {
		buf.WriteByte(byte(r >> 8))
		buf.WriteByte(byte(r))
	}
	return buf.Bytes()
}
