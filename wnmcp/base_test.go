package wnmcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
)

// MockMCPClient 是一个模拟的MCP客户端
type MockMCPClient struct {
	mock.Mock
}

// CallTool 模拟CallTool方法
func (m *MockMCPClient) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

// ListTools 模拟ListTools方法
func (m *MockMCPClient) ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ListToolsResult), args.Error(1)
}

// Initialize 模拟Initialize方法
func (m *MockMCPClient) Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.InitializeResult), args.Error(1)
}

// Ping 模拟Ping方法
func (m *MockMCPClient) Ping(ctx context.Context) error {
	return nil
}

// ListResources 模拟ListResources方法
func (m *MockMCPClient) ListResources(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return nil, nil
}

// ListResourceTemplates 模拟ListResourceTemplates方法
func (m *MockMCPClient) ListResourceTemplates(ctx context.Context, request mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return nil, nil
}

// ReadResource 模拟ReadResource方法
func (m *MockMCPClient) ReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, nil
}

// Subscribe 模拟Subscribe方法
func (m *MockMCPClient) Subscribe(ctx context.Context, request mcp.SubscribeRequest) error {
	return nil
}

// Unsubscribe 模拟Unsubscribe方法
func (m *MockMCPClient) Unsubscribe(ctx context.Context, request mcp.UnsubscribeRequest) error {
	return nil
}

// ListPrompts 模拟ListPrompts方法
func (m *MockMCPClient) ListPrompts(ctx context.Context, request mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	return nil, nil
}

// GetPrompt 模拟GetPrompt方法
func (m *MockMCPClient) GetPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, nil
}

// SetLevel 模拟SetLevel方法
func (m *MockMCPClient) SetLevel(ctx context.Context, request mcp.SetLevelRequest) error {
	return nil
}

// Complete 模拟Complete方法
func (m *MockMCPClient) Complete(ctx context.Context, request mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return nil, nil
}

// Close 模拟Close方法
func (m *MockMCPClient) Close() error {
	return nil
}

// OnNotification 模拟OnNotification方法
func (m *MockMCPClient) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
}
