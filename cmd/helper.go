package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/config"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/share"
	"github.com/sjzsdu/wn/wnmcp"
)

func GetProject() *project.Project {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		return nil
	}
	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}
	project, _ := project.BuildProjectTree(targetPath, options)
	return project
}

func GetMcpHost() *wnmcp.Host {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		return nil
	}
	mcpConfig, err := wnmcp.LoadMCPConfig(targetPath, configFile)

	project := GetProject()

	host, err := wnmcp.NewHost(mcpConfig, project)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return nil
	}

	return host
}

func GetChatOptions() *aigc.ChatOptions {
	// 设置默认值
	if llmName == "" {
		// 优先从配置中获取默认提供商
		llmName = config.GetConfig("default_provider")
		if llmName == "" {
			llmName = share.DEFAULT_LLM_NAME
		}
	}
	if llmModel == "" {
		// 根据提供商获取对应的默认模型
		modelKey := fmt.Sprintf("%s_model", llmName)
		llmModel = config.GetConfig(modelKey)
		if llmModel == "" {
			llmModel = share.DEFAULT_LLM_MODEL
		}
	}
	if llmAgent == "" {
		llmAgent = share.DEFAULT_LLM_AGENT
	}
	if llmMessageLimit <= 0 {
		llmMessageLimit = share.DEFAULT_LLM_MESSAGES_LIMIT
	}

	return &aigc.ChatOptions{
		ProviderName: llmName,
		MessageLimit: llmMessageLimit,
		UseAgent:     llmAgent,
		Hooks:        &aigc.Hooks{},
		Request: llm.CompletionRequest{
			Model:          llmModel,
			MaxTokens:      0,
			ResponseFormat: "text",
		},
	}
}
