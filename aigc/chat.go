package aigc

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
)

// defaultOptions 返回默认的聊天选项
func defaultOptions() ChatOptions {
	return ChatOptions{
		ProviderName: "",
		Model:        "",
		MaxTokens:    0,
		UseAgent:     "",
		MessageLimit: 2,
		Hooks:        &Hooks{}, // 初始化空钩子
	}
}

// mergeOptions 合并选项，后面的选项会覆盖前面的选项的非零值
func mergeOptions(base, override ChatOptions) ChatOptions {
	if override.ProviderName != "" {
		base.ProviderName = override.ProviderName
	}
	if override.Model != "" {
		base.Model = override.Model
	}
	if override.MaxTokens != 0 {
		base.MaxTokens = override.MaxTokens
	}
	if override.UseAgent != "" {
		base.UseAgent = override.UseAgent
	}
	if override.MessageLimit != 0 {
		base.MessageLimit = override.MessageLimit
	}
	return base
}

// NewChat 创建新的聊天实例
func NewChat(opts ChatOptions) (*Chat, error) {
	// 合并默认选项
	options := mergeOptions(defaultOptions(), opts)

	provider, err := llm.GetProvider(options.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %v", err)
	}

	return &Chat{
		options:    options,
		msgManager: message.New(),
		provider:   provider,
	}, nil
}

// SendMessage 发送消息并获取响应
func (c *Chat) SendMessage(ctx context.Context, content string) (string, error) {
	msg := &llm.Message{
		Role:    "user",
		Content: content,
	}

	// 执行发送前钩子
	if c.options.Hooks.BeforeSend != nil {
		if err := c.options.Hooks.BeforeSend(ctx, msg); err != nil {
			return "", err
		}
	}

	c.msgManager.Append(*msg)

	// 执行发送后钩子
	if c.options.Hooks.AfterSend != nil {
		if err := c.options.Hooks.AfterSend(ctx, msg); err != nil {
			return "", err
		}
	}

	var response strings.Builder
	req := llm.CompletionRequest{
		Model:     c.options.Model,
		Messages:  c.getContextMessages(),
		MaxTokens: c.options.MaxTokens,
	}

	// 执行响应前钩子
	if c.options.Hooks.BeforeResponse != nil {
		if err := c.options.Hooks.BeforeResponse(ctx, &req); err != nil {
			return "", err
		}
	}

	err := c.provider.CompleteStream(ctx, req, func(resp llm.StreamResponse) {
		if !resp.Done {
			response.WriteString(resp.Content)
		}
	})

	if err != nil {
		return "", err
	}

	responseText := response.String()

	// 执行响应后钩子
	if c.options.Hooks.AfterResponse != nil {
		if err := c.options.Hooks.AfterResponse(ctx, responseText); err != nil {
			return "", err
		}
	}

	c.msgManager.Append(llm.Message{
		Role:    "assistant",
		Content: responseText,
	})

	return responseText, nil
}

func (c *Chat) getContextMessages() []llm.Message {
	messages := append([]llm.Message{}, agent.GetAgentMessages(c.options.UseAgent)...)
	messages = append(messages, c.msgManager.GetAll()...)

	if len(messages) == 0 {
		return messages
	}

	// 如果消息数量超过限制，只取最后几条
	if len(messages) > c.options.MessageLimit {
		messages = messages[len(messages)-c.options.MessageLimit:]
	}

	// 执行获取上下文钩子
	if c.options.Hooks.BeforeGetContext != nil {
		messages = c.options.Hooks.BeforeGetContext(context.Background(), messages)
	}

	return messages
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
