package query

import (
	"testing"
	"time"

	"github.com/mkelk/ticks-melk/internal/tick"
)

func TestApplyFilter(t *testing.T) {
	base := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "a", Owner: "alice", Status: tick.StatusOpen, Priority: 1, Type: tick.TypeBug, Labels: []string{"backend"}, Parent: "epic1", CreatedAt: base},
		{ID: "b", Owner: "bob", Status: tick.StatusClosed, Priority: 2, Type: tick.TypeTask, Labels: []string{"frontend"}, Parent: "epic2", CreatedAt: base.Add(time.Minute)},
	}

	prio := 1
	filtered := Apply(items, Filter{Owner: "alice", Priority: &prio})
	if len(filtered) != 1 || filtered[0].ID != "a" {
		t.Fatalf("unexpected filter result: %+v", filtered)
	}

	filtered = Apply(items, Filter{Label: "frontend"})
	if len(filtered) != 1 || filtered[0].ID != "b" {
		t.Fatalf("unexpected label filter result: %+v", filtered)
	}
}

func TestSortByPriorityCreatedAt(t *testing.T) {
	base := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "b", Priority: 2, CreatedAt: base},
		{ID: "a", Priority: 1, CreatedAt: base.Add(time.Minute)},
		{ID: "c", Priority: 1, CreatedAt: base},
	}

	SortByPriorityCreatedAt(items)
	if items[0].ID != "c" || items[1].ID != "a" || items[2].ID != "b" {
		t.Fatalf("unexpected order: %v, %v, %v", items[0].ID, items[1].ID, items[2].ID)
	}
}

func TestSortInProgressFirst(t *testing.T) {
	base := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "a", Status: tick.StatusOpen, Priority: 1, CreatedAt: base},
		{ID: "b", Status: tick.StatusInProgress, Priority: 2, CreatedAt: base}, // lower priority but in_progress
		{ID: "c", Status: tick.StatusOpen, Priority: 1, CreatedAt: base.Add(time.Minute)},
	}

	SortByPriorityCreatedAt(items)
	// in_progress should come first, even though it has lower priority
	if items[0].ID != "b" {
		t.Fatalf("in_progress task should be first, got: %v", items[0].ID)
	}
	// then open tasks by priority, then created_at
	if items[1].ID != "a" || items[2].ID != "c" {
		t.Fatalf("unexpected order for open tasks: %v, %v", items[1].ID, items[2].ID)
	}
}

func TestApplyFilterByProject(t *testing.T) {
	base := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "a", Project: "proj-a", CreatedAt: base},
		{ID: "b", Project: "proj-b", CreatedAt: base.Add(time.Minute)},
		{ID: "c", Project: "", CreatedAt: base.Add(2 * time.Minute)}, // no project
	}

	// Filter by project should return only matching ticks
	filtered := Apply(items, Filter{Project: "proj-a"})
	if len(filtered) != 1 || filtered[0].ID != "a" {
		t.Fatalf("expected only tick 'a', got: %+v", filtered)
	}

	// Ticks with empty project should be excluded when filter is set
	filtered = Apply(items, Filter{Project: "proj-b"})
	if len(filtered) != 1 || filtered[0].ID != "b" {
		t.Fatalf("expected only tick 'b', got: %+v", filtered)
	}

	// No project filter should return all
	filtered = Apply(items, Filter{})
	if len(filtered) != 3 {
		t.Fatalf("expected 3 ticks, got: %d", len(filtered))
	}

	// Non-matching project should return empty
	filtered = Apply(items, Filter{Project: "nonexistent"})
	if len(filtered) != 0 {
		t.Fatalf("expected 0 ticks, got: %d", len(filtered))
	}
}

func TestApplyFilterProjectCombinedWithOtherFilters(t *testing.T) {
	base := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	items := []tick.Tick{
		{ID: "a", Project: "proj-a", Type: tick.TypeBug, Status: tick.StatusOpen, CreatedAt: base},
		{ID: "b", Project: "proj-a", Type: tick.TypeTask, Status: tick.StatusOpen, CreatedAt: base.Add(time.Minute)},
		{ID: "c", Project: "proj-b", Type: tick.TypeBug, Status: tick.StatusOpen, CreatedAt: base.Add(2 * time.Minute)},
		{ID: "d", Project: "proj-a", Type: tick.TypeBug, Status: tick.StatusClosed, CreatedAt: base.Add(3 * time.Minute)},
	}

	// Combine project filter with type filter
	filtered := Apply(items, Filter{Project: "proj-a", Type: tick.TypeBug})
	if len(filtered) != 2 {
		t.Fatalf("expected 2 ticks (bugs in proj-a), got: %d", len(filtered))
	}
	for _, tick := range filtered {
		if tick.Project != "proj-a" || tick.Type != "bug" {
			t.Fatalf("unexpected tick: %+v", tick)
		}
	}

	// Combine project filter with status filter
	filtered = Apply(items, Filter{Project: "proj-a", Status: tick.StatusOpen})
	if len(filtered) != 2 {
		t.Fatalf("expected 2 ticks (open in proj-a), got: %d", len(filtered))
	}
	for _, tick := range filtered {
		if tick.Project != "proj-a" || tick.Status != "open" {
			t.Fatalf("unexpected tick: %+v", tick)
		}
	}
}
