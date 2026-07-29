package license

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
)

// embeddedPublicKeyHex 编译期内置的 Ed25519 公钥（hex）
// 由 license-gen -init-keys 生成，对应的私钥仅软件作者持有
// 公钥只能验签，无法签发 License
const embeddedPublicKeyHex = "1aceaf0f0fccb6e801181690706cf2a06b0edd6a8fdba14f2101cd85c8e426a1"

// publicKey 解析内置公钥
func publicKey() (ed25519.PublicKey, error) {
	if embeddedPublicKeyHex == "" {
		return nil, errors.New("未配置内置公钥")
	}
	b, err := hex.DecodeString(embeddedPublicKeyHex)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, errors.New("公钥长度错误")
	}
	return ed25519.PublicKey(b), nil
}
