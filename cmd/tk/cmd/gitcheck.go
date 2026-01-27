package cmd

import (
	"os/exec"
)

// IsTickDirGitignored checks if .tick/ is covered by .gitignore.
// Uses git check-ignore to handle complex gitignore patterns.
// Returns true if the .tick directory would be ignored by git.
func IsTickDirGitignored(repoRoot string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", ".tick/")
	cmd.Dir = repoRoot
	err := cmd.Run()
	// Exit 0 means the path is ignored
	return err == nil
}
