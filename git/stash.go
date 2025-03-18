package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CreateStash 创建一个新的 stash
func CreateStash(message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	return ExecGitCommand("git", args...)
}

// ListStash 列出所有的 stash
func ListStash() error {
	return ExecGitCommand("git", "stash", "list")
}

// ApplyStash 应用指定的 stash
// index: stash 的索引，如 "0" 表示最新的 stash
// keepStash: 是否保留 stash
func ApplyStash(index string, keepStash bool) error {
	if index == "" {
		index = "0"
	}

	stashRef := fmt.Sprintf("stash@{%s}", index)

	if keepStash {
		return ExecGitCommand("git", "stash", "apply", stashRef)
	}
	return ExecGitCommand("git", "stash", "pop", stashRef)
}

// DropStash 删除指定的 stash
// index: stash 的索引，如 "0" 表示最新的 stash
// dropAll: 是否删除所有 stash
func DropStash(index string, dropAll bool) error {
	if dropAll {
		return ExecGitCommand("git", "stash", "clear")
	}

	if index == "" {
		index = "0"
	}

	stashRef := fmt.Sprintf("stash@{%s}", index)
	return ExecGitCommand("git", "stash", "drop", stashRef)
}

// ShowStash 显示指定 stash 的详细内容
func ShowStash(index string) error {
	if index == "" {
		index = "0"
	}

	stashRef := fmt.Sprintf("stash@{%s}", index)
	return ExecGitCommand("git", "stash", "show", "-p", stashRef)
}

// BranchFromStash 从指定的 stash 创建新分支
func BranchFromStash(branchName string, index string) error {
	if index == "" {
		index = "0"
	}

	stashRef := fmt.Sprintf("stash@{%s}", index)
	return ExecGitCommand("git", "stash", "branch", branchName, stashRef)
}

// ApplyStashByMessage 通过 message 查找并应用 stash
func ApplyStashByMessage(message string, keepStash bool) error {
	// 执行 git stash list 命令并获取输出
	cmd := exec.Command("git", "stash", "list")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// 按行分割输出
	stashList := strings.Split(string(output), "\n")

	// 查找包含指定 message 的 stash
	for i, stash := range stashList {
		if strings.Contains(stash, message) {
			// 找到匹配的 stash，调用 ApplyStash
			return ApplyStash(fmt.Sprintf("%d", i), keepStash)
		}
	}

	return fmt.Errorf("no stash found with message: %s", message)
}
