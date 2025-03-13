package llm

import (
	"fmt"

	"github.com/sjzsdu/wn/config"
)

// 首先定义 NewProvider 类型
type NewProvider func(options map[string]interface{}) (Provider, error)

// providers 存储所有注册的提供商工厂方法
var providers = make(map[string]NewProvider)

// Register 注册一个大模型提供商的工厂方法
func Register(name string, newProvider NewProvider) {
	if newProvider == nil {
		panic("llm: Register newProvider is nil")
	}
	if _, dup := providers[name]; dup {
		panic("llm: Register called twice for provider " + name)
	}
	providers[name] = newProvider
}

// GetProvider 获取指定名称的大模型提供商
func CreateProvider(name string, options map[string]interface{}) (Provider, error) {
	// 移除了锁相关代码

	if name == "" {
		return nil, fmt.Errorf("llm: no provider specified and no default provider configured")
	}

	newProvider, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("llm: unknown provider %q", name)
	}

	// 从配置中获取所有参数
	configMap := config.GetConfigMap()

	// 创建最终的配置map
	params := make(map[string]interface{})

	// 将配置文件中的参数添加到finalOptions
	for k, v := range configMap {
		params[k] = v
	}

	// 如果传入了options，覆盖或添加到finalOptions
	for k, v := range options {
		params[k] = v
	}
	provider, err := newProvider(params)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// Providers 返回所有已注册的提供商名称
func Providers() []string {
	// 移除了锁相关代码
	var list []string
	for name := range providers {
		list = append(list, name)
	}
	return list
}
