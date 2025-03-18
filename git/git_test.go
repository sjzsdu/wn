package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// 测试辅助函数，用于创建临时 git 仓库
func setupTestRepo(t *testing.T) (string, func()) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// 初始化 git 仓库
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "test"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = tmpDir
		if err := c.Run(); err != nil {
			cleanup()
			t.Fatalf("执行命令 %v 失败: %v", cmd, err)
		}
	}

	return tmpDir, cleanup
}

// 测试辅助函数，用于创建测试提交
func createTestCommit(t *testing.T, repoPath, fileName, content string) string {
	filePath := filepath.Join(repoPath, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	cmds := [][]string{
		{"git", "add", fileName},
		{"git", "commit", "-m", "test commit: " + fileName},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = repoPath
		if err := c.Run(); err != nil {
			t.Fatalf("执行命令 %v 失败: %v", cmd, err)
		}
	}

	// 获取提交的 hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	hash, err := cmd.Output()
	if err != nil {
		t.Fatalf("获取提交 hash 失败: %v", err)
	}

	return string(hash[:40])
}

func TestRebaseBranch(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// 切换到测试仓库目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("切换到测试仓库目录失败: %v", err)
	}

	// 创建初始提交在主分支上
	createTestCommit(t, repoPath, "initial.txt", "initial content")

	// 确保主分支存在并指向初始提交
	ExecGitCommand("git", "branch", "-M", "master")

	// 创建并切换到新分支
	ExecGitCommand("git", "checkout", "-b", "feature")

	// 在主分支上创建一个提交
	ExecGitCommand("git", "checkout", "master")
	mainCommit := createTestCommit(t, repoPath, "main.txt", "main content")

	// 切回特性分支创建提交
	ExecGitCommand("git", "checkout", "feature")
	createTestCommit(t, repoPath, "feature.txt", "feature content")

	// 执行 rebase
	RebaseBranch("master", mainCommit)

	// 验证 rebase 结果
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}

	if !strings.Contains(string(output), "feature.txt") {
		t.Error("Feature commit should be present after rebase")
	}
}

func TestExecGitCommand(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// 切换到测试仓库目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("切换到测试仓库目录失败: %v", err)
	}

	// 测试成功的情况
	if err := ExecGitCommand("git", "status"); err != nil {
		t.Errorf("ExecGitCommand 执行 git status 失败: %v", err)
	}

	// 测试失败的情况
	if err := ExecGitCommand("git", "invalid-command"); err == nil {
		t.Error("ExecGitCommand 应该返回错误，但没有")
	}
}
