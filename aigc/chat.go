package aigc

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
)

// ChatOptions 配置聊天选项
type ChatOptions struct {
	ProviderName  string
	Model        string
	MaxTokens    int
	UseAgent     string
	MessageLimit int
}

// Chat 表示一个AI聊天会话
type Chat struct {
	options    ChatOptions
	msgManager *message.Manager
	provider   llm.Provider
}

// NewChat 创建新的聊天实例
func NewChat(opts ChatOptions) (*Chat, error) {
	provider, err := llm.GetProvider(opts.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %v", err)
	}

	return &Chat{
		options:    opts,
		msgManager: message.New(),
		provider:   provider,
	}, nil
}

// SendMessage 发送消息并获取响应
func (c *Chat) SendMessage(ctx context.Context, content string) (string, error) {
	c.msgManager.Append(llm.Message{
		Role:    "user",
		Content: content,
	})

	var response strings.Builder
	err := c.provider.CompleteStream(ctx, llm.CompletionRequest{
		Model:     c.options.Model,
		Messages:  c.getContextMessages(),
		MaxTokens: c.options.MaxTokens,
	}, func(resp llm.StreamResponse) {
		if !resp.Done {
			response.WriteString(resp.Content)
		}
	})

	if err != nil {
		return "", err
	}

	responseText := response.String()
	c.msgManager.Append(llm.Message{
		Role:    "assistant",
		Content: responseText,
	})

	return responseText, nil
}

// GetMessages 获取聊天历史
func (c *Chat) GetMessages() []llm.Message {
	return c.msgManager.GetAll()
}

// SetMessages 设置聊天历史
func (c *Chat) SetMessages(msgs []llm.Message) {
	for _, msg := range msgs {
		c.msgManager.Append(msg)
	}
}

func (c *Chat) getContextMessages() []llm.Message {
	contextMessages := agent.GetAgentMessages(c.options.UseAgent)
	messages := c.msgManager.GetAll()

	if len(messages) == 0 {
		return contextMessages
	}

	// 如果消息数量超过限制，只取最后几条
	start := 0
	if len(messages) > c.options.MessageLimit {
		start = len(messages) - c.options.MessageLimit
	}

	return append(contextMessages, messages[start:]...)
}