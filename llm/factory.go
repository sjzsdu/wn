package llm

import (
	"fmt"
	"sync"
	
	"github.com/sjzsdu/wn/config"
)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]Provider)
)

// Register 注册一个大模型提供商
func Register(name string, provider Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if provider == nil {
		panic("llm: Register provider is nil")
	}
	if _, dup := providers[name]; dup {
		panic("llm: Register called twice for provider " + name)
	}
	providers[name] = provider
}

// GetProvider 获取指定名称的大模型提供商
// 如果name为空，则使用配置中的默认提供商
func GetProvider(name string) (Provider, error) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	
	// 如果name为空，则使用配置中的默认LLM提供商
	if name == "" {
		name = config.GetConfig("llm")
		// 如果配置中也没有设置默认提供商，返回错误
		if name == "" {
			return nil, fmt.Errorf("llm: no provider specified and no default provider configured")
		}
	}
	
	provider, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("llm: unknown provider %q", name)
	}
	return provider, nil
}

// Providers 返回所有已注册的提供商名称
func Providers() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()
	var list []string
	for name := range providers {
		list = append(list, name)
	}
	return list
}