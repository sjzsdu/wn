package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/spf13/cobra"
	_ "github.com/sjzsdu/wn/llm/providers/openai"
	_ "github.com/sjzsdu/wn/llm/providers/deepseek"
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
	aiCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", lang.T("LLM model to use"))
	aiCmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 2000, lang.T("Maximum tokens for response"))
	aiCmd.Flags().BoolVar(&listProviders, "providers", false, lang.T("List available LLM providers"))
	aiCmd.Flags().BoolVar(&listModels, "models", false, lang.T("List available models for current provider"))
	rootCmd.AddCommand(aiCmd)
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
	fmt.Println(lang.T("Using model")+":", model)

	scanner := bufio.NewScanner(os.Stdin)
	messages := make([]llm.Message, 0)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
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

		// 发送请求
		resp, err := provider.Complete(context.Background(), llm.CompletionRequest{
			Model:     model,
			Messages:  messages,
			MaxTokens: maxTokens,
		})

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
