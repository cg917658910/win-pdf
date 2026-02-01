package license

import (
	"crypto"
	"crypto/rsa"
	sha "crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	pubMu    sync.RWMutex
	cachedPK *rsa.PublicKey
)

// SetEmbeddedPublicKey 允许通过将公钥 PEM 字节设置到客户端（例如通过 go:embed）来加载公钥。
func SetEmbeddedPublicKey(pemBytes []byte) error {
	pk, err := parsePublicKeyFromPEM(pemBytes)
	if err != nil {
		return err
	}
	pubMu.Lock()
	cachedPK = pk
	pubMu.Unlock()
	return nil
}

// LoadPublicKeyFromFile 从磁盘读取公钥 PEM 并加载到缓存。
func LoadPublicKeyFromFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败: %w", err)
	}
	return SetEmbeddedPublicKey(b)
}

func parsePublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("公钥 PEM 解码失败")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		if pk, ok := pub.(*rsa.PublicKey); ok {
			return pk, nil
		}
		return nil, errors.New("解析到的公钥不是 RSA 类型")
	}
	// 有些 PEM 可能直接是 PKCS1 公钥（少见）
	if pk, err2 := x509.ParsePKCS1PublicKey(block.Bytes); err2 == nil {
		return pk, nil
	}
	return nil, fmt.Errorf("解析公钥失败: %v", err)
}

// verifyRSASignature 使用已加载的公钥验证 base64 编码的签名。返回 (true,nil) 表示验证成功，(false,nil) 表示签名不匹配。
func verifyRSASignature(data []byte, sigB64 string) (bool, error) {
	pubMu.RLock()
	pk := cachedPK
	pubMu.RUnlock()
	if pk == nil {
		return false, errors.New("公钥未加载")
	}

	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false, fmt.Errorf("签名 base64 解码失败: %w", err)
	}

	h := sha.Sum256(data)
	if err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, h[:], sig); err != nil {
		// 验证不通过
		return false, nil
	}
	return true, nil
}
