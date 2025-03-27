package wnmcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/share"
)

// NewInitializeRequest 创建一个新的初始化请求
func NewInitializeRequest() mcp.InitializeRequest {
	return mcp.InitializeRequest{
		Request: mcp.Request{
			Method: string(mcp.MethodInitialize),
		},
		Params: struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    mcp.ClientCapabilities `json:"capabilities"`
			ClientInfo      mcp.Implementation     `json:"clientInfo"`
		}{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    share.MCP_CLIENT_NAME,
				Version: share.VERSION,
			},
		},
	}
}

// NewReadResourceRequest 创建一个新的资源读取请求
func NewReadResourceRequest(uri string, args map[string]interface{}) mcp.ReadResourceRequest {
	return mcp.ReadResourceRequest{
		Request: mcp.Request{
			Method: string(mcp.MethodResourcesRead),
		},
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI:       uri,
			Arguments: args,
		},
	}
}
