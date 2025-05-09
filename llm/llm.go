package llm

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/config"
)

var llms = make(map[string]Provider)

// defaultProvider 存储默认的LLM提供商实例
var defaultProvider Provider

func init() {
}

func Init() {
	if provider := config.GetConfig("default_provider"); provider != "" {
		if p, err := CreateProvider(provider, nil); err == nil {
			defaultProvider = p
		} else {
			fmt.Printf("Failed to create default provider %s: %v\n", provider, err)
		}
	}
}

func GetProvider(name string, options map[string]interface{}) (Provider, error) {
	if name == "" {
		if defaultProvider == nil {
			return nil, fmt.Errorf("no default provider set")
		}
		return defaultProvider, nil
	}
	_, ok := llms[name]
	if !ok {
		provider, err := CreateProvider(name, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", name, err)
		}
		llms[name] = provider
		return provider, err
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

// SetDefaultProvider 设置默认的LLM提供商
func SetDefaultProvider(name string, options map[string]interface{}) error {
	provider, err := CreateProvider(name, options)
	if err != nil {
		return fmt.Errorf("failed to set default provider: %w", err)
	}
	defaultProvider = provider
	return nil
}

// GetDefaultProvider 获取默认的LLM提供商
func GetDefaultProvider() Provider {
	return defaultProvider
}

// Complete 使用默认提供商发送请求
func Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if defaultProvider == nil {
		return nil, fmt.Errorf("no default provider set")
	}
	return defaultProvider.Complete(ctx, req)
}
