package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func RebaseBranch(branch string, commitHash string) error {
	// 1. 获取从 commitHash 到最新的所有 commit
	cmd := exec.Command("git", "rev-list", commitHash+"..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("获取 commit 列表失败: %v", err)
	}
	
	// 将输出转换为 commit 数组，并反转顺序（因为 rev-list 是从新到旧）
	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	needCherryPick := make([]string, 0, len(commits))
	for i := len(commits) - 1; i >= 0; i-- {
		needCherryPick = append(needCherryPick, commits[i])
	}
	// 添加起始的 commitHash
	needCherryPick = append([]string{commitHash}, needCherryPick...)

	// 2. reset --hard 到指定分支
	cmd = exec.Command("git", "reset", "--hard", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("reset 到分支 %s 失败: %v", branch, err)
	}

	// 3. 依次 cherry-pick 所有 commit
	for _, commit := range needCherryPick {
		cmd = exec.Command("git", "cherry-pick", commit)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cherry-pick commit %s 失败: %v", commit, err)
		}
	}

	return nil
}
