package tick

import (
	"testing"
	"time"
)

func TestProcessVerdict(t *testing.T) {
	tests := []struct {
		name         string
		awaiting     *string
		verdict      *string
		wantClosed   bool
		wantStatus   string
		wantAwaiting *string
		wantVerdict  *string
	}{
		// No-op cases
		{
			name:         "nil verdict and nil awaiting",
			awaiting:     nil,
			verdict:      nil,
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "nil verdict with awaiting",
			awaiting:     ptr(AwaitingApproval),
			verdict:      nil,
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: ptr(AwaitingApproval), // Not cleared when no verdict
			wantVerdict:  nil,
		},
		{
			name:         "verdict with nil awaiting",
			awaiting:     nil,
			verdict:      ptr(VerdictApproved),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  ptr(VerdictApproved), // Not cleared when no awaiting
		},

		// Terminal states - close on approved
		{
			name:         "work approved closes",
			awaiting:     ptr(AwaitingWork),
			verdict:      ptr(VerdictApproved),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "work rejected does not close",
			awaiting:     ptr(AwaitingWork),
			verdict:      ptr(VerdictRejected),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "approval approved closes",
			awaiting:     ptr(AwaitingApproval),
			verdict:      ptr(VerdictApproved),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "approval rejected does not close",
			awaiting:     ptr(AwaitingApproval),
			verdict:      ptr(VerdictRejected),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "review approved closes",
			awaiting:     ptr(AwaitingReview),
			verdict:      ptr(VerdictApproved),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "review rejected does not close",
			awaiting:     ptr(AwaitingReview),
			verdict:      ptr(VerdictRejected),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "content approved closes",
			awaiting:     ptr(AwaitingContent),
			verdict:      ptr(VerdictApproved),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "content rejected does not close",
			awaiting:     ptr(AwaitingContent),
			verdict:      ptr(VerdictRejected),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},

		// Non-terminal states - close on rejected
		{
			name:         "input approved does not close",
			awaiting:     ptr(AwaitingInput),
			verdict:      ptr(VerdictApproved),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "input rejected closes",
			awaiting:     ptr(AwaitingInput),
			verdict:      ptr(VerdictRejected),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "escalation approved does not close",
			awaiting:     ptr(AwaitingEscalation),
			verdict:      ptr(VerdictApproved),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "escalation rejected closes",
			awaiting:     ptr(AwaitingEscalation),
			verdict:      ptr(VerdictRejected),
			wantClosed:   true,
			wantStatus:   StatusClosed,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},

		// Checkpoint - never closes
		{
			name:         "checkpoint approved does not close",
			awaiting:     ptr(AwaitingCheckpoint),
			verdict:      ptr(VerdictApproved),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
		{
			name:         "checkpoint rejected does not close",
			awaiting:     ptr(AwaitingCheckpoint),
			verdict:      ptr(VerdictRejected),
			wantClosed:   false,
			wantStatus:   StatusInProgress,
			wantAwaiting: nil,
			wantVerdict:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tick := &Tick{
				ID:        "test",
				Title:     "Test Tick",
				Status:    StatusInProgress,
				Priority:  2,
				Type:      TypeTask,
				Owner:     "agent",
				CreatedBy: "agent",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Awaiting:  tt.awaiting,
				Verdict:   tt.verdict,
			}

			closed, err := ProcessVerdict(tick)
			if err != nil {
				t.Fatalf("ProcessVerdict() error = %v", err)
			}

			if closed != tt.wantClosed {
				t.Errorf("ProcessVerdict() closed = %v, want %v", closed, tt.wantClosed)
			}

			if tick.Status != tt.wantStatus {
				t.Errorf("tick.Status = %v, want %v", tick.Status, tt.wantStatus)
			}

			if !ptrEqual(tick.Awaiting, tt.wantAwaiting) {
				t.Errorf("tick.Awaiting = %v, want %v", ptrStr(tick.Awaiting), ptrStr(tt.wantAwaiting))
			}

			if !ptrEqual(tick.Verdict, tt.wantVerdict) {
				t.Errorf("tick.Verdict = %v, want %v", ptrStr(tick.Verdict), ptrStr(tt.wantVerdict))
			}

			// If closed, ClosedAt should be set
			if closed && tick.ClosedAt == nil {
				t.Error("tick.ClosedAt should be set when closed")
			}
			if !closed && tick.ClosedAt != nil {
				t.Error("tick.ClosedAt should not be set when not closed")
			}
		})
	}
}

func TestProcessVerdictDoesNotReturnError(t *testing.T) {
	// ProcessVerdict currently never returns an error, but the signature
	// allows for future validation or other error conditions
	tick := &Tick{
		ID:        "test",
		Title:     "Test",
		Status:    StatusInProgress,
		Priority:  2,
		Type:      TypeTask,
		Owner:     "agent",
		CreatedBy: "agent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Awaiting:  ptr(AwaitingApproval),
		Verdict:   ptr(VerdictApproved),
	}

	_, err := ProcessVerdict(tick)
	if err != nil {
		t.Errorf("ProcessVerdict() should not return error, got: %v", err)
	}
}

// Helper functions
func ptr(s string) *string {
	return &s
}

func ptrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
