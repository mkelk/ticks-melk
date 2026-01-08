package query

import (
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/tick"
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
