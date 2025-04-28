package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/helper/renders"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	MaxRecentMessages = 2 // 保留最近的消息数量
)

// 定义 AI 命令的结构体
type aiCommand struct {
	cmd           *cobra.Command
	providerName  string
	model         string
	maxTokens     int
	listProviders bool
	listModels    bool
	useAgent      string
	renderer      renders.Renderer
}

func newAICommand() *aiCommand {
	cmd := &aiCommand{
		renderer: helper.GetDefaultRenderer(),
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
		providers := aigc.GetAvailableProviders()
		fmt.Println(lang.T("Available LLM providers") + ":")
		for _, p := range providers {
			fmt.Printf("- %s\n", p)
		}
		return
	}

	if c.listModels {
		models, err := aigc.GetAvailableModels(c.providerName)
		if err != nil {
			fmt.Printf(lang.T("Error getting models")+": %v\n", err)
			return
		}
		fmt.Printf(lang.T("Available models for provider")+" (%s):\n", c.providerName)
		for _, m := range models {
			fmt.Printf("- %s\n", m)
		}
		return
	}

	chat, err := aigc.NewChat(aigc.ChatOptions{
		ProviderName: c.providerName,
		MessageLimit: MaxRecentMessages,
		UseAgent:     c.useAgent,
		Request: llm.CompletionRequest{
			Model:          c.model,
			MaxTokens:      c.maxTokens,
			ResponseFormat: "text",
		},
	}, nil)
	if err != nil {
		fmt.Printf(lang.T("Failed to initialize chat: %v\n"), err)
		return
	}

	// 处理管道输入
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		content, err := helper.ReadPipeContent()
		if err != nil {
			fmt.Printf(lang.T("Failed to handle pipe input: %v\n"), err)
			return
		}
		if content != "" {
			chat.SetMessages([]llm.Message{{
				Role:    "user",
				Content: content,
			}})
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动交互式会话
	err = chat.StartInteractiveSession(ctx, aigc.InteractiveOptions{
		Renderer: c.renderer,
		Debug:    false,
	})
	if err != nil {
		fmt.Printf(lang.T("Chat session ended with error: %v\n"), err)
	}
}

func init() {
	aiCmd := newAICommand()
	rootCmd.AddCommand(aiCmd.cmd)
	llm.Init()
}
