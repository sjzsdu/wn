package cmd

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: lang.T("start chat"),
	Long:  lang.T("start chat with specified configuration"),
	Run:   runChat,
}

var configFile string

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringVar(&configFile, "config", "", lang.T("Config file"))
}

func runChat(cmd *cobra.Command, args []string) {
	// 使用 GetMcpHost 获取 host
	host := GetMcpHost()
	if host == nil {
		return
	}

	tools := host.GetTools(context.Background(), mcp.ListToolsRequest{})
	chat, _ := aigc.NewChat(aigc.ChatOptions{
		UseAgent: "blog",
		Hooks:    &aigc.Hooks{},
		Tools:    tools,
	})
	// 启动交互式会话
	ctx := context.Background()
	chat.StartInteractiveSession(ctx, aigc.InteractiveOptions{
		Renderer: helper.GetDefaultRenderer(),
		Debug:    false,
	})

}
