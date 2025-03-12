package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: lang.T("Print version information"),
	Long:  lang.T("Print detailed version information of wn"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wn version 1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
