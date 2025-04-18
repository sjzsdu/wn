package cmd

import (
	"fmt"

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

	traverser := project.NewBaseChatter(doc)
	traverser.ChatWithLLM()
	data := doc.GetLLMResponse()
	if data == "" {
		fmt.Printf("failed to get llm response: %v\n", err)
		return
	}

	// 创建 AI 命令实例并设置消息
	aiCmd := newAICommand()
	aiCmd.SetMessage([]llm.Message{
		{
			Role:    "user",
			Content: data,
		},
	}).SetAgent("project")
	// 获取默认的 provider
	provider, err := llm.GetProvider("", nil)
	if err != nil {
		fmt.Printf("failed to get provider: %v\n", err)
		return
	}

	// 直接开始聊天
	aiCmd.startChat(provider)
}
