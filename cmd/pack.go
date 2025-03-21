package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/output/file"
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

	// 检查项目树是否为空
	if doc.IsEmpty() {
		fmt.Printf("No files to pack\n")
		return
	}

	// 根据输出文件扩展名选择导出格式
	switch filepath.Ext(output) {
	case ".md":
		exporter := file.NewMarkdownExporter(doc)
		err = exporter.Export(output)
	case ".pdf":
		exporter, err := file.NewPDFExporter(doc)
		if err != nil {
			fmt.Printf("Error creating PDF exporter: %v\n", err)
			return
		}
		exporter.Export(output)
	case ".xml":
		exporter := file.NewXMLExporter(doc)
		err = exporter.Export(output)
	default:
		fmt.Printf("Unsupported output format: %s\n", filepath.Ext(output))
		return
	}

	if err != nil {
		fmt.Printf("Error packing files: %v\n", err)
		return
	}

	fmt.Printf("Successfully packed files into %s\n", output)
}
