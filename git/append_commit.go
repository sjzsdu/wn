package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
)

func AppendCommit(commitHash string, modifiedFiles string) {
	if err := helper.CheckFilesExist(modifiedFiles); err != nil {
		fmt.Printf(lang.T(err.Error()) + "\n")
		return
	}

	if err := ExecGitCommand("git", "cat-file", "-e", commitHash); err != nil {
		fmt.Printf(lang.T("Commit not found: %s")+"\n", commitHash)
		return
	}

	currentBranchName := CurrentBranch

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
	if err := ExecGitCommand("git", "checkout", "-b", tempBranch, commitHash); err != nil {
		return
	}

	// 如果是指定文件，则从当前分支复制这些文件
	if modifiedFiles != "." {
		for _, file := range strings.Split(modifiedFiles, ",") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			if err := ExecGitCommand("git", "checkout", currentBranchName, "--", file); err != nil {
				ExecGitCommand("git", "checkout", currentBranchName)
				ExecGitCommand("git", "branch", "-D", tempBranch)
				return
			}
		}
	} else {
		// 如果是修改所有文件，直接从当前分支复制所有改动
		if err := ExecGitCommand("git", "checkout", currentBranchName, "--", "."); err != nil {
			ExecGitCommand("git", "checkout", currentBranchName)
			ExecGitCommand("git", "branch", "-D", tempBranch)
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
			if err := ExecGitCommand("git", "add", file); err != nil {
				ExecGitCommand("git", "checkout", currentBranchName)
				ExecGitCommand("git", "branch", "-D", tempBranch)
				return
			}
		}
	} else {
		if err := ExecGitCommand("git", "add", "."); err != nil {
			ExecGitCommand("git", "checkout", currentBranchName)
			ExecGitCommand("git", "branch", "-D", tempBranch)
			return
		}
	}

	// 提交修改
	if err := ExecGitCommand("git", "commit", "--amend", "--no-edit"); err != nil {
		ExecGitCommand("git", "checkout", currentBranchName)
		ExecGitCommand("git", "branch", "-D", tempBranch)
		return
	}

	// 应用后续的提交
	for _, commit := range needCherryPicks {
		if commit == "" {
			continue
		}
		if err := ExecGitCommand("git", "cherry-pick", commit); err != nil {
			// cherry-pick 失败，可能是冲突
			fmt.Printf(lang.T("Failed to cherry-pick commit %s, please fix it manually")+"\n", commit)
			ExecGitCommand("git", "cherry-pick", "--abort")
			ExecGitCommand("git", "checkout", currentBranchName)
			ExecGitCommand("git", "branch", "-D", tempBranch)
			return
		}
	}

	// 切回原分支并重置
	if err := ExecGitCommand("git", "checkout", currentBranchName); err != nil {
		return
	}

	if err := ExecGitCommand("git", "reset", "--hard", tempBranch); err != nil {
		return
	}

	// 清理
	ExecGitCommand("git", "branch", "-D", tempBranch)
}
