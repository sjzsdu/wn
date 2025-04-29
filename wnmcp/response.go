package wnmcp

import "github.com/mark3labs/mcp-go/mcp"

// TODO 这类还需要处理其他的类型
func ToolCallResultToString(resp *mcp.CallToolResult) string {
	if resp == nil {
		return ""
	}

	// 如果没有内容，返回空字符串
	if len(resp.Content) == 0 {
		return ""
	}

	// 遍历所有内容并拼接
	var result string
	for _, content := range resp.Content {
		// 根据 Content 的定义，它可以是 TextContent, ImageContent, 或 EmbeddedResource
		// 这里我们主要处理文本内容
		if textContent, ok := content.(mcp.TextContent); ok {
			result += textContent.Text
		}
	}

	return result
}
