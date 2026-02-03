package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	sha "crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

func GenerateRegCodeWithRES(machine string, expiry int64) (string, error) {
	// machine 传入时可能是带 4-4 分隔的形式，这里统一去掉 '-'
	m := strings.ReplaceAll(strings.TrimSpace(machine), "-", "")
	if len(m) == 0 {
		return "", fmt.Errorf("机器码为空")
	}

	if cachedPK == nil {
		return "", fmt.Errorf("未加载私钥")
	}
	payload := fmt.Sprintf("%s|%d", m, expiry)
	h := sha.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, cachedPK, crypto.SHA256, h[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "签名失败: %v\n", err)
		os.Exit(1)
	}

	sigB64 := base64.StdEncoding.EncodeToString(sig)
	full := fmt.Sprintf("%s|%s", payload, sigB64)
	reg := base64.StdEncoding.EncodeToString([]byte(full))
	return reg, nil
}
func GenerateRegCodeWithB32(machine string, expiry int64) (string, error) {
	m := strings.ReplaceAll(strings.TrimSpace(machine), "-", "")
	if len(m) == 0 {
		return "", fmt.Errorf("机器码为空")
	}

	// payload 用于确定性生成短码：机器码(无-) + 过期时间
	payload := fmt.Sprintf("%s|%d", m, expiry)

	// 生成短码：SHA256 -> Base32(无填充) -> 截断 20 -> 4-4-4-4-4
	h := sha.Sum256([]byte(payload))
	code := b32.EncodeToString(h[:])

	if len(code) > 20 {
		code = code[:20]
	}

	var parts []string
	for i := 0; i < len(code); i += 4 {
		end := i + 4
		if end > len(code) {
			end = len(code)
		}
		parts = append(parts, code[i:end])
	}
	return strings.Join(parts, "-"), nil
}

func GenerateRegCode(machine string, expiry int64) (string, error) {
	return GenerateRegCodeWithB32(machine, expiry)
}
