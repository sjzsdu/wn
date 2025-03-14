package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
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

	messages := make([]llm.Message, 0)

	for {
		// 使用 readline 读取输入
		input, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			break
		}

		// 添加用户消息
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		// 显示加载动画
		fmt.Println() // 确保在新行显示动画
		done := make(chan bool)
		go showLoadingAnimation(done)

		// 发送请求
		resp, err := provider.Complete(context.Background(), llm.CompletionRequest{
			Model:     model,
			Messages:  messages,
			MaxTokens: maxTokens,
		})

		// 停止加载动画
		done <- true
		time.Sleep(100 * time.Millisecond) // 确保动画完全停止

		if err != nil {
			fmt.Printf(lang.T("Error")+": %v\n", err)
			continue
		}

		// 添加 AI 回复到消息历史
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		fmt.Printf("\n%s\n", resp.Content)
	}
}

func showLoadingAnimation(done chan bool) {
	spinChars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	i := 0

	for {
		select {
		case <-done:
			// 使用更多空格确保完全清除动画行
			fmt.Print("\r                                                                \r")
			return
		default:
			// 确保每次更新都完全覆盖前一次的输出
			fmt.Printf("\r%-50s", fmt.Sprintf("%s "+lang.T("Thinking")+"... ", spinChars[i]))
			i = (i + 1) % len(spinChars)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
