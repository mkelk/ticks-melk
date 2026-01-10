// Package beads provides import functionality for beads issue tracker.
package beads

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"
)

// Issue represents a beads issue from JSONL export.
type Issue struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description,omitempty"`
	Notes              string       `json:"notes,omitempty"`
	Status             string       `json:"status"`
	Priority           int          `json:"priority"`
	IssueType          string       `json:"issue_type"`
	Assignee           string       `json:"assignee,omitempty"`
	Labels             []string     `json:"labels,omitempty"`
	Dependencies       []Dependency `json:"dependencies,omitempty"`
	AcceptanceCriteria string       `json:"acceptance_criteria,omitempty"`
	Design             string       `json:"design,omitempty"`
	ExternalRef        string       `json:"external_ref,omitempty"`
	DeferUntil         *time.Time   `json:"defer_until,omitempty"`
	DueAt              *time.Time   `json:"due_at,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	CreatedBy          string       `json:"created_by,omitempty"`
	UpdatedAt          time.Time    `json:"updated_at"`
	ClosedAt           *time.Time   `json:"closed_at,omitempty"`
	CloseReason        string       `json:"close_reason,omitempty"`
	DeletedAt          *time.Time   `json:"deleted_at,omitempty"`
}

// Dependency represents a relationship between issues.
type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
}

// ParseFile reads a beads JSONL file and returns all issues.
func ParseFile(path string) ([]Issue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parse reads beads issues from a JSONL reader.
func Parse(r io.Reader) ([]Issue, error) {
	var issues []Issue
	scanner := bufio.NewScanner(r)
	// Increase buffer for large issues
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var issue Issue
		if err := json.Unmarshal(line, &issue); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, scanner.Err()
}

// FilterImportable returns issues that should be imported (not closed, not deleted).
func FilterImportable(issues []Issue) []Issue {
	var out []Issue
	for _, issue := range issues {
		if issue.Status == "closed" || issue.Status == "tombstone" {
			continue
		}
		if issue.DeletedAt != nil {
			continue
		}
		out = append(out, issue)
	}
	return out
}
