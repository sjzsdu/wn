package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var (
	cmdPath          string
	extensions       []string
	output           string
	excludes         []string
	gitURL           string
	disableGitIgnore bool
)

var RootCmd = rootCmd

var rootCmd = &cobra.Command{
	Use:   share.BUILDNAME,
	Short: lang.T("Wn command line tool"),
	Long:  lang.T("Wn command line tool"),
	Args:  cobra.MinimumNArgs(1),
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, lang.T("One or more arguments are not correct"), args)
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
	lang.Init()

	rootCmd.PersistentFlags().StringVarP(&cmdPath, "workPath", "p", "", lang.T("work directory"))
	rootCmd.PersistentFlags().StringSliceVarP(&extensions, "exts", "e", []string{"*"}, lang.T("File extensions to include"))
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "output.md", lang.T("Output file name"))
	rootCmd.PersistentFlags().StringSliceVarP(&excludes, "excludes", "x", []string{}, lang.T("Glob patterns to exclude"))
	rootCmd.PersistentFlags().StringVarP(&gitURL, "git-url", "g", "", lang.T("Git repository URL to clone and pack"))
	rootCmd.PersistentFlags().BoolVarP(&disableGitIgnore, "disable-gitignore", "d", false, lang.T("Disable .gitignore rules"))
}
