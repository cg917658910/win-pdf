package license

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

const appName = "win-pdf"
const activationFileName = "activation.json"

// ActivationInfo 保存激活信息
type ActivationInfo struct {
	MachineCode string     `json:"machine_code"`
	RegCode     string     `json:"reg_code"`
	ActivatedAt time.Time  `json:"activated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// GetMachineCode 生成基于主机信息的机器码（不可逆的哈希）
func GetMachineCode() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("无法获取主机名: %w", err)
	}

	ifs, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("无法获取网络接口: %w", err)
	}

	macs := make([]string, 0, len(ifs))
	for _, ifi := range ifs {
		if ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		ma := ifi.HardwareAddr.String()
		if ma == "" {
			continue
		}
		macs = append(macs, ma)
	}

	sort.Strings(macs)

	b := strings.Builder{}
	b.WriteString(host)
	b.WriteString("|")
	b.WriteString(strings.Join(macs, ","))

	h := sha256.Sum256([]byte(b.String()))
	// 返回前 16 字节（32 个十六进制字符）以缩短机器码长度，同时保持较高的唯一性
	return strings.ToUpper(hex.EncodeToString(h[:16])), nil
}

// ValidateRegCode 验证注册码是否合法并返回解析后的激活信息
func ValidateRegCode(machineCode, regCode string) (*ActivationInfo, error) {
	if machineCode == "" || regCode == "" {
		return nil, errors.New("参数为空")
	}

	raw, err := base64.StdEncoding.DecodeString(regCode)
	if err != nil {
		return nil, fmt.Errorf("注册码解码失败: %w", err)
	}

	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return nil, errors.New("注册码格式错误")
	}

	mc := parts[0]
	if mc != machineCode {
		return nil, errors.New("注册码与当前机器码不匹配")
	}

	expiryUnix, err := parseInt64(parts[1])
	if err != nil {
		return nil, fmt.Errorf("解析过期时间失败: %w", err)
	}

	sigB64 := parts[2]
	payload := fmt.Sprintf("%s|%d", mc, expiryUnix)
	ok, vErr := verifyRSASignature([]byte(payload), sigB64)
	if vErr != nil {
		return nil, fmt.Errorf("签名验证出错: %w", vErr)
	}
	if !ok {
		return nil, errors.New("注册码签名验证失败")
	}

	var expiresAt *time.Time
	if expiryUnix > 0 {
		t := time.Unix(expiryUnix, 0)
		expiresAt = &t
		if time.Now().After(t) {
			return nil, errors.New("注册码已过期")
		}
	}

	ai := &ActivationInfo{
		MachineCode: mc,
		RegCode:     regCode,
		ActivatedAt: time.Now(),
		ExpiresAt:   expiresAt,
	}
	return ai, nil
}

// ActivateWithRegCode 使用注册码激活并将激活信息持久化到用户配置目录
func ActivateWithRegCode(regCode string) error {
	mc, err := GetMachineCode()
	if err != nil {
		return err
	}

	ai, err := ValidateRegCode(mc, regCode)
	if err != nil {
		return err
	}

	if err := saveActivation(ai); err != nil {
		return fmt.Errorf("保存激活信息失败: %w", err)
	}
	return nil
}

// IsActivated 检查当前程序是否已激活，返回激活信息
func IsActivated() (bool, *ActivationInfo, error) {
	mc, err := GetMachineCode()
	if err != nil {
		return false, nil, err
	}

	ai, err := loadActivation()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, err
	}

	if ai.MachineCode != mc {
		return false, nil, nil
	}

	if ai.ExpiresAt != nil && time.Now().After(*ai.ExpiresAt) {
		return false, ai, nil
	}

	return true, ai, nil
}

// Deactivate 删除本地保存的激活信息（反激活）
func Deactivate() error {
	p, err := activationFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	return nil
}

// GetActivationInfo 返回当前保存的激活信息（不校验机器码或过期）
func GetActivationInfo() (*ActivationInfo, error) {
	return loadActivation()
}

// ----------------- 内部小工具 -----------------

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
