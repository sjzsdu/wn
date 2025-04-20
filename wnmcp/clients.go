package wnmcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/project"
)

type Clients struct {
	clients map[string]*Client
	project *project.Project
}

func createMCPClient(config MCPServerConfig) (client.MCPClient, error) {
	switch config.TransportType {
	case "sse":
		return client.NewSSEMCPClient(config.Url)
	case "stdio":
		return client.NewStdioMCPClient(
			config.Command,
			config.Env,
			config.Args...,
		)
	default:
		return nil, fmt.Errorf("不支持的传输类型: %s", config.TransportType)
	}
}

func NewClients(config *MCPConfig, project *project.Project) (*Clients, error) {
	if config == nil {
		return nil, nil
	}

	clients := &Clients{
		clients: make(map[string]*Client),
		project: project,
	}

	for name, serverConfig := range config.MCPServers {
		if serverConfig.Disabled {
			continue
		}

		mcpClient, err := createMCPClient(serverConfig)
		if err != nil {
			fmt.Printf("创建客户端 %s 失败: %v\n", name, err)
			continue
		}

		clients.clients[name] = NewClient(mcpClient, project)
	}

	return clients, nil
}

func (c *Clients) Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	var lastErr error
	var lastResult *mcp.InitializeResult

	for name, client := range c.clients {
		result, err := client.conn.Initialize(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 初始化失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) Ping(ctx context.Context) error {
	var lastErr error
	for name, client := range c.clients {
		if err := client.conn.Ping(ctx); err != nil {
			fmt.Printf("客户端 %s Ping 失败: %v\n", name, err)
			lastErr = err
		}
	}
	return lastErr
}

func (c *Clients) ListResources(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	var lastErr error
	var lastResult *mcp.ListResourcesResult

	for name, client := range c.clients {
		result, err := client.conn.ListResources(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 获取资源列表失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) ListResourceTemplates(ctx context.Context, request mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	var lastErr error
	var lastResult *mcp.ListResourceTemplatesResult

	for name, client := range c.clients {
		result, err := client.conn.ListResourceTemplates(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 获取资源模板列表失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) ReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	var lastErr error
	var lastResult *mcp.ReadResourceResult

	for name, client := range c.clients {
		result, err := client.conn.ReadResource(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 读取资源失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) Subscribe(ctx context.Context, request mcp.SubscribeRequest) error {
	var lastErr error
	for name, client := range c.clients {
		if err := client.conn.Subscribe(ctx, request); err != nil {
			fmt.Printf("客户端 %s 订阅失败: %v\n", name, err)
			lastErr = err
		}
	}
	return lastErr
}

func (c *Clients) Unsubscribe(ctx context.Context, request mcp.UnsubscribeRequest) error {
	var lastErr error
	for name, client := range c.clients {
		if err := client.conn.Unsubscribe(ctx, request); err != nil {
			fmt.Printf("客户端 %s 取消订阅失败: %v\n", name, err)
			lastErr = err
		}
	}
	return lastErr
}

func (c *Clients) ListPrompts(ctx context.Context, request mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	var lastErr error
	var lastResult *mcp.ListPromptsResult

	for name, client := range c.clients {
		result, err := client.conn.ListPrompts(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 获取提示列表失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) GetPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var lastErr error
	var lastResult *mcp.GetPromptResult

	for name, client := range c.clients {
		result, err := client.conn.GetPrompt(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 获取提示失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	var lastErr error
	var lastResult *mcp.ListToolsResult

	for name, client := range c.clients {
		result, err := client.conn.ListTools(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 获取工具列表失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var lastErr error
	var lastResult *mcp.CallToolResult

	for name, client := range c.clients {
		result, err := client.conn.CallTool(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 调用工具失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) SetLevel(ctx context.Context, request mcp.SetLevelRequest) error {
	var lastErr error
	for name, client := range c.clients {
		if err := client.conn.SetLevel(ctx, request); err != nil {
			fmt.Printf("客户端 %s 设置日志级别失败: %v\n", name, err)
			lastErr = err
		}
	}
	return lastErr
}

func (c *Clients) Complete(ctx context.Context, request mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	var lastErr error
	var lastResult *mcp.CompleteResult

	for name, client := range c.clients {
		result, err := client.conn.Complete(ctx, request)
		if err != nil {
			fmt.Printf("客户端 %s 自动完成失败: %v\n", name, err)
			lastErr = err
			continue
		}
		lastResult = result
	}
	return lastResult, lastErr
}

func (c *Clients) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
	for _, client := range c.clients {
		client.conn.OnNotification(handler)
	}
}

func (c *Clients) GetClient(name string) *Client {
	if c == nil {
		return nil
	}
	return c.clients[name]
}

func (c *Clients) GetAllClients() map[string]*Client {
	if c == nil {
		return nil
	}
	return c.clients
}

func (c *Clients) Close() error {
	if c == nil {
		return nil
	}
	var lastErr error
	for name, client := range c.clients {
		if client != nil && client.conn != nil {
			if err := client.conn.Close(); err != nil {
				fmt.Printf("关闭客户端 %s 失败: %v\n", name, err)
				lastErr = err
			}
		}
	}
	return lastErr
}
