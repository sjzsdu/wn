package aigc

import (
	"context"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
)

// Hooks 定义聊天过程中的各个生命周期钩子
type Hooks struct {
	// 发送消息前的钩子
	BeforeSend func(ctx context.Context, msg *llm.Message) error
	// 发送消息后的钩子
	AfterSend func(ctx context.Context, msg *llm.Message) error
	// 接收响应前的钩子
	BeforeResponse func(ctx context.Context, req *llm.CompletionRequest) error
	// 接收响应后的钩子
	AfterResponse func(ctx context.Context, resp string) error
	// 获取上下文消息时的钩子
	BeforeGetContext func(ctx context.Context, agentMessages []llm.Message, historyMessages []llm.Message) []llm.Message
}

// ChatOptions 配置聊天选项
type ChatOptions struct {
	ProviderName string
	Model        string
	MaxTokens    int
	UseAgent     string
	MessageLimit int
	Tools        []llm.Tool
	Hooks        *Hooks // 添加钩子配置
}

// Chat 表示一个AI聊天会话
type Chat struct {
	options    ChatOptions
	msgManager *message.Manager
	provider   llm.Provider
}
