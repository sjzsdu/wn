package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/agent"
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
	aiCmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 2000, lang.T("Maximum tokens for response"))
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

	// 创建一个父 context 用于处理整个会话的取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化消息列表
	var messages []llm.Message
	if useAgent != "" {
		messages = agent.GetAgentMessages(useAgent)
	} else {
		messages = make([]llm.Message, 0)
	}

	// 检查是否有管道输入
	isPipe := !terminal.IsTerminal(int(os.Stdin.Fd()))
	if isPipe {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return
		}
		cleanContent := stripAnsiCodes(string(stdinContent))
		if cleanContent != "" {
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: cleanContent,
			})
			// 处理管道输入，但不退出
			processChatRequest(ctx, provider, messages, model, maxTokens)
		}
		
		// 重新打开终端设备用于后续交互
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Printf(lang.T("Failed to open terminal: %v\n"), err)
			return
		}
		defer tty.Close()
		
		// 将标准输入重定向到终端
		os.Stdin = tty
	}

	// 继续进入交互式聊天
	startInteractiveChat(ctx, provider, messages, model, maxTokens)
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

// handlePipeInput 处理管道输入的情况
func handlePipeInput(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) {
	stdinContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf(lang.T("Failed to read input: %v\n"), err)
		return
	}
	cleanContent := stripAnsiCodes(string(stdinContent))
	if cleanContent == "" {
		return
	}
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: cleanContent,
	})

	// 直接处理单次请求
	handleSingleRequest(ctx, provider, messages, model, maxTokens)
}

// startInteractiveChat 启动交互式聊天会话
func startInteractiveChat(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) {
	fmt.Println(lang.T("Start chatting with AI") + " (" + lang.T("Enter 'quit' or 'exit' to end the conversation") + ")")
	targetModel := provider.SetModel(model)
	fmt.Println(lang.T("Using model")+":", targetModel)

	// 使用 readline 替代 bufio.Scanner
	rl, err := readline.New("> ")
	if err != nil {
		fmt.Printf(lang.T("Error initializing readline")+": %v\n", err)
		return
	}
	defer rl.Close()

	// 设置 readline 的中断处理
	rl.SetVimMode(false)
	rl.Config.InterruptPrompt = "^C"
	rl.Config.EOFPrompt = "exit"

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\n" + lang.T("Chat session terminated, thanks for using!"))
				return
			}
			fmt.Printf(lang.T("Error reading input")+": %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println(lang.T("Chat session terminated, thanks for using!"))
			return
		}

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		// 处理单次对话请求
		processChatRequest(ctx, provider, messages, model, maxTokens)
	}
}

// processChatRequest 处理单次对话请求并更新消息历史
func processChatRequest(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) {
	fmt.Println()
	responseStarted := false
	loadingDone := make(chan bool, 1)
	completed := make(chan error, 1)

	// 为每个请求创建一个带超时的子 context
	requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)
	defer requestCancel()

	// 先启动加载动画
	go showLoadingAnimation(loadingDone)

	// 使用流式输出
	go func() {
		var fullContent strings.Builder
		err := provider.CompleteStream(requestCtx, llm.CompletionRequest{
			Model:     model,
			Messages:  messages,
			MaxTokens: maxTokens,
		}, func(resp llm.StreamResponse) {
			if !responseStarted {
				loadingDone <- true
				responseStarted = true
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
		})
		completed <- err
		if !responseStarted {
			loadingDone <- true // 确保在错误情况下也能停止加载动画
		}
	}()

	// 等待完成或超时
	streamErr := <-completed

	// 错误处理
	if streamErr != nil {
		if !responseStarted {
			fmt.Print("\r                                                                \r")
		}
		fmt.Print("\n")
		switch {
		case streamErr == context.Canceled || strings.Contains(streamErr.Error(), "context canceled"):
			fmt.Println(lang.T("Operation canceled"))
			return
		case streamErr == context.DeadlineExceeded || requestCtx.Err() == context.DeadlineExceeded:
			fmt.Printf(lang.T("Request timeout, reason: %v\n"), streamErr)
		default:
			fmt.Printf(lang.T("Request failed: %v\n"), streamErr)
		}
		return
	}

	fmt.Println()
}

// showLoadingAnimation 函数也需要优化以支持取消
func showLoadingAnimation(done chan bool) {
	spinChars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Print("\r                                                                \r")
			return
		case <-ticker.C:
			fmt.Printf("\r%-50s", fmt.Sprintf("%s "+lang.T("Thinking")+"... ", spinChars[i]))
			i = (i + 1) % len(spinChars)
		}
	}
}

// 添加新的辅助函数来清理 ANSI 转义序列
func stripAnsiCodes(s string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansi.ReplaceAllString(s, "")
}

// 添加新的函数处理单次请求
func handleSingleRequest(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) {
	responseStarted := false
	loadingDone := make(chan bool, 1)
	completed := make(chan error, 1)

	requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)
	defer requestCancel()

	go showLoadingAnimation(loadingDone)

	go func() {
		var fullContent strings.Builder
		err := provider.CompleteStream(requestCtx, llm.CompletionRequest{
			Model:     model,
			Messages:  messages,
			MaxTokens: maxTokens,
		}, func(resp llm.StreamResponse) {
			if !responseStarted {
				loadingDone <- true
				responseStarted = true
			}
			if !resp.Done {
				fmt.Print(resp.Content)
				fullContent.WriteString(resp.Content)
			}
		})
		completed <- err
		if !responseStarted {
			loadingDone <- true
		}
	}()

	streamErr := <-completed
	if streamErr != nil {
		if !responseStarted {
			fmt.Print("\r                                                                \r")
		}
		fmt.Print("\n")
		fmt.Printf(lang.T("Request failed: %v\n"), streamErr)
		return
	}
	fmt.Println()
}
