package cmd

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
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
	if cmdPath == "" {
		fmt.Println("the output paramter is required")
		return
	}

	blogContent, err := helper.GetFileContent(cmdPath)
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return
	}

	// 创建聊天实例
	chat, err := aigc.NewChat(aigc.ChatOptions{
		UseAgent: "blog",
	})
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return
	}

	// 设置初始消息
	chat.SetMessages([]llm.Message{
		{
			Role:    "user",
			Content: blogContent,
		},
	})

	// 启动交互式会话
	ctx := context.Background()
	err = chat.StartInteractiveSession(ctx, aigc.InteractiveOptions{
		Renderer: helper.GetDefaultRenderer(),
		Debug:    false,
	})
	if err != nil {
		fmt.Printf("chat session ended with error: %v\n", err)
	}
}
