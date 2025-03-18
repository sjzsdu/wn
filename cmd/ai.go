package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
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

	targetModel := provider.SetModel(model)
	fmt.Println(lang.T("Start chatting with AI") + " (" + lang.T("Enter 'quit' or 'exit' to end the conversation") + ")")
	fmt.Println(lang.T("Tips: Press Ctrl+Enter for new line, Enter to submit"))
	fmt.Println(lang.T("Using model")+":", targetModel)

	// 检查是否有管道输入
	isPipe := !terminal.IsTerminal(int(os.Stdin.Fd()))
	if isPipe {
		var pipeContent []byte
		// 先读取管道内容
		pipeContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return
		}

		// 重新设置标准输入为终端
		tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err != nil {
			fmt.Printf(lang.T("Failed to reopen terminal: %v\n"), err)
			return
		}
		os.Stdin = tty

		// 重新初始化终端
		if err := helper.InitTerminal(); err != nil {
			fmt.Printf(lang.T("Failed to initialize terminal: %v\n"), err)
			return
		}

		// 确保终端已经准备好
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Println(lang.T("Failed to initialize terminal"))
			return
		}

		// 处理管道输入
		cleanContent := stripAnsiCodes(string(pipeContent))
		if cleanContent != "" {
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: cleanContent,
			})
			if err := processChatRequest(ctx, provider, messages, targetModel, maxTokens); err != nil {
				return
			}
		}
	}

	// 继续进入交互式聊天
	startInteractiveChat(ctx, provider, messages, targetModel, maxTokens)
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
func startInteractiveChat(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) {
	for {
		input, err := helper.ReadFromTerminal("> ")
		if err != nil {
			// 在 startInteractiveChat 函数中
			if err == readline.ErrInterrupt || err == io.EOF {
			    fmt.Println("\n" + lang.T("Chat session terminated, thanks for using!"))
			    return
			}
			
			// 在 input == "quit" || input == "exit" 的条件分支中
			if input == "quit" || input == "exit" {
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
		if input == "quit" || input == "exit" {
			fmt.Println(lang.T("Chat session terminated, thanks for using!"))
			return
		}

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		if err := processChatRequest(ctx, provider, messages, model, maxTokens); err != nil {
			// 如果是取消操作，直接返回
			if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
				return
			}
			// 其他错误继续等待用户输入
			continue
		}
	}
}

// processChatRequest 处理单次对话请求并更新消息历史
func processChatRequest(ctx context.Context, provider llm.Provider, messages []llm.Message, model string, maxTokens int) error {
    responseStarted := false
    loadingDone := make(chan bool, 1)
    completed := make(chan error, 1)

    requestCtx, requestCancel := context.WithTimeout(ctx, share.TIMEOUT)
    defer requestCancel()

    go helper.ShowLoadingAnimation(loadingDone)

    go func() {
        var fullContent strings.Builder
        contextMessages := getContextMessages(messages)
        err := provider.CompleteStream(requestCtx, llm.CompletionRequest{
            Model:     model,
            Messages:  contextMessages,
            MaxTokens: maxTokens,
        }, func(resp llm.StreamResponse) {
            if !responseStarted {
                loadingDone <- true
                responseStarted = true
                fmt.Print("\n")
            }
            if !resp.Done {
                // 直接输出原始内容，不做任何处理
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
            loadingDone <- true
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
			return streamErr
		case streamErr == context.DeadlineExceeded || requestCtx.Err() == context.DeadlineExceeded:
			fmt.Printf(lang.T("Request timeout, reason: %v\n"), streamErr)
			return streamErr
		default:
			fmt.Printf(lang.T("Request failed: %v\n"), streamErr)
			return streamErr
		}
	}

	fmt.Println()
	return nil
}

// 添加新的辅助函数来清理 ANSI 转义序列
// 修改 stripAnsiCodes 函数，确保正确处理 git diff 输出
func stripAnsiCodes(s string) string {
	// 处理 git diff 常见的颜色代码和格式控制符
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)
	return strings.TrimSpace(ansi.ReplaceAllString(s, ""))
}

// 在文件开头添加配置
const (
	MaxContextMessages = 5 // 保留最近5轮对话
)

// 在发送请求前处理消息历史
func getContextMessages(messages []llm.Message) []llm.Message {
	if len(messages) <= MaxContextMessages {
		return messages
	}

	// 保留系统消息（如果有的话）
	var contextMessages []llm.Message
	for _, msg := range messages {
		if msg.Role == "system" {
			contextMessages = append(contextMessages, msg)
		}
	}

	// 添加最近的对话
	start := len(messages) - MaxContextMessages
	contextMessages = append(contextMessages, messages[start:]...)

	return contextMessages
}
