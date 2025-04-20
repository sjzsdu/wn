package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
	"github.com/sjzsdu/wn/wnmcp/servers"
	"github.com/spf13/cobra"
)

var (
	mcpLayer string
	mcpPort  string

	mcpServer string
	mcpAction string
	mcpArgs   []string
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
	mcpCmd.PersistentFlags().StringVar(&mcpPort, "port", "9595", lang.T("MCP sse port"))
	mcpCmd.AddCommand(mcpServerCmd)
	mcpCmd.AddCommand(mcpClientCmd)

	mcpClientCmd.PersistentFlags().StringSliceVar(&mcpArgs, "args", nil, lang.T("MCP server command arguments"))
	mcpClientCmd.PersistentFlags().StringVar(&mcpAction, "action", "", lang.T("MCP server action"))
	mcpClientCmd.PersistentFlags().StringVar(&mcpServer, "server", "", lang.T("MCP server name"))
}

func runMcpServer(cmd *cobra.Command, args []string) {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		fmt.Printf("failed to get target path: %v\n", ferr)
		return
	}

	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}

	// 构建项目树
	project, perr := project.BuildProjectTree(targetPath, options)
	if perr != nil {
		fmt.Printf("failed to build project tree: %v\n", perr)
		return
	}
	servers.NewResource(project)
	servers.NewTool(project)
	servers.NewPrompt(project)

	fmt.Printf("Starting MCP server at %s...\n", targetPath)

	var err error
	switch mcpLayer {
	case "sse":
		if !helper.IsValidPort(mcpPort) {
			log.Fatalf("无效的端口号: %s", mcpPort)
		}
		sseServer := server.NewSSEServer(wnmcp.McpServer())
		err = sseServer.Start(":" + mcpPort)
	case "stdio":
		err = server.ServeStdio(wnmcp.McpServer())
	default:
		log.Fatalf("不支持的传输层: %s", mcpLayer)
	}
	if err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

func runMcpClient(cmd *cobra.Command, args []string) {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		fmt.Printf("failed to get target path: %v\n", ferr)
		return
	}

	// 加载 MCP 配置
	mcpConfig, err := wnmcp.LoadMCPConfig(targetPath)
	if err != nil {
		fmt.Printf("加载 MCP 配置失败: %v\n", err)
		return
	}

	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}
	project, _ := project.BuildProjectTree(targetPath, options)

	// 创建多个客户端
	clients, err := wnmcp.NewClients(mcpConfig, project)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}
	// 关闭所有客户端
	defer clients.Close()

	// 执行指定的操作
	if mcpAction == "" {
		fmt.Println("未指定操作，执行默认操作...")
		executeDefaultActions(clients)
		return
	}

	// 准备参数
	ctx := context.Background()
	var actionErr error

	// 根据 mcpServer 选择客户端
	if mcpServer != "" {
		client := clients.GetClient(mcpServer)
		if client == nil {
			fmt.Printf("未找到指定的服务器 %s\n", mcpServer)
			return
		}
		actionErr = executeAction(ctx, client, mcpAction, mcpArgs)
	} else {
		// 对所有客户端执行操作
		for name, client := range clients.GetAllClients() {
			fmt.Printf("\n执行 %s 客户端操作...\n", name)
			if err := executeAction(ctx, client, mcpAction, mcpArgs); err != nil {
				actionErr = err
			}
		}
	}

	if actionErr != nil {
		fmt.Printf("执行操作失败: %v\n", actionErr)
	}
}

// executeAction 执行指定的操作
func executeAction(ctx context.Context, client *wnmcp.Client, action string, args []string) error {
	switch action {
	case "ping":
		return client.Ping()
	case "list-resources":
		return client.ListResources()
	case "read-resources":
		client.ReadResources()
		return nil
	case "list-templates":
		return client.ListResourceTemplates()
	case "list-prompts":
		return client.ListPrompts()
	case "list-tools":
		return client.ListTools()
	case "read-resource":
		if len(args) < 1 {
			return fmt.Errorf("read-resource 需要指定资源路径")
		}
		return client.ReadResource(args[0], nil)
	case "call-tool":
		if len(args) < 2 {
			return fmt.Errorf("call-tool 需要指定工具名称和参数")
		}
		toolArgs := make(map[string]interface{})
		if err := json.Unmarshal([]byte(args[1]), &toolArgs); err != nil {
			return fmt.Errorf("解析工具参数失败: %v", err)
		}
		return client.CallTool(args[0], toolArgs)
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}
}

// executeDefaultActions 执行默认的操作集
func executeDefaultActions(clients *wnmcp.Clients) {
	for name, client := range clients.GetAllClients() {
		fmt.Printf("\n执行 %s 客户端操作...\n", name)
		client.Ping()
		client.ListResources()
		client.ReadResources()
		client.ListResourceTemplates()
		client.ListPrompts()
		client.ListTools()
	}
}
