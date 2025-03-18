package git

import (
	"fmt"

	"github.com/sjzsdu/wn/lang"
)

func RebaseBranch(branch string, commitHash string) {
	tempBranch := "temp_rebase_" + CurrentBranch
	if err := ExecGitCommand("git", "branch", tempBranch); err != nil {
		fmt.Printf("创建临时分支失败: %v", err)
		return
	}
	defer func() {
		ExecGitCommand("git", "branch", "-D", tempBranch)
	}()

	needCherryPicks, err := GetCommitsBetween(commitHash, "HEAD")
	if err != nil {
		fmt.Printf(lang.T("Failed to get commit list: %s")+"\n", err)
		return
	}
	fmt.Printf(lang.T("Commit list: %s")+"\n", needCherryPicks)

	if err := ExecGitCommand("git", "reset", "--hard", branch); err != nil {
		fmt.Printf("reset 到分支 %s 失败: %v", branch, err)
		return
	}
	ApplyCherryPicks(needCherryPicks)
}
