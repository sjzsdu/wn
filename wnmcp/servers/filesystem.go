package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
)

func NewProject(project *project.Project) {
	name := project.GetName()
	// 添加项目文件列表资源
	fileListResource := mcp.NewResource(
		"files://"+name,
		"Project: "+name+"(Files List)",
		mcp.WithResourceDescription("List all files in the project: "+name),
		mcp.WithMIMEType("application/json"),
	)

	wnmcp.McpServer().AddResource(fileListResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		files, _ := project.GetAllFiles()

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

	templateSource(project, name)
}

func templateSource(project *project.Project, name string) {

	// 添加动态资源模板
	template := mcp.NewResourceTemplate(
		name+"://{path}",
		"Project Files",
		mcp.WithTemplateDescription("Access project files"),
		mcp.WithTemplateMIMEType("text/plain"),
	)

	// 添加资源处理器
	wnmcp.McpServer().AddResourceTemplate(template, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// 从 URI 中提取文件路径
		fmt.Println("request.Params.URI: ", request.Params.URI)
		filePath := request.Params.URI
		content, err := project.ReadFile(filePath)

		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}

		// 确定文件的 MIME 类型
		mimeType := helper.GetMimeType(filepath.Ext(filePath))

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: mimeType,
				Text:     string(content),
			},
		}, nil
	})
}
