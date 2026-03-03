package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// EncodeKubeconfigBase64 明文 kubeconfig → base64
func EncodeKubeconfigBase64(plain string) (string, error) {
	s := strings.TrimSpace(plain)
	if s == "" {
		return "", errors.New("empty kubeconfig")
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

// DecodeKubeconfigBase64 base64 → 明文 kubeconfig
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
