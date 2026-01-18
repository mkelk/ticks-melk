// Package runrecord provides storage for completed agent run records.
// Run records are stored as JSON files in .tick/logs/records/<tick-id>.json
//
// This is distinct from the internal/runlog package which writes JSONL
// event streams to .tick/logs/runs/ for debugging and replay purposes.
package runrecord

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pengelbrecht/ticks/internal/agent"
)

// Store manages run record files in the .tick/runrecords/ directory.
type Store struct {
	dir string
}

// ErrNotFound is returned when a run record doesn't exist.
var ErrNotFound = errors.New("run record not found")

// NewStore creates a store for the given tick root directory.
// The tick root should contain a .tick/ directory.
func NewStore(tickRoot string) *Store {
	return &Store{
		dir: filepath.Join(tickRoot, ".tick", "logs", "records"),
	}
}

// Write saves a run record for the given tick ID.
// Overwrites any existing record for that tick.
func (s *Store) Write(tickID string, record *agent.RunRecord) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("create runrecords dir: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run record: %w", err)
	}

	path := s.path(tickID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write run record: %w", err)
	}

	return nil
}

// Read loads a run record for the given tick ID.
// Returns ErrNotFound if no record exists.
func (s *Store) Read(tickID string) (*agent.RunRecord, error) {
	path := s.path(tickID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read run record: %w", err)
	}

	var record agent.RunRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("unmarshal run record: %w", err)
	}

	return &record, nil
}

// Exists checks if a run record exists for the given tick ID.
func (s *Store) Exists(tickID string) bool {
	_, err := os.Stat(s.path(tickID))
	return err == nil
}

// Delete removes a run record for the given tick ID.
// Does not return an error if the record doesn't exist.
func (s *Store) Delete(tickID string) error {
	path := s.path(tickID)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete run record: %w", err)
	}
	return nil
}

// List returns all tick IDs that have run records.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read runrecords dir: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only include .json files, skip .live.json (future: in-progress runs)
		if filepath.Ext(name) == ".json" && !isLiveFile(name) {
			id := name[:len(name)-5] // strip .json
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// path returns the file path for a tick's run record.
func (s *Store) path(tickID string) string {
	return filepath.Join(s.dir, tickID+".json")
}

// isLiveFile checks if a filename is a live record (ends with .live.json).
func isLiveFile(name string) bool {
	return len(name) > 10 && name[len(name)-10:] == ".live.json"
}
