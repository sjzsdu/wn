package wnmcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/project"
)

// Client 实现 MCPClient 接口
type Client struct {
	conn    client.MCPClient
	project *project.Project
}

func NewClient(conn client.MCPClient, project *project.Project) *Client {
	client := &Client{
		conn:    conn,
		project: project,
	}
	client.Initialize()
	return client
}

func (c *Client) Initialize() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化客户端并打印服务器信息
	initResult, err := c.conn.Initialize(ctx, NewInitializeRequest())
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	fmt.Printf("连接到服务器: %s %s\n\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
}

func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.conn.Ping(ctx)
}

func (c *Client) ListResources() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Listing available resources...")
	resourcesRequest := mcp.ListResourcesRequest{}
	resources, err := c.conn.ListResources(ctx, resourcesRequest)
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}
	for _, resource := range resources.Resources {
		fmt.Printf("资源: %s\n", resource.URI)
		fmt.Printf("名称: %s\n", resource.Name)
		if resource.Description != "" {
			fmt.Printf("描述: %s\n", resource.Description)
		}
		fmt.Printf("MIME类型: %s\n", resource.MIMEType)
		fmt.Println()
	}
	return err
}

func (c *Client) ListResourceTemplates() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("列出可用的资源模板...")
	request := mcp.ListResourceTemplatesRequest{}
	result, err := c.conn.ListResourceTemplates(ctx, request)
	if err != nil {
		log.Fatalf("获取资源模板失败: %v", err)
	}
	for _, template := range result.ResourceTemplates {
		fmt.Printf("名称: %s\n", template.Name)
		if template.MIMEType != "" {
			fmt.Printf("MIME类型: %s\n", template.MIMEType)
		}
		if template.Description != "" {
			fmt.Printf("描述: %s\n", template.Description)
		}
		fmt.Println()
	}
	return err
}

func (c *Client) ReadResource(uri string, args map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("读取资源: %s\n", uri)

	request := NewReadResourceRequest(uri, args)
	result, err := c.conn.ReadResource(ctx, request)
	if err != nil {
		log.Fatalf("读取资源失败: %v", err)
	}
	fmt.Printf("资源内容Meta:\n%v\n", result.Meta)
	fmt.Printf("资源内容:\n%v\n", result.Contents)
	return err
}

func (c *Client) ListPrompts() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("列出可用的提示...")
	request := mcp.ListPromptsRequest{}
	result, err := c.conn.ListPrompts(ctx, request)
	if err != nil {
		log.Fatalf("获取提示列表失败: %v", err)
	}
	for _, prompt := range result.Prompts {
		fmt.Printf("arguments: %v\n", prompt.Arguments)
		fmt.Printf("名称: %s\n", prompt.Name)
		if prompt.Description != "" {
			fmt.Printf("描述: %s\n", prompt.Description)
		}
		fmt.Println()
	}
	return err
}

func (c *Client) GetPrompt(name string, args map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("获取提示: %s\n", name)
	request := NewPromptRequest(name, args)
	result, err := c.conn.GetPrompt(ctx, request)
	if err != nil {
		log.Fatalf("获取提示失败: %v", err)
	}
	fmt.Printf("提示内容:\n%s\n", result.Description)
	fmt.Printf("提示内容:\n%v\n", result.Messages)
	fmt.Printf("提示内容:\n%v\n", result.Meta)
	return err
}

func (c *Client) ListTools() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("列出可用的工具...")
	request := mcp.ListToolsRequest{}
	result, err := c.conn.ListTools(ctx, request)
	if err != nil {
		log.Fatalf("获取工具列表失败: %v", err)
	}
	for _, tool := range result.Tools {
		fmt.Printf("名称: %s\n", tool.Name)
		if tool.Description != "" {
			fmt.Printf("描述: %s\n", tool.Description)
		}
		fmt.Printf("InputSchema: %s\n", tool.InputSchema)
		fmt.Printf("Type: %s\n", tool.InputSchema.Type)
		fmt.Println()
	}
	return err
}

func (c *Client) CallTool(name string, args map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("调用工具: %s\n", name)
	request := NewToolCallRequest(name, args)
	result, err := c.conn.CallTool(ctx, request)
	if err != nil {
		log.Fatalf("工具调用失败: %v", err)
	}
	fmt.Printf("执行结果:\n%s\n", result.Content)
	fmt.Printf("执行结果:\n%v\n", result.Meta)
	return err
}
