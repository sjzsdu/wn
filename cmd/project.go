package cmd

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/project"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: lang.T("Project files"),
	Long:  lang.T("Project files with specified extensions into a single output file"),
	Run:   runproject,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}

func runproject(cmd *cobra.Command, args []string) {
	targetPath, err := helper.GetTargetPath(cmdPath, gitURL)
	if err != nil {
		fmt.Printf("failed to get target path: %v\n", err)
		return
	}

	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}

	// 构建项目树
	doc, err := project.BuildProjectTree(targetPath, options)
	if err != nil {
		fmt.Printf("failed to build project tree: %v\n", err)
		return
	}

	// 获取项目分析结果
	traverser := project.NewBaseChatter(doc)
	traverser.ChatWithLLM()
	data := doc.GetLLMResponse()
	if data == "" {
		fmt.Printf("failed to get llm response\n")
		return
	}

	// 创建聊天实例
	chat, err := aigc.NewChat(aigc.ChatOptions{
		ProviderName: "", // 使用默认 provider
		MessageLimit: llmMessageLimit,
		UseAgent:     "project",
		Request: llm.CompletionRequest{
			Model:          "", // 使用默认 model
			MaxTokens:      0,  // 使用默认 token 限制
			ResponseFormat: "text",
		},
	}, nil)
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return
	}

	// 设置初始消息
	chat.SetMessages([]llm.Message{
		{
			Role:    "user",
			Content: data,
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
