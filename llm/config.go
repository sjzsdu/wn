package llm

import (
	"encoding/json"
	"os"
)

// Config 表示LLM配置
type Config struct {
	DefaultProvider string                       `json:"default_provider"`
	Providers       map[string]ProviderConfig    `json:"providers"`
}

// ProviderConfig 表示提供商特定的配置
type ProviderConfig struct {
	APIKey      string            `json:"api_key"`
	APIEndpoint string            `json:"api_endpoint,omitempty"`
	DefaultModel string           `json:"default_model"`
	Models      []string          `json:"models,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}