package tick

import (
	"encoding/json"
	"testing"
	"time"
)

// TestBackwardsCompatibility_HelperMethods tests the helper methods that handle
// backwards compatibility between Manual field and Awaiting field.
func TestBackwardsCompatibility_HelperMethods(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	baseTick := func() Tick {
		return Tick{
			ID:        "a1b",
			Title:     "Test tick",
			Status:    StatusOpen,
			Priority:  2,
			Type:      TypeTask,
			Owner:     "petere",
			CreatedBy: "petere",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	t.Run("IsAwaitingHuman_Manual_true_Awaiting_nil", func(t *testing.T) {
		// Manual=true, Awaiting=nil → true
		tick := baseTick()
		tick.Manual = true
		tick.Awaiting = nil

		if !tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman() should return true when Manual=true, Awaiting=nil")
		}
	})

	t.Run("IsAwaitingHuman_Manual_false_Awaiting_set", func(t *testing.T) {
		// Manual=false, Awaiting="work" → true
		tick := baseTick()
		tick.Manual = false
		awaiting := AwaitingWork
		tick.Awaiting = &awaiting

		if !tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman() should return true when Manual=false, Awaiting=work")
		}
	})

	t.Run("IsAwaitingHuman_both_set_Awaiting_wins", func(t *testing.T) {
		// Manual=true, Awaiting="approval" → true (awaiting wins)
		tick := baseTick()
		tick.Manual = true
		awaiting := AwaitingApproval
		tick.Awaiting = &awaiting

		if !tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman() should return true when both Manual and Awaiting are set")
		}
	})

	t.Run("IsAwaitingHuman_neither_set", func(t *testing.T) {
		// Manual=false, Awaiting=nil → false
		tick := baseTick()
		tick.Manual = false
		tick.Awaiting = nil

		if tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman() should return false when neither Manual nor Awaiting is set")
		}
	})

	t.Run("GetAwaitingType_Manual_true_returns_work", func(t *testing.T) {
		// Manual=true → "work"
		tick := baseTick()
		tick.Manual = true
		tick.Awaiting = nil

		if got := tick.GetAwaitingType(); got != AwaitingWork {
			t.Errorf("GetAwaitingType() should return %q for Manual=true, got %q", AwaitingWork, got)
		}
	})

	t.Run("GetAwaitingType_Awaiting_approval_returns_approval", func(t *testing.T) {
		// Awaiting="approval" → "approval"
		tick := baseTick()
		tick.Manual = false
		awaiting := AwaitingApproval
		tick.Awaiting = &awaiting

		if got := tick.GetAwaitingType(); got != AwaitingApproval {
			t.Errorf("GetAwaitingType() should return %q, got %q", AwaitingApproval, got)
		}
	})

	t.Run("GetAwaitingType_both_set_Awaiting_takes_precedence", func(t *testing.T) {
		// Manual=true, Awaiting="approval" → "approval" (Awaiting takes precedence)
		tick := baseTick()
		tick.Manual = true
		awaiting := AwaitingApproval
		tick.Awaiting = &awaiting

		if got := tick.GetAwaitingType(); got != AwaitingApproval {
			t.Errorf("GetAwaitingType() should return %q (Awaiting takes precedence), got %q", AwaitingApproval, got)
		}
	})

	t.Run("GetAwaitingType_neither_set_returns_empty", func(t *testing.T) {
		// Neither set → ""
		tick := baseTick()
		tick.Manual = false
		tick.Awaiting = nil

		if got := tick.GetAwaitingType(); got != "" {
			t.Errorf("GetAwaitingType() should return empty string, got %q", got)
		}
	})

	t.Run("GetAwaitingType_all_awaiting_types", func(t *testing.T) {
		// Test all valid awaiting types return their value
		for _, awaitingType := range ValidAwaitingValues {
			t.Run(awaitingType, func(t *testing.T) {
				tick := baseTick()
				aw := awaitingType
				tick.Awaiting = &aw

				if got := tick.GetAwaitingType(); got != awaitingType {
					t.Errorf("GetAwaitingType() should return %q, got %q", awaitingType, got)
				}
			})
		}
	})
}

// TestBackwardsCompatibility_Migration tests that setting awaiting clears manual
// and that the --manual flag behavior is correct.
func TestBackwardsCompatibility_Migration(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	baseTick := func() Tick {
		return Tick{
			ID:        "a1b",
			Title:     "Test tick",
			Status:    StatusOpen,
			Priority:  2,
			Type:      TypeTask,
			Owner:     "petere",
			CreatedBy: "petere",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	t.Run("SetAwaiting_clears_manual", func(t *testing.T) {
		// Setting awaiting clears manual
		tick := baseTick()
		tick.Manual = true

		tick.SetAwaiting(AwaitingApproval)

		if tick.Manual {
			t.Error("SetAwaiting should clear Manual field")
		}
		if tick.Awaiting == nil || *tick.Awaiting != AwaitingApproval {
			t.Errorf("SetAwaiting should set Awaiting to %q", AwaitingApproval)
		}
	})

	t.Run("SetAwaiting_work_clears_manual", func(t *testing.T) {
		// Setting awaiting=work should also clear manual (migration path)
		tick := baseTick()
		tick.Manual = true

		tick.SetAwaiting(AwaitingWork)

		if tick.Manual {
			t.Error("SetAwaiting(work) should clear Manual field")
		}
		if tick.Awaiting == nil || *tick.Awaiting != AwaitingWork {
			t.Errorf("SetAwaiting should set Awaiting to %q", AwaitingWork)
		}
	})

	t.Run("SetAwaiting_empty_clears_both", func(t *testing.T) {
		// SetAwaiting("") should clear both
		tick := baseTick()
		tick.Manual = true
		awaiting := AwaitingApproval
		tick.Awaiting = &awaiting

		tick.SetAwaiting("")

		if tick.Manual {
			t.Error("SetAwaiting(\"\") should clear Manual field")
		}
		if tick.Awaiting != nil {
			t.Error("SetAwaiting(\"\") should clear Awaiting field")
		}
	})

	t.Run("ClearAwaiting_clears_both_fields", func(t *testing.T) {
		// ClearAwaiting should clear both Manual and Awaiting
		tick := baseTick()
		tick.Manual = true
		awaiting := AwaitingWork
		tick.Awaiting = &awaiting

		tick.ClearAwaiting()

		if tick.Manual {
			t.Error("ClearAwaiting should clear Manual field")
		}
		if tick.Awaiting != nil {
			t.Error("ClearAwaiting should clear Awaiting field")
		}
	})

	t.Run("ClearAwaiting_only_manual_set", func(t *testing.T) {
		// ClearAwaiting when only Manual is set
		tick := baseTick()
		tick.Manual = true
		tick.Awaiting = nil

		tick.ClearAwaiting()

		if tick.Manual {
			t.Error("ClearAwaiting should clear Manual field")
		}
		if tick.Awaiting != nil {
			t.Error("ClearAwaiting should leave Awaiting nil")
		}
	})

	t.Run("ClearAwaiting_only_awaiting_set", func(t *testing.T) {
		// ClearAwaiting when only Awaiting is set
		tick := baseTick()
		tick.Manual = false
		awaiting := AwaitingApproval
		tick.Awaiting = &awaiting

		tick.ClearAwaiting()

		if tick.Manual {
			t.Error("ClearAwaiting should leave Manual false")
		}
		if tick.Awaiting != nil {
			t.Error("ClearAwaiting should clear Awaiting field")
		}
	})
}

// TestBackwardsCompatibility_RoundTrip tests that loading a tick with Manual=true,
// updating other fields, and saving preserves or migrates correctly.
func TestBackwardsCompatibility_RoundTrip(t *testing.T) {
	t.Run("unmarshal_Manual_true_preserved", func(t *testing.T) {
		// Load Manual=true tick, Manual field should be preserved
		jsonStr := `{
			"id": "xyz",
			"title": "Manual tick",
			"status": "open",
			"priority": 1,
			"type": "task",
			"owner": "agent",
			"created_by": "human",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z",
			"manual": true
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonStr), &tick); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if !tick.Manual {
			t.Error("Manual field should be preserved when unmarshaling")
		}
		if tick.Awaiting != nil {
			t.Errorf("Awaiting should be nil, got %v", *tick.Awaiting)
		}
	})

	t.Run("load_update_other_field_save_preserves_Manual", func(t *testing.T) {
		// Load Manual=true tick, update other field, save → Manual preserved
		jsonStr := `{
			"id": "xyz",
			"title": "Manual tick",
			"status": "open",
			"priority": 1,
			"type": "task",
			"owner": "agent",
			"created_by": "human",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z",
			"manual": true
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonStr), &tick); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		// Update other field
		tick.Title = "Updated Manual tick"
		tick.Priority = 2

		// Marshal back
		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// Unmarshal to verify
		var reloaded Tick
		if err := json.Unmarshal(data, &reloaded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if !reloaded.Manual {
			t.Error("Manual field should be preserved after round-trip with other field updates")
		}
		if reloaded.Title != "Updated Manual tick" {
			t.Errorf("Title should be updated, got %q", reloaded.Title)
		}
	})

	t.Run("load_Manual_true_set_awaiting_save_clears_Manual", func(t *testing.T) {
		// Load Manual=true tick, set awaiting, save → Manual cleared
		jsonStr := `{
			"id": "xyz",
			"title": "Manual tick",
			"status": "open",
			"priority": 1,
			"type": "task",
			"owner": "agent",
			"created_by": "human",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z",
			"manual": true
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonStr), &tick); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		// Set awaiting using SetAwaiting helper (should clear Manual)
		tick.SetAwaiting(AwaitingApproval)

		// Marshal back
		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// Unmarshal to verify
		var reloaded Tick
		if err := json.Unmarshal(data, &reloaded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if reloaded.Manual {
			t.Error("Manual field should be cleared after SetAwaiting")
		}
		if reloaded.Awaiting == nil || *reloaded.Awaiting != AwaitingApproval {
			t.Errorf("Awaiting should be %q, got %v", AwaitingApproval, reloaded.Awaiting)
		}
	})

	t.Run("marshal_Manual_true_includes_field", func(t *testing.T) {
		// When Manual=true, it should be included in JSON
		now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
		tick := Tick{
			ID:        "xyz",
			Title:     "Manual tick",
			Status:    StatusOpen,
			Priority:  1,
			Type:      TypeTask,
			Owner:     "agent",
			CreatedBy: "human",
			CreatedAt: now,
			UpdatedAt: now,
			Manual:    true,
		}

		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// Parse as raw JSON to check field presence
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		manual, ok := raw["manual"]
		if !ok {
			t.Error("manual field should be present in JSON when Manual=true")
		}
		if manual != true {
			t.Errorf("manual field should be true, got %v", manual)
		}
	})

	t.Run("marshal_Manual_false_omits_field", func(t *testing.T) {
		// When Manual=false, it should be omitted from JSON (omitempty)
		now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
		tick := Tick{
			ID:        "xyz",
			Title:     "Non-manual tick",
			Status:    StatusOpen,
			Priority:  1,
			Type:      TypeTask,
			Owner:     "agent",
			CreatedBy: "human",
			CreatedAt: now,
			UpdatedAt: now,
			Manual:    false,
		}

		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// Parse as raw JSON to check field presence
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if _, ok := raw["manual"]; ok {
			t.Error("manual field should be omitted from JSON when Manual=false")
		}
	})
}

// TestBackwardsCompatibility_IsTerminalAwaiting tests that IsTerminalAwaiting
// correctly handles the Manual flag for backwards compatibility.
func TestBackwardsCompatibility_IsTerminalAwaiting(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	baseTick := func() Tick {
		return Tick{
			ID:        "a1b",
			Title:     "Test tick",
			Status:    StatusOpen,
			Priority:  2,
			Type:      TypeTask,
			Owner:     "petere",
			CreatedBy: "petere",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	t.Run("Manual_true_is_terminal", func(t *testing.T) {
		// Manual=true maps to awaiting=work, which is terminal
		tick := baseTick()
		tick.Manual = true

		if !tick.IsTerminalAwaiting() {
			t.Error("IsTerminalAwaiting should return true for Manual=true (maps to work)")
		}
	})

	t.Run("Awaiting_work_is_terminal", func(t *testing.T) {
		tick := baseTick()
		awaiting := AwaitingWork
		tick.Awaiting = &awaiting

		if !tick.IsTerminalAwaiting() {
			t.Error("IsTerminalAwaiting should return true for awaiting=work")
		}
	})

	t.Run("Manual_true_Awaiting_input_not_terminal", func(t *testing.T) {
		// When both set, Awaiting takes precedence - input is not terminal
		tick := baseTick()
		tick.Manual = true
		awaiting := AwaitingInput
		tick.Awaiting = &awaiting

		if tick.IsTerminalAwaiting() {
			t.Error("IsTerminalAwaiting should return false when Awaiting=input (Awaiting takes precedence over Manual)")
		}
	})
}

// TestBackwardsCompatibility_JSONSerialization tests JSON serialization with
// various combinations of Manual and Awaiting fields.
func TestBackwardsCompatibility_JSONSerialization(t *testing.T) {
	now := time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)

	t.Run("both_Manual_and_Awaiting_roundtrip", func(t *testing.T) {
		// Test that both fields can coexist in JSON
		awaiting := AwaitingApproval
		tick := Tick{
			ID:        "xyz",
			Title:     "Both fields",
			Status:    StatusOpen,
			Priority:  1,
			Type:      TypeTask,
			Owner:     "agent",
			CreatedBy: "human",
			CreatedAt: now,
			UpdatedAt: now,
			Manual:    true,
			Awaiting:  &awaiting,
		}

		data, err := json.Marshal(tick)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		var reloaded Tick
		if err := json.Unmarshal(data, &reloaded); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if !reloaded.Manual {
			t.Error("Manual should be preserved")
		}
		if reloaded.Awaiting == nil || *reloaded.Awaiting != AwaitingApproval {
			t.Errorf("Awaiting should be preserved, got %v", reloaded.Awaiting)
		}

		// GetAwaitingType should return Awaiting value (takes precedence)
		if got := reloaded.GetAwaitingType(); got != AwaitingApproval {
			t.Errorf("GetAwaitingType should return %q, got %q", AwaitingApproval, got)
		}
	})

	t.Run("unmarshal_legacy_JSON_without_awaiting", func(t *testing.T) {
		// Test that old JSON without awaiting field works
		jsonStr := `{
			"id": "xyz",
			"title": "Legacy tick",
			"status": "open",
			"priority": 1,
			"type": "task",
			"owner": "agent",
			"created_by": "human",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z",
			"manual": true
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonStr), &tick); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if !tick.Manual {
			t.Error("Manual should be true")
		}
		if tick.Awaiting != nil {
			t.Error("Awaiting should be nil for legacy JSON")
		}

		// Helper methods should work correctly
		if !tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman should return true for Manual=true")
		}
		if got := tick.GetAwaitingType(); got != AwaitingWork {
			t.Errorf("GetAwaitingType should return %q for Manual=true, got %q", AwaitingWork, got)
		}
	})

	t.Run("unmarshal_new_JSON_with_awaiting", func(t *testing.T) {
		// Test that new JSON with awaiting field works
		jsonStr := `{
			"id": "xyz",
			"title": "New tick",
			"status": "open",
			"priority": 1,
			"type": "task",
			"owner": "agent",
			"created_by": "human",
			"created_at": "2025-01-08T10:30:00Z",
			"updated_at": "2025-01-08T10:30:00Z",
			"awaiting": "approval"
		}`

		var tick Tick
		if err := json.Unmarshal([]byte(jsonStr), &tick); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if tick.Manual {
			t.Error("Manual should be false for new JSON without manual field")
		}
		if tick.Awaiting == nil || *tick.Awaiting != AwaitingApproval {
			t.Errorf("Awaiting should be %q", AwaitingApproval)
		}

		// Helper methods should work correctly
		if !tick.IsAwaitingHuman() {
			t.Error("IsAwaitingHuman should return true for Awaiting set")
		}
		if got := tick.GetAwaitingType(); got != AwaitingApproval {
			t.Errorf("GetAwaitingType should return %q, got %q", AwaitingApproval, got)
		}
	})
}
