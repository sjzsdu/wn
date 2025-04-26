package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"

	_ "github.com/sjzsdu/wn/llm/providers/claude"
	_ "github.com/sjzsdu/wn/llm/providers/deepseek"
	_ "github.com/sjzsdu/wn/llm/providers/openai"
)

var (
	cmdPath          string
	extensions       []string
	output           string
	excludes         []string
	gitURL           string
	disableGitIgnore bool
	inDebug          bool
	llmName          string
	llmModel         string
	llmAgent         string
	llmMessageLimit  int
)

var RootCmd = rootCmd

var rootCmd = &cobra.Command{
	Use:   share.BUILDNAME,
	Short: lang.T("Wn command line tool"),
	Long:  lang.T("A versatile command line tool for development"),
	// 移除 Args 限制，允许无参数调用
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有参数，显示帮助信息
		if len(args) == 0 {
			cmd.Help()
			return
		}
		fmt.Fprintln(os.Stderr, lang.T("Invalid arguments")+": ", args)
		os.Exit(1)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 确保在初始化时已经加载了语言包
	rootCmd.PersistentFlags().StringVarP(&cmdPath, "workPath", "p", "", lang.T("Work directory path"))
	rootCmd.PersistentFlags().StringSliceVarP(&extensions, "exts", "e", []string{"*"}, lang.T("File extensions to include"))
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", lang.T("Output file name"))
	rootCmd.PersistentFlags().StringSliceVarP(&excludes, "excludes", "x", []string{}, lang.T("Glob patterns to exclude"))
	rootCmd.PersistentFlags().StringVarP(&gitURL, "git-url", "g", "", lang.T("Git repository URL to clone and pack"))
	rootCmd.PersistentFlags().BoolVarP(&disableGitIgnore, "disable-gitignore", "i", false, lang.T("Disable .gitignore rules"))
	rootCmd.PersistentFlags().BoolVarP(&inDebug, "debug", "d", false, lang.T("Debug mode"))

	rootCmd.PersistentFlags().StringVarP(&llmName, "llm-name", "n", "", lang.T("LLM model Provider"))
	rootCmd.PersistentFlags().StringVarP(&llmModel, "llm-model", "m", "", lang.T("LLM model to use"))
	rootCmd.PersistentFlags().StringVarP(&llmAgent, "llm-agent", "a", "", lang.T("AI use agent name"))
	rootCmd.PersistentFlags().IntVarP(&llmMessageLimit, "llm-message-limit", "l", 1000, lang.T("LLM message limit"))
	// 设置全局 debug 模式
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		share.SetDebug(inDebug)
	}
}
