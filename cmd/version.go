package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: L("Wn version"),
	Long:  "Show Wn tool version",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Wn version:  %s\n", share.VERSION)
}
