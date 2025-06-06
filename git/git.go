package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sjzsdu/wn/lang"
)

var (
	// CurrentBranch 当前分支名称
	CurrentBranch string
	// LatestCommit 最新提交的 hash 值
	LatestCommit string
	// GitRoot git 仓库根目录
	GitRoot string
	// IsGitRepo 是否是 git 仓库
	IsGitRepo bool
)

type EditCommitOptions struct {
	CommitHash    string
	ModifiedFiles []string
}

// Init 初始化 git 相关变量
func init() {
	// 检查是否在 git 仓库中
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		IsGitRepo = false
		return
	}
	IsGitRepo = true

	// 获取当前分支
	if output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		CurrentBranch = strings.TrimSpace(string(output))
	}

	// 获取最新提交
	if output, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		LatestCommit = strings.TrimSpace(string(output))
	}

	// 获取 git 仓库根目录
	if output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		GitRoot = strings.TrimSpace(string(output))
	}
}

func ExecGitCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(output)
		return err
	}
	return nil
}

func GetCommitsBetween(from, to string) ([]string, error) {
	cmd := exec.Command("git", "rev-list", "--reverse", fmt.Sprintf("%s..%s", from, to))
	commits, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取提交列表失败: %v", err)
	}
	commitList := strings.Split(strings.TrimSpace(string(commits)), "\n")
	if len(commitList) == 0 || (len(commitList) == 1 && commitList[0] == "") {
		return nil, fmt.Errorf("没有找到需要重放的提交")
	}
	return commitList, nil
}

func ApplyCherryPicks(commits []string) bool {
	// 过滤掉空的 commits
	var validCommits []string
	for _, commit := range commits {
		if commit != "" {
			validCommits = append(validCommits, commit)
		}
	}

	if len(validCommits) == 0 {
		return true
	}

	// 一次性 cherry-pick 所有 commits
	args := append([]string{"cherry-pick"}, validCommits...)
	if err := ExecGitCommand("git", args...); err != nil {
		fmt.Printf(lang.T("Failed to cherry-pick commits, please fix it manually") + "\n")
		ExecGitCommand("git", "cherry-pick", "--abort")
		return false
	}
	return true
}
