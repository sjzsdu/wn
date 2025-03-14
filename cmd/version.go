package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: lang.T("Print version information"),
	Long:  lang.T("Print detailed version information of wn"),
	Run: func(cmd *cobra.Command, args []string) {
		// 使用简单的字符串拼接替代模板
		fmt.Printf("%s: %s\n", lang.T("wn version"), share.VERSION)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
