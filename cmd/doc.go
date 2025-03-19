package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: lang.T("Generate documentation"),
	Long: lang.T(`Generate documentation for all commands.
The documentation will be generated in Markdown format and saved in the docs directory.`),
	Run: func(cmd *cobra.Command, args []string) {
		// 创建中文和英文文档目录
		dirs := []string{"./docs/zh", "./docs/en"}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Failed to create directory %s: %v\n", dir, err)
				return
			}
		}

		// 生成中文文档
		indexZh := "# WN CLI 文档\n\n欢迎使用 WN CLI 工具。"
		if err := os.WriteFile("./docs/zh/index.md", []byte(indexZh), 0644); err != nil {
			fmt.Printf("Failed to create Chinese index.md: %v\n", err)
			return
		}
		if err := doc.GenMarkdownTree(rootCmd, "./docs/zh"); err != nil {
			fmt.Printf("Failed to generate Chinese documentation: %v\n", err)
			return
		}

		// 生成英文文档
		indexEn := "# WN CLI Documentation\n\nWelcome to WN CLI documentation."
		if err := os.WriteFile("./docs/en/index.md", []byte(indexEn), 0644); err != nil {
			fmt.Printf("Failed to create English index.md: %v\n", err)
			return
		}
		if err := doc.GenMarkdownTree(rootCmd, "./docs/en"); err != nil {
			fmt.Printf("Failed to generate English documentation: %v\n", err)
			return
		}

		fmt.Println("Documentation generated successfully in ./docs/zh and ./docs/en directories")
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
