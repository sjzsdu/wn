package servers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
)

func NewPrompt(project *project.Project) {
	name := project.GetName()

	greeting(project, name)
	codeReview(project, name)
	queryBuilder(project, name)
}

func greeting(project *project.Project, name string) {
	greetingPrompt := mcp.NewPrompt("greeting",
		mcp.WithPromptDescription("友好的问候提示"),
		mcp.WithArgument("name",
			mcp.ArgumentDescription("要问候的人的名字"),
		),
	)

	wnmcp.McpServer().AddPrompt(greetingPrompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		userName := request.Params.Arguments["name"]
		if userName == "" {
			userName = "朋友"
		}

		result := &mcp.GetPromptResult{
			Description: "友好的问候",
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleAssistant,
					Content: mcp.TextContent{
						Text: fmt.Sprintf("你好，%s！今天我能帮你什么？", userName),
					},
				},
			},
		}
		return result, nil
	})
}

func codeReview(project *project.Project, name string) {
	codeReviewPrompt := mcp.NewPrompt("code_review",
		mcp.WithPromptDescription("代码审查辅助"),
		mcp.WithArgument("pr_number",
			mcp.ArgumentDescription("要审查的PR编号"),
			mcp.RequiredArgument(),
		),
	)

	wnmcp.McpServer().AddPrompt(codeReviewPrompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prNumber := request.Params.Arguments["pr_number"]
		if prNumber == "" {
			return nil, fmt.Errorf("需要提供PR编号")
		}

		result := &mcp.GetPromptResult{
			Description: "代码审查辅助",
			Messages: []mcp.PromptMessage{
				{
					Role: "system",
					Content: mcp.TextContent{
						Text: "我是一个代码审查助手。我会审查变更并提供建设性的反馈。",
					},
				},
			},
		}
		return result, nil
	})
}

func queryBuilder(project *project.Project, name string) {
	queryBuilderPrompt := mcp.NewPrompt("query_builder",
		mcp.WithPromptDescription("SQL查询构建辅助"),
		mcp.WithArgument("table",
			mcp.ArgumentDescription("要查询的表名"),
			mcp.RequiredArgument(),
		),
	)

	wnmcp.McpServer().AddPrompt(queryBuilderPrompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		tableName := request.Params.Arguments["table"]
		if tableName == "" {
			return nil, fmt.Errorf("需要提供表名")
		}

		result := &mcp.GetPromptResult{
			Description: "SQL查询构建辅助",
			Messages: []mcp.PromptMessage{
				{
					Role: "system",
					Content: mcp.TextContent{
						Text: "我是一个SQL专家。我会帮助构建高效和安全的查询。",
					},
				},
			},
		}
		return result, nil
	})
}
