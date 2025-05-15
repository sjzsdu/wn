package base

// NewProvider 创建新的Provider实例
func NewProvider(name, apiKey, endpoint, model string, config RequestConfig) *Provider {
	return &Provider{
		HTTPHandler: NewHTTPHandler(apiKey, endpoint, config),
		Model:       model,
		Name:        name,
		MaxTokens:   2048,
	}
}

// GetName 返回提供商名称
func (p *Provider) GetName() string {
	return p.Name
}

// SetModel 设置当前使用的模型
func (p *Provider) SetModel(model string) string {
	p.Model = model
	return p.Model
}

// SetModel 设置当前使用的模型
func (p *Provider) GetModel() string {
	return p.Model
}
