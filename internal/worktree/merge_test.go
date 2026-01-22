package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMergeManager(t *testing.T) {
	t.Run("detects main branch", func(t *testing.T) {
		dir := createTempGitRepo(t)

		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}
		if mm == nil {
			t.Fatal("NewMergeManager() returned nil")
		}
		// Default branch from git init is usually "master" but may vary
		if mm.MainBranch() != "main" && mm.MainBranch() != "master" {
			t.Errorf("MainBranch() = %q, want 'main' or 'master'", mm.MainBranch())
		}
	})

	t.Run("returns error for repo without main/master", func(t *testing.T) {
		dir := t.TempDir()

		// Initialize bare git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to init git repo: %v", err)
		}

		// No commits yet, so no main/master branch
		_, err := NewMergeManager(dir)
		if err == nil {
			t.Error("NewMergeManager() should fail for repo without main/master")
		}
	})
}

func TestMergeManager_Merge(t *testing.T) {
	t.Run("successful fast-forward merge", func(t *testing.T) {
		dir := createTempGitRepo(t)
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		// Create worktree
		wt, err := wm.Create("ff-epic")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Make a commit in the worktree
		newFile := filepath.Join(wt.Path, "new-file.txt")
		if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, wt.Path, "add", "new-file.txt")
		runGit(t, wt.Path, "commit", "-m", "Add new file")

		// Create merge manager and merge
		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if !result.Success {
			t.Errorf("Merge() Success = false, want true. Error: %s", result.ErrorMessage)
		}
		if !result.Merged {
			t.Error("Merge() Merged = false, want true")
		}
		if result.MergeCommit == "" {
			t.Error("Merge() MergeCommit should not be empty")
		}
		if len(result.Conflicts) > 0 {
			t.Errorf("Merge() Conflicts = %v, want empty", result.Conflicts)
		}

		// Verify file exists on main branch
		mainFile := filepath.Join(dir, "new-file.txt")
		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			t.Error("new-file.txt should exist on main after merge")
		}
	})

	t.Run("successful merge with commits on both sides", func(t *testing.T) {
		dir := createTempGitRepo(t)
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		// Create worktree
		wt, err := wm.Create("both-sides")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Make commit in worktree (different file)
		wtFile := filepath.Join(wt.Path, "worktree-file.txt")
		if err := os.WriteFile(wtFile, []byte("worktree content"), 0644); err != nil {
			t.Fatalf("failed to write worktree file: %v", err)
		}
		runGit(t, wt.Path, "add", "worktree-file.txt")
		runGit(t, wt.Path, "commit", "-m", "Add worktree file")

		// Make commit in main (different file)
		mainFile := filepath.Join(dir, "main-file.txt")
		if err := os.WriteFile(mainFile, []byte("main content"), 0644); err != nil {
			t.Fatalf("failed to write main file: %v", err)
		}
		runGit(t, dir, "add", "main-file.txt")
		runGit(t, dir, "commit", "-m", "Add main file")

		// Merge
		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if !result.Success {
			t.Errorf("Merge() Success = false, want true. Error: %s", result.ErrorMessage)
		}
		if !result.Merged {
			t.Error("Merge() Merged = false, want true")
		}

		// Verify both files exist on main
		if _, err := os.Stat(filepath.Join(dir, "worktree-file.txt")); os.IsNotExist(err) {
			t.Error("worktree-file.txt should exist on main after merge")
		}
		if _, err := os.Stat(filepath.Join(dir, "main-file.txt")); os.IsNotExist(err) {
			t.Error("main-file.txt should exist on main after merge")
		}
	})

	t.Run("conflict detection", func(t *testing.T) {
		dir := createTempGitRepo(t)
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		// Create worktree
		wt, err := wm.Create("conflict-epic")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Modify same file in worktree
		conflictFile := filepath.Join(wt.Path, "initial.txt")
		if err := os.WriteFile(conflictFile, []byte("worktree version"), 0644); err != nil {
			t.Fatalf("failed to write conflict file in worktree: %v", err)
		}
		runGit(t, wt.Path, "add", "initial.txt")
		runGit(t, wt.Path, "commit", "-m", "Worktree change to initial.txt")

		// Modify same file in main
		mainConflictFile := filepath.Join(dir, "initial.txt")
		if err := os.WriteFile(mainConflictFile, []byte("main version"), 0644); err != nil {
			t.Fatalf("failed to write conflict file in main: %v", err)
		}
		runGit(t, dir, "add", "initial.txt")
		runGit(t, dir, "commit", "-m", "Main change to initial.txt")

		// Attempt merge - should fail with conflict
		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if result.Success {
			t.Error("Merge() Success = true, want false (conflict)")
		}
		if len(result.Conflicts) == 0 {
			t.Error("Merge() Conflicts should not be empty")
		}
		if len(result.Conflicts) > 0 && result.Conflicts[0] != "initial.txt" {
			t.Errorf("Merge() Conflicts = %v, want [initial.txt]", result.Conflicts)
		}

		// Clean up the conflict
		if mm.HasConflict() {
			mm.AbortMerge()
		}
	})
}

func TestMergeManager_AbortMerge(t *testing.T) {
	t.Run("aborts in-progress merge", func(t *testing.T) {
		dir := createTempGitRepo(t)
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		// Create worktree and cause conflict
		wt, err := wm.Create("abort-epic")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Create conflict
		conflictFile := filepath.Join(wt.Path, "initial.txt")
		if err := os.WriteFile(conflictFile, []byte("worktree version"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, wt.Path, "add", "initial.txt")
		runGit(t, wt.Path, "commit", "-m", "Worktree change")

		mainFile := filepath.Join(dir, "initial.txt")
		if err := os.WriteFile(mainFile, []byte("main version"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, dir, "add", "initial.txt")
		runGit(t, dir, "commit", "-m", "Main change")

		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		// Trigger conflict
		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if result.Success {
			t.Skip("Expected conflict but merge succeeded")
		}

		// Verify conflict exists
		if !mm.HasConflict() {
			t.Fatal("HasConflict() = false, want true after failed merge")
		}

		// Abort the merge
		if err := mm.AbortMerge(); err != nil {
			t.Fatalf("AbortMerge() error = %v", err)
		}

		// Verify no conflict after abort
		if mm.HasConflict() {
			t.Error("HasConflict() = true, want false after AbortMerge()")
		}
	})

	t.Run("returns error when no merge in progress", func(t *testing.T) {
		dir := createTempGitRepo(t)

		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		err = mm.AbortMerge()
		if err != ErrNoMergeInProgress {
			t.Errorf("AbortMerge() error = %v, want %v", err, ErrNoMergeInProgress)
		}
	})
}

func TestMergeManager_HasConflict(t *testing.T) {
	t.Run("returns false when no merge in progress", func(t *testing.T) {
		dir := createTempGitRepo(t)

		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		if mm.HasConflict() {
			t.Error("HasConflict() = true, want false (no merge in progress)")
		}
	})

	t.Run("returns true during merge conflict", func(t *testing.T) {
		dir := createTempGitRepo(t)
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		// Create worktree
		wt, err := wm.Create("has-conflict-epic")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Create conflict scenario
		conflictFile := filepath.Join(wt.Path, "initial.txt")
		if err := os.WriteFile(conflictFile, []byte("worktree version"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, wt.Path, "add", "initial.txt")
		runGit(t, wt.Path, "commit", "-m", "Worktree change")

		mainFile := filepath.Join(dir, "initial.txt")
		if err := os.WriteFile(mainFile, []byte("main version"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, dir, "add", "initial.txt")
		runGit(t, dir, "commit", "-m", "Main change")

		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		// Trigger conflict
		result, _ := mm.Merge(wt)
		if result.Success {
			t.Skip("Expected conflict but merge succeeded")
		}

		if !mm.HasConflict() {
			t.Error("HasConflict() = false, want true during conflict")
		}

		// Clean up
		mm.AbortMerge()
	})
}

func TestMergeManager_MergesToParentBranch(t *testing.T) {
	t.Run("merges to parent branch when it exists", func(t *testing.T) {
		dir := createTempGitRepo(t)

		// Create a feature branch from main
		runGit(t, dir, "checkout", "-b", "feature-branch")
		featureFile := filepath.Join(dir, "feature-base.txt")
		if err := os.WriteFile(featureFile, []byte("feature base content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, dir, "add", "feature-base.txt")
		runGit(t, dir, "commit", "-m", "Add feature base")

		// Create worktree manager and create worktree from feature branch
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		wt, err := wm.Create("parent-test")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify parent branch is set
		if wt.ParentBranch != "feature-branch" {
			t.Errorf("ParentBranch = %q, want %q", wt.ParentBranch, "feature-branch")
		}

		// Make a commit in the worktree
		newFile := filepath.Join(wt.Path, "worktree-change.txt")
		if err := os.WriteFile(newFile, []byte("worktree content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, wt.Path, "add", "worktree-change.txt")
		runGit(t, wt.Path, "commit", "-m", "Add worktree change")

		// Create merge manager and merge
		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}

		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if !result.Success {
			t.Errorf("Merge() Success = false, want true. Error: %s", result.ErrorMessage)
		}

		// Verify we're on feature-branch (not main)
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = dir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}
		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != "feature-branch" {
			t.Errorf("Current branch = %q, want %q", currentBranch, "feature-branch")
		}

		// Verify the worktree change is on feature-branch
		if _, err := os.Stat(filepath.Join(dir, "worktree-change.txt")); os.IsNotExist(err) {
			t.Error("worktree-change.txt should exist on feature-branch after merge")
		}

		// Verify main branch does NOT have the worktree change
		runGit(t, dir, "checkout", mm.MainBranch())
		if _, err := os.Stat(filepath.Join(dir, "worktree-change.txt")); !os.IsNotExist(err) {
			t.Error("worktree-change.txt should NOT exist on main branch")
		}
	})
}

func TestMergeManager_FallbackToMain(t *testing.T) {
	t.Run("falls back to main when parent branch is deleted", func(t *testing.T) {
		dir := createTempGitRepo(t)

		// Create a feature branch from main
		runGit(t, dir, "checkout", "-b", "temp-feature")
		tempFile := filepath.Join(dir, "temp-feature.txt")
		if err := os.WriteFile(tempFile, []byte("temp feature content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, dir, "add", "temp-feature.txt")
		runGit(t, dir, "commit", "-m", "Add temp feature")

		// Create worktree manager and create worktree from temp-feature branch
		wm, err := NewManager(dir)
		if err != nil {
			t.Fatalf("NewManager() error = %v", err)
		}

		wt, err := wm.Create("fallback-test")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify parent branch is set
		if wt.ParentBranch != "temp-feature" {
			t.Errorf("ParentBranch = %q, want %q", wt.ParentBranch, "temp-feature")
		}

		// Make a commit in the worktree
		newFile := filepath.Join(wt.Path, "fallback-change.txt")
		if err := os.WriteFile(newFile, []byte("fallback content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		runGit(t, wt.Path, "add", "fallback-change.txt")
		runGit(t, wt.Path, "commit", "-m", "Add fallback change")

		// Switch to main and delete the parent branch
		mm, err := NewMergeManager(dir)
		if err != nil {
			t.Fatalf("NewMergeManager() error = %v", err)
		}
		runGit(t, dir, "checkout", mm.MainBranch())
		runGit(t, dir, "branch", "-D", "temp-feature")

		// Merge should fall back to main
		result, err := mm.Merge(wt)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if !result.Success {
			t.Errorf("Merge() Success = false, want true. Error: %s", result.ErrorMessage)
		}

		// Verify we're on main branch
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = dir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}
		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != mm.MainBranch() {
			t.Errorf("Current branch = %q, want %q", currentBranch, mm.MainBranch())
		}

		// Verify the fallback change is on main
		if _, err := os.Stat(filepath.Join(dir, "fallback-change.txt")); os.IsNotExist(err) {
			t.Error("fallback-change.txt should exist on main after merge")
		}
	})
}

// runGit runs a git command in the specified directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s: %v", args, string(output), err)
	}
}
