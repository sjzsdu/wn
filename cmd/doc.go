package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag" // 添加 pflag 包导入
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

		// 保存当前环境变量
		originalLang := os.Getenv("WN_LANG")
		defer os.Setenv("WN_LANG", originalLang)

		// 生成中文文档
		os.Setenv("WN_LANG", "zh")
		lang.SetupI18n("")

		// 创建中文版本的根命令副本
		zhRootCmd := *rootCmd
		// 更新所有命令的本地化文本
		updateCommandsLocalization(&zhRootCmd)

		indexZh := "# WN CLI 文档\n\n欢迎使用 WN CLI 工具。"
		if err := os.WriteFile("./docs/zh/index.md", []byte(indexZh), 0644); err != nil {
			fmt.Printf("Failed to create Chinese index.md: %v\n", err)
			return
		}
		if err := doc.GenMarkdownTree(&zhRootCmd, "./docs/zh"); err != nil {
			fmt.Printf("Failed to generate Chinese documentation: %v\n", err)
			return
		}

		// 生成英文文档
		os.Setenv("WN_LANG", "en")
		lang.SetupI18n("")
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

// 更新命令及其子命令的本地化文本
func updateCommandsLocalization(cmd *cobra.Command) {
	// 更新当前命令的本地化文本
	if cmd.Short != "" {
		cmd.Short = lang.T(cmd.Short)
	}
	if cmd.Long != "" {
		cmd.Long = lang.T(cmd.Long)
	}
	if cmd.Example != "" {
		cmd.Example = lang.T(cmd.Example)
	}

	// 更新所有标志的本地化文本
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Usage != "" {
			f.Usage = lang.T(f.Usage)
		}
	})

	// 递归更新所有子命令
	for _, subCmd := range cmd.Commands() {
		updateCommandsLocalization(subCmd)
	}
}

func init() {
	rootCmd.AddCommand(docCmd)
}
