package wnmcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/sjzsdu/wn/share"
)

var mcpServer *server.MCPServer = server.NewMCPServer(
	share.MCP_SERVER_NAME,
	share.VERSION,
)

func McpServer() *server.MCPServer {
	return mcpServer
}
