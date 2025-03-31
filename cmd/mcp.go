package cmd

import (
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
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

	mcpCommand string
	mcpEnv     []string
	mcpArgs    []string
	mcpAction  string
	sourceUrl  string
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

	mcpClientCmd.PersistentFlags().StringVar(&mcpCommand, "cmd", "wn", lang.T("MCP server command"))
	mcpClientCmd.PersistentFlags().StringSliceVar(&mcpEnv, "env", nil, lang.T("MCP server environtment"))
	mcpClientCmd.PersistentFlags().StringSliceVar(&mcpArgs, "args", nil, lang.T("MCP server command arguments"))
	mcpClientCmd.PersistentFlags().StringVar(&mcpAction, "action", "", lang.T("MCP server command arguments"))
	mcpClientCmd.PersistentFlags().StringVar(&sourceUrl, "path", "", lang.T("Source URL"))
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
	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}

	// 构建项目树
	project, _ := project.BuildProjectTree(targetPath, options)

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

	wnClient := wnmcp.NewClient(mcpClient, project)

	wnClient.Ping()

	// 列出所有资源模板
	wnClient.ListResources()
	wnClient.ReadResources()
	if sourceUrl != "" {
		wnClient.ReadResource(sourceUrl, map[string]interface{}{})
	}

	wnClient.ListResourceTemplates()

	// 列出所有提示
	wnClient.ListPrompts()

	// 列出所有工具
	wnClient.ListTools()

}
