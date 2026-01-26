// Package taskrunner provides a shared abstraction for running Claude agent tasks.
// It handles run records, live streaming, and consistent agent invocation options
// so that both the engine (ralph mode) and pool mode share the same behavior.
package taskrunner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pengelbrecht/ticks/internal/agent"
	"github.com/pengelbrecht/ticks/internal/runrecord"
	"github.com/pengelbrecht/ticks/internal/ticks"
)

// Runner provides a consistent interface for running Claude agent tasks.
// It wraps the agent with run records, live streaming, and state callbacks.
type Runner struct {
	agent       agent.Agent
	tickClient  *ticks.Client
	recordStore *runrecord.Store

	// Optional callbacks
	onAgentState func(snap agent.AgentStateSnapshot)
	onOutput     func(chunk string)

	// Config
	workDir string
	timeout time.Duration
	debug   bool
}

// Config holds the configuration for creating a Runner.
type Config struct {
	// Required
	Agent      agent.Agent
	TickClient *ticks.Client

	// Optional
	RecordStore  *runrecord.Store
	OnAgentState func(snap agent.AgentStateSnapshot)
	OnOutput     func(chunk string)

	// Defaults
	WorkDir string
	Timeout time.Duration
	Debug   bool
}

// Result contains the outcome of running a task.
type Result struct {
	Output    string
	TokensIn  int
	TokensOut int
	Cost      float64
	Duration  time.Duration
	Error     error
	IsTimeout bool
}

// New creates a new Runner with the given configuration.
func New(cfg Config) *Runner {
	return &Runner{
		agent:        cfg.Agent,
		tickClient:   cfg.TickClient,
		recordStore:  cfg.RecordStore,
		onAgentState: cfg.OnAgentState,
		onOutput:     cfg.OnOutput,
		workDir:      cfg.WorkDir,
		timeout:      cfg.Timeout,
		debug:        cfg.Debug,
	}
}

// Run executes the agent for a given task with the provided prompt.
// It handles:
// - Live streaming via StateCallback (writes .live.json files)
// - Run record finalization (renames .live.json to .json)
// - Legacy stream channel forwarding (if OnOutput callback is set)
// - Persisting RunRecord to the tick store
func (r *Runner) Run(ctx context.Context, taskID string, prompt string) Result {
	result := Result{}
	startTime := time.Now()

	// Build agent options
	opts := agent.RunOpts{
		Timeout: r.timeout,
		WorkDir: r.workDir,
	}

	// Set up rich streaming callback with live file tracking
	if r.onAgentState != nil || r.recordStore != nil {
		opts.StateCallback = func(snap agent.AgentStateSnapshot) {
			// Call user-provided callback if set
			if r.onAgentState != nil {
				r.onAgentState(snap)
			}
			// Write to .live.json file for external watchers (e.g., ticks board)
			if r.recordStore != nil {
				if err := r.recordStore.WriteLive(taskID, snap); err != nil {
					if r.debug {
						fmt.Fprintf(os.Stderr, "[DEBUG] WriteLive error for %s: %v\n", taskID, err)
					}
				} else if r.debug {
					fmt.Fprintf(os.Stderr, "[DEBUG] WriteLive success for %s (output len=%d)\n", taskID, len(snap.Output))
				}
			}
		}
	}

	// Set up legacy streaming if callback is configured (backward compat)
	var streamChan chan string
	if r.onOutput != nil {
		streamChan = make(chan string, 100)
		opts.Stream = streamChan

		// Forward stream to callback
		go func() {
			for chunk := range streamChan {
				r.onOutput(chunk)
			}
		}()
	}

	// Run the agent
	agentResult, err := r.agent.Run(ctx, prompt, opts)

	// Finalize live record if store is configured
	// This renames .live.json to .json (or deletes on error)
	if r.recordStore != nil {
		_ = r.recordStore.FinalizeLive(taskID)
	}

	// Close stream channel
	if streamChan != nil {
		close(streamChan)
	}

	result.Duration = time.Since(startTime)

	// Handle timeout specially - capture partial output
	if err == agent.ErrTimeout {
		result.IsTimeout = true
		if agentResult != nil {
			result.Output = agentResult.Output
			result.TokensIn = agentResult.TokensIn
			result.TokensOut = agentResult.TokensOut
			result.Cost = agentResult.Cost
			if agentResult.Record != nil && r.tickClient != nil {
				_ = r.tickClient.SetRunRecord(taskID, agentResult.Record)
			}
		}
		return result
	}

	if err != nil {
		result.Error = fmt.Errorf("agent run: %w", err)
		return result
	}

	result.Output = agentResult.Output
	result.TokensIn = agentResult.TokensIn
	result.TokensOut = agentResult.TokensOut
	result.Cost = agentResult.Cost

	// Persist RunRecord to task (enables viewing historical run data)
	if agentResult.Record != nil && r.tickClient != nil {
		_ = r.tickClient.SetRunRecord(taskID, agentResult.Record)
	}

	return result
}

// RunSimple executes the agent without any run record tracking.
// This is useful for lightweight tasks where tracking overhead is not needed.
func (r *Runner) RunSimple(ctx context.Context, prompt string) (*agent.Result, error) {
	opts := agent.RunOpts{
		Timeout: r.timeout,
		WorkDir: r.workDir,
	}
	return r.agent.Run(ctx, prompt, opts)
}
