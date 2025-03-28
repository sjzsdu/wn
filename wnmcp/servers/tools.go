package servers

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
)

func NewTool(project *project.Project) {
	name := project.GetName()

	read(project, name)
	write(project, name)
}

func read(project *project.Project, name string) {
	readFileTool := mcp.NewTool("readFile",
		mcp.WithDescription("读取指定文件的内容"),
		mcp.WithString("path",
			mcp.Description("文件路径"),
			mcp.Required(),
		),
	)

	wnmcp.McpServer().AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, ok := request.Params.Arguments["path"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid path parameter")
		}

		content, err := project.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取文件失败: %v", err)
		}
		return mcp.NewToolResultText(string(content)), nil
	})
}

func write(project *project.Project, name string) {
	writeFileTool := mcp.NewTool("writeFile",
		mcp.WithDescription("写入内容到指定文件"),
		mcp.WithString("path",
			mcp.Description("文件路径"),
			mcp.Required(),
		),
		mcp.WithString("content",
			mcp.Description("要写入的内容"),
			mcp.Required(),
		),
	)

	wnmcp.McpServer().AddTool(writeFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, ok := request.Params.Arguments["path"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid path parameter")
		}

		content, ok := request.Params.Arguments["content"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid content parameter")
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("写入文件失败: %v", err)
		}

		return mcp.NewToolResultText(fmt.Sprintf("成功写入文件: %s", path)), nil
	})
}
