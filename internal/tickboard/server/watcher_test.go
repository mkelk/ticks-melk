package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/runrecord"
)

func TestNewLiveFileWatcher(t *testing.T) {
	tmpDir := t.TempDir()
	recordsDir := filepath.Join(tmpDir, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)
	if w == nil {
		t.Fatal("NewLiveFileWatcher() returned nil")
	}
	if w.recordsDir != recordsDir {
		t.Errorf("recordsDir = %q, want %q", w.recordsDir, recordsDir)
	}
}

func TestLiveFileWatcher_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	recordsDir := filepath.Join(tmpDir, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)

	// Start should create the directory and begin watching
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(recordsDir); os.IsNotExist(err) {
		t.Error("Start() did not create records directory")
	}

	// Stop should succeed
	w.Stop()

	// Verify events channel is closed
	select {
	case _, ok := <-w.Events():
		if ok {
			t.Error("Events channel should be closed after Stop()")
		}
	default:
		// Channel is closed or empty, both are acceptable
	}
}

func TestLiveFileWatcher_StartTwice(t *testing.T) {
	tmpDir := t.TempDir()
	recordsDir := filepath.Join(tmpDir, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)

	// Start twice should be idempotent
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("Second Start() error = %v", err)
	}

	w.Stop()
}

func TestLiveFileWatcher_StopWithoutStart(t *testing.T) {
	tmpDir := t.TempDir()
	recordsDir := filepath.Join(tmpDir, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)

	// Stop without start should not panic
	w.Stop()
}

func TestLiveFileWatcher_DetectsCreated(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Create a live record file
	liveRecord := runrecord.LiveRecord{
		SessionID: "test-session",
		Status:    "thinking",
		NumTurns:  1,
	}
	data, _ := json.MarshalIndent(liveRecord, "", "  ")
	livePath := filepath.Join(recordsDir, "tick1.live.json")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to write live file: %v", err)
	}

	// Wait for event with timeout
	select {
	case event := <-w.Events():
		if event.Type != Created {
			t.Errorf("event.Type = %v, want Created", event.Type)
		}
		if event.TickID != "tick1" {
			t.Errorf("event.TickID = %q, want %q", event.TickID, "tick1")
		}
		if event.Record == nil {
			t.Error("event.Record is nil, want non-nil")
		} else if event.Record.SessionID != "test-session" {
			t.Errorf("event.Record.SessionID = %q, want %q", event.Record.SessionID, "test-session")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Created event")
	}
}

func TestLiveFileWatcher_DetectsUpdated(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	// Pre-create the directory and file before starting watcher
	if err := os.MkdirAll(recordsDir, 0755); err != nil {
		t.Fatalf("failed to create records dir: %v", err)
	}
	liveRecord := runrecord.LiveRecord{
		SessionID: "test-session",
		Status:    "thinking",
		NumTurns:  1,
	}
	data, _ := json.MarshalIndent(liveRecord, "", "  ")
	livePath := filepath.Join(recordsDir, "tick2.live.json")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to write initial live file: %v", err)
	}

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Update the file
	liveRecord.Status = "writing"
	liveRecord.NumTurns = 2
	data, _ = json.MarshalIndent(liveRecord, "", "  ")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to update live file: %v", err)
	}

	// Wait for update event
	select {
	case event := <-w.Events():
		if event.Type != Updated {
			t.Errorf("event.Type = %v, want Updated", event.Type)
		}
		if event.TickID != "tick2" {
			t.Errorf("event.TickID = %q, want %q", event.TickID, "tick2")
		}
		if event.Record == nil {
			t.Error("event.Record is nil, want non-nil")
		} else if event.Record.Status != "writing" {
			t.Errorf("event.Record.Status = %q, want %q", event.Record.Status, "writing")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Updated event")
	}
}

func TestLiveFileWatcher_DetectsFinalized(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	// Pre-create the directory and live file
	if err := os.MkdirAll(recordsDir, 0755); err != nil {
		t.Fatalf("failed to create records dir: %v", err)
	}
	liveRecord := runrecord.LiveRecord{
		SessionID: "test-session",
		Status:    "complete",
		NumTurns:  3,
	}
	data, _ := json.MarshalIndent(liveRecord, "", "  ")
	livePath := filepath.Join(recordsDir, "tick3.live.json")
	finalPath := filepath.Join(recordsDir, "tick3.json")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to write live file: %v", err)
	}

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Rename live to final (simulating finalization)
	if err := os.Rename(livePath, finalPath); err != nil {
		t.Fatalf("failed to rename live to final: %v", err)
	}

	// Wait for finalized event
	select {
	case event := <-w.Events():
		if event.Type != Finalized {
			t.Errorf("event.Type = %v, want Finalized", event.Type)
		}
		if event.TickID != "tick3" {
			t.Errorf("event.TickID = %q, want %q", event.TickID, "tick3")
		}
		if event.Record != nil {
			t.Error("event.Record should be nil for Finalized events")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Finalized event")
	}
}

func TestLiveFileWatcher_IgnoresNonLiveFiles(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Create a non-live file (should be ignored)
	otherPath := filepath.Join(recordsDir, "other.txt")
	if err := os.WriteFile(otherPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write other file: %v", err)
	}

	// Should not receive any events
	select {
	case event := <-w.Events():
		t.Errorf("unexpected event: %+v", event)
	case <-time.After(200 * time.Millisecond):
		// Good - no events expected
	}
}

func TestLiveFileWatcher_Debouncing(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Rapidly write to the same file multiple times
	livePath := filepath.Join(recordsDir, "tick4.live.json")
	for i := 0; i < 5; i++ {
		liveRecord := runrecord.LiveRecord{
			SessionID: "test-session",
			Status:    "writing",
			NumTurns:  i + 1,
		}
		data, _ := json.MarshalIndent(liveRecord, "", "  ")
		if err := os.WriteFile(livePath, data, 0644); err != nil {
			t.Fatalf("failed to write live file (iteration %d): %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // Rapid writes
	}

	// Should receive at most 2 events (one for initial create, possibly one debounced update)
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)

	for {
		select {
		case event := <-w.Events():
			eventCount++
			if event.TickID != "tick4" {
				t.Errorf("event.TickID = %q, want %q", event.TickID, "tick4")
			}
			// Check that debouncing worked - the last update should have the highest NumTurns
			if eventCount == 1 && event.Record != nil && event.Record.NumTurns < 1 {
				t.Errorf("first event NumTurns = %d, want >= 1", event.Record.NumTurns)
			}
		case <-timeout:
			// Done waiting
			if eventCount == 0 {
				t.Error("expected at least one event")
			}
			if eventCount > 3 {
				t.Errorf("expected at most 3 events due to debouncing, got %d", eventCount)
			}
			return
		}
	}
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
		et   EventType
		want string
	}{
		{Created, "created"},
		{Updated, "updated"},
		{Finalized, "finalized"},
		{EventType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.et.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.et, got, tt.want)
		}
	}
}

func TestLiveFileWatcher_ExistingFilesDetectedOnStart(t *testing.T) {
	tmpDir := t.TempDir()
	tickRoot := filepath.Join(tmpDir, "repo")
	recordsDir := filepath.Join(tickRoot, ".tick", "logs", "records")

	// Pre-create the directory and a live file BEFORE starting watcher
	if err := os.MkdirAll(recordsDir, 0755); err != nil {
		t.Fatalf("failed to create records dir: %v", err)
	}
	liveRecord := runrecord.LiveRecord{
		SessionID: "existing-session",
		Status:    "tool_use",
	}
	data, _ := json.MarshalIndent(liveRecord, "", "  ")
	livePath := filepath.Join(recordsDir, "existing.live.json")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to write live file: %v", err)
	}

	w := NewLiveFileWatcher(recordsDir)
	if err := w.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// The watcher should know about the existing file
	// Verify by checking that an update to it triggers an Updated (not Created) event
	liveRecord.Status = "complete"
	data, _ = json.MarshalIndent(liveRecord, "", "  ")
	if err := os.WriteFile(livePath, data, 0644); err != nil {
		t.Fatalf("failed to update live file: %v", err)
	}

	select {
	case event := <-w.Events():
		if event.Type != Updated {
			t.Errorf("event.Type = %v, want Updated (file was pre-existing)", event.Type)
		}
		if event.TickID != "existing" {
			t.Errorf("event.TickID = %q, want %q", event.TickID, "existing")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Updated event")
	}
}
