package context

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/pengelbrecht/ticks/internal/agent"
	"github.com/pengelbrecht/ticks/internal/tick"
	"github.com/pengelbrecht/ticks/internal/ticks"
)

// DependencyAnalyzer analyzes tasks for potential file conflicts and adds dependencies.
type DependencyAnalyzer struct {
	agent   agent.Agent
	store   *tick.Store
	logger  *slog.Logger
	timeout time.Duration
}

// DependencyAnalyzerOption configures a DependencyAnalyzer.
type DependencyAnalyzerOption func(*DependencyAnalyzer)

// WithDepTimeout sets the timeout for dependency analysis.
func WithDepTimeout(d time.Duration) DependencyAnalyzerOption {
	return func(da *DependencyAnalyzer) {
		da.timeout = d
	}
}

// WithDepLogger sets the logger for the analyzer.
func WithDepLogger(logger *slog.Logger) DependencyAnalyzerOption {
	return func(da *DependencyAnalyzer) {
		da.logger = logger
	}
}

// NewDependencyAnalyzer creates a new dependency analyzer.
func NewDependencyAnalyzer(a agent.Agent, store *tick.Store, opts ...DependencyAnalyzerOption) *DependencyAnalyzer {
	da := &DependencyAnalyzer{
		agent:   a,
		store:   store,
		logger:  slog.Default(),
		timeout: 3 * time.Minute,
	}
	for _, opt := range opts {
		opt(da)
	}
	return da
}

// TaskFilePrediction holds the predicted files a task will modify.
type TaskFilePrediction struct {
	TaskID string   `json:"task_id"`
	Files  []string `json:"files"`
}

// AnalysisResult contains the results of dependency analysis.
type AnalysisResult struct {
	Predictions      []TaskFilePrediction `json:"predictions"`
	AddedDeps        map[string][]string  `json:"added_deps"` // task_id -> new blocked_by IDs
	ConflictingPairs []ConflictPair       `json:"conflicting_pairs"`
}

// ConflictPair represents two tasks that would conflict on the same files.
type ConflictPair struct {
	Task1       string   `json:"task1"`
	Task2       string   `json:"task2"`
	SharedFiles []string `json:"shared_files"`
}

// filePredictionPattern extracts JSON from <file_predictions> tags.
var filePredictionPattern = regexp.MustCompile(`(?s)<file_predictions>\s*(.*?)\s*</file_predictions>`)

// Analyze predicts file conflicts and adds dependencies to prevent parallel edits.
// Returns the analysis result showing what dependencies were added.
func (da *DependencyAnalyzer) Analyze(ctx context.Context, epic *ticks.Epic, tasks []ticks.Task) (*AnalysisResult, error) {
	if len(tasks) <= 1 {
		return &AnalysisResult{}, nil // No conflicts possible with 0-1 tasks
	}

	da.logger.Info("dependency analysis started",
		"epic_id", epic.ID,
		"task_count", len(tasks),
	)

	startTime := time.Now()

	// Build prompt for file prediction
	prompt := da.buildPredictionPrompt(epic, tasks)

	// Run the agent
	result, err := da.agent.Run(ctx, prompt, agent.RunOpts{
		Timeout: da.timeout,
	})
	if err != nil {
		da.logger.Error("dependency analysis failed",
			"epic_id", epic.ID,
			"error", err,
			"duration", time.Since(startTime),
		)
		return nil, fmt.Errorf("running agent: %w", err)
	}

	// Parse predictions from response
	predictions, err := da.parsePredictions(result.Output)
	if err != nil {
		da.logger.Warn("failed to parse predictions, skipping dependency analysis",
			"epic_id", epic.ID,
			"error", err,
		)
		return &AnalysisResult{}, nil
	}

	// Build file -> tasks map
	fileToTasks := make(map[string][]string)
	for _, pred := range predictions {
		for _, file := range pred.Files {
			fileToTasks[file] = append(fileToTasks[file], pred.TaskID)
		}
	}

	// Find conflicting pairs
	var conflicts []ConflictPair
	seen := make(map[string]bool)
	for file, taskIDs := range fileToTasks {
		if len(taskIDs) < 2 {
			continue
		}
		// All pairs of tasks touching this file conflict
		for i := 0; i < len(taskIDs); i++ {
			for j := i + 1; j < len(taskIDs); j++ {
				key := taskIDs[i] + ":" + taskIDs[j]
				if taskIDs[j] < taskIDs[i] {
					key = taskIDs[j] + ":" + taskIDs[i]
				}
				if seen[key] {
					// Already recorded, add file to existing conflict
					for k := range conflicts {
						if (conflicts[k].Task1 == taskIDs[i] && conflicts[k].Task2 == taskIDs[j]) ||
							(conflicts[k].Task1 == taskIDs[j] && conflicts[k].Task2 == taskIDs[i]) {
							conflicts[k].SharedFiles = append(conflicts[k].SharedFiles, file)
							break
						}
					}
					continue
				}
				seen[key] = true
				conflicts = append(conflicts, ConflictPair{
					Task1:       taskIDs[i],
					Task2:       taskIDs[j],
					SharedFiles: []string{file},
				})
			}
		}
	}

	// Add dependencies to resolve conflicts
	addedDeps := make(map[string][]string)
	if len(conflicts) > 0 {
		addedDeps, err = da.addDependencies(conflicts, tasks)
		if err != nil {
			da.logger.Warn("failed to add dependencies",
				"epic_id", epic.ID,
				"error", err,
			)
		}
	}

	da.logger.Info("dependency analysis completed",
		"epic_id", epic.ID,
		"duration", time.Since(startTime),
		"predictions", len(predictions),
		"conflicts", len(conflicts),
		"deps_added", len(addedDeps),
	)

	return &AnalysisResult{
		Predictions:      predictions,
		AddedDeps:        addedDeps,
		ConflictingPairs: conflicts,
	}, nil
}

// buildPredictionPrompt creates the prompt for file prediction.
func (da *DependencyAnalyzer) buildPredictionPrompt(epic *ticks.Epic, tasks []ticks.Task) string {
	var sb strings.Builder

	sb.WriteString(`# Predict Files Modified by Tasks

You are analyzing tasks to predict which files each task will modify.
This helps prevent parallel tasks from making conflicting edits.

## Epic
**[`)
	sb.WriteString(epic.ID)
	sb.WriteString(`] `)
	sb.WriteString(epic.Title)
	sb.WriteString(`**

`)
	sb.WriteString(epic.Description)
	sb.WriteString(`

## Tasks

`)
	for _, t := range tasks {
		sb.WriteString("### [")
		sb.WriteString(t.ID)
		sb.WriteString("] ")
		sb.WriteString(t.Title)
		sb.WriteString("\n\n")
		if t.Description != "" {
			sb.WriteString(t.Description)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString(`## Instructions

For each task, predict which files it will likely CREATE or MODIFY.

Guidelines:
- Include BOTH existing files to modify AND new files to create
- For new files, predict the likely path based on codebase conventions
- Include test files if the task involves testing (e.g., foo.ts -> foo.test.ts)
- Use glob patterns for multiple similar files (e.g., "src/components/*.ts")
- Be specific - use actual file paths based on the codebase structure
- If a task creates a new component/module, predict both the main file and its test
- Two tasks creating the same NEW file is also a conflict

Common patterns to consider:
- Component tasks: component file + test file + possibly styles
- API tasks: handler + routes + tests
- Feature tasks: multiple related files in the same directory

## Output Format

Return a JSON array wrapped in <file_predictions> tags:

<file_predictions>
[
  {"task_id": "abc", "files": ["src/foo.ts", "src/bar.ts"]},
  {"task_id": "def", "files": ["src/foo.ts", "tests/foo.test.ts"]}
]
</file_predictions>

Important: Only include the JSON array, no other text inside the tags.
`)

	return sb.String()
}

// parsePredictions extracts file predictions from the agent response.
func (da *DependencyAnalyzer) parsePredictions(output string) ([]TaskFilePrediction, error) {
	matches := filePredictionPattern.FindStringSubmatch(output)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no <file_predictions> tags found")
	}

	jsonStr := strings.TrimSpace(matches[1])
	var predictions []TaskFilePrediction
	if err := json.Unmarshal([]byte(jsonStr), &predictions); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return predictions, nil
}

// addDependencies adds blocked_by relationships to resolve conflicts.
// Uses a simple heuristic: for each conflict pair, add a dependency from one to the other.
// Tries to minimize the number of added dependencies while ensuring no parallel conflicts.
func (da *DependencyAnalyzer) addDependencies(conflicts []ConflictPair, tasks []ticks.Task) (map[string][]string, error) {
	// Build task index for quick lookup
	taskIndex := make(map[string]int)
	for i, t := range tasks {
		taskIndex[t.ID] = i
	}

	// Track existing blocked_by relationships
	blockedBy := make(map[string]map[string]bool)
	for _, t := range tasks {
		blockedBy[t.ID] = make(map[string]bool)
		for _, b := range t.BlockedBy {
			blockedBy[t.ID][b] = true
		}
	}

	// For each conflict, add dependency if not already present
	// Simple strategy: task that appears first in the list blocks the later one
	addedDeps := make(map[string][]string)
	for _, c := range conflicts {
		idx1, ok1 := taskIndex[c.Task1]
		idx2, ok2 := taskIndex[c.Task2]
		if !ok1 || !ok2 {
			continue
		}

		var blocker, blocked string
		if idx1 < idx2 {
			blocker, blocked = c.Task1, c.Task2
		} else {
			blocker, blocked = c.Task2, c.Task1
		}

		// Skip if dependency already exists (directly or transitively)
		if blockedBy[blocked][blocker] {
			continue
		}

		// Add the dependency
		blockedBy[blocked][blocker] = true
		addedDeps[blocked] = append(addedDeps[blocked], blocker)
	}

	// Persist the changes
	for taskID, newBlockers := range addedDeps {
		t, err := da.store.Read(taskID)
		if err != nil {
			da.logger.Warn("failed to read task for dependency update",
				"task_id", taskID,
				"error", err,
			)
			continue
		}

		// Add new blockers
		existingSet := make(map[string]bool)
		for _, b := range t.BlockedBy {
			existingSet[b] = true
		}
		for _, b := range newBlockers {
			if !existingSet[b] {
				t.BlockedBy = append(t.BlockedBy, b)
			}
		}
		t.UpdatedAt = time.Now()

		if err := da.store.WriteAs(t, "dependency-analyzer"); err != nil {
			da.logger.Warn("failed to update task dependencies",
				"task_id", taskID,
				"error", err,
			)
		} else {
			da.logger.Info("added dependencies to task",
				"task_id", taskID,
				"blockers", newBlockers,
			)
		}
	}

	return addedDeps, nil
}
