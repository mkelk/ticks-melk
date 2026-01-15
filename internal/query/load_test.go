package query

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mkelk/ticks-melk/internal/tick"
)

func TestLoadTicksParallel(t *testing.T) {
	dir := t.TempDir()
	issuesDir := filepath.Join(dir, "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "a1b", Title: "A", Status: tick.StatusOpen, Priority: 2, Type: tick.TypeTask, Owner: "alice", CreatedBy: "alice", CreatedAt: now, UpdatedAt: now},
		{ID: "b2c", Title: "B", Status: tick.StatusOpen, Priority: 2, Type: tick.TypeTask, Owner: "bob", CreatedBy: "bob", CreatedAt: now, UpdatedAt: now},
	}

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		path := filepath.Join(issuesDir, item.ID+".json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	loaded, err := LoadTicksParallel(issuesDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 ticks, got %d", len(loaded))
	}
}
