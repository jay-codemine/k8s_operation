package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key: must be 16, 24, or 32 bytes")
	ErrInvalidCiphertext = errors.New("invalid ciphertext: too short")
	ErrDecryptionFailed  = errors.New("decryption failed: data may be corrupted or key is wrong")
)

// CryptoService 加密服务
type CryptoService struct {
	key []byte
}

// NewCryptoService 创建加密服务
// key 可以是任意长度的字符串，内部会通过 SHA-256 转换为 32 字节密钥
func NewCryptoService(secretKey string) *CryptoService {
	// 使用 SHA-256 将任意长度的密钥转换为固定 32 字节
	hash := sha256.Sum256([]byte(secretKey))
	return &CryptoService{key: hash[:]}
}

// Encrypt 使用 AES-256-GCM 加密数据
// 返回 base64 编码的密文
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	if len(plaintext) == 0 {
		return "", nil
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据，nonce 作为前缀存储
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用 AES-256-GCM 解密数据
// 输入为 base64 编码的密文
func (c *CryptoService) Decrypt(ciphertextB64 string) (string, error) {
	if len(ciphertextB64) == 0 {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// EncryptKubeConfig 加密 KubeConfig（带前缀标识）
// 加密后的数据格式: ENC:base64(ciphertext)
func (c *CryptoService) EncryptKubeConfig(kubeconfig string) (string, error) {
	if len(kubeconfig) == 0 {
		return "", nil
	}

	encrypted, err := c.Encrypt(kubeconfig)
	if err != nil {
		return "", err
	}

	// 添加前缀标识这是加密数据
	return "ENC:" + encrypted, nil
}

// DecryptKubeConfig 解密 KubeConfig
// 支持自动识别：
// - ENC:xxx 格式：加密数据，需要解密
// - 其他格式：假设是未加密的明文（向后兼容）
func (c *CryptoService) DecryptKubeConfig(data string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	// 检查是否是加密数据
	if len(data) > 4 && data[:4] == "ENC:" {
		return c.Decrypt(data[4:])
	}

	// 未加密数据，直接返回（向后兼容旧数据）
	return data, nil
}

// IsEncrypted 检查数据是否已加密
func IsEncrypted(data string) bool {
	return len(data) > 4 && data[:4] == "ENC:"
}

// ========== 密码加密（bcrypt）==========

// HashPassword 使用 bcrypt 加密密码
// cost 参数控制计算强度，默认 10，范围 4-31
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// HashPasswordWithCost 使用指定 cost 加密密码
func HashPasswordWithCost(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPassword 验证密码是否匹配
// hashedPassword 是存储在数据库中的哈希值
// password 是用户输入的明文密码
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// IsPasswordHashed 检查密码是否已经是 bcrypt 哈希格式
// bcrypt 哈希以 $2a$, $2b$, $2y$ 开头
func IsPasswordHashed(password string) bool {
	if len(password) < 4 {
		return false
	}
	prefix := password[:4]
	return prefix == "$2a$" || prefix == "$2b$" || prefix == "$2y$"
}

// CheckPasswordSmart 智能密码验证（兼容旧明文密码）
// 返回值：
// - matched: 密码是否匹配
// - needMigrate: 是否需要迁移到 bcrypt
func CheckPasswordSmart(storedPassword, inputPassword string) (matched bool, needMigrate bool) {
	// 1) 如果存储的是 bcrypt 哈希，使用 bcrypt 验证
	if IsPasswordHashed(storedPassword) {
		return CheckPassword(storedPassword, inputPassword), false
	}

	// 2) 存储的是明文密码，直接比对
	if storedPassword == inputPassword {
		return true, true // 匹配成功，需要迁移
	}

	return false, false
}

// ========== 全局加密服务实例 ==========

var globalCrypto *CryptoService

// InitGlobalCrypto 初始化全局加密服务
func InitGlobalCrypto(secretKey string) {
	globalCrypto = NewCryptoService(secretKey)
}

// GetGlobalCrypto 获取全局加密服务
func GetGlobalCrypto() *CryptoService {
	return globalCrypto
}

// GlobalEncrypt 使用全局加密服务加密
func GlobalEncrypt(plaintext string) (string, error) {
	if globalCrypto == nil {
		return "", errors.New("global crypto service not initialized")
	}
	return globalCrypto.Encrypt(plaintext)
}

// GlobalDecrypt 使用全局加密服务解密
func GlobalDecrypt(ciphertext string) (string, error) {
	if globalCrypto == nil {
		return "", errors.New("global crypto service not initialized")
	}
	return globalCrypto.Decrypt(ciphertext)
}

// GlobalEncryptKubeConfig 使用全局加密服务加密 KubeConfig
func GlobalEncryptKubeConfig(kubeconfig string) (string, error) {
	if globalCrypto == nil {
		return "", errors.New("global crypto service not initialized")
	}
	return globalCrypto.EncryptKubeConfig(kubeconfig)
}

// GlobalDecryptKubeConfig 使用全局加密服务解密 KubeConfig
func GlobalDecryptKubeConfig(data string) (string, error) {
	if globalCrypto == nil {
		return "", errors.New("global crypto service not initialized")
	}
	return globalCrypto.DecryptKubeConfig(data)
}
