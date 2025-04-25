package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/project"
	"github.com/sjzsdu/wn/wnmcp"
)

func GetProject() *project.Project {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		return nil
	}
	options := helper.WalkDirOptions{
		DisableGitIgnore: disableGitIgnore,
		Extensions:       extensions,
		Excludes:         excludes,
	}
	project, _ := project.BuildProjectTree(targetPath, options)
	return project
}

func GetMcpHost() *wnmcp.Host {
	targetPath, ferr := helper.GetTargetPath(cmdPath, gitURL)
	if ferr != nil {
		return nil
	}
	mcpConfig, err := wnmcp.LoadMCPConfig(targetPath, configFile)

	project := GetProject()

	host, err := wnmcp.NewHost(mcpConfig, project)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return nil
	}

	defer host.Close()
	return host
}
