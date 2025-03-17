package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
)

func init() {
	rootCmd.AddCommand(gitCmd)
	// 修复 StringVar 参数顺序：变量指针、参数名、短参数名、默认值、描述
	gitCmd.Flags().StringVarP(&commitHash, "commit", "c", "", lang.T("commit hash"))
	gitCmd.Flags().StringVarP(&modifiedFiles, "files", "f", ".", lang.T("files to modify"))
}

// 避免初始化循环，将 runGit 定义为变量
var runGit = func(cmd *cobra.Command, args []string) {
	// 修改判断逻辑，两个参数都需要提供
	if commitHash != "" && modifiedFiles != "" {
		editCommit()
		return
	}
	if commitHash != "" || modifiedFiles != "" {
		// 如果只提供了其中一个参数，提示错误
		fmt.Println(lang.T("Both commit hash and files are required"))
		return
	}
	cmd.Help()
}

func editCommit() {
	// 基础检查保持不变
	if modifiedFiles != "." {
		// 检查每个文件是否存在
		for _, file := range strings.Split(modifiedFiles, ",") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			if _, err := os.Stat(file); os.IsNotExist(err) {
				fmt.Printf(lang.T("File not found: %s")+"\n", file)
				return
			}
		}
	}

	if err := execGitCommand("git", "cat-file", "-e", commitHash); err != nil {
		fmt.Printf(lang.T("Commit not found: %s")+"\n", commitHash)
		return
	}

	// 获取当前分支名
	currentBranch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Printf(lang.T("Failed to get current branch: %s")+"\n", err)
		return
	}
	currentBranchName := strings.TrimSpace(string(currentBranch))

	// 获取当前分支到目标提交之后的所有提交
	revListCmd := exec.Command("git", "rev-list", "--reverse", commitHash+"..HEAD")
	revListOutput, err := revListCmd.Output()
	if err != nil {
		fmt.Printf(lang.T("Failed to get commit list: %s")+"\n", err)
		return
	}
	needCherryPicks := strings.Split(strings.TrimSpace(string(revListOutput)), "\n")

	// 创建并切换到临时分支
	tempBranch := fmt.Sprintf("temp-edit-%s", commitHash[:8])
	if err := execGitCommand("git", "checkout", "-b", tempBranch, commitHash); err != nil {
		return
	}

	// 如果是指定文件，则从当前分支复制这些文件
	if modifiedFiles != "." {
		for _, file := range strings.Split(modifiedFiles, ",") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			if err := execGitCommand("git", "checkout", currentBranchName, "--", file); err != nil {
				execGitCommand("git", "checkout", currentBranchName)
				execGitCommand("git", "branch", "-D", tempBranch)
				return
			}
		}
	} else {
		// 如果是修改所有文件，直接从当前分支复制所有改动
		if err := execGitCommand("git", "checkout", currentBranchName, "--", "."); err != nil {
			execGitCommand("git", "checkout", currentBranchName)
			execGitCommand("git", "branch", "-D", tempBranch)
			return
		}
	}

	// 提交修改
	// 添加所有修改过的文件
	if modifiedFiles != "." {
		for _, file := range strings.Split(modifiedFiles, ",") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			if err := execGitCommand("git", "add", file); err != nil {
				execGitCommand("git", "checkout", currentBranchName)
				execGitCommand("git", "branch", "-D", tempBranch)
				return
			}
		}
	} else {
		if err := execGitCommand("git", "add", "."); err != nil {
			execGitCommand("git", "checkout", currentBranchName)
			execGitCommand("git", "branch", "-D", tempBranch)
			return
		}
	}

	// 提交修改
	if err := execGitCommand("git", "commit", "--amend", "--no-edit"); err != nil {
		execGitCommand("git", "checkout", currentBranchName)
		execGitCommand("git", "branch", "-D", tempBranch)
		return
	}

	// 应用后续的提交
	for _, commit := range needCherryPicks {
		if commit == "" {
			continue
		}
		if err := execGitCommand("git", "cherry-pick", commit); err != nil {
			// cherry-pick 失败，可能是冲突
			fmt.Printf(lang.T("Failed to cherry-pick commit %s, please fix it manually")+"\n", commit)
			execGitCommand("git", "cherry-pick", "--abort")
			execGitCommand("git", "checkout", currentBranchName)
			execGitCommand("git", "branch", "-D", tempBranch)
			return
		}
	}

	// 切回原分支并重置
	if err := execGitCommand("git", "checkout", currentBranchName); err != nil {
		return
	}

	if err := execGitCommand("git", "reset", "--hard", tempBranch); err != nil {
		return
	}

	// 清理
	execGitCommand("git", "branch", "-D", tempBranch)
}

func execGitCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf(lang.T("Git command execution failed: %s")+"\n", err)
		return err
	}
	return nil
}
