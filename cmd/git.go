package cmd

import (
	"fmt"

	"github.com/sjzsdu/wn/git"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: lang.T("git command"),
	Long:  lang.T("Extend the git command"),
	Run:   runGit,
}

var (
	commitHash    string
	modifiedFiles string
	branchName    string
)

func init() {
	rootCmd.AddCommand(gitCmd)
	// 修复 StringVar 参数顺序：变量指针、参数名、短参数名、默认值、描述
	gitCmd.Flags().StringVarP(&commitHash, "commit", "c", git.LatestCommit, lang.T("commit hash"))
	gitCmd.Flags().StringVarP(&modifiedFiles, "files", "f", ".", lang.T("files to modify"))
	gitCmd.Flags().StringVarP(&branchName, "branch", "b", ".", lang.T("branch to rebase"))
}

// 避免初始化循环，将 runGit 定义为变量
var runGit = func(cmd *cobra.Command, args []string) {
	if !git.IsGitRepo {
		fmt.Println(lang.T("Current directory is not a git repository"))
		return
	}
	// 修改判断逻辑，两个参数都需要提供
	if commitHash != "" && branchName != "" {
		git.RebaseBranch(branchName, commitHash)
		return
	}
	if commitHash != "" || modifiedFiles != "" {
		// 如果只提供了其中一个参数，提示错误
		fmt.Println(lang.T("Both commit hash and files are required"))
		return
	}
	cmd.Help()
}
