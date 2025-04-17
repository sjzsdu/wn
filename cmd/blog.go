package cmd

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var blogCmd = &cobra.Command{
	Use:   "blog",
	Short: lang.T("Blog Writer"),
	Long:  lang.T("Blog Writer is a tool for writing blog"),
	Run:   runBlog,
}

var blogContent = ""

func init() {
	rootCmd.AddCommand(blogCmd)
}

func runBlog(cmd *cobra.Command, args []string) {
	if output == "" {
		fmt.Println("the output paramter is required")
		return
	}

	blogContent, err := helper.GetFileContent(output)
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return
	}
	helper.UpdatePreviewContent(blogContent)

	// 启动预览服务器
	previewURL := helper.StartPreviewServer(share.SERVER_PORT)
	fmt.Printf("预览地址: %s\n", previewURL)

	// 创建聊天实例
	chat, err := aigc.NewChat(aigc.ChatOptions{
		UseAgent: "blog",
		Hooks: &aigc.Hooks{
			AfterResponse: func(ctx context.Context, resp string) error {
				helper.UpdatePreviewContent(resp)
				// 更新全局变量
				blogContent = resp
				writeError := helper.WriteFileContent(output, content)
				return writeError
			},
			BeforeGetContext: func(ctx context.Context, agentMessages []llm.Message, historyMessages []llm.Message) []llm.Message {
				blogMessages := getBlogMessages(blogContent)
				messages := append(agentMessages, blogMessages...)

				if len(historyMessages) > 0 {
					lastMsg := historyMessages[len(historyMessages)-1]
					messages = append(messages, lastMsg)
				}

				return messages
			},
		},
	})
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return
	}

	// 启动交互式会话
	ctx := context.Background()
	err = chat.StartInteractiveSession(ctx, aigc.InteractiveOptions{
		Renderer: helper.GetDefaultRenderer(),
		Debug:    false,
	})
	if err != nil {
		fmt.Printf("chat session ended with error: %v\n", err)
	}
	helper.StopPreviewServer()
}

func getBlogMessages(content string) []llm.Message {
	if content == "" {
		return []llm.Message{}
	}
	return []llm.Message{
		{
			Role:    "system",
			Content: "这是用户的blog内容，请在这个基础上更改：\n" + content,
		},
	}
}
