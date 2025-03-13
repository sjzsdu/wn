package llm

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/config"
)

// Message 表示对话中的一条消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest 表示请求大模型的参数
type CompletionRequest struct {
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Model     string    `json:"model,omitempty"`
	// 其他通用参数...
}

// CompletionResponse 表示大模型的响应
type CompletionResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage"`
}

// Usage 表示token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider 定义大模型提供商的接口
type Provider interface {
	// Complete 发送请求到大模型并获取回复
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)

	// Name 返回提供商名称
	Name() string

	// AvailableModels 返回该提供商支持的模型列表
	AvailableModels() []string
}

var llms = make(map[string]Provider)

// defaultProvider 存储默认的LLM提供商实例
var defaultProvider Provider

func init() {
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
func Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if defaultProvider == nil {
		return CompletionResponse{}, fmt.Errorf("no default provider set")
	}
	return defaultProvider.Complete(ctx, req)
}
