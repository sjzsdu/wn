package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

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

	outputExt := strings.ToLower(filepath.Ext(output))

	switch outputExt {
	case ".pdf":
		if err := doc.ExportToPDF(output); err != nil {
			fmt.Printf("failed to export PDF: %v\n", err)
			return
		}
	case ".md":
		if err := doc.ExportToMarkdown(output); err != nil {
			fmt.Printf("failed to export Markdown: %v\n", err)
			return
		}
	case ".xml":
		if err := doc.ExportToXML(output); err != nil {
			fmt.Printf("failed to export xml: %v\n", err)
			return
		}
	default:
		fmt.Printf("Output file format only support pdf, md, xml")
		return
	}

	if err != nil {
		fmt.Printf("Error packing files: %v\n", err)
		return
	}

	fmt.Printf("Successfully packed files into %s\n", output)
}
