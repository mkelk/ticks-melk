package runrecord

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/agent"
)

func TestStore_WriteRead(t *testing.T) {
	dir := t.TempDir()

	store := NewStore(dir)

	record := &agent.RunRecord{
		SessionID: "test-session-123",
		Model:     "claude-sonnet-4-20250514",
		StartedAt: time.Now().Add(-5 * time.Minute),
		EndedAt:   time.Now(),
		Output:    "Task completed successfully",
		Thinking:  "Let me analyze this...",
		Tools: []agent.ToolRecord{
			{Name: "Read", Input: "file.go", Duration: 100, IsError: false},
			{Name: "Edit", Input: "file.go", Duration: 200, IsError: false},
		},
		Metrics: agent.MetricsRecord{
			InputTokens:  1000,
			OutputTokens: 500,
			CostUSD:      0.05,
			DurationMS:   30000,
		},
		Success:  true,
		NumTurns: 3,
	}

	// Write
	err := store.Write("abc", record)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, ".tick", "logs", "records", "abc.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Run record file not created")
	}

	// Read back
	got, err := store.Read("abc")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got.SessionID != record.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", got.SessionID, record.SessionID)
	}
	if got.Model != record.Model {
		t.Errorf("Model mismatch: got %q, want %q", got.Model, record.Model)
	}
	if got.Success != record.Success {
		t.Errorf("Success mismatch: got %v, want %v", got.Success, record.Success)
	}
	if got.NumTurns != record.NumTurns {
		t.Errorf("NumTurns mismatch: got %d, want %d", got.NumTurns, record.NumTurns)
	}
	if len(got.Tools) != len(record.Tools) {
		t.Errorf("Tools length mismatch: got %d, want %d", len(got.Tools), len(record.Tools))
	}
}

func TestStore_ReadNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Read("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if store.Exists("abc") {
		t.Error("Exists returned true for nonexistent record")
	}

	record := &agent.RunRecord{SessionID: "test"}
	if err := store.Write("abc", record); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if !store.Exists("abc") {
		t.Error("Exists returned false for existing record")
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	record := &agent.RunRecord{SessionID: "test"}
	if err := store.Write("abc", record); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := store.Delete("abc"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if store.Exists("abc") {
		t.Error("Record still exists after delete")
	}

	// Delete nonexistent should not error
	if err := store.Delete("nonexistent"); err != nil {
		t.Errorf("Delete nonexistent returned error: %v", err)
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Empty list initially
	ids, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty list, got %d items", len(ids))
	}

	// Write some records
	for _, id := range []string{"abc", "def", "xyz"} {
		record := &agent.RunRecord{SessionID: id}
		if err := store.Write(id, record); err != nil {
			t.Fatalf("Write %s failed: %v", id, err)
		}
	}

	ids, err = store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("Expected 3 items, got %d", len(ids))
	}
}

func TestStore_ListSkipsLiveFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create records directory
	runrecordsDir := filepath.Join(dir, ".tick", "logs", "records")
	if err := os.MkdirAll(runrecordsDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	// Create a regular record
	if err := os.WriteFile(filepath.Join(runrecordsDir, "abc.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create a live record (should be skipped)
	if err := os.WriteFile(filepath.Join(runrecordsDir, "def.live.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	ids, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(ids) != 1 {
		t.Errorf("Expected 1 item (excluding .live.json), got %d", len(ids))
	}
	if len(ids) > 0 && ids[0] != "abc" {
		t.Errorf("Expected 'abc', got %q", ids[0])
	}
}
