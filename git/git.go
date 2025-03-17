package git

import (
	"os"
	"os/exec"
	"strings"
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
	if err := ExecGitCommand("git", "rev-parse", "--git-dir"); err != nil {
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
