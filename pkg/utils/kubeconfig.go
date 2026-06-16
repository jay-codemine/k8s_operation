package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// EncodeKubeconfigBase64 明文 kubeconfig → base64
// 已废弃：请使用 EncodeKubeconfigSecure
func EncodeKubeconfigBase64(plain string) (string, error) {
	s := strings.TrimSpace(plain)
	if s == "" {
		return "", errors.New("empty kubeconfig")
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

// DecodeKubeconfigBase64 base64 → 明文 kubeconfig
// 已废弃：请使用 DecodeKubeconfigSmart
func DecodeKubeconfigBase64(b64 string) (string, error) {
	s := strings.TrimSpace(b64)
	if s == "" {
		return "", errors.New("empty kubeconfig_b64")
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("kube_config base64 decode failed: %w", err)
	}
	return string(raw), nil
}

// ========== 新的加密方法 ==========

// EncodeKubeconfigSecure 明文 kubeconfig → AES 加密
// 返回格式: ENC:base64(ciphertext)
func EncodeKubeconfigSecure(plain string) (string, error) {
	s := strings.TrimSpace(plain)
	if s == "" {
		return "", errors.New("empty kubeconfig")
	}
	return GlobalEncryptKubeConfig(s)
}

// DecodeKubeconfigSmart 智能解码 kubeconfig
// - ENC: 前缀 → AES 解密
// - 以 api 或 { 开头 → 已是明文（YAML/JSON）
// - 其他 → 尝试 base64 解码（向后兼容旧数据）
func DecodeKubeconfigSmart(data string) (string, error) {
	s := strings.TrimSpace(data)
	if s == "" {
		return "", errors.New("empty kubeconfig data")
	}

	// 1) 加密数据：ENC: 前缀
	if IsEncrypted(s) {
		return GlobalDecryptKubeConfig(s)
	}

	// 2) 已是明文 YAML/JSON
	if strings.HasPrefix(s, "apiVersion:") || strings.HasPrefix(s, "{") {
		return s, nil
	}

	// 3) 尝试 base64 解码（向后兼容）
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// 解码失败，可能就是明文
		return s, nil
	}

	decoded := string(raw)

	// 4) 验证 base64 解码后是否是有效的 kubeconfig
	if !strings.Contains(decoded, "apiVersion:") && !strings.HasPrefix(decoded, "{") {
		// 解码后不是 kubeconfig 格式，返回原始数据
		return s, nil
	}

	return decoded, nil
}
