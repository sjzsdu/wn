package aigc

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
	"github.com/sjzsdu/wn/share"
)

// defaultOptions 返回默认的聊天选项
func defaultOptions() ChatOptions {
	return ChatOptions{
		ProviderName: "",
		MessageLimit: 2,
		Hooks:        &Hooks{},
		Request: llm.CompletionRequest{
			Model:          "",
			MaxTokens:      0,
			ResponseFormat: "text",
		},
	}
}

// mergeOptions 合并选项，后面的选项会覆盖前面的选项的非零值
func mergeOptions(base, override ChatOptions) ChatOptions {
	if override.ProviderName != "" {
		base.ProviderName = override.ProviderName
	}
	if override.Request.Model != "" {
		base.Request.Model = override.Request.Model
	}
	if override.Request.MaxTokens != 0 {
		base.Request.MaxTokens = override.Request.MaxTokens
	}
	if override.Request.ResponseFormat != "" {
		base.Request.ResponseFormat = override.Request.ResponseFormat
	}
	if override.Request.Tools != nil {
		base.Request.Tools = override.Request.Tools
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
	req := c.options.Request
	req.Messages = c.getContextMessages()

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
	agentMessages := agent.GetAgentMessages(c.options.UseAgent)
	historyMessages := c.msgManager.GetRecentMessages(c.options.MessageLimit)

	// 执行获取上下文钩子
	if c.options.Hooks.BeforeGetContext != nil {
		messages := c.options.Hooks.BeforeGetContext(context.Background(), agentMessages, historyMessages)
		if share.GetDebug() {
			PrintMessages(messages)
		}
		return messages
	}
	messages := make([]llm.Message, 0, len(agentMessages)+len(historyMessages))
	messages = append(messages, agentMessages...)
	messages = append(messages, historyMessages...)
	if share.GetDebug() {
		PrintMessages(messages)
	}
	return messages
}

// GetMessages 获取聊天历史
func (c *Chat) GetMessages() []llm.Message {
	return c.msgManager.GetAll()
}

// PrintMessages 打印聊天历史，每条消息内容最多显示20个字符
func PrintMessages(messages []llm.Message) {
	fmt.Println("\n=== 聊天历史 ===")
	for i, msg := range messages {
		fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, helper.SubString(msg.Content, 40))
	}
	fmt.Println("==============")
}

// SetMessages 设置聊天历史
func (c *Chat) SetMessages(msgs []llm.Message) {
	for _, msg := range msgs {
		c.msgManager.Append(msg)
	}
}
