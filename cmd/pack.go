package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/project"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: lang.T("Pack files"),
	Long:  lang.T("Pack files with specified extensions into a single output file"),
	Run:   runPack,
}

func init() {
	rootCmd.AddCommand(packCmd)
}

func runPack(cmd *cobra.Command, args []string) {
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

	err = doc.Export(output)

	if err != nil {
		fmt.Printf("Error packing files: %v\n", err)
		return
	}

	fmt.Printf("Successfully packed files into %s\n", output)
}
