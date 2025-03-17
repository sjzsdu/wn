package git

import (
	"fmt"
)

func RebaseBranch(branch string, commitHash string) error {
	_currentBranch := CurrentBranch
	_head := LatestCommit

	// 创建临时分支保存当前状态
	tempBranch := "temp_rebase_" + _currentBranch
	if err := ExecGitCommand("git", "branch", tempBranch); err != nil {
		return fmt.Errorf("创建临时分支失败: %v", err)
	}
	defer func() {
		ExecGitCommand("git", "branch", "-D", tempBranch)
	}()

	commitList, err := GetCommitsBetween(commitHash, "HEAD")
	if err != nil {
		return err
	}

	ExecGitCommand("git", "checkout", _currentBranch)
	// reset --hard 到目标分支
	if err := ExecGitCommand("git", "reset", "--hard", branch); err != nil {
		return fmt.Errorf("reset 到分支 %s 失败: %v", branch, err)
	}

	// 一次性 cherry-pick 所有提交
	args := append([]string{"cherry-pick"}, commitList...)
	if err := ExecGitCommand("git", args...); err != nil {
		// 发生错误时，中止 cherry-pick 并回滚到原始状态
		ExecGitCommand("git", "cherry-pick", "--abort")
		ExecGitCommand("git", "reset", "--hard", _head)
		return fmt.Errorf("cherry-pick 提交失败: %v", err)
	}

	return nil
}
