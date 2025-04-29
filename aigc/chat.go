package aigc

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
	"github.com/sjzsdu/wn/share"
	"github.com/sjzsdu/wn/wnmcp"
)

// defaultOptions 返回默认的聊天选项
func defaultOptions() ChatOptions {
	return ChatOptions{
		ProviderName: "",
		MessageLimit: 2,
		Hooks:        &Hooks{},
		UseAgent:     share.DEFAULT_LLM_AGENT,
		Request: llm.CompletionRequest{
			Model:          "",
			MaxTokens:      0,
			ResponseFormat: "text",
		},
	}
}

// NewChat 创建新的聊天实例
func NewChat(opts ChatOptions, host *wnmcp.Host) (*Chat, error) {
	// 合并默认选项
	options := helper.MergeStruct(defaultOptions(), opts)

	provider, err := llm.GetProvider(options.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %v", err)
	}

	return &Chat{
		options:    options,
		msgManager: message.New(),
		provider:   provider,
		host:       host,
	}, nil
}

func (c *Chat) Complete(ctx context.Context, content string) (string, error) {
	if content != "" {
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
	}

	req := c.options.Request
	req.Messages = c.getContextMessages()
	// 执行响应前钩子
	if c.options.Hooks.BeforeResponse != nil {
		if err := c.options.Hooks.BeforeResponse(ctx, &req); err != nil {
			return "", err
		}
	}
	resp, compErr := c.provider.Complete(ctx, req)
	if compErr != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("Completion error", compErr.Error())
		}
		return "", compErr
	}
	if (resp.ToolCalls != nil) && (len(resp.ToolCalls) > 0) && c.host != nil {
		msg := &llm.Message{
			Role:      "assistant",
			Content:   "",
			ToolCalls: resp.ToolCalls,
		}
		c.msgManager.Append(*msg)
		for _, toolCall := range resp.ToolCalls {
			toolContent, _ := c.host.CallTool(ctx, wnmcp.NewToolCallRequest(toolCall.Function, toolCall.Arguments))
			helper.PrintWithLabel("Tool call response", toolContent)
			msg := &llm.Message{
				Role:       "tool",
				Content:    wnmcp.ToolCallResultToString(toolContent),
				ToolCallId: toolCall.ID,
			}
			c.msgManager.Append(*msg)
		}
		return c.Complete(ctx, "")
	}
	if c.options.Hooks.AfterResponse != nil {
		if err := c.options.Hooks.AfterResponse(ctx, &req, &resp); err != nil {
			return "", err
		}
	}
	if compErr != nil {
		return "", compErr
	}
	c.msgManager.Append(llm.Message{
		Role:    "assistant",
		Content: resp.Content,
	})
	if share.GetDebug() {
		helper.PrintWithLabel("Completion response", resp)
	}
	return resp.Content, nil
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
