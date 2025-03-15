package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/output/ai"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"

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
	Short: "Chat with AI", // 先使用英文，后面动态更新
	Long:  "Start an interactive chat session with AI using configured LLM provider",
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
	if listProviders {
		providers := llm.Providers()
		fmt.Println(lang.T("Available LLM providers") + ":")
		for _, p := range providers {
			fmt.Printf("- %s\n", p)
		}
		return
	}

	provider, err := llm.GetProvider(providerName, nil)
	if err != nil {
		fmt.Printf(lang.T("Error getting LLM provider")+": %v\n", err)
		return
	}

	if listModels {
		models := provider.AvailableModels()
		fmt.Printf(lang.T("Available models for provider")+" (%s):\n", provider.Name())
		for _, m := range models {
			fmt.Printf("- %s\n", m)
		}
		return
	}

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
	// 创建一个父 context 用于处理整个会话的取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置 readline 的中断处理
	rl.SetVimMode(false)
	rl.Config.InterruptPrompt = "^C"
	rl.Config.EOFPrompt = "exit"

	// 替换原来的消息初始化
	var messages []llm.Message
	if useAgent != "" {
		messages = agent.GetAgentMessages(useAgent)
	} else {
		messages = make([]llm.Message, 0)
	}

	for {
		input, err := rl.Readline()
		if err != nil {
			// 处理中断信号，给出更友好的提示
			if err == readline.ErrInterrupt {
				fmt.Println("\n" + lang.T("Chat session terminated, thanks for using!"))
			}
			cancel() // 取消所有操作
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			ai.Output(output, messages)
			break
		}

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		fmt.Println()
		responseStarted := false
		loadingDone := make(chan bool)
		completed := make(chan error, 1)

		// 为每个请求创建一个带超时的子 context
		requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)

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
					time.Sleep(50 * time.Millisecond)
				}
				if !resp.Done {
					fmt.Print(resp.Content)
					fullContent.WriteString(resp.Content)
				} else {
					fmt.Println()
					messages = append(messages, llm.Message{
						Role:    "assistant",
						Content: fullContent.String(),
					})
				}
			})
			completed <- err
		}()

		// 等待完成或超时
		streamErr := <-completed
		requestCancel()

		// 错误处理
		if streamErr != nil {
			fmt.Print("\n")
			switch {
			case ctx.Err() == context.Canceled:
				fmt.Println(lang.T("操作已取消"))
				return
			case requestCtx.Err() == context.DeadlineExceeded:
				fmt.Printf(lang.T("请求超时，原因：%v\n"), streamErr)
			default:
				fmt.Printf(lang.T("请求失败：%v\n"), streamErr)
			}
			continue
		}
	}
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
