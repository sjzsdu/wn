package aigc

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/helper/renders"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
	"github.com/sjzsdu/wn/wnmcp"
)

type InteractiveOptions struct {
	Renderer renders.Renderer
	Debug    bool
}

// StartInteractiveSession 启动交互式会话
func (c *Chat) StartInteractiveSession(ctx context.Context, opts InteractiveOptions) error {
	if opts.Renderer == nil {
		opts.Renderer = helper.GetDefaultRenderer()
	}

	fmt.Println(lang.T("Start chatting with AI") + " (" + lang.T("Enter 'quit' or 'q' to end the conversation") + ")")
	fmt.Println(lang.T("Tips: Type 'vim' or press Ctrl+V to open vim for multi-line input"))
	fmt.Println(lang.T("Using model")+":", c.provider.SetModel(""))

	messages := c.GetMessages()
	if len(messages) > 0 {
		if err := c.processInteraction(ctx, "", opts); err != nil {
			return err
		}
	}

	for {
		input, err := helper.InputString("> ")
		if err != nil {
			fmt.Printf(lang.T("Error reading input")+": %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 特殊命令处理
		switch input {
		case "debug":
			c.outputDebug()
			return nil
		case "quit", "q":
			fmt.Println(lang.T("Chat session terminated, thanks for using!"))
			return nil
		}

		if err := c.processInteraction(ctx, input, opts); err != nil {
			if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
				return err
			}
			continue
		}
	}
}

func (c *Chat) processInteraction(ctx context.Context, input string, opts InteractiveOptions) error {

	if input != "" {
		msg := &llm.Message{
			Role:    "user",
			Content: input,
		}
		if c.options.Hooks.BeforeSend != nil {
			if err := c.options.Hooks.BeforeSend(ctx, msg); err != nil {
				return err
			}
		}
		c.msgManager.Append(*msg)
		if c.options.Hooks.AfterSend != nil {
			if err := c.options.Hooks.AfterSend(ctx, msg); err != nil {
				return err
			}
		}
	}

	responseStarted := false
	loadingDone := make(chan bool, 1)
	completed := make(chan error, 1)

	go helper.ShowLoadingAnimation(loadingDone)

	go func() {
		var fullContent strings.Builder
		req := c.options.Request
		req.Messages = c.getContextMessages()

		// 执行响应前钩子
		if c.options.Hooks.BeforeResponse != nil {
			if err := c.options.Hooks.BeforeResponse(ctx, &req); err != nil {
				return
			}
		}

		err := c.provider.CompleteStream(ctx, req, func(resp llm.StreamResponse) {
			if !responseStarted {
				loadingDone <- true
				responseStarted = true
				<-loadingDone
			}
			if !resp.Done {
				fullContent.WriteString(resp.Content)
				if err := opts.Renderer.WriteStream(resp.Content); err != nil {
					fmt.Print(resp.Content)
				}
			} else {
				opts.Renderer.Done()
				if share.GetDebug() {
					helper.PrintWithLabel("Stream response", resp.Response)
				}
				// 检查是否有工具调用
				if (resp.Response.ToolCalls != nil) && (len(resp.Response.ToolCalls) > 0) && c.host != nil {
					msg := &llm.Message{
						Role:      "assistant",
						Content:   "",
						ToolCalls: resp.Response.ToolCalls,
					}
					c.msgManager.Append(*msg)
					for _, toolCall := range resp.Response.ToolCalls {
						helper.PrintWithLabel("Tool call", toolCall)
						toolContent, _ := c.host.CallTool(ctx, wnmcp.NewToolCallRequest(toolCall.Function, toolCall.Arguments))
						toolResult := wnmcp.ToolCallResultToString(toolContent)
						helper.PrintWithLabel("Tool call result", toolResult)
						msg := &llm.Message{
							Role:       "tool",
							Content:    toolResult,
							ToolCallId: toolCall.ID,
						}
						c.msgManager.Append(*msg)
					}
					// 递归调用以获取最终响应
					if err := c.processInteraction(ctx, "", opts); err != nil {
						completed <- err
						return
					}
					return
				}

				if c.options.Hooks.AfterResponse != nil {
					c.options.Hooks.AfterResponse(ctx, &req, resp.Response)
				}
				c.msgManager.Append(llm.Message{
					Role:    "assistant",
					Content: fullContent.String(),
				})
			}
		})
		completed <- err
		if !responseStarted {
			loadingDone <- true
		}
	}()

	if err := <-completed; err != nil {
		return c.handleStreamError(err, responseStarted)
	}

	fmt.Println()
	return nil
}

func (c *Chat) handleStreamError(err error, responseStarted bool) error {
	if err == nil {
		return nil
	}

	if !responseStarted {
		fmt.Print("\r                                                                \r")
	}
	fmt.Print("\n")

	switch {
	case err == context.Canceled || strings.Contains(err.Error(), "context canceled"):
		fmt.Println(lang.T("Operation canceled"))
		return err
	case err == context.DeadlineExceeded:
		fmt.Printf(lang.T("Request timeout, reason: %v\n"), err)
		return err
	default:
		fmt.Printf(lang.T("Request failed: %v\n"), err)
		return err
	}
}

func (c *Chat) outputDebug() {
	messages := c.GetMessages()
	fmt.Println("\n=== Debug: Messages History ===")
	for i, msg := range messages {
		fmt.Printf("\n[%d] Role: %s\n", i, msg.Role)
		if msg.Role == "assistant" {
			fmt.Println("Content:")
			fmt.Println("----------------------------------------")
			fmt.Println(msg.Content)
			fmt.Println("----------------------------------------")
		} else {
			fmt.Printf("Content: %s\n", msg.Content)
		}
	}
	fmt.Println("\nTotal messages:", len(messages))
}
