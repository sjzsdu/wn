package cmd

import (
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
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: lang.T("start mcp server"),
	Long:  lang.T("start mcp server with specified configuration"),
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&mcpLayer, "layer", "stdio", lang.T("MCP transfer layer"))
	serverCmd.Flags().StringVar(&mcpPort, "port", "9595", lang.T("MCP sse port"))
}

func runServer(cmd *cobra.Command, args []string) {
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