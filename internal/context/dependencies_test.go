package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/agent"
	"github.com/pengelbrecht/ticks/internal/tick"
	"github.com/pengelbrecht/ticks/internal/ticks"
)

func TestNewDependencyAnalyzer(t *testing.T) {
	mock := &mockAgent{name: "test"}
	store := tick.NewStore(t.TempDir())

	da := NewDependencyAnalyzer(mock, store)

	if da.agent != mock {
		t.Error("agent not set correctly")
	}

	if da.store != store {
		t.Error("store not set correctly")
	}

	if da.timeout != 3*time.Minute {
		t.Errorf("timeout = %v, want 3m", da.timeout)
	}
}

func TestNewDependencyAnalyzer_WithOptions(t *testing.T) {
	mock := &mockAgent{name: "test"}
	store := tick.NewStore(t.TempDir())
	customTimeout := 5 * time.Minute

	da := NewDependencyAnalyzer(mock, store, WithDepTimeout(customTimeout))

	if da.timeout != customTimeout {
		t.Errorf("timeout = %v, want %v", da.timeout, customTimeout)
	}
}

func TestDependencyAnalyzer_Analyze_SingleTask(t *testing.T) {
	mock := &mockAgent{name: "test"}
	store := tick.NewStore(t.TempDir())
	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{{ID: "t1", Title: "Only task"}}

	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Single task should return empty result (no conflicts possible)
	if len(result.Predictions) != 0 {
		t.Errorf("expected no predictions for single task, got %d", len(result.Predictions))
	}
}

func TestDependencyAnalyzer_Analyze_NoConflicts(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			// Return predictions with no overlapping files
			return &agent.Result{
				Output: `<file_predictions>
[
  {"task_id": "t1", "files": ["src/a.go"]},
  {"task_id": "t2", "files": ["src/b.go"]}
]
</file_predictions>`,
			}, nil
		},
	}

	store := tick.NewStore(t.TempDir())
	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
	}

	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(result.Predictions) != 2 {
		t.Errorf("expected 2 predictions, got %d", len(result.Predictions))
	}

	if len(result.ConflictingPairs) != 0 {
		t.Errorf("expected no conflicts, got %d", len(result.ConflictingPairs))
	}

	if len(result.AddedDeps) != 0 {
		t.Errorf("expected no added deps, got %d", len(result.AddedDeps))
	}
}

func TestDependencyAnalyzer_Analyze_DetectsConflicts(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			// Return predictions with overlapping files
			return &agent.Result{
				Output: `<file_predictions>
[
  {"task_id": "t1", "files": ["src/shared.go", "src/a.go"]},
  {"task_id": "t2", "files": ["src/shared.go", "src/b.go"]}
]
</file_predictions>`,
			}, nil
		},
	}

	// Create temp store with task files
	tmpDir := t.TempDir()
	store := tick.NewStore(tmpDir)
	if err := store.Ensure(); err != nil {
		t.Fatalf("store.Ensure() error = %v", err)
	}

	// Create the tasks in the store
	now := time.Now()
	task1 := tick.Tick{
		ID:        "t1",
		Title:     "Task 1",
		Status:    tick.StatusOpen,
		Type:      tick.TypeTask,
		Owner:     "test",
		CreatedBy: "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	task2 := tick.Tick{
		ID:        "t2",
		Title:     "Task 2",
		Status:    tick.StatusOpen,
		Type:      tick.TypeTask,
		Owner:     "test",
		CreatedBy: "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.Write(task1); err != nil {
		t.Fatalf("store.Write(task1) error = %v", err)
	}
	if err := store.Write(task2); err != nil {
		t.Fatalf("store.Write(task2) error = %v", err)
	}

	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
	}

	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should detect conflict on shared.go
	if len(result.ConflictingPairs) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(result.ConflictingPairs))
	}

	if len(result.ConflictingPairs) > 0 {
		conflict := result.ConflictingPairs[0]
		if len(conflict.SharedFiles) != 1 || conflict.SharedFiles[0] != "src/shared.go" {
			t.Errorf("expected shared file 'src/shared.go', got %v", conflict.SharedFiles)
		}
	}

	// Should add dependency: t2 blocked by t1 (t1 comes first in list)
	if len(result.AddedDeps) != 1 {
		t.Errorf("expected 1 added dep, got %d", len(result.AddedDeps))
	}

	if blockers, ok := result.AddedDeps["t2"]; ok {
		if len(blockers) != 1 || blockers[0] != "t1" {
			t.Errorf("expected t2 blocked by t1, got %v", blockers)
		}
	} else {
		t.Error("expected t2 to have added dependencies")
	}

	// Verify the task was actually updated in the store
	updatedTask, err := store.Read("t2")
	if err != nil {
		t.Fatalf("store.Read(t2) error = %v", err)
	}
	if len(updatedTask.BlockedBy) != 1 || updatedTask.BlockedBy[0] != "t1" {
		t.Errorf("task t2 should be blocked by t1, got %v", updatedTask.BlockedBy)
	}
}

func TestDependencyAnalyzer_Analyze_SkipsExistingDeps(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			return &agent.Result{
				Output: `<file_predictions>
[
  {"task_id": "t1", "files": ["src/shared.go"]},
  {"task_id": "t2", "files": ["src/shared.go"]}
]
</file_predictions>`,
			}, nil
		},
	}

	tmpDir := t.TempDir()
	store := tick.NewStore(tmpDir)
	if err := store.Ensure(); err != nil {
		t.Fatalf("store.Ensure() error = %v", err)
	}

	now := time.Now()
	// Task 2 already blocked by Task 1
	task1 := tick.Tick{
		ID: "t1", Title: "Task 1", Status: tick.StatusOpen, Type: tick.TypeTask,
		Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now,
	}
	task2 := tick.Tick{
		ID: "t2", Title: "Task 2", Status: tick.StatusOpen, Type: tick.TypeTask,
		Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now,
		BlockedBy: []string{"t1"}, // Already has the dependency
	}
	if err := store.Write(task1); err != nil {
		t.Fatalf("store.Write(task1) error = %v", err)
	}
	if err := store.Write(task2); err != nil {
		t.Fatalf("store.Write(task2) error = %v", err)
	}

	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2", BlockedBy: []string{"t1"}},
	}

	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should still detect the conflict
	if len(result.ConflictingPairs) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(result.ConflictingPairs))
	}

	// But should NOT add new deps since it already exists
	if len(result.AddedDeps) != 0 {
		t.Errorf("expected 0 added deps (already exists), got %d", len(result.AddedDeps))
	}
}

func TestDependencyAnalyzer_Analyze_MultipleConflicts(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			// Three tasks, two conflicts: t1-t2 on shared1.go, t2-t3 on shared2.go
			return &agent.Result{
				Output: `<file_predictions>
[
  {"task_id": "t1", "files": ["shared1.go"]},
  {"task_id": "t2", "files": ["shared1.go", "shared2.go"]},
  {"task_id": "t3", "files": ["shared2.go"]}
]
</file_predictions>`,
			}, nil
		},
	}

	tmpDir := t.TempDir()
	store := tick.NewStore(tmpDir)
	store.Ensure()

	now := time.Now()
	for _, id := range []string{"t1", "t2", "t3"} {
		store.Write(tick.Tick{
			ID: id, Title: "Task " + id, Status: tick.StatusOpen, Type: tick.TypeTask,
			Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now,
		})
	}

	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
		{ID: "t3", Title: "Task 3"},
	}

	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should detect 2 conflicts
	if len(result.ConflictingPairs) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(result.ConflictingPairs))
	}

	// Should add 2 dependencies: t2 blocked by t1, t3 blocked by t2
	if len(result.AddedDeps) != 2 {
		t.Errorf("expected 2 added deps, got %d: %v", len(result.AddedDeps), result.AddedDeps)
	}
}

func TestDependencyAnalyzer_ParsePredictions(t *testing.T) {
	da := &DependencyAnalyzer{}

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name: "valid predictions",
			input: `Some preamble text
<file_predictions>
[
  {"task_id": "t1", "files": ["a.go", "b.go"]},
  {"task_id": "t2", "files": ["c.go"]}
]
</file_predictions>
Some postamble`,
			want:    2,
			wantErr: false,
		},
		{
			name:    "no tags",
			input:   `Just some text without tags`,
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid JSON",
			input: `<file_predictions>
not valid json
</file_predictions>`,
			want:    0,
			wantErr: true,
		},
		{
			name: "empty array",
			input: `<file_predictions>
[]
</file_predictions>`,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predictions, err := da.parsePredictions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePredictions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(predictions) != tt.want {
				t.Errorf("parsePredictions() got %d predictions, want %d", len(predictions), tt.want)
			}
		})
	}
}

func TestDependencyAnalyzer_BuildPredictionPrompt(t *testing.T) {
	da := &DependencyAnalyzer{}

	epic := &ticks.Epic{
		ID:          "epic1",
		Title:       "Test Epic",
		Description: "Epic description",
	}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1", Description: "Do thing 1"},
		{ID: "t2", Title: "Task 2", Description: "Do thing 2"},
	}

	prompt := da.buildPredictionPrompt(epic, tasks)

	// Check prompt contains key elements
	checks := []string{
		"[epic1] Test Epic",
		"Epic description",
		"[t1] Task 1",
		"Do thing 1",
		"[t2] Task 2",
		"<file_predictions>",
		"task_id",
		"files",
	}

	for _, check := range checks {
		if !contains(prompt, check) {
			t.Errorf("prompt should contain %q", check)
		}
	}
}

func TestDependencyAnalyzer_Analyze_AgentError(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			return nil, context.DeadlineExceeded
		},
	}

	store := tick.NewStore(t.TempDir())
	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
	}

	_, err := da.Analyze(context.Background(), epic, tasks)
	if err == nil {
		t.Fatal("Analyze() should return error when agent fails")
	}
}

func TestDependencyAnalyzer_Analyze_InvalidResponse(t *testing.T) {
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			// Return response without valid predictions
			return &agent.Result{
				Output: "I couldn't determine the files.",
			}, nil
		},
	}

	store := tick.NewStore(t.TempDir())
	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{ID: "e1", Title: "Test"}
	tasks := []ticks.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
	}

	// Should not error, just return empty result
	result, err := da.Analyze(context.Background(), epic, tasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v, should handle gracefully", err)
	}

	if len(result.Predictions) != 0 {
		t.Errorf("expected empty predictions on invalid response, got %d", len(result.Predictions))
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration test that uses the real file system
func TestDependencyAnalyzer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test verifies the full flow with a mock agent
	mock := &mockAgent{
		name: "test",
		runFunc: func(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
			// Simulate realistic agent response
			return &agent.Result{
				Output: `Based on the task descriptions, here are my predictions:

<file_predictions>
[
  {"task_id": "task-add-button", "files": ["src/components/Button.tsx", "src/components/Button.test.tsx"]},
  {"task_id": "task-style-button", "files": ["src/components/Button.tsx", "src/styles/button.css"]},
  {"task_id": "task-add-form", "files": ["src/components/Form.tsx", "src/components/Form.test.tsx"]}
]
</file_predictions>

Note: task-add-button and task-style-button both modify Button.tsx, so they should not run in parallel.`,
				TokensIn:  500,
				TokensOut: 200,
				Cost:      0.02,
			}, nil
		},
	}

	// Set up temp directory with tick store
	tmpDir := t.TempDir()
	tickDir := filepath.Join(tmpDir, ".tick")
	os.MkdirAll(filepath.Join(tickDir, "issues"), 0755)

	store := tick.NewStore(tickDir)

	// Create tasks
	now := time.Now()
	tasks := []tick.Tick{
		{ID: "task-add-button", Title: "Add Button component", Status: tick.StatusOpen, Type: tick.TypeTask, Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now},
		{ID: "task-style-button", Title: "Style Button component", Status: tick.StatusOpen, Type: tick.TypeTask, Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now},
		{ID: "task-add-form", Title: "Add Form component", Status: tick.StatusOpen, Type: tick.TypeTask, Owner: "test", CreatedBy: "test", CreatedAt: now, UpdatedAt: now},
	}
	for _, task := range tasks {
		if err := store.Write(task); err != nil {
			t.Fatalf("store.Write() error = %v", err)
		}
	}

	da := NewDependencyAnalyzer(mock, store)

	epic := &ticks.Epic{
		ID:          "epic-ui",
		Title:       "Build UI Components",
		Description: "Create reusable UI components",
	}
	ticksTasks := []ticks.Task{
		{ID: "task-add-button", Title: "Add Button component", Description: "Create a new Button component"},
		{ID: "task-style-button", Title: "Style Button component", Description: "Add CSS styles to Button"},
		{ID: "task-add-form", Title: "Add Form component", Description: "Create a new Form component"},
	}

	result, err := da.Analyze(context.Background(), epic, ticksTasks)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Verify predictions
	if len(result.Predictions) != 3 {
		t.Errorf("expected 3 predictions, got %d", len(result.Predictions))
	}

	// Verify conflict detected between add-button and style-button
	if len(result.ConflictingPairs) != 1 {
		t.Errorf("expected 1 conflict (Button.tsx), got %d", len(result.ConflictingPairs))
	}

	// Verify dependency was added
	if len(result.AddedDeps) != 1 {
		t.Errorf("expected 1 added dependency, got %d", len(result.AddedDeps))
	}

	// Verify task was updated in store
	styleTask, err := store.Read("task-style-button")
	if err != nil {
		t.Fatalf("store.Read() error = %v", err)
	}
	if len(styleTask.BlockedBy) != 1 || styleTask.BlockedBy[0] != "task-add-button" {
		t.Errorf("task-style-button should be blocked by task-add-button, got %v", styleTask.BlockedBy)
	}

	// Verify form task was NOT modified (no conflicts)
	formTask, err := store.Read("task-add-form")
	if err != nil {
		t.Fatalf("store.Read() error = %v", err)
	}
	if len(formTask.BlockedBy) != 0 {
		t.Errorf("task-add-form should have no blockers, got %v", formTask.BlockedBy)
	}
}
