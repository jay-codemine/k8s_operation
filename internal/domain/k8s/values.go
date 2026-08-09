package k8s

import (
	"fmt"
	"strings"

	"k8soperation/pkg/utils"
)

// ——— 值对象：不可变，自验证 ———

// ClusterName 集群名称
type ClusterName struct{ val string }

func NewClusterName(s string) (ClusterName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ClusterName{}, fmt.Errorf("集群名称不能为空")
	}
	if len(s) > 100 {
		return ClusterName{}, fmt.Errorf("集群名称最长 100 字符")
	}
	return ClusterName{val: s}, nil
}

func (n ClusterName) String() string { return n.val }

// KubeConfig 加密的 KubeConfig 配置
type KubeConfig struct{ encrypted string }

// NewKubeConfigFromPlain 从明文加密创建
func NewKubeConfigFromPlain(plain string) (KubeConfig, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return KubeConfig{}, fmt.Errorf("kubeconfig 不能为空")
	}
	encrypted, err := utils.EncodeKubeconfigSecure(plain)
	if err != nil {
		return KubeConfig{}, fmt.Errorf("加密 kubeconfig 失败: %w", err)
	}
	return KubeConfig{encrypted: encrypted}, nil
}

// NewKubeConfigFromEncrypted 从已加密字符串创建（不验证，假设已是加密态）
func NewKubeConfigFromEncrypted(encrypted string) KubeConfig {
	return KubeConfig{encrypted: encrypted}
}

// Encrypted 返回加密后的值（落库用）
func (k KubeConfig) Encrypted() string { return k.encrypted }

// Decrypt 解密为明文
func (k KubeConfig) Decrypt() (string, error) {
	if k.encrypted == "" {
		return "", nil
	}
	return utils.DecodeKubeconfigSmart(k.encrypted)
}

// IsEncrypted 检查是否已加密
func (k KubeConfig) IsEncrypted() bool {
	return utils.IsEncrypted(k.encrypted)
}

// ClusterVersion 集群版本号
type ClusterVersion struct{ val string }

func NewClusterVersion(s string) (ClusterVersion, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ClusterVersion{val: "unknown"}, nil
	}
	return ClusterVersion{val: s}, nil
}

func (v ClusterVersion) String() string { return v.val }
