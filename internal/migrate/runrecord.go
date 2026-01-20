// Package migrate provides one-time data migrations for ticks.
package migrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pengelbrecht/ticks/internal/agent"
	"github.com/pengelbrecht/ticks/internal/runrecord"
)

// RunRecordMigration migrates run records from tick JSON files to the
// separate .tick/logs/records/ directory.
type RunRecordMigration struct {
	tickDir string
	store   *runrecord.Store
	dryRun  bool
}

// MigrationResult contains the results of running a migration.
type MigrationResult struct {
	Migrated int      // Number of records migrated
	Skipped  int      // Number of ticks without run records
	Errors   []string // Errors encountered (tick ID: error message)
}

// NewRunRecordMigration creates a new migration for the given .tick directory.
func NewRunRecordMigration(tickDir string) *RunRecordMigration {
	// The runrecord store expects the project root, not the .tick dir
	projectRoot := filepath.Dir(tickDir)
	return &RunRecordMigration{
		tickDir: tickDir,
		store:   runrecord.NewStore(projectRoot),
		dryRun:  false,
	}
}

// SetDryRun enables dry-run mode (no files are modified).
func (m *RunRecordMigration) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
}

// Run executes the migration, returning the results.
func (m *RunRecordMigration) Run() (*MigrationResult, error) {
	result := &MigrationResult{}

	issuesDir := filepath.Join(m.tickDir, "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil // No issues dir = nothing to migrate
		}
		return nil, fmt.Errorf("read issues dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		tickID := entry.Name()[:len(entry.Name())-5] // strip .json
		tickPath := filepath.Join(issuesDir, entry.Name())

		migrated, err := m.migrateTick(tickID, tickPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", tickID, err))
			continue
		}

		if migrated {
			result.Migrated++
		} else {
			result.Skipped++
		}
	}

	return result, nil
}

// migrateTick migrates a single tick's run record if present.
// Returns true if a run record was migrated, false if the tick had no run record.
func (m *RunRecordMigration) migrateTick(tickID, tickPath string) (bool, error) {
	// Read the raw JSON
	data, err := os.ReadFile(tickPath)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}

	// Parse as a generic map to access the 'run' field
	var tickMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &tickMap); err != nil {
		return false, fmt.Errorf("parse JSON: %w", err)
	}

	// Check if there's a 'run' field
	runJSON, hasRun := tickMap["run"]
	if !hasRun {
		return false, nil // No run record to migrate
	}

	// Parse the run record
	var runRecord agent.RunRecord
	if err := json.Unmarshal(runJSON, &runRecord); err != nil {
		return false, fmt.Errorf("parse run record: %w", err)
	}

	// Validate the run record has required fields
	if runRecord.SessionID == "" {
		return false, errors.New("run record has no session_id")
	}

	if m.dryRun {
		return true, nil
	}

	// Write to the new location
	if err := m.store.Write(tickID, &runRecord); err != nil {
		return false, fmt.Errorf("write to store: %w", err)
	}

	// Remove the 'run' field from the tick JSON
	delete(tickMap, "run")

	// Re-serialize the tick without the run field
	newData, err := json.MarshalIndent(tickMap, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal updated tick: %w", err)
	}

	// Write back atomically
	tmpPath := tickPath + ".tmp"
	if err := os.WriteFile(tmpPath, newData, 0644); err != nil {
		return false, fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, tickPath); err != nil {
		os.Remove(tmpPath) // Clean up on failure
		return false, fmt.Errorf("rename temp file: %w", err)
	}

	return true, nil
}

// NeedsMigration checks if any tick files have embedded run records.
func NeedsMigration(tickDir string) (bool, error) {
	issuesDir := filepath.Join(tickDir, "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		tickPath := filepath.Join(issuesDir, entry.Name())
		data, err := os.ReadFile(tickPath)
		if err != nil {
			continue
		}

		var tickMap map[string]json.RawMessage
		if err := json.Unmarshal(data, &tickMap); err != nil {
			continue
		}

		if _, hasRun := tickMap["run"]; hasRun {
			return true, nil
		}
	}

	return false, nil
}
