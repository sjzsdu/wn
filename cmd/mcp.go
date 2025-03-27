package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
	"github.com/spf13/cobra"
)

var (
	mcpLayer string
	mcpPort  string

	mcpCommand string
	mcpEnv     []string
	mcpArgs    []string
	mcpAction  string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: lang.T("mcp commands"),
	Long:  lang.T("mcp server and client commands for this project"),
}

var mcpServerCmd = &cobra.Command{
	Use:   "server",
	Short: lang.T("start mcp server"),
	Long:  lang.T("start mcp server with specified configuration"),
	Run:   runMcpServer,
}

var mcpClientCmd = &cobra.Command{
	Use:   "client",
	Short: lang.T("start mcp client"),
	Long:  lang.T("start mcp client with specified configuration"),
	Run:   runMcpClient,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.PersistentFlags().StringVar(&mcpLayer, "layer", "stdio", lang.T("MCP transfer layer"))
	mcpCmd.PersistentFlags().StringVar(&mcpPort, "port", "8080", lang.T("MCP sse port"))
	mcpCmd.AddCommand(mcpServerCmd)
	mcpCmd.AddCommand(mcpClientCmd)

	mcpClientCmd.PersistentFlags().StringVar(&mcpCommand, "cmd", "wn", lang.T("MCP server command"))
	mcpClientCmd.PersistentFlags().StringSliceVar(&mcpEnv, "env", nil, lang.T("MCP server environtment"))
	mcpClientCmd.PersistentFlags().StringSliceVar(&mcpArgs, "args", nil, lang.T("MCP server command arguments"))
	mcpClientCmd.PersistentFlags().StringVar(&mcpAction, "action", "", lang.T("MCP server command arguments"))
}

func runMcpServer(cmd *cobra.Command, args []string) {
	targetPath, err := helper.GetTargetPath(cmdPath, gitURL)
	if err != nil {
		fmt.Printf("failed to get target path: %v\n", err)
		return
	}

	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}

	// 构建项目树
	project, err := project.BuildProjectTree(targetPath, options)
	if err != nil {
		fmt.Printf("failed to build project tree: %v\n", err)
		return
	}

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"WN MCP Server: file system",
		"1.0.0",
	)

	// 添加项目文件列表资源
	fileListResource := mcp.NewResource(
		"files://list",
		"Project Files List",
		mcp.WithResourceDescription("List all files in the project"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(fileListResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		files, err := project.GetAllFiles()

		fileList, err := json.Marshal(files)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal file list: %v", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "files://list",
				MIMEType: "application/json",
				Text:     string(fileList),
			},
		}, nil
	})

	// 添加动态资源模板
	template := mcp.NewResourceTemplate(
		"file://{path}",
		"Project Files",
		mcp.WithTemplateDescription("Access project files"),
		mcp.WithTemplateMIMEType("text/plain"),
	)

	// 添加资源处理器
	s.AddResourceTemplate(template, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// 从 URI 中提取文件路径
		filePath := filepath.Join(targetPath, request.Params.URI)

		// 读取文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}

		// 确定文件的 MIME 类型
		mimeType := "text/plain"
		switch filepath.Ext(filePath) {
		case ".md":
			mimeType = "text/markdown"
		case ".json":
			mimeType = "application/json"
		case ".go":
			mimeType = "text/x-go"
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: mimeType,
				Text:     string(content),
			},
		}, nil
	})

	fmt.Printf("Starting MCP server at %s...\n", targetPath)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runMcpClient(cmd *cobra.Command, args []string) {
	// 初始化 MCP 客户端
	mcpClient, err := client.NewStdioMCPClient(
		mcpCommand,
		mcpEnv,
		append(mcpArgs, "mcp", "server")...,
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer mcpClient.Close()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化客户端并打印服务器信息
	initResult, err := mcpClient.Initialize(ctx, wnmcp.NewInitializeRequest())
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	fmt.Printf("连接到服务器: %s %s\n\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// List Tools
	// fmt.Println("Listing available tools...")
	// toolsRequest := mcp.ListToolsRequest{}
	// tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	// if err != nil {
	// 	log.Fatalf("Failed to list tools: %v", err)
	// }
	// for _, tool := range tools.Tools {
	// 	fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	// }
	// fmt.Println()

	// List Resources
	fmt.Println("Listing available resources...")
	resourcesRequest := mcp.ListResourcesRequest{}
	resources, err := mcpClient.ListResources(context.Background(), resourcesRequest)
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

	// 列出所有资源模板
	fmt.Println("=== 资源模板列表 ===")
	templates, err := mcpClient.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
	if err != nil {
		log.Fatalf("获取资源模板列表失败: %v", err)
	}
	for _, template := range templates.ResourceTemplates {
		fmt.Printf("名称: %s\n", template.Name)
		if template.Description != "" {
			fmt.Printf("描述: %s\n", template.Description)
		}
		fmt.Printf("MIME类型: %s\n", template.MIMEType)
		fmt.Println()
	}

	// 尝试读取文件列表
	fmt.Println("=== 项目文件列表 ===")
	fileList, err := mcpClient.ReadResource(ctx, wnmcp.NewReadResourceRequest("files://list", nil))
	if err != nil {
		log.Printf("读取文件列表失败: %v\n", err)
	} else {
		for _, content := range fileList.Contents {
			if textContent, ok := content.(mcp.TextResourceContents); ok {
				var files []string
				if err := json.Unmarshal([]byte(textContent.Text), &files); err != nil {
					log.Printf("解析文件列表失败: %v\n", err)
				} else {
					for _, file := range files {
						fmt.Printf("- %s\n", file)
					}
				}
			}
		}
	}
	fmt.Println()

	// 如果指定了具体文件，则读取文件内容
	if len(args) >= 2 && args[0] == "resources/read" {
		fmt.Printf("=== 读取文件: %s ===\n", args[1])
		result, err := mcpClient.ReadResource(ctx, wnmcp.NewReadResourceRequest(args[1], nil))
		if err != nil {
			log.Fatalf("读取文件失败: %v", err)
		}
		for _, content := range result.Contents {
			if textContent, ok := content.(mcp.TextResourceContents); ok {
				fmt.Printf("MIME类型: %s\n", textContent.MIMEType)
				fmt.Printf("内容:\n%s\n", textContent.Text)
			}
		}
	}
}
