package license

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
)

var (
	pubMu    sync.RWMutex
	cachedPK *rsa.PrivateKey
)

func SetEmbeddedPrivateKey(pemBytes []byte) error {
	pk, err := parsePrivateKey(pemBytes)
	if err != nil {
		return err
	}
	pubMu.Lock()
	cachedPK = pk
	pubMu.Unlock()
	return nil
}

func LoadPrivateKeyFromFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败: %w", err)
	}
	return SetEmbeddedPrivateKey(b)
}

func parsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("PEM 解码失败")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return priv, nil
	}
	// 尝试 PKCS8
	pk, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err2 == nil {
		if r, ok := pk.(*rsa.PrivateKey); ok {
			return r, nil
		}
	}
	return nil, fmt.Errorf("解析私钥失败: %v / %v", err, err2)
}
