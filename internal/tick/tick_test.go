package tick

import (
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

	// Test invalid requires
	invalidRequires := "invalid_gate"
	badRequires := base
	badRequires.Requires = &invalidRequires
	if err := badRequires.Validate(); err == nil || !strings.Contains(err.Error(), "invalid requires") {
		t.Fatalf("expected invalid requires error, got %v", err)
	}

	// Test invalid awaiting
	invalidAwaiting := "invalid_state"
	badAwaiting := base
	badAwaiting.Awaiting = &invalidAwaiting
	if err := badAwaiting.Validate(); err == nil || !strings.Contains(err.Error(), "invalid awaiting") {
		t.Fatalf("expected invalid awaiting error, got %v", err)
	}

	// Test invalid verdict
	invalidVerdict := "invalid_verdict"
	badVerdict := base
	badVerdict.Verdict = &invalidVerdict
	if err := badVerdict.Validate(); err == nil || !strings.Contains(err.Error(), "invalid verdict") {
		t.Fatalf("expected invalid verdict error, got %v", err)
	}
}

func TestTickValidateRequires(t *testing.T) {
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

	// nil requires should be valid
	if err := base.Validate(); err != nil {
		t.Fatalf("nil requires should be valid, got error: %v", err)
	}

	// valid requires values
	validRequires := []string{RequiresApproval, RequiresReview, RequiresContent}
	for _, r := range validRequires {
		t.Run(r, func(t *testing.T) {
			tick := base
			req := r
			tick.Requires = &req
			if err := tick.Validate(); err != nil {
				t.Fatalf("requires=%q should be valid, got error: %v", r, err)
			}
		})
	}
}

func TestTickValidateAwaiting(t *testing.T) {
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

	// nil awaiting should be valid
	if err := base.Validate(); err != nil {
		t.Fatalf("nil awaiting should be valid, got error: %v", err)
	}

	// valid awaiting values
	validAwaiting := []string{AwaitingWork, AwaitingApproval, AwaitingInput, AwaitingReview, AwaitingContent, AwaitingEscalation, AwaitingCheckpoint}
	for _, a := range validAwaiting {
		t.Run(a, func(t *testing.T) {
			tick := base
			aw := a
			tick.Awaiting = &aw
			if err := tick.Validate(); err != nil {
				t.Fatalf("awaiting=%q should be valid, got error: %v", a, err)
			}
		})
	}
}

func TestTickValidateVerdict(t *testing.T) {
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

	// nil verdict should be valid
	if err := base.Validate(); err != nil {
		t.Fatalf("nil verdict should be valid, got error: %v", err)
	}

	// valid verdict values
	validVerdicts := []string{VerdictApproved, VerdictRejected}
	for _, v := range validVerdicts {
		t.Run(v, func(t *testing.T) {
			tick := base
			vd := v
			tick.Verdict = &vd
			if err := tick.Validate(); err != nil {
				t.Fatalf("verdict=%q should be valid, got error: %v", v, err)
			}
		})
	}
}

func TestValidValueSlices(t *testing.T) {
	// Test ValidRequiresValues contains all valid requires values
	expectedRequires := []string{RequiresApproval, RequiresReview, RequiresContent}
	if len(ValidRequiresValues) != len(expectedRequires) {
		t.Errorf("ValidRequiresValues has %d elements, expected %d", len(ValidRequiresValues), len(expectedRequires))
	}
	for i, v := range expectedRequires {
		if ValidRequiresValues[i] != v {
			t.Errorf("ValidRequiresValues[%d] = %q, expected %q", i, ValidRequiresValues[i], v)
		}
	}

	// Test ValidAwaitingValues contains all valid awaiting values
	expectedAwaiting := []string{AwaitingWork, AwaitingApproval, AwaitingInput, AwaitingReview, AwaitingContent, AwaitingEscalation, AwaitingCheckpoint}
	if len(ValidAwaitingValues) != len(expectedAwaiting) {
		t.Errorf("ValidAwaitingValues has %d elements, expected %d", len(ValidAwaitingValues), len(expectedAwaiting))
	}
	for i, v := range expectedAwaiting {
		if ValidAwaitingValues[i] != v {
			t.Errorf("ValidAwaitingValues[%d] = %q, expected %q", i, ValidAwaitingValues[i], v)
		}
	}

	// Test ValidVerdictValues contains all valid verdict values
	expectedVerdict := []string{VerdictApproved, VerdictRejected}
	if len(ValidVerdictValues) != len(expectedVerdict) {
		t.Errorf("ValidVerdictValues has %d elements, expected %d", len(ValidVerdictValues), len(expectedVerdict))
	}
	for i, v := range expectedVerdict {
		if ValidVerdictValues[i] != v {
			t.Errorf("ValidVerdictValues[%d] = %q, expected %q", i, ValidVerdictValues[i], v)
		}
	}
}

func TestIsAwaitingHuman(t *testing.T) {
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

	// No awaiting or manual - not waiting
	if base.IsAwaitingHuman() {
		t.Error("expected IsAwaitingHuman() to be false when neither Awaiting nor Manual is set")
	}

	// With Awaiting set
	awaiting := AwaitingApproval
	withAwaiting := base
	withAwaiting.Awaiting = &awaiting
	if !withAwaiting.IsAwaitingHuman() {
		t.Error("expected IsAwaitingHuman() to be true when Awaiting is set")
	}

	// With Manual set (backwards compat)
	withManual := base
	withManual.Manual = true
	if !withManual.IsAwaitingHuman() {
		t.Error("expected IsAwaitingHuman() to be true when Manual is set")
	}

	// With both set
	withBoth := base
	withBoth.Awaiting = &awaiting
	withBoth.Manual = true
	if !withBoth.IsAwaitingHuman() {
		t.Error("expected IsAwaitingHuman() to be true when both Awaiting and Manual are set")
	}
}

func TestGetAwaitingType(t *testing.T) {
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

	// No awaiting or manual - empty string
	if got := base.GetAwaitingType(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}

	// With Awaiting set - returns awaiting value
	awaiting := AwaitingApproval
	withAwaiting := base
	withAwaiting.Awaiting = &awaiting
	if got := withAwaiting.GetAwaitingType(); got != AwaitingApproval {
		t.Errorf("expected %q, got %q", AwaitingApproval, got)
	}

	// With Manual set (backwards compat) - returns work
	withManual := base
	withManual.Manual = true
	if got := withManual.GetAwaitingType(); got != AwaitingWork {
		t.Errorf("expected %q for Manual=true, got %q", AwaitingWork, got)
	}

	// With both set - Awaiting takes precedence
	withBoth := base
	input := AwaitingInput
	withBoth.Awaiting = &input
	withBoth.Manual = true
	if got := withBoth.GetAwaitingType(); got != AwaitingInput {
		t.Errorf("expected %q (Awaiting takes precedence), got %q", AwaitingInput, got)
	}
}

func TestHasRequiredGate(t *testing.T) {
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

	// No requires - false
	if base.HasRequiredGate() {
		t.Error("expected HasRequiredGate() to be false when Requires is nil")
	}

	// With requires set - true
	requires := RequiresApproval
	withRequires := base
	withRequires.Requires = &requires
	if !withRequires.HasRequiredGate() {
		t.Error("expected HasRequiredGate() to be true when Requires is set")
	}
}

func TestIsTerminalAwaiting(t *testing.T) {
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

	// Terminal awaiting types
	terminalTypes := []string{AwaitingApproval, AwaitingReview, AwaitingContent, AwaitingWork}
	for _, awType := range terminalTypes {
		t.Run("terminal_"+awType, func(t *testing.T) {
			tick := base
			a := awType
			tick.Awaiting = &a
			if !tick.IsTerminalAwaiting() {
				t.Errorf("expected IsTerminalAwaiting() to be true for %q", awType)
			}
		})
	}

	// Non-terminal awaiting types
	nonTerminalTypes := []string{AwaitingInput, AwaitingEscalation, AwaitingCheckpoint}
	for _, awType := range nonTerminalTypes {
		t.Run("non_terminal_"+awType, func(t *testing.T) {
			tick := base
			a := awType
			tick.Awaiting = &a
			if tick.IsTerminalAwaiting() {
				t.Errorf("expected IsTerminalAwaiting() to be false for %q", awType)
			}
		})
	}

	// Manual flag (backwards compat) should be terminal since it maps to work
	withManual := base
	withManual.Manual = true
	if !withManual.IsTerminalAwaiting() {
		t.Error("expected IsTerminalAwaiting() to be true for Manual=true (maps to work)")
	}

	// No awaiting - not terminal
	if base.IsTerminalAwaiting() {
		t.Error("expected IsTerminalAwaiting() to be false when not awaiting")
	}
}
