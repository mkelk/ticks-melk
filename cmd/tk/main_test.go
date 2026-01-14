package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIWorkflow(t *testing.T) {
	repo := t.TempDir()
	if err := runGit(repo, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := runGit(repo, "remote", "add", "origin", "https://github.com/petere/chefswiz.git"); err != nil {
		t.Fatalf("git remote add: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := os.Setenv("TICK_OWNER", "tester"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("TICK_OWNER") })

	if code := run([]string{"tk", "init"}); code != exitSuccess {
		t.Fatalf("expected init exit %d, got %d", exitSuccess, code)
	}

	out, code := captureStdout(func() int {
		return run([]string{"tk", "create", "Test", "tick", "-t", "bug", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var created map[string]any
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected id in create output")
	}
	if created["type"] != "bug" {
		t.Fatalf("expected type bug, got %v", created["type"])
	}

	showOut, code := captureStdout(func() int {
		return run([]string{"tk", "show", "--json", id})
	})
	if code != exitSuccess {
		t.Fatalf("expected show exit %d, got %d", exitSuccess, code)
	}
	var shown map[string]any
	if err := json.Unmarshal([]byte(showOut), &shown); err != nil {
		t.Fatalf("parse show json: %v", err)
	}
	if shown["id"] != id {
		t.Fatalf("expected show id %s, got %v", id, shown["id"])
	}

	listOut, code := captureStdout(func() int {
		return run([]string{"tk", "list", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected list exit %d, got %d", exitSuccess, code)
	}
	if !bytes.Contains([]byte(listOut), []byte(id)) {
		t.Fatalf("expected list to include id %s", id)
	}

	if _, err := os.Stat(filepath.Join(repo, ".tick", "issues", id+".json")); err != nil {
		t.Fatalf("expected tick file: %v", err)
	}
}

func captureStdout(fn func() int) (string, int) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := fn()
	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	return buf.String(), code
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func TestCreateProjectFlag(t *testing.T) {
	repo := t.TempDir()
	if err := runGit(repo, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := runGit(repo, "remote", "add", "origin", "https://github.com/petere/chefswiz.git"); err != nil {
		t.Fatalf("git remote add: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := os.Setenv("TICK_OWNER", "tester"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("TICK_OWNER") })

	if code := run([]string{"tk", "init"}); code != exitSuccess {
		t.Fatalf("expected init exit %d, got %d", exitSuccess, code)
	}

	// Test 1: Create with explicit project
	out, code := captureStdout(func() int {
		return run([]string{"tk", "create", "Task with project", "-project", "test-proj-123", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var result1 map[string]any
	if err := json.Unmarshal([]byte(out), &result1); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	if result1["project"] != "test-proj-123" {
		t.Fatalf("expected project test-proj-123, got %v", result1["project"])
	}

	// Test 2: Create without project
	out, code = captureStdout(func() int {
		return run([]string{"tk", "create", "Task without project", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var result2 map[string]any
	if err := json.Unmarshal([]byte(out), &result2); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	if result2["project"] != nil && result2["project"] != "" {
		t.Fatalf("expected empty project, got %v", result2["project"])
	}

	// Test 3: Create parent epic with project, then child inherits it
	out, code = captureStdout(func() int {
		return run([]string{"tk", "create", "Parent epic", "-t", "epic", "-project", "epic-proj", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var result3 map[string]any
	if err := json.Unmarshal([]byte(out), &result3); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	epicID := result3["id"].(string)
	if result3["project"] != "epic-proj" {
		t.Fatalf("expected project epic-proj, got %v", result3["project"])
	}

	// Child without explicit project should inherit from parent
	out, code = captureStdout(func() int {
		return run([]string{"tk", "create", "Child task", "-parent", epicID, "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var result4 map[string]any
	if err := json.Unmarshal([]byte(out), &result4); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	if result4["project"] != "epic-proj" {
		t.Fatalf("expected inherited project epic-proj, got %v", result4["project"])
	}

	// Test 4: Child that overrides parent project
	out, code = captureStdout(func() int {
		return run([]string{"tk", "create", "Child with override", "-parent", epicID, "-project", "different-proj", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var result5 map[string]any
	if err := json.Unmarshal([]byte(out), &result5); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	if result5["project"] != "different-proj" {
		t.Fatalf("expected project different-proj, got %v", result5["project"])
	}
}

func TestUpdateProjectFlag(t *testing.T) {
	repo := t.TempDir()
	if err := runGit(repo, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := runGit(repo, "remote", "add", "origin", "https://github.com/petere/chefswiz.git"); err != nil {
		t.Fatalf("git remote add: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := os.Setenv("TICK_OWNER", "tester"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("TICK_OWNER") })

	if code := run([]string{"tk", "init"}); code != exitSuccess {
		t.Fatalf("expected init exit %d, got %d", exitSuccess, code)
	}

	// Create a tick without a project
	out, code := captureStdout(func() int {
		return run([]string{"tk", "create", "Task without project", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected create exit %d, got %d", exitSuccess, code)
	}
	var created map[string]any
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("parse create json: %v", err)
	}
	id := created["id"].(string)

	// Test 1: Set project on tick without project
	out, code = captureStdout(func() int {
		return run([]string{"tk", "update", id, "--project", "new-proj", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected update exit %d, got %d", exitSuccess, code)
	}
	var result1 map[string]any
	if err := json.Unmarshal([]byte(out), &result1); err != nil {
		t.Fatalf("parse update json: %v", err)
	}
	if result1["project"] != "new-proj" {
		t.Fatalf("expected project new-proj, got %v", result1["project"])
	}

	// Test 2: Change project on tick with project
	out, code = captureStdout(func() int {
		return run([]string{"tk", "update", id, "--project", "changed-proj", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected update exit %d, got %d", exitSuccess, code)
	}
	var result2 map[string]any
	if err := json.Unmarshal([]byte(out), &result2); err != nil {
		t.Fatalf("parse update json: %v", err)
	}
	if result2["project"] != "changed-proj" {
		t.Fatalf("expected project changed-proj, got %v", result2["project"])
	}

	// Test 3: Clear project with empty string
	out, code = captureStdout(func() int {
		return run([]string{"tk", "update", id, "--project", "", "--json"})
	})
	if code != exitSuccess {
		t.Fatalf("expected update exit %d, got %d", exitSuccess, code)
	}
	var result3 map[string]any
	if err := json.Unmarshal([]byte(out), &result3); err != nil {
		t.Fatalf("parse update json: %v", err)
	}
	if result3["project"] != nil && result3["project"] != "" {
		t.Fatalf("expected empty project, got %v", result3["project"])
	}
}
