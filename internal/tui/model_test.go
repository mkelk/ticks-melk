package tui

import (
	"testing"

	"github.com/pengelbrecht/ticks/internal/tick"
)

func TestApplyAwaitingFilter(t *testing.T) {
	// Create test ticks with various awaiting states
	work := tick.AwaitingWork
	approval := tick.AwaitingApproval
	review := tick.AwaitingReview

	ticks := []tick.Tick{
		{ID: "t1", Title: "Not awaiting"},
		{ID: "t2", Title: "Awaiting work", Awaiting: &work},
		{ID: "t3", Title: "Awaiting approval", Awaiting: &approval},
		{ID: "t4", Title: "Legacy manual", Manual: true},
		{ID: "t5", Title: "Awaiting review", Awaiting: &review},
		{ID: "t6", Title: "Also not awaiting"},
	}

	t.Run("filter off returns all ticks", func(t *testing.T) {
		result := applyAwaitingFilter(ticks, awaitingFilterOff, "")
		if len(result) != 6 {
			t.Errorf("expected 6 ticks, got %d", len(result))
		}
	})

	t.Run("human only filter returns awaiting ticks", func(t *testing.T) {
		result := applyAwaitingFilter(ticks, awaitingFilterHumanOnly, "")
		if len(result) != 4 {
			t.Errorf("expected 4 awaiting ticks, got %d", len(result))
		}
		// Verify all results are awaiting
		for _, tk := range result {
			if !tk.IsAwaitingHuman() {
				t.Errorf("expected tick %s to be awaiting human", tk.ID)
			}
		}
	})

	t.Run("agent only filter returns non-awaiting ticks", func(t *testing.T) {
		result := applyAwaitingFilter(ticks, awaitingFilterAgentOnly, "")
		if len(result) != 2 {
			t.Errorf("expected 2 non-awaiting ticks, got %d", len(result))
		}
		// Verify none are awaiting
		for _, tk := range result {
			if tk.IsAwaitingHuman() {
				t.Errorf("expected tick %s to NOT be awaiting human", tk.ID)
			}
		}
	})

	t.Run("by type filter returns specific type", func(t *testing.T) {
		result := applyAwaitingFilter(ticks, awaitingFilterByType, tick.AwaitingApproval)
		if len(result) != 1 {
			t.Errorf("expected 1 tick awaiting approval, got %d", len(result))
		}
		if result[0].ID != "t3" {
			t.Errorf("expected tick t3, got %s", result[0].ID)
		}
	})

	t.Run("by type filter for work includes legacy manual", func(t *testing.T) {
		result := applyAwaitingFilter(ticks, awaitingFilterByType, tick.AwaitingWork)
		if len(result) != 2 {
			t.Errorf("expected 2 ticks awaiting work (including legacy), got %d", len(result))
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		result := applyAwaitingFilter([]tick.Tick{}, awaitingFilterHumanOnly, "")
		if len(result) != 0 {
			t.Errorf("expected 0 ticks, got %d", len(result))
		}
	})
}
