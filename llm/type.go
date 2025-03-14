package llm

import "context"

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
