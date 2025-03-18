package git

import (
	"fmt"
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
		appendCommit(commitHash, modifiedFiles)
		return
	}

	appendCommit(commitHash, strings.Join(strings.Split(modifiedFiles, ","), " "))
}

func appendCommit(commitHash string, files string) {
	allFiles := files == "."
	tempBranch := fmt.Sprintf("temp-edit-%s", commitHash[:8])
	tempStash := fmt.Sprintf("stash-edit-%s", commitHash[:8])
	CreateStash(tempStash)

	needCherryPicks, err := GetCommitsBetween(commitHash, "HEAD")
	if err != nil {
		fmt.Printf(lang.T("Failed to get commit list: %s")+"\n", err)
		return
	}
	fmt.Printf(lang.T("Commit list: %s")+"\n", needCherryPicks)

	if err := ExecGitCommand("git", "checkout", commitHash); err != nil {
		return
	}

	if err := ExecGitCommand("git", "checkout", "-b", tempBranch); err != nil {
		return
	}

	defer cleanup(tempBranch)

	ApplyStashByMessage(tempStash, true)

	if err := ExecGitCommand("git", "add", files); err != nil {
		return
	}

	if err := ExecGitCommand("git", "commit", "--amend", "--no-edit"); err != nil {
		return
	}

	if !allFiles {
		CreateStash(tempStash)
		defer ApplyStashByMessage(tempStash, true)
	}

	if !ApplyCherryPicks(needCherryPicks) {
		return
	}

	finalizeBranch(tempBranch)
}

func cleanup(tempBranch string) {
	ExecGitCommand("git", "branch", "-D", tempBranch)
}

func finalizeBranch(tempBranch string) {
	if err := ExecGitCommand("git", "checkout", CurrentBranch); err != nil {
		return
	}

	if err := ExecGitCommand("git", "reset", "--hard", tempBranch); err != nil {
		return
	}
}
