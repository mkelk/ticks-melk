package tick

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTickValidateValid(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	valid := Tick{
		ID:        "a1b",
		Title:     "Fix auth",
		Status:    StatusOpen,
		Priority:  2,
		Type:      TypeBug,
		Owner:     "petere",
		CreatedBy: "petere",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid tick, got error: %v", err)
	}
}

func TestTickValidateRequiredFields(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	base := Tick{
		ID:        "a1b",
		Title:     "Fix auth",
		Status:    StatusOpen,
		Priority:  2,
		Type:      TypeBug,
		Owner:     "petere",
		CreatedBy: "petere",
		CreatedAt: now,
		UpdatedAt: now,
	}

	cases := []struct {
		name     string
		mutate   func(t Tick) Tick
		expected string
	}{
		{"missing id", func(t Tick) Tick { t.ID = ""; return t }, "id is required"},
		{"missing title", func(t Tick) Tick { t.Title = ""; return t }, "title is required"},
		{"missing status", func(t Tick) Tick { t.Status = ""; return t }, "status is required"},
		{"missing type", func(t Tick) Tick { t.Type = ""; return t }, "type is required"},
		{"missing owner", func(t Tick) Tick { t.Owner = ""; return t }, "owner is required"},
		{"missing created_by", func(t Tick) Tick { t.CreatedBy = ""; return t }, "created_by is required"},
		{"missing created_at", func(t Tick) Tick { t.CreatedAt = time.Time{}; return t }, "created_at is required"},
		{"missing updated_at", func(t Tick) Tick { t.UpdatedAt = time.Time{}; return t }, "updated_at is required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := tc.mutate(base)
			err := mutated.Validate()
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.expected) {
				t.Fatalf("expected error to contain %q, got %q", tc.expected, err.Error())
			}
		})
	}
}

func TestTickValidateEnums(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	base := Tick{
		ID:        "a1b",
		Title:     "Fix auth",
		Status:    StatusOpen,
		Priority:  2,
		Type:      TypeBug,
		Owner:     "petere",
		CreatedBy: "petere",
		CreatedAt: now,
		UpdatedAt: now,
	}

	invalidStatus := base
	invalidStatus.Status = "broken"
	if err := invalidStatus.Validate(); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}

	invalidType := base
	invalidType.Type = "unknown"
	if err := invalidType.Validate(); err == nil || !strings.Contains(err.Error(), "invalid type") {
		t.Fatalf("expected invalid type error, got %v", err)
	}

	lowPriority := base
	lowPriority.Priority = -1
	if err := lowPriority.Validate(); err == nil || !strings.Contains(err.Error(), "priority") {
		t.Fatalf("expected priority error, got %v", err)
	}

	highPriority := base
	highPriority.Priority = 5
	if err := highPriority.Validate(); err == nil || !strings.Contains(err.Error(), "priority") {
		t.Fatalf("expected priority error, got %v", err)
	}
}

func TestTickProjectJSONMarshalUnmarshal(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)

	t.Run("marshal with project", func(t *testing.T) {
		tick := Tick{
			ID:        "abc",
			Title:     "Test task",
			Status:    StatusOpen,
			Priority:  2,
			Type:      TypeTask,
			Owner:     "petere",
			Project:   "2026-01-14-5464-project-dim",
			CreatedBy: "petere",
			CreatedAt: now,
			UpdatedAt: now,
		}

		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal tick: %v", err)
		}

		if !strings.Contains(string(data), `"project":"2026-01-14-5464-project-dim"`) {
			t.Fatalf("expected JSON to contain project field, got: %s", string(data))
		}
	})

	t.Run("marshal without project omits field", func(t *testing.T) {
		tick := Tick{
			ID:        "abc",
			Title:     "Test task",
			Status:    StatusOpen,
			Priority:  2,
			Type:      TypeTask,
			Owner:     "petere",
			CreatedBy: "petere",
			CreatedAt: now,
			UpdatedAt: now,
		}

		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal tick: %v", err)
		}

		if strings.Contains(string(data), `"project"`) {
			t.Fatalf("expected JSON to omit project field when empty, got: %s", string(data))
		}
	})

	t.Run("unmarshal with project", func(t *testing.T) {
		jsonData := `{
			"id": "xyz",
			"title": "From JSON",
			"status": "open",
			"priority": 1,
			"type": "feature",
			"owner": "alice",
			"project": "my-project-123",
			"created_by": "alice",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z"
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonData), &tick); err != nil {
			t.Fatalf("failed to unmarshal tick: %v", err)
		}

		if tick.Project != "my-project-123" {
			t.Fatalf("expected project 'my-project-123', got '%s'", tick.Project)
		}
	})

	t.Run("unmarshal without project defaults to empty", func(t *testing.T) {
		jsonData := `{
			"id": "xyz",
			"title": "From JSON",
			"status": "open",
			"priority": 1,
			"type": "feature",
			"owner": "alice",
			"created_by": "alice",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z"
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonData), &tick); err != nil {
			t.Fatalf("failed to unmarshal tick: %v", err)
		}

		if tick.Project != "" {
			t.Fatalf("expected empty project, got '%s'", tick.Project)
		}
	})
}
