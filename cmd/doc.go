package cmd

import (
	"fmt"

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
		err := doc.GenMarkdownTree(rootCmd, "./docs")
		if err != nil {
			fmt.Println(lang.T("Failed to generate documentation:"), err)
			return
		}
		fmt.Println(lang.T("Documentation generated successfully in ./docs directory"))
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
