package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	sha "crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	privPath := flag.String("priv", "server_private.pem", "path to RSA private key PEM file")
	pubPath := flag.String("pub", "server_public.pem", "path to RSA public key PEM file to output")
	gen := flag.Bool("gen", false, "generate a new RSA private key and save to -priv (and write public key to -pub)")
	machine := flag.String("machine", "", "machine code to bind the reg code to")
	days := flag.Int("days", 0, "valid days (0 means permanent)")
	flag.Parse()

	if *gen {
		if err := genKey(*privPath); err != nil {
			fmt.Fprintf(os.Stderr, "生成私钥失败: %v\n", err)
			os.Exit(1)
		}
		// 读取新生成的私钥并写出公钥
		privPem, err := os.ReadFile(*privPath)
		if err == nil {
			priv, err2 := parsePrivateKey(privPem)
			if err2 == nil {
				if err3 := writePublicKeyFromPrivate(priv, *pubPath); err3 != nil {
					fmt.Fprintf(os.Stderr, "写出公钥失败: %v\n", err3)
				}
			} else {
				fmt.Fprintf(os.Stderr, "解析私钥失败: %v\n", err2)
			}
		} else {
			fmt.Fprintf(os.Stderr, "读取新私钥失败: %v\n", err)
		}
		// 同时打印公钥 PEM 到 stdout 便于 embed
		if b, err := os.ReadFile(*pubPath); err == nil {
			fmt.Println(string(b))
		}

		fmt.Printf("已生成私钥: %s，公钥: %s\n", *privPath, *pubPath)
		os.Exit(0)
	}

	if *machine == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -machine")
		flag.Usage()
		os.Exit(2)
	}

	privPem, err := os.ReadFile(*privPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取私钥失败: %v\n", err)
		os.Exit(1)
	}

	priv, err := parsePrivateKey(privPem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析私钥失败: %v\n", err)
		os.Exit(1)
	}

	// 在签名之前，可选择导出公钥文件并打印，便于客户端 embed 使用
	if *pubPath != "" {
		if err := writePublicKeyFromPrivate(priv, *pubPath); err != nil {
			fmt.Fprintf(os.Stderr, "写出公钥失败: %v\n", err)
		} else {
			if b, err := os.ReadFile(*pubPath); err == nil {
				fmt.Println(string(b))
			}
		}
	}

	var expiry int64 = 0
	if *days > 0 {
		expiry = time.Now().Add(time.Duration(*days) * 24 * time.Hour).Unix()
	}

	payload := fmt.Sprintf("%s|%d", *machine, expiry)
	h := sha.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "签名失败: %v\n", err)
		os.Exit(1)
	}

	sigB64 := base64.StdEncoding.EncodeToString(sig)
	full := fmt.Sprintf("%s|%s", payload, sigB64)
	reg := base64.StdEncoding.EncodeToString([]byte(full))
	fmt.Println(reg)
}

func genKey(path string) error {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	b := x509.MarshalPKCS1PrivateKey(k)
	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return err
	}
	return nil
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

// writePublicKeyFromPrivate 从私钥生成公钥 PEM 并写入文件；返回写入的字节或错误
func writePublicKeyFromPrivate(priv *rsa.PrivateKey, path string) error {
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return fmt.Errorf("序列化公钥失败: %w", err)
	}
	pemBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	pemBytes := pem.EncodeToMemory(pemBlock)
	if err := os.WriteFile(path, pemBytes, 0644); err != nil {
		return fmt.Errorf("写公钥文件失败: %w", err)
	}
	return nil
}
