package openai

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"k8soperation/pkg/setting"
)

// ProviderInfo 提供商信息（前端展示用）
type ProviderInfo struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Icon   string      `json:"icon"`
	Models []ModelInfo `json:"models"`
}

// ModelInfo 模型信息（前端展示用）
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Capability  string `json:"capability"`
	ProviderID  string `json:"provider_id"`
}

// Registry 多模型提供商注册中心
type Registry struct {
	mu        sync.RWMutex
	providers map[string]*providerEntry // key = provider ID
	defaultID string                     // 默认提供商 ID
}

type providerEntry struct {
	config  setting.AIProviderConfig
	models  map[string]setting.AIModelConfig // key = model ID
	clients map[string]*Client              // key = model ID，懒加载缓存
}

// NewRegistry 从配置构建注册中心
func NewRegistry(cfg *setting.AIAssistantSettingS) *Registry {
	r := &Registry{
		providers: make(map[string]*providerEntry),
		defaultID: cfg.DefaultProvider,
	}

	if len(cfg.Providers) > 0 {
		// 新版多提供商配置
		for _, p := range cfg.Providers {
			if p.APIKey == "" {
				continue
			}
			entry := &providerEntry{
				config:  p,
				models:  make(map[string]setting.AIModelConfig),
				clients: make(map[string]*Client),
			}
			for _, m := range p.Models {
				entry.models[m.ID] = m
			}
			r.providers[p.ID] = entry
		}
		// 如果没设默认提供商，用第一个
		if r.defaultID == "" && len(cfg.Providers) > 0 {
			r.defaultID = cfg.Providers[0].ID
		}
	} else if cfg.APIKey != "" {
		// 兼容旧版单提供商配置
		model := cfg.Model
		if model == "" {
			model = "gpt-4o-mini"
		}
		entry := &providerEntry{
			config: setting.AIProviderConfig{
				ID:          "openai",
				Name:        "OpenAI",
				Icon:        "openai",
				APIKey:      cfg.APIKey,
				BaseURL:     cfg.BaseURL,
				MaxTokens:   cfg.MaxTokens,
				Temperature: cfg.Temperature,
				Models: []setting.AIModelConfig{
					{ID: model, Name: model},
				},
			},
			models:  map[string]setting.AIModelConfig{model: {ID: model, Name: model}},
			clients: make(map[string]*Client),
		}
		r.providers["openai"] = entry
		r.defaultID = "openai"
	}

	return r
}

// GetClient 获取指定提供商+模型的客户端（懒加载，线程安全）
func (r *Registry) GetClient(providerID, modelID, systemPrompt string) (*Client, error) {
	r.mu.RLock()

	// 空值回退到默认
	origProvider := providerID
	origModel := modelID
	if providerID == "" {
		providerID = r.defaultID
	}
	entry, ok := r.providers[providerID]
	if !ok {
		r.mu.RUnlock()
		log().Warn("[AI-Registry] 提供商不存在",
			zap.String("requested_provider", origProvider),
			zap.String("resolved_provider", providerID),
		)
		return nil, fmt.Errorf("未知的 AI 提供商: %s", providerID)
	}

	// 空模型 → 使用提供商的第一个模型
	if modelID == "" && len(entry.config.Models) > 0 {
		modelID = entry.config.Models[0].ID
	}

	log().Debug("[AI-Registry] 解析客户端",
		zap.String("requested_provider", origProvider),
		zap.String("resolved_provider", providerID),
		zap.String("requested_model", origModel),
		zap.String("resolved_model", modelID),
	)

	// 检查缓存（仅缓存无自定义 systemPrompt 的客户端）
	cacheKey := modelID
	if systemPrompt == "" {
		if cached, ok := entry.clients[cacheKey]; ok {
			r.mu.RUnlock()
			return cached, nil
		}
	}
	r.mu.RUnlock()

	// 构建配置
	maxTokens := entry.config.MaxTokens
	if m, ok := entry.models[modelID]; ok && m.MaxTokens > 0 {
		maxTokens = m.MaxTokens
	}
	temp := entry.config.Temperature

	cfg := Config{
		APIKey:       entry.config.APIKey,
		BaseURL:      entry.config.BaseURL,
		Model:        modelID,
		MaxTokens:    maxTokens,
		Temperature:  temp,
		SystemPrompt: systemPrompt,
	}

	client := NewClient(cfg)

	// 缓存（无自定义 systemPrompt 时）
	if systemPrompt == "" {
		r.mu.Lock()
		entry.clients[cacheKey] = client
		r.mu.Unlock()
	}

	return client, nil
}

// GetDefaultProviderID 获取默认提供商 ID
func (r *Registry) GetDefaultProviderID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultID
}

// GetDefaultModelID 获取默认提供商的第一个模型 ID
func (r *Registry) GetDefaultModelID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.providers[r.defaultID]
	if !ok || len(entry.config.Models) == 0 {
		return ""
	}
	return entry.config.Models[0].ID
}

// ListProviders 列出所有可用提供商和模型（前端展示用，隐藏敏感信息）
func (r *Registry) ListProviders() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ProviderInfo
	for _, entry := range r.providers {
		p := ProviderInfo{
			ID:   entry.config.ID,
			Name: entry.config.Name,
			Icon: entry.config.Icon,
		}
		for _, m := range entry.config.Models {
			p.Models = append(p.Models, ModelInfo{
				ID:          m.ID,
				Name:        m.Name,
				Description: m.Description,
				Capability:  m.Capability,
				ProviderID:  entry.config.ID,
			})
		}
		result = append(result, p)
	}
	return result
}
