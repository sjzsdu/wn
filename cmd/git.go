package cmd

import (
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/git"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: lang.T("git command"),
	Long:  lang.T("Extend the git command"),
}

var (
	commitHash    string
	modifiedFiles string
	branchName    string
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: lang.T("Append changes to previous commit"),
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepo {
			fmt.Println(lang.T("Current directory is not a git repository"))
			return
		}

		targetCommit := commitHash // 默认使用 flag 中的值
		if len(args) > 0 {
			targetCommit = args[0] // 如果提供了位置参数，则使用位置参数
		}
		if targetCommit == "" {
			fmt.Println(lang.T("Commit hash name is required"))
			return
		}
		git.AppendCommit(commitHash, modifiedFiles)
	},
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase [branch]",
	Short: lang.T("Rebase current branch to target commit"),
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepo {
			fmt.Println(lang.T("Current directory is not a git repository"))
			return
		}

		targetBranch := branchName
		targetCommit := commitHash

		// 根据参数数量处理
		switch len(args) {
		case 2:
			targetBranch = args[0]
			targetCommit = args[1]
		case 1:
			targetCommit = args[0]
		}

		if targetBranch == "" {
			fmt.Println(lang.T("Branch name is required"))
			return
		}

		git.RebaseBranch(targetBranch, targetCommit)
	},
}

var listCmd = &cobra.Command{
	Use:   "list [from] [to]",
	Short: lang.T("List commits between two commits"),
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepo {
			fmt.Println(lang.T("Current directory is not a git repository"))
			return
		}

		fromCommit := commitHash
		toCommit := git.LatestCommit

		// 根据参数数量处理
		switch len(args) {
		case 2:
			fromCommit = args[0]
			toCommit = args[1]
		case 1:
			fromCommit = args[0]
		}

		commits, err := git.GetCommitsBetween(fromCommit, toCommit)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(strings.Join(commits, " "))
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	// 为命令添加全局标志
	gitCmd.PersistentFlags().StringVarP(&commitHash, "commit", "c", "", "指定提交哈希")
	gitCmd.PersistentFlags().StringVarP(&modifiedFiles, "files", "f", "", "指定修改的文件，多个文件用逗号分隔")
	gitCmd.PersistentFlags().StringVarP(&branchName, "branch", "b", "", "指定分支名称")

	// 添加子命令
	gitCmd.AddCommand(commitCmd)
	gitCmd.AddCommand(rebaseCmd)
	gitCmd.AddCommand(listCmd)
}
