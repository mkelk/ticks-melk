package taskrunner

import (
	"context"
	"testing"
	"time"

	"github.com/pengelbrecht/ticks/internal/agent"
)

// mockAgent implements agent.Agent for testing
type mockAgent struct {
	runResult *agent.Result
	runErr    error
}

func (m *mockAgent) Name() string {
	return "mock"
}

func (m *mockAgent) Available() bool {
	return true
}

func (m *mockAgent) Run(ctx context.Context, prompt string, opts agent.RunOpts) (*agent.Result, error) {
	// Call StateCallback if provided to test streaming
	if opts.StateCallback != nil {
		opts.StateCallback(agent.AgentStateSnapshot{
			SessionID: "test-session",
			Model:     "test-model",
			Output:    "test output",
			Status:    agent.StatusWriting,
			NumTurns:  1,
		})
	}
	return m.runResult, m.runErr
}

func TestRunner_Run(t *testing.T) {
	mock := &mockAgent{
		runResult: &agent.Result{
			Output:    "test output",
			TokensIn:  100,
			TokensOut: 50,
			Cost:      0.01,
		},
	}

	runner := New(Config{
		Agent:   mock,
		Timeout: 5 * time.Minute,
	})

	result := runner.Run(context.Background(), "test-task-id", "test prompt")

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Output != "test output" {
		t.Errorf("expected output 'test output', got %q", result.Output)
	}
	if result.TokensIn != 100 {
		t.Errorf("expected TokensIn 100, got %d", result.TokensIn)
	}
	if result.TokensOut != 50 {
		t.Errorf("expected TokensOut 50, got %d", result.TokensOut)
	}
	if result.Cost != 0.01 {
		t.Errorf("expected Cost 0.01, got %f", result.Cost)
	}
}

func TestRunner_RunSimple(t *testing.T) {
	mock := &mockAgent{
		runResult: &agent.Result{
			Output:    "simple output",
			TokensIn:  10,
			TokensOut: 5,
			Cost:      0.001,
		},
	}

	runner := New(Config{
		Agent:   mock,
		Timeout: 5 * time.Minute,
	})

	result, err := runner.RunSimple(context.Background(), "test prompt")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "simple output" {
		t.Errorf("expected output 'simple output', got %q", result.Output)
	}
}

func TestRunner_StateCallback(t *testing.T) {
	mock := &mockAgent{
		runResult: &agent.Result{
			Output: "output",
		},
	}

	var callbackCalled bool
	var receivedSnap agent.AgentStateSnapshot

	runner := New(Config{
		Agent:   mock,
		Timeout: 5 * time.Minute,
		OnAgentState: func(snap agent.AgentStateSnapshot) {
			callbackCalled = true
			receivedSnap = snap
		},
	})

	result := runner.Run(context.Background(), "test-task-id", "test prompt")

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if !callbackCalled {
		t.Error("expected state callback to be called")
	}
	if receivedSnap.SessionID != "test-session" {
		t.Errorf("expected session ID 'test-session', got %q", receivedSnap.SessionID)
	}
}
