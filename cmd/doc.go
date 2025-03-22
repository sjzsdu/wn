package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

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

		zhRootCmd := *rootCmd
		updateCommandsLocalization(&zhRootCmd)

		// 复制 README.md 作为中文文档索引
		readmeContent, err := os.ReadFile("README.md")
		if err != nil {
			fmt.Printf("Failed to read README.md: %v\n", err)
			return
		}
		if err := os.WriteFile("./docs/zh/index.md", readmeContent, 0644); err != nil {
			fmt.Printf("Failed to create Chinese index.md: %v\n", err)
			return
		}

		// 遍历所有命令生成中文文档
		if err := generateDocsForCommand(&zhRootCmd, "zh"); err != nil {
			fmt.Printf("Failed to generate Chinese documentation: %v\n", err)
			return
		}

		// 生成英文文档
		os.Setenv("WN_LANG", "en")
		lang.SetupI18n("")

		// 复制 README.en.md 作为英文文档索引
		readmeEnContent, err := os.ReadFile("README.en.md")
		if err != nil {
			fmt.Printf("Failed to read README.en.md: %v\n", err)
			return
		}
		if err := os.WriteFile("./docs/en/index.md", readmeEnContent, 0644); err != nil {
			fmt.Printf("Failed to create English index.md: %v\n", err)
			return
		}

		// 生成英文索引
		indexEn := "# WN CLI Documentation\n\nWelcome to WN CLI documentation."
		if err := os.WriteFile("./docs/en/index.md", []byte(indexEn), 0644); err != nil {
			fmt.Printf("Failed to create English index.md: %v\n", err)
			return
		}

		// 遍历所有命令生成英文文档
		if err := generateDocsForCommand(rootCmd, "en"); err != nil {
			fmt.Printf("Failed to generate English documentation: %v\n", err)
			return
		}

		fmt.Println("Documentation generated successfully in ./docs/zh and ./docs/en directories")
	},
}

// 添加新的辅助函数用于递归生成文档
func generateDocsForCommand(cmd *cobra.Command, langCode string) error {
	// 先处理当前命令
	if err := generateDocWithTemplate(cmd, langCode); err != nil {
		return fmt.Errorf("error generating doc for command %s: %v", cmd.Name(), err)
	}

	// 递归处理所有子命令
	for _, subCmd := range cmd.Commands() {
		if err := generateDocsForCommand(subCmd, langCode); err != nil {
			return err
		}
	}

	return nil
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

// 添加辅助函数
func getCommandExamples(cmd *cobra.Command) []string {
	if cmd.Example == "" {
		return nil
	}
	return strings.Split(cmd.Example, "\n")
}

func getDetailedDescription(cmd *cobra.Command) string {
	if cmd.Long != "" {
		return cmd.Long
	}
	return cmd.Short
}

// 修改模板处理函数
func generateDocWithTemplate(cmd *cobra.Command, langCode string) error {
	// 获取模板文件
	templatePath := filepath.Join("templates", "docs", langCode, cmd.Name()+".tmpl")
	if _, err := os.Stat(templatePath); err == nil {
		// 如果存在模板，使用模板生成
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return err
		}

		data := struct {
			Command       *cobra.Command
			UsageExamples []string
			DetailedDesc  string
		}{
			Command:       cmd,
			UsageExamples: getCommandExamples(cmd),
			DetailedDesc:  getDetailedDescription(cmd),
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return err
		}

		// 写入生成的文档
		outputPath := filepath.Join("docs", langCode, cmd.Name()+".md")
		return os.WriteFile(outputPath, buf.Bytes(), 0644)
	}

	// 如果没有模板，使用默认的 cobra 文档生成
	outputPath := filepath.Join("docs", langCode, cmd.Name()+".md")
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return doc.GenMarkdown(cmd, f)
}
