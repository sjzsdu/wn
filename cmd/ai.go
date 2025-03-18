package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	_ "github.com/sjzsdu/wn/llm/providers/claude"
	_ "github.com/sjzsdu/wn/llm/providers/deepseek"
	_ "github.com/sjzsdu/wn/llm/providers/openai"
)

var (
	providerName  string
	model         string
	maxTokens     int
	listProviders bool
	listModels    bool
	useAgent      string // 新增 agent 参数
	messages      []llm.Message
)

const (
	MaxRecentMessages = 2 // 保留最近的消息数量
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: lang.T("Chat with AI"), // 先使用英文，后面动态更新
	Long:  lang.T("Start an interactive chat session with AI using configured LLM provider"),
	Run:   runAI,
}

func init() {
	aiCmd.Flags().StringVarP(&providerName, "provider", "c", "", lang.T("LLM model Provider"))
	aiCmd.Flags().StringVarP(&model, "model", "m", "", lang.T("LLM model to use"))
	aiCmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 0, lang.T("Maximum tokens for response"))
	aiCmd.Flags().BoolVar(&listProviders, "providers", false, lang.T("List available LLM providers"))
	aiCmd.Flags().BoolVar(&listModels, "models", false, lang.T("List available models for current provider"))
	aiCmd.Flags().StringVarP(&useAgent, "agent", "a", "", lang.T("AI use agent name"))
	rootCmd.AddCommand(aiCmd)

	llm.Init()
}

func runAI(cmd *cobra.Command, args []string) {
	// 处理列出提供商的请求
	if listProviders {
		listAvailableProviders()
		return
	}

	provider, err := llm.GetProvider(providerName, nil)
	if err != nil {
		fmt.Printf(lang.T("Error getting LLM provider")+": %v\n", err)
		return
	}

	// 处理列出模型的请求
	if listModels {
		listAvailableModels(provider)
		return
	}

	startChat(provider)
}

func startChat(provider llm.Provider) {
	// 创建一个父 context 用于处理整个会话的取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化消息列表
	messages = make([]llm.Message, 0)

	targetModel := provider.SetModel(model)
	fmt.Println(lang.T("Start chatting with AI") + " (" + lang.T("Enter 'quit' or 'exit' to end the conversation") + ")")
	fmt.Println(lang.T("Tips: Press Ctrl+Enter for new line, Enter to submit"))
	fmt.Println(lang.T("Using model")+":", targetModel)

	// 检查是否有管道输入
	isPipe := !terminal.IsTerminal(int(os.Stdin.Fd()))
	if isPipe {
		content, err := helper.ReadPipeContent()
		if err != nil {
			fmt.Printf(lang.T("Failed to handle pipe input: %v\n"), err)
			return
		}

		if content != "" {
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: content,
			})
		}
	}

	startInteractiveChat(ctx, provider)
}

// listAvailableProviders 列出所有可用的LLM提供商
func listAvailableProviders() {
	providers := llm.Providers()
	fmt.Println(lang.T("Available LLM providers") + ":")
	for _, p := range providers {
		fmt.Printf("- %s\n", p)
	}
}

// listAvailableModels 列出指定提供商的所有可用模型
func listAvailableModels(provider llm.Provider) {
	models := provider.AvailableModels()
	fmt.Printf(lang.T("Available models for provider")+" (%s):\n", provider.Name())
	for _, m := range models {
		fmt.Printf("- %s\n", m)
	}
}

// startInteractiveChat 启动交互式聊天会话
func startInteractiveChat(ctx context.Context, provider llm.Provider) {
	if len(messages) > 0 {
		processChatRequest(ctx, provider)
	}
	for {
		input, err := helper.ReadFromTerminal("> ")
		if err != nil {
			// 在 startInteractiveChat 函数中
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\n" + lang.T("Chat session terminated, thanks for using!"))
				return
			}

			if input == "quit" || input == "exit" || input == "q" {
				fmt.Println(lang.T("Chat session terminated, thanks for using!"))
				return
			}
			fmt.Printf(lang.T("Error reading input")+": %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 添加 debug 命令处理
		if input == "debug" {
			outputDebug()
			return
		}
		if input == "quit" || input == "exit" || input == "q" {
			fmt.Println(lang.T("Chat session terminated, thanks for using!"))
			return
		}

		if inDebug {
			fmt.Println(input)
		}

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		if err := processChatRequest(ctx, provider); err != nil {
			// 如果是取消操作，直接返回
			if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
				return
			}
			continue
		}
	}
}

func outputDebug() {
	fmt.Println("\n=== Debug: Messages History ===")
	for i, msg := range messages {
		fmt.Printf("\n[%d] Role: %s\n", i, msg.Role)
		if msg.Role == "assistant" {
			// 对于 AI 回复，使用不同的格式显示
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

// processChatRequest 处理单次对话请求并更新消息历史
func processChatRequest(ctx context.Context, provider llm.Provider) error {
	responseStarted := false
	loadingDone := make(chan bool, 1)
	completed := make(chan error, 1)

	requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)
	defer requestCancel()

	go helper.ShowLoadingAnimation(loadingDone)

	go func() {
		var fullContent strings.Builder
		contextMessages := getContextMessages()
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
			Model:     model,
			Messages:  contextMessages,
			MaxTokens: maxTokens,
		}, func(resp llm.StreamResponse) {
			handleStreamResponse(resp, &responseStarted, loadingDone, &fullContent)
		})
		completed <- err
		if !responseStarted {
			loadingDone <- true
		}
	}()

	// 等待完成或超时
	if err := handleStreamError(<-completed, responseStarted, requestCtx); err != nil {
		return err
	}

	fmt.Println()
	return nil
}

// 处理流式响应的回调函数
func handleStreamResponse(resp llm.StreamResponse, responseStarted *bool, loadingDone chan bool, fullContent *strings.Builder) {
	if !*responseStarted {
		loadingDone <- true
		*responseStarted = true
		fmt.Print("\n")
	}
	if !resp.Done {
		fmt.Print(resp.Content)
		fullContent.WriteString(resp.Content)
	} else {
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: fullContent.String(),
		})
	}
}

// 处理流式请求的错误
func handleStreamError(streamErr error, responseStarted bool, requestCtx context.Context) error {
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

// getContextMessages 返回系统提示和最近的对话记录
func getContextMessages() []llm.Message {
	// 获取 agent 系统提示
	contextMessages := agent.GetAgentMessages(useAgent)

	// 如果没有历史消息，直接返回系统提示
	if len(messages) == 0 {
		return contextMessages
	}

	// 获取最近的消息
	start := len(messages)
	if start > MaxRecentMessages {
		start = len(messages) - MaxRecentMessages
	}

	// 添加最近的对话记录
	contextMessages = append(contextMessages, messages[start:]...)

	return contextMessages
}
