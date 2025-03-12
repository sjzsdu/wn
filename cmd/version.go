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
		fmt.Println(lang.T("wn version {{.Version}}", map[string]interface{}{
			"Version": share.VERSION,
		}))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
