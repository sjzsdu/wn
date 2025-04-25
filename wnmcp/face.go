package wnmcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/llm"
)

func (c *Host) GetTools(ctx context.Context, request mcp.ListToolsRequest) []llm.Tool {
	results, lastError := c.ListTools(ctx, request)

	if lastError != nil {
		return nil
	}
	var tools []llm.Tool
	for _, toolsResult := range results {
		for _, tool := range toolsResult.Tools {
			tools = append(tools, llm.Tool{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			})
		}
	}
	return tools
}
