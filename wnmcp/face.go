package wnmcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (c *Host) GetTools(ctx context.Context, request mcp.ListToolsRequest) []mcp.Tool {
	results, lastError := c.ListTools(ctx, request)

	if lastError != nil {
		return nil
	}
	var tools []mcp.Tool
	for _, toolsResult := range results {
		for _, tool := range toolsResult.Tools {
			tools = append(tools, tool)
		}
	}
	return tools
}
