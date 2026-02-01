package license

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// configDir 返回用于存放应用配置的目录（优先 UserConfigDir，回退到 ~/.config）
func configDir() (string, error) {
	ud, err := os.UserConfigDir()
	if err != nil {
		hd, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", err
		}
		return filepath.Join(hd, ".config", appName), nil
	}
	return filepath.Join(ud, appName), nil
}

// activationFilePath 返回 activation.json 的绝对路径并确保目录存在
func activationFilePath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(d, activationFileName), nil
}

// saveActivation 将 ActivationInfo 序列化为 JSON 并写入文件
func saveActivation(ai *ActivationInfo) error {
	p, err := activationFilePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(ai)
}

// loadActivation 从文件读取并反序列化 ActivationInfo
func loadActivation() (*ActivationInfo, error) {
	p, err := activationFilePath()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loading activation info from %s\n", p)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var ai ActivationInfo
	if err := json.Unmarshal(b, &ai); err != nil {
		return nil, err
	}
	return &ai, nil
}
