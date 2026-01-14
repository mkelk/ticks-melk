package query

import (
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/tick"
)

// TestBackwardsCompatibility_Ready tests that Ready() correctly excludes
// ticks with Manual=true (legacy) just like it excludes ticks with Awaiting set.
func TestBackwardsCompatibility_Ready(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)

	t.Run("Ready_excludes_Manual_true_ticks", func(t *testing.T) {
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 1 {
			t.Fatalf("expected 1 ready tick, got %d", len(ready))
		}
		if ready[0].ID != "ready" {
			t.Fatalf("expected 'ready' tick, got %s", ready[0].ID)
		}
	})

	t.Run("Ready_excludes_Manual_true_same_as_Awaiting", func(t *testing.T) {
		// Both Manual=true and Awaiting="work" should be treated the same
		awaiting := tick.AwaitingWork
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-work", Status: tick.StatusOpen, Awaiting: &awaiting, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 1 {
			t.Fatalf("expected 1 ready tick, got %d", len(ready))
		}
		if ready[0].ID != "ready" {
			t.Fatalf("expected 'ready' tick, got %s", ready[0].ID)
		}
	})

	t.Run("Ready_excludes_Manual_true_with_other_awaiting_types", func(t *testing.T) {
		// Test that Manual=true is excluded alongside all other awaiting types
		approval := tick.AwaitingApproval
		input := tick.AwaitingInput
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-approval", Status: tick.StatusOpen, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-input", Status: tick.StatusOpen, Awaiting: &input, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 1 {
			t.Fatalf("expected 1 ready tick, got %d", len(ready))
		}
	})
}

// TestBackwardsCompatibility_ListFilter tests that tk list --awaiting includes
// Manual=true ticks when filtering for awaiting=work.
func TestBackwardsCompatibility_ListFilter(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)

	t.Run("Filter_Awaiting_work_includes_Manual_true", func(t *testing.T) {
		work := tick.AwaitingWork
		items := []tick.Tick{
			{ID: "not-awaiting", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-work", Status: tick.StatusOpen, Awaiting: &work, CreatedAt: now, UpdatedAt: now},
		}

		// Filter for awaiting=work should match both manual and awaiting-work
		filter := tick.AwaitingWork
		filtered := Apply(items, Filter{Awaiting: &filter})
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ticks (manual + awaiting-work), got %d", len(filtered))
		}

		ids := map[string]bool{}
		for _, f := range filtered {
			ids[f.ID] = true
		}
		if !ids["manual"] || !ids["awaiting-work"] {
			t.Fatalf("expected both manual and awaiting-work, got %v", ids)
		}
	})

	t.Run("Filter_Awaiting_approval_excludes_Manual_true", func(t *testing.T) {
		approval := tick.AwaitingApproval
		items := []tick.Tick{
			{ID: "not-awaiting", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-approval", Status: tick.StatusOpen, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
		}

		// Filter for awaiting=approval should NOT include manual (which maps to work)
		filter := tick.AwaitingApproval
		filtered := Apply(items, Filter{Awaiting: &filter})
		if len(filtered) != 1 {
			t.Fatalf("expected 1 tick (awaiting-approval only), got %d", len(filtered))
		}
		if filtered[0].ID != "awaiting-approval" {
			t.Fatalf("expected awaiting-approval tick, got %s", filtered[0].ID)
		}
	})

	t.Run("Filter_Awaiting_any_with_work_includes_Manual", func(t *testing.T) {
		approval := tick.AwaitingApproval
		items := []tick.Tick{
			{ID: "not-awaiting", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-approval", Status: tick.StatusOpen, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
		}

		// AwaitingAny with work should include manual
		filtered := Apply(items, Filter{AwaitingAny: []string{tick.AwaitingApproval, tick.AwaitingWork}})
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ticks, got %d", len(filtered))
		}

		ids := map[string]bool{}
		for _, f := range filtered {
			ids[f.ID] = true
		}
		if !ids["manual"] || !ids["awaiting-approval"] {
			t.Fatalf("expected manual and awaiting-approval, got %v", ids)
		}
	})

	t.Run("Filter_Awaiting_empty_shows_all_awaiting_including_Manual", func(t *testing.T) {
		approval := tick.AwaitingApproval
		empty := ""
		items := []tick.Tick{
			{ID: "not-awaiting", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-approval", Status: tick.StatusOpen, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
		}

		// Filter with Awaiting="" should show only ticks NOT awaiting
		// (empty string means "filter for ticks with no awaiting")
		filtered := Apply(items, Filter{Awaiting: &empty})
		if len(filtered) != 1 {
			t.Fatalf("expected 1 tick (not awaiting), got %d", len(filtered))
		}
		if filtered[0].ID != "not-awaiting" {
			t.Fatalf("expected not-awaiting tick, got %s", filtered[0].ID)
		}
	})
}

// TestBackwardsCompatibility_NextAwaiting tests that tk next --awaiting returns
// Manual=true ticks.
func TestBackwardsCompatibility_NextAwaiting(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)

	t.Run("ReadyIncludeAwaiting_includes_Manual_true", func(t *testing.T) {
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
		}

		ready := ReadyIncludeAwaiting(items)
		if len(ready) != 2 {
			t.Fatalf("expected 2 ticks, got %d", len(ready))
		}
	})

	t.Run("ReadyIncludeAwaiting_includes_all_awaiting_types", func(t *testing.T) {
		approval := tick.AwaitingApproval
		work := tick.AwaitingWork
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-approval", Status: tick.StatusOpen, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
			{ID: "awaiting-work", Status: tick.StatusOpen, Awaiting: &work, CreatedAt: now, UpdatedAt: now},
		}

		ready := ReadyIncludeAwaiting(items)
		if len(ready) != 4 {
			t.Fatalf("expected 4 ticks, got %d", len(ready))
		}
	})
}

// TestBackwardsCompatibility_MixedState tests edge cases where both Manual
// and Awaiting are set (shouldn't happen in practice, but test the precedence).
func TestBackwardsCompatibility_MixedState(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)

	t.Run("Filter_prefers_Awaiting_over_Manual_when_both_set", func(t *testing.T) {
		// Tick has both Manual=true and Awaiting=approval
		// Filter for approval should match (Awaiting takes precedence)
		approval := tick.AwaitingApproval
		items := []tick.Tick{
			{ID: "both", Status: tick.StatusOpen, Manual: true, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
		}

		// Filter for approval
		filter := tick.AwaitingApproval
		filtered := Apply(items, Filter{Awaiting: &filter})
		if len(filtered) != 1 {
			t.Fatalf("expected 1 tick, got %d", len(filtered))
		}

		// Filter for work should NOT match (Awaiting=approval takes precedence)
		filterWork := tick.AwaitingWork
		filteredWork := Apply(items, Filter{Awaiting: &filterWork})
		if len(filteredWork) != 0 {
			t.Fatalf("expected 0 ticks (Awaiting=approval takes precedence over Manual=true), got %d", len(filteredWork))
		}
	})

	t.Run("Ready_excludes_when_both_Manual_and_Awaiting_set", func(t *testing.T) {
		// Tick with both set should still be excluded from Ready
		approval := tick.AwaitingApproval
		items := []tick.Tick{
			{ID: "ready", Status: tick.StatusOpen, CreatedAt: now, UpdatedAt: now},
			{ID: "both", Status: tick.StatusOpen, Manual: true, Awaiting: &approval, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 1 {
			t.Fatalf("expected 1 ready tick, got %d", len(ready))
		}
		if ready[0].ID != "ready" {
			t.Fatalf("expected 'ready' tick, got %s", ready[0].ID)
		}
	})
}

// TestBackwardsCompatibility_StatusInteraction tests that Manual/Awaiting
// behavior is correct across different tick statuses.
func TestBackwardsCompatibility_StatusInteraction(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)

	t.Run("Ready_excludes_closed_Manual_true", func(t *testing.T) {
		// Closed ticks should not be in Ready, regardless of Manual
		items := []tick.Tick{
			{ID: "closed-manual", Status: tick.StatusClosed, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "open-manual", Status: tick.StatusOpen, Manual: true, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 0 {
			t.Fatalf("expected 0 ready ticks (manual ticks excluded), got %d", len(ready))
		}
	})

	t.Run("Ready_excludes_in_progress_Manual_true", func(t *testing.T) {
		// In-progress ticks with Manual should also be excluded
		items := []tick.Tick{
			{ID: "in-progress-manual", Status: tick.StatusInProgress, Manual: true, CreatedAt: now, UpdatedAt: now},
			{ID: "in-progress-regular", Status: tick.StatusInProgress, CreatedAt: now, UpdatedAt: now},
		}

		ready := Ready(items)
		if len(ready) != 1 {
			t.Fatalf("expected 1 ready tick, got %d", len(ready))
		}
		if ready[0].ID != "in-progress-regular" {
			t.Fatalf("expected in-progress-regular, got %s", ready[0].ID)
		}
	})
}
