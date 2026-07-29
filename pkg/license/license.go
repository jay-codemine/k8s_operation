// Package license 平台 License 授权校验
//
// 授权模型：
//   - 软件作者使用 license-gen 工具（cmd/license-gen）+ Ed25519 私钥离线签发 License
//   - 平台二进制内置 Ed25519 公钥（key.go），只能验签、无法签发，
//     客户即使拿到全部代码与 License 文件也无法伪造新授权
//   - License 绑定机器码，过期后平台立即锁定（除登录/激活接口外全部拒绝）
//
// License 文本格式（单行字符串，便于复制粘贴）：
//
//	K8SOP-LICENSE.<base64url(payload JSON)>.<base64url(ed25519 signature)>
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Prefix License 文本前缀，用于快速识别
const Prefix = "K8SOP-LICENSE"

// 校验错误（区分前端提示场景）
var (
	ErrFormat          = errors.New("license 格式无效")
	ErrSignature       = errors.New("license 签名校验失败")
	ErrExpired         = errors.New("license 已过期")
	ErrMachineMismatch = errors.New("license 与本机机器码不匹配")
	ErrNotYetValid     = errors.New("license 尚未生效")
)

// Payload License 载荷（签名保护的内容）
type Payload struct {
	Licensee  string `json:"licensee"`   // 被授权方（客户名称）
	Edition   string `json:"edition"`    // 版本：enterprise / standard / trial
	MachineID string `json:"machine_id"` // 绑定的机器码（必填）
	IssuedAt  int64  `json:"issued_at"`  // 签发时间（unix 秒）
	ExpireAt  int64  `json:"expire_at"`  // 到期时间（unix 秒）
}

// Encode 将载荷与签名编码为 License 文本（license-gen 签发时使用）
func Encode(payload []byte, sig []byte) string {
	return Prefix + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(sig)
}

// Decode 解析 License 文本，返回原始载荷字节与签名（不校验签名）
func Decode(text string) (payloadRaw []byte, sig []byte, err error) {
	text = strings.TrimSpace(text)
	parts := strings.Split(text, ".")
	if len(parts) != 3 || parts[0] != Prefix {
		return nil, nil, ErrFormat
	}
	payloadRaw, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, ErrFormat
	}
	sig, err = base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, nil, ErrFormat
	}
	return payloadRaw, sig, nil
}

// Verify 完整校验 License 文本：格式 → 签名 → 机器码 → 有效期
// 返回解析后的载荷；任一环节失败返回对应错误
func Verify(text string, machineID string) (*Payload, error) {
	payloadRaw, sig, err := Decode(text)
	if err != nil {
		return nil, err
	}

	// 1) Ed25519 验签（公钥编译期内置，见 key.go）
	pub, err := publicKey()
	if err != nil {
		return nil, fmt.Errorf("内置公钥异常: %w", err)
	}
	if !ed25519.Verify(pub, payloadRaw, sig) {
		return nil, ErrSignature
	}

	var p Payload
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		return nil, ErrFormat
	}

	// 2) 机器码绑定校验（大小写不敏感）
	if !strings.EqualFold(strings.TrimSpace(p.MachineID), strings.TrimSpace(machineID)) {
		return nil, ErrMachineMismatch
	}

	// 3) 有效期校验
	now := time.Now().Unix()
	if p.IssuedAt > 0 && now < p.IssuedAt-86400 { // 允许 1 天时钟偏差
		return nil, ErrNotYetValid
	}
	if p.ExpireAt > 0 && now > p.ExpireAt {
		return nil, ErrExpired
	}

	return &p, nil
}
