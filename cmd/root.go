package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var cmdPath string

var lang = os.Getenv("WN_LANG")
var langs = map[string]string{
	"One or more arguments are not correct": "参数错误",
	"work directory":                        "工作目录",
	"Pack files":                            "打包文件",
	"Pack files with specified extensions into a single output file": "将指定扩展名的文件打包成单个输出文件",
}

// L Language switch
func L(words string) string {
	if lang == "" {
		return words
	}

	if trans, has := langs[words]; has {
		return trans
	}
	return words
}

var RootCmd = rootCmd

var rootCmd = &cobra.Command{
	Use:   share.BUILDNAME,
	Short: "Wn command line tool",
	Long:  `Wn command line tool`,
	Args:  cobra.MinimumNArgs(1),
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, L("One or more arguments are not correct"), args)
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
	rootCmd.PersistentFlags().StringVarP(&cmdPath, "workPath", "p", "", L("work directory"))
}
