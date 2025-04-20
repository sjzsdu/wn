package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
	"github.com/spf13/cobra"
)

var (
	mcpServer string
	mcpAction string
	mcpArgs   []string
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: lang.T("start mcp client"),
	Long:  lang.T("start mcp client with specified configuration"),
	Run:   runClient,
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringSliceVar(&mcpArgs, "args", nil, lang.T("MCP server command arguments"))
	clientCmd.Flags().StringVar(&mcpAction, "action", "", lang.T("MCP server action"))
	clientCmd.Flags().StringVar(&mcpServer, "server", "", lang.T("MCP server name"))
}

func runClient(cmd *cobra.Command, args []string) {
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