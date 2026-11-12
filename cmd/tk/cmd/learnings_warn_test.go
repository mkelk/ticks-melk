package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pengelbrecht/ticks/internal/tick"
)

// captureStderrArgs runs ExecuteArgs while capturing stderr, returning the
// captured output and the command error.
func captureStderrArgs(t *testing.T, args []string) (string, error) {
	t.Helper()

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	runErr := ExecuteArgs(args)

	_ = w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	return buf.String(), runErr
}

// writeOverBudgetLearnings writes a head over 4096 bytes and a body over 16384
// bytes (of small entries, so only the body's budget warns) into repoDir/.tick.
func writeOverBudgetLearnings(t *testing.T, repoDir string) {
	t.Helper()
	tickDir := filepath.Join(repoDir, ".tick")
	if err := os.MkdirAll(tickDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	head := strings.Repeat("h", tick.LearningsHeadBudgetBytes+1)
	body := "## H\n\n" + strings.Repeat("small\n\n", tick.LearningsBodyBudgetBytes/7+1)
	for name, content := range map[string]string{tick.LearningsHeadFile: head, tick.LearningsBodyFile: body} {
		if err := os.WriteFile(filepath.Join(tickDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

// assertPerFileLearningsWarnings checks stderr holds exactly one warning line
// naming the head and exactly one naming the body.
func assertPerFileLearningsWarnings(t *testing.T, stderr string) {
	t.Helper()
	var head, body int
	for _, line := range strings.Split(stderr, "\n") {
		switch {
		case strings.HasPrefix(line, "warning: .tick/learnings.md "):
			head++
		case strings.HasPrefix(line, "warning: .tick/learnings-ref.md "):
			body++
		}
	}
	if head != 1 || body != 1 {
		t.Errorf("stderr has %d head and %d body warning lines, want 1 and 1:\n%s", head, body, stderr)
	}
}

func TestStatusWarnsLearningsPerFile(t *testing.T) {
	repoDir, _ := setupTestRepo(t)
	writeOverBudgetLearnings(t, repoDir)

	stderr, err := captureStderrArgs(t, []string{"status"})
	if err != nil {
		t.Fatalf("status: %v (warnings must never block)", err)
	}
	assertPerFileLearningsWarnings(t, stderr)
}

func TestCloseEpicWarnsLearningsPerFile(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	epic := makeTestEpic("lw1")
	if err := store.Write(epic); err != nil {
		t.Fatalf("write epic: %v", err)
	}
	writeOverBudgetLearnings(t, repoDir)

	stderr, err := captureStderrArgs(t, []string{"close", epic.ID, "--reason", "done"})
	if err != nil {
		t.Fatalf("close: %v (warnings must never block)", err)
	}
	assertPerFileLearningsWarnings(t, stderr)

	closed, err := store.Read(epic.ID)
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != tick.StatusClosed {
		t.Errorf("epic status = %q, want closed", closed.Status)
	}
}
