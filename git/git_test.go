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

	// 保存原始目录并切换到测试仓库目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("切换工作目录失败: %v", err)
	}

	// 创建初始提交
	baseCommit := createTestCommit(t, repoPath, "file1.txt", "initial")

	// 创建 feature 分支，但保持在 master 分支上
	cmd := exec.Command("git", "branch", "feature", baseCommit)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("创建分支失败: %v", err)
	}

	// 在 master 分支上创建新提交
	mainCommit := createTestCommit(t, repoPath, "file3.txt", "main")

	// 切换到 feature 分支并创建提交
	cmd = exec.Command("git", "checkout", "feature")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("切换到 feature 分支失败: %v", err)
	}

	// 在 feature 分支上创建提交
	createTestCommit(t, repoPath, "file2.txt", "feature")

	// 在执行 rebase 之前，确保我们在正确的分支上
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("获取当前分支失败: %v", err)
	}
	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != "feature" {
		t.Fatalf("当前不在 feature 分支上，是在 %s 分支上", currentBranch)
	}

	// 测试 RebaseBranch
	if err := RebaseBranch("master", mainCommit); err != nil {
		// 如果失败，输出更多调试信息
		cmd = exec.Command("git", "status")
		cmd.Dir = repoPath
		status, _ := cmd.Output()
		t.Errorf("RebaseBranch 失败: %v\nGit Status:\n%s", err, status)
		return
	}

	// 验证结果：检查文件是否存在且内容正确
	files := map[string]string{
		"file1.txt": "initial",
		"file2.txt": "feature",
		"file3.txt": "main",
	}

	for file, expectedContent := range files {
		content, err := os.ReadFile(filepath.Join(repoPath, file))
		if err != nil {
			t.Errorf("无法读取文件 %s: %v", file, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("文件 %s 的内容不正确，期望 %q，实际 %q", file, expectedContent, string(content))
		}
	}

	// 验证提交历史
	cmd = exec.Command("git", "log", "--pretty=format:%s", "--reverse")
	cmd.Dir = repoPath
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("获取提交历史失败: %v", err)
	}

	commits := strings.Split(string(output), "\n")
	expectedCommits := []string{
		"test commit: file1.txt",
		"test commit: file3.txt",
		"test commit: file2.txt",
	}

	if len(commits) != len(expectedCommits) {
		t.Errorf("提交数量不匹配，期望 %d，实际 %d", len(expectedCommits), len(commits))
		t.Errorf("实际提交: %v", commits)
		return
	}

	for i, expectedCommit := range expectedCommits {
		if commits[i] != expectedCommit {
			t.Errorf("提交顺序不匹配\n期望: %v\n实际: %v", expectedCommits, commits)
			return
		}
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
