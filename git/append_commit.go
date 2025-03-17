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

	if modifiedFiles == "." {
		appendCommitAll(commitHash)
		return
	}

	appendCommitFiles(commitHash, strings.Split(modifiedFiles, ","))
}

func appendCommitAll(commitHash string) {
	needCherryPicks, err := GetCommitsBetween(commitHash, "HEAD")
	if err != nil {
		fmt.Printf(lang.T("Failed to get commit list: %s")+"\n", err)
		return
	}

	hasStash := stashIfNeeded()
	tempBranch := fmt.Sprintf("temp-edit-%s", commitHash[:8])

	if err := ExecGitCommand("git", "checkout", "-b", tempBranch, commitHash); err != nil {
		restoreStash(hasStash)
		return
	}

	defer cleanup(tempBranch, hasStash)

	if err := ExecGitCommand("git", "checkout", CurrentBranch, "--", "."); err != nil {
		return
	}

	if err := ExecGitCommand("git", "add", "."); err != nil {
		return
	}

	if err := ExecGitCommand("git", "commit", "--amend", "--no-edit"); err != nil {
		return
	}

	if !applyCherryPicks(needCherryPicks) {
		return
	}

	finalizeBranch(tempBranch)
}

func appendCommitFiles(commitHash string, files []string) {
	needCherryPicks, err := GetCommitsBetween(commitHash, "HEAD")
	if err != nil {
		fmt.Printf(lang.T("Failed to get commit list: %s")+"\n", err)
		return
	}

	hasStash := stashIfNeeded()
	tempBranch := fmt.Sprintf("temp-edit-%s", commitHash[:8])

	if err := ExecGitCommand("git", "checkout", "-b", tempBranch, commitHash); err != nil {
		restoreStash(hasStash)
		return
	}

	defer cleanup(tempBranch, hasStash)

	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		if err := ExecGitCommand("git", "checkout", CurrentBranch, "--", file); err != nil {
			return
		}
		if err := ExecGitCommand("git", "add", file); err != nil {
			return
		}
	}

	if err := ExecGitCommand("git", "commit", "--amend", "--no-edit"); err != nil {
		return
	}

	if !applyCherryPicks(needCherryPicks) {
		return
	}

	finalizeBranch(tempBranch)
}

func stashIfNeeded() bool {
	if output, err := exec.Command("git", "status", "--porcelain").Output(); err == nil && len(output) > 0 {
		if err := ExecGitCommand("git", "stash", "push", "-u"); err == nil {
			return true
		}
	}
	return false
}

func cleanup(tempBranch string, hasStash bool) {
	ExecGitCommand("git", "checkout", CurrentBranch)
	ExecGitCommand("git", "branch", "-D", tempBranch)
	restoreStash(hasStash)
}

func restoreStash(hasStash bool) {
	if hasStash {
		ExecGitCommand("git", "stash", "pop")
	}
}

func applyCherryPicks(commits []string) bool {
	for _, commit := range commits {
		if commit == "" {
			continue
		}
		if err := ExecGitCommand("git", "cherry-pick", commit); err != nil {
			fmt.Printf(lang.T("Failed to cherry-pick commit %s, please fix it manually")+"\n", commit)
			ExecGitCommand("git", "cherry-pick", "--abort")
			return false
		}
	}
	return true
}

func finalizeBranch(tempBranch string) {
	if err := ExecGitCommand("git", "checkout", CurrentBranch); err != nil {
		return
	}

	if err := ExecGitCommand("git", "reset", "--hard", tempBranch); err != nil {
		return
	}
}
