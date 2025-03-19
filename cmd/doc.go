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
		// 确保目录存在
		if err := os.MkdirAll("./docs", 0755); err != nil {
			fmt.Printf("Failed to create docs directory: %v\n", err)
			return
		}

		// 生成默认的 index.md
		indexContent := "# WN CLI Documentation\n\nWelcome to WN CLI documentation."
		if err := os.WriteFile("./docs/index.md", []byte(indexContent), 0644); err != nil {
			fmt.Printf("Failed to create index.md: %v\n", err)
			return
		}

		// 生成命令文档
		err := doc.GenMarkdownTree(rootCmd, "./docs")
		if err != nil {
			fmt.Printf("Failed to generate documentation: %v\n", err)
			return
		}
		fmt.Println("Documentation generated successfully in ./docs directory")
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
