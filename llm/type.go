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

// StreamResponse 定义流式响应的结构
type StreamResponse struct {
	Content      string
	FinishReason string
	Done         bool
}

// StreamHandler 处理流式响应的回调函数
type StreamHandler func(StreamResponse)

// 在 Provider 接口中添加流式方法
type Provider interface {
	// Complete 发送请求到大模型并获取回复
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)

	// Name 返回提供商名称
	Name() string

	// AvailableModels 返回该提供商支持的模型列表
	AvailableModels() []string

	SetModel(model string) string
	
	// CompleteStream 发送流式请求到大模型并通过回调处理响应
	CompleteStream(ctx context.Context, req CompletionRequest, handler StreamHandler) error
}
