package github

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureGitAttributes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitattributes")

	if err := os.WriteFile(path, []byte("*.go text\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := EnsureGitAttributes(dir); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	contents := string(data)
	if !containsLine(contents, mergeAttributeLine) {
		t.Fatalf("expected merge driver line in .gitattributes")
	}

	if err := EnsureGitAttributes(dir); err != nil {
		t.Fatalf("ensure second time: %v", err)
	}
}

func containsLine(contents, line string) bool {
	for _, candidate := range splitLines(contents) {
		if candidate == line {
			return true
		}
	}
	return false
}

func splitLines(contents string) []string {
	var lines []string
	start := 0
	for i, r := range contents {
		if r == '\n' {
			lines = append(lines, contents[start:i])
			start = i + 1
		}
	}
	if start < len(contents) {
		lines = append(lines, contents[start:])
	}
	return lines
}
