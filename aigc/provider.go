package aigc

import (
	"github.com/sjzsdu/wn/llm"
)

// GetAvailableProviders 获取所有可用的AI提供商
func GetAvailableProviders() []string {
	return llm.Providers()
}

// GetAvailableModels 获取指定提供商的所有可用模型
func GetAvailableModels(providerName string) ([]string, error) {
	provider, err := llm.GetProvider(providerName, nil)
	if err != nil {
		return nil, err
	}
	return provider.AvailableModels(), nil
}