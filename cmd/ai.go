package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/helper/renders"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/message"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	MaxRecentMessages = 2 // 保留最近的消息数量
)

// 定义 AI 命令的结构体
type aiCommand struct {
	cmd           *cobra.Command
	msgManager    *message.Manager
	providerName  string
	model         string
	maxTokens     int
	listProviders bool
	listModels    bool
	useAgent      string
	renderer      renders.Renderer // 使用我们自己定义的 Renderer 接口
}

func newAICommand() *aiCommand {
	cmd := &aiCommand{
		msgManager: message.New(),
		renderer:   helper.GetDefaultRenderer(),
	}

	cmd.cmd = &cobra.Command{
		Use:   "ai",
		Short: lang.T("Chat with AI"),
		Long:  lang.T("Start an interactive chat session with AI using configured LLM provider"),
		Run:   cmd.runAI,
	}

	cmd.initFlags()
	return cmd
}

func (c *aiCommand) initFlags() {
	c.cmd.Flags().StringVarP(&c.providerName, "provider", "c", "", lang.T("LLM model Provider"))
	c.cmd.Flags().StringVarP(&c.model, "model", "m", "", lang.T("LLM model to use"))
	c.cmd.Flags().IntVarP(&c.maxTokens, "max-tokens", "t", 0, lang.T("Maximum tokens for response"))
	c.cmd.Flags().BoolVar(&c.listProviders, "providers", false, lang.T("List available LLM providers"))
	c.cmd.Flags().BoolVar(&c.listModels, "models", false, lang.T("List available models for current provider"))
	c.cmd.Flags().StringVarP(&c.useAgent, "agent", "a", "", lang.T("AI use agent name"))
}

func (c *aiCommand) runAI(cmd *cobra.Command, args []string) {
	if c.listProviders {
		c.listAvailableProviders()
		return
	}

	provider, err := llm.GetProvider(c.providerName, nil)
	if err != nil {
		fmt.Printf(lang.T("Error getting LLM provider")+": %v\n", err)
		return
	}

	if c.listModels {
		c.listAvailableModels(provider)
		return
	}

	c.startChat(provider)
}

func (c *aiCommand) startChat(provider llm.Provider) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targetModel := provider.SetModel(c.model)
	fmt.Println(lang.T("Start chatting with AI") + " (" + lang.T("Enter 'quit' or 'q' to end the conversation") + ")")
	fmt.Println(lang.T("Tips: Type 'vim' or press Ctrl+V to open vim for multi-line input"))
	fmt.Println(lang.T("Using model")+":", targetModel)

	isPipe := !terminal.IsTerminal(int(os.Stdin.Fd()))
	if isPipe {
		content, err := helper.ReadPipeContent()
		if err != nil {
			fmt.Printf(lang.T("Failed to handle pipe input: %v\n"), err)
			return
		}

		if content != "" {
			c.msgManager.Append(llm.Message{
				Role:    "user",
				Content: content,
			})
		}
	}

	c.startInteractiveChat(ctx, provider)
}

func (c *aiCommand) listAvailableProviders() {
	providers := llm.Providers()
	fmt.Println(lang.T("Available LLM providers") + ":")
	for _, p := range providers {
		fmt.Printf("- %s\n", p)
	}
}

func (c *aiCommand) listAvailableModels(provider llm.Provider) {
	models := provider.AvailableModels()
	fmt.Printf(lang.T("Available models for provider")+" (%s):\n", provider.Name())
	for _, m := range models {
		fmt.Printf("- %s\n", m)
	}
}

func (c *aiCommand) startInteractiveChat(ctx context.Context, provider llm.Provider) {
	messages := c.msgManager.GetAll()
	if len(messages) > 0 {
		c.processChatRequest(ctx, provider)
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
			return
		case "quit", "q":
			fmt.Println(lang.T("Chat session terminated, thanks for using!"))
			return
		}

		if inDebug {
			fmt.Println(input)
		}

		c.msgManager.Append(llm.Message{
			Role:    "user",
			Content: input,
		})

		if err := c.processChatRequest(ctx, provider); err != nil {
			if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
				return
			}
			continue
		}
	}
}

func (c *aiCommand) outputDebug() {
	messages := c.msgManager.GetAll()
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
	fmt.Println(lang.T("Chat session terminated, thanks for using!"))
}

func (c *aiCommand) processChatRequest(ctx context.Context, provider llm.Provider) error {
	responseStarted := false
	loadingDone := make(chan bool, 1)
	completed := make(chan error, 1)

	requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)
	defer requestCancel()

	go helper.ShowLoadingAnimation(loadingDone)

	go func() {
		var fullContent strings.Builder
		contextMessages := c.getContextMessages()
		if inDebug {
			fmt.Println("\n=== Debug: Context Messages ===")
			for i, msg := range contextMessages {
				fmt.Printf("\n[%d] Role: %s\n", i, msg.Role)
				fmt.Println("Content:")
				fmt.Println("----------------------------------------")
				fmt.Println(msg.Content)
				fmt.Println("----------------------------------------")
			}
			fmt.Printf("\nTotal context messages: %d\n\n", len(contextMessages))
		}

		err := provider.CompleteStream(requestCtx, llm.CompletionRequest{
			Model:     c.model,
			Messages:  contextMessages,
			MaxTokens: c.maxTokens,
		}, func(resp llm.StreamResponse) {
			c.handleStreamResponse(resp, &responseStarted, loadingDone, &fullContent)
		})
		completed <- err
		if !responseStarted {
			loadingDone <- true
		}
	}()

	if err := c.handleStreamError(<-completed, responseStarted, requestCtx); err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func (c *aiCommand) handleStreamResponse(resp llm.StreamResponse, responseStarted *bool, loadingDone chan bool, fullContent *strings.Builder) {
	if !*responseStarted {
		loadingDone <- true
		*responseStarted = true
		<-loadingDone
	}
	if !resp.Done {
		fullContent.WriteString(resp.Content)
		if err := c.renderer.WriteStream(resp.Content); err != nil {
			// 如果渲染失败，退回到普通输出
			fmt.Print(resp.Content)
		}
	} else {
		// 完成时，清空缓冲区并添加到消息历史
		c.renderer.Done()
		c.msgManager.Append(llm.Message{
			Role:    "assistant",
			Content: fullContent.String(),
		})
	}
}

func (c *aiCommand) handleStreamError(streamErr error, responseStarted bool, requestCtx context.Context) error {
	if streamErr == nil {
		return nil
	}

	if !responseStarted {
		fmt.Print("\r                                                                \r")
	}
	fmt.Print("\n")

	switch {
	case streamErr == context.Canceled || strings.Contains(streamErr.Error(), "context canceled"):
		fmt.Println(lang.T("Operation canceled"))
		return streamErr
	case streamErr == context.DeadlineExceeded || requestCtx.Err() == context.DeadlineExceeded:
		fmt.Printf(lang.T("Request timeout, reason: %v\n"), streamErr)
		return streamErr
	default:
		fmt.Printf(lang.T("Request failed: %v\n"), streamErr)
		return streamErr
	}
}

func (c *aiCommand) getContextMessages() []llm.Message {
	contextMessages := agent.GetAgentMessages(c.useAgent)
	messages := c.msgManager.GetAll()

	if len(messages) == 0 {
		return contextMessages
	}

	// 计算开始位置：如果消息数量超过限制，只取最后 MaxRecentMessages 条
	start := 0
	if len(messages) > MaxRecentMessages {
		start = len(messages) - MaxRecentMessages
	}

	return append(contextMessages, messages[start:]...)
}

func (c *aiCommand) SetMessage(msgs []llm.Message) *aiCommand {
	for _, msg := range msgs {
		c.msgManager.Append(msg)
	}
	return c
}

func (c *aiCommand) SetAgent(agent string) *aiCommand {
	c.useAgent = agent
	return c
}

func init() {
	aiCmd := newAICommand()
	rootCmd.AddCommand(aiCmd.cmd)
	llm.Init()
}
