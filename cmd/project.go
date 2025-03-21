package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/project"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: lang.T("Project files"),
	Long:  lang.T("Project files with specified extensions into a single output file"),
	Run:   runproject,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}

func runproject(cmd *cobra.Command, args []string) {
	if output == "" {
		fmt.Printf("Output is required")
		return
	}
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
	doc, err := project.BuildProjectTree(targetPath, options)
	if err != nil {
		fmt.Printf("failed to build project tree: %v\n", err)
		return
	}

	// 导出为 PDF 文件
	fmt.Println(doc)

	fmt.Printf("Successfully exported project to %s\n", output)
}
