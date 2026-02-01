package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	sha "crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
)

func GenerateRegCode(machine string, expiry int64) (string, error) {

	if cachedPK == nil {
		return "", fmt.Errorf("未加载私钥")
	}
	payload := fmt.Sprintf("%s|%d", machine, expiry)
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
