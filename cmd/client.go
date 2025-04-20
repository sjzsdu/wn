package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
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
	host, err := wnmcp.NewHost(mcpConfig, project)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}
	// 关闭所有客户端
	defer host.Close()

	// 执行指定的操作
	if mcpAction == "" {
		fmt.Println("未指定操作，执行默认操作...")
		executeDefaultActions(host)
		return
	}

	// 准备参数
	ctx := context.Background()
	var actionErr error

	// 根据 mcpServer 选择客户端
	if mcpServer != "" {
		client := host.GetClient(mcpServer)
		if client == nil {
			fmt.Printf("未找到指定的服务器 %s\n", mcpServer)
			return
		}
		_, actionErr = executeAction(ctx, client, mcpAction, mcpArgs)
	} else {
		// 对所有客户端执行操作
		for name, client := range host.GetAllClients() {
			fmt.Printf("\n执行 %s 客户端操作...\n", name)
			if _, err := executeAction(ctx, client, mcpAction, mcpArgs); err != nil {
				actionErr = err
			}
		}
	}

	if actionErr != nil {
		fmt.Printf("执行操作失败: %v\n", actionErr)
	}
}

func executeAction(ctx context.Context, client *wnmcp.Client, action string, args []string) (interface{}, error) {
	switch action {
	case "ping":
		return nil, client.Ping(ctx)
	case "list-resources":
		return client.ListResources(ctx, mcp.ListResourcesRequest{})
	case "read-resources":
		return client.ReadResource(ctx, mcp.ReadResourceRequest{
			Params: struct {
				URI       string                 `json:"uri"`
				Arguments map[string]interface{} `json:"arguments,omitempty"`
			}{
				URI: "files://" + client.GetProjectName(),
			},
		})
	case "list-templates":
		return client.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
	case "list-prompts":
		return client.ListPrompts(ctx, mcp.ListPromptsRequest{})
	case "list-tools":
		return client.ListTools(ctx, mcp.ListToolsRequest{})
	case "read-resource":
		if len(args) < 1 {
			return nil, fmt.Errorf("read-resource 需要指定资源路径")
		}
		return client.ReadResource(ctx, mcp.ReadResourceRequest{
			Params: struct {
				URI       string                 `json:"uri"`
				Arguments map[string]interface{} `json:"arguments,omitempty"`
			}{
				URI: args[0],
			},
		})
	case "call-tool":
		if len(args) < 2 {
			return nil, fmt.Errorf("call-tool 需要指定工具名称和参数")
		}
		toolArgs := make(map[string]interface{})
		if err := json.Unmarshal([]byte(args[1]), &toolArgs); err != nil {
			return nil, fmt.Errorf("解析工具参数失败: %v", err)
		}
		return client.CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{},
			Params:  mcp.CallToolRequest{}.Params, // 使用零值初始化内嵌结构体
		})
	default:
		return nil, fmt.Errorf("不支持的操作: %s", action)
	}
}

func executeDefaultActions(host *wnmcp.Host) {
	ctx := context.Background()
	for name, client := range host.GetAllClients() {
		fmt.Printf("\n执行 %s 客户端操作...\n", name)

		// 列出资源
		if result, err := client.ListResources(ctx, mcp.ListResourcesRequest{}); err != nil {
			fmt.Printf("ListResources 失败: %v\n", err)
		} else {
			fmt.Printf("获取到 %d 个资源\n", len(result.Resources))
		}

		// 读取资源
		if result, err := client.ReadResource(ctx, mcp.ReadResourceRequest{
			Params: struct {
				URI       string                 `json:"uri"`
				Arguments map[string]interface{} `json:"arguments,omitempty"`
			}{
				URI: "files://" + client.GetProjectName(),
			},
		}); err != nil {
			fmt.Printf("ReadResource 失败: %v\n", err)
		} else {
			fmt.Printf("读取资源成功: %v\n", result)
		}

		// 列出资源模板
		if result, err := client.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{}); err != nil {
			fmt.Printf("ListResourceTemplates 失败: %v\n", err)
		} else {
			fmt.Printf("获取到 %d 个资源模板\n", len(result.ResourceTemplates))
		}

		// 列出提示
		if result, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{}); err != nil {
			fmt.Printf("ListPrompts 失败: %v\n", err)
		} else {
			fmt.Printf("获取到 %d 个提示\n", len(result.Prompts))
		}

		// 列出工具
		if result, err := client.ListTools(ctx, mcp.ListToolsRequest{}); err != nil {
			fmt.Printf("ListTools 失败: %v\n", err)
		} else {
			fmt.Printf("获取到 %d 个工具\n", len(result.Tools))
		}
	}
}
