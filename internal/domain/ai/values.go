package ai

import (
	"fmt"
	"strings"
)

// ProviderName AI 提供商名称
type ProviderName struct{ val string }

func NewProviderName(s string) (ProviderName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ProviderName{}, fmt.Errorf("AI 提供商名称不能为空")
	}
	return ProviderName{val: s}, nil
}

func (n ProviderName) String() string { return n.val }

// ModelName AI 模型名称
type ModelName struct{ val string }

func NewModelName(s string) (ModelName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ModelName{}, fmt.Errorf("AI 模型名称不能为空")
	}
	return ModelName{val: s}, nil
}

func (n ModelName) String() string { return n.val }
