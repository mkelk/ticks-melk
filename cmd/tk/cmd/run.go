package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/agent"
	"github.com/pengelbrecht/ticks/internal/budget"
	"github.com/pengelbrecht/ticks/internal/checkpoint"
	"github.com/pengelbrecht/ticks/internal/engine"
	"github.com/pengelbrecht/ticks/internal/gc"
	"github.com/pengelbrecht/ticks/internal/runrecord"
	"github.com/pengelbrecht/ticks/internal/ticks"
)

var runCmd = &cobra.Command{
	Use:   "run [epic-id...]",
	Short: "Run AI agent on epics",
	Long: `Run AI agent on one or more epics until tasks are complete.

If no epic-id is specified, use --auto to auto-select the next ready epic.

Examples:
  tk run abc123                     # Run agent on epic abc123
  tk run abc123 def456              # Run agent on multiple epics (sequential)
  tk run --auto                     # Auto-select next ready epic
  tk run abc123 --max-iterations 10 # Limit to 10 iterations per task
  tk run abc123 --max-cost 5.00     # Stop if cost exceeds $5.00
  tk run abc123 --worktree          # Run in isolated git worktree
  tk run abc123 --watch             # Watch mode - restart when tasks ready
  tk run abc123 --jsonl             # Output JSONL format for parsing`,
	RunE: runRun,
}

var (
	runMaxIterations     int
	runMaxCost           float64
	runCheckpointEvery   int
	runMaxTaskRetries    int
	runAuto              bool
	runJSONL             bool
	runSkipVerify        bool
	runVerifyOnly        bool
	runWorktree          bool
	runParallel          int
	runWatch             bool
	runTimeout           time.Duration
	runPoll              time.Duration
	runDebounce          time.Duration
	runIncludeStandalone bool
	runIncludeOrphans    bool
	runAll               bool
)

func init() {
	runCmd.Flags().IntVar(&runMaxIterations, "max-iterations", 50, "maximum iterations per task")
	runCmd.Flags().Float64Var(&runMaxCost, "max-cost", 0, "maximum cost in USD (0=unlimited)")
	runCmd.Flags().IntVar(&runCheckpointEvery, "checkpoint-interval", 5, "checkpoint every N iterations")
	runCmd.Flags().IntVar(&runMaxTaskRetries, "max-task-retries", 3, "max retries for failed tasks")
	runCmd.Flags().BoolVar(&runAuto, "auto", false, "auto-select next ready epic if none specified")
	runCmd.Flags().BoolVar(&runJSONL, "jsonl", false, "output JSONL format for parsing")
	runCmd.Flags().BoolVar(&runSkipVerify, "skip-verify", false, "skip verification after task completion")
	runCmd.Flags().BoolVar(&runVerifyOnly, "verify-only", false, "only run verification, no agent")
	runCmd.Flags().BoolVar(&runWorktree, "worktree", false, "use git worktree for parallel runs")
	runCmd.Flags().IntVar(&runParallel, "parallel", 1, "number of parallel tasks")
	runCmd.Flags().BoolVar(&runWatch, "watch", false, "watch mode - restart when tasks become ready")
	runCmd.Flags().DurationVar(&runTimeout, "timeout", 30*time.Minute, "task timeout duration")
	runCmd.Flags().DurationVar(&runPoll, "poll", 10*time.Second, "poll interval for watch mode")
	runCmd.Flags().DurationVar(&runDebounce, "debounce", 0, "debounce interval for file changes")
	runCmd.Flags().BoolVar(&runIncludeStandalone, "include-standalone", false, "include tasks without parent epic")
	runCmd.Flags().BoolVar(&runIncludeOrphans, "include-orphans", false, "include orphaned tasks")
	runCmd.Flags().BoolVar(&runAll, "all", false, "run all ready tasks, not just first")

	rootCmd.AddCommand(runCmd)
}

// runOutput is the JSONL output format for run results.
type runOutput struct {
	EpicID         string   `json:"epic_id"`
	Iterations     int      `json:"iterations"`
	TotalTokens    int      `json:"total_tokens"`
	TotalCost      float64  `json:"total_cost"`
	DurationSec    float64  `json:"duration_sec"`
	CompletedTasks []string `json:"completed_tasks"`
	ExitReason     string   `json:"exit_reason"`
	Signal         string   `json:"signal,omitempty"`
	SignalReason   string   `json:"signal_reason,omitempty"`
}

func runRun(cmd *cobra.Command, args []string) error {
	// Start async garbage collection
	go func() {
		root, err := repoRoot()
		if err != nil {
			return
		}
		_, _ = gc.Cleanup(root, gc.DefaultMaxAge)
	}()

	root, err := repoRoot()
	if err != nil {
		return NewExitError(ExitNoRepo, "not in a git repository: %v", err)
	}

	// Determine epic IDs to run
	epicIDs := args
	if len(epicIDs) == 0 {
		if !runAuto {
			return NewExitError(ExitUsage, "specify epic-id(s) or use --auto")
		}
		// Auto-select next ready epic
		client := ticks.NewClient(filepath.Join(root, ".tick"))
		epic, err := client.NextReadyEpic()
		if err != nil {
			return NewExitError(ExitGeneric, "failed to find ready epic: %v", err)
		}
		if epic == nil {
			if runJSONL {
				// Output empty result
				output := runOutput{ExitReason: "no ready epics"}
				enc := json.NewEncoder(os.Stdout)
				_ = enc.Encode(output)
				return nil
			}
			fmt.Println("No ready epics")
			return nil
		}
		epicIDs = []string{epic.ID}
	}

	// Verify-only mode not implemented yet
	if runVerifyOnly {
		return NewExitError(ExitUsage, "--verify-only is not yet implemented")
	}

	// Parallel mode not implemented yet
	if runParallel > 1 {
		return NewExitError(ExitUsage, "--parallel > 1 is not yet implemented")
	}

	// Create the agent
	claudeAgent := agent.NewClaudeAgent()
	if !claudeAgent.Available() {
		return NewExitError(ExitGeneric, "claude CLI not found - install from https://claude.ai/code")
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		if !runJSONL {
			fmt.Fprintln(os.Stderr, "\nInterrupted - finishing current iteration...")
		}
		cancel()
	}()

	// Run each epic sequentially
	for _, epicID := range epicIDs {
		result, err := runEpic(ctx, root, epicID, claudeAgent)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled - output partial result if we have one
				if result != nil {
					outputResult(result)
				}
				return nil
			}
			return NewExitError(ExitGeneric, "run failed for epic %s: %v", epicID, err)
		}

		outputResult(result)

		// Stop if context cancelled
		if ctx.Err() != nil {
			break
		}
	}

	return nil
}

func runEpic(ctx context.Context, root, epicID string, agentImpl agent.Agent) (*engine.RunResult, error) {
	// Create dependencies
	ticksClient := ticks.NewClient(filepath.Join(root, ".tick"))
	budgetTracker := budget.NewTracker(budget.Limits{
		MaxIterations: runMaxIterations,
		MaxCost:       runMaxCost,
	})
	checkpointMgr := checkpoint.NewManager()

	// Create engine
	eng := engine.NewEngine(agentImpl, ticksClient, budgetTracker, checkpointMgr)

	// Enable live run record streaming for tickboard
	runRecordStore := runrecord.NewStore(root)
	eng.SetRunRecordStore(runRecordStore)

	// Enable verification unless skipped
	if !runSkipVerify {
		eng.EnableVerification()
	}

	// Set up output streaming for non-JSONL mode
	if !runJSONL {
		eng.OnOutput = func(chunk string) {
			fmt.Print(chunk)
		}
		eng.OnIterationStart = func(ctx engine.IterationContext) {
			fmt.Printf("\n=== Iteration %d: %s (%s) ===\n", ctx.Iteration, ctx.Task.ID, ctx.Task.Title)
		}
		eng.OnIterationEnd = func(result *engine.IterationResult) {
			fmt.Printf("\n--- Iteration %d complete (tokens: %d in, %d out, cost: $%.4f) ---\n",
				result.Iteration, result.TokensIn, result.TokensOut, result.Cost)
		}
	}

	// Build run config
	config := engine.RunConfig{
		EpicID:            epicID,
		MaxIterations:     runMaxIterations,
		MaxCost:           runMaxCost,
		CheckpointEvery:   runCheckpointEvery,
		MaxTaskRetries:    runMaxTaskRetries,
		AgentTimeout:      runTimeout,
		SkipVerify:        runSkipVerify,
		UseWorktree:       runWorktree,
		RepoRoot:          root,
		Watch:             runWatch,
		WatchPollInterval: runPoll,
		DebounceInterval:  runDebounce,
	}

	// Run the engine
	return eng.Run(ctx, config)
}

func outputResult(result *engine.RunResult) {
	if runJSONL {
		output := runOutput{
			EpicID:         result.EpicID,
			Iterations:     result.Iterations,
			TotalTokens:    result.TotalTokens,
			TotalCost:      result.TotalCost,
			DurationSec:    result.Duration.Seconds(),
			CompletedTasks: result.CompletedTasks,
			ExitReason:     result.ExitReason,
		}
		if result.Signal != engine.SignalNone {
			output.Signal = result.Signal.String()
			output.SignalReason = result.SignalReason
		}
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(output)
	} else {
		fmt.Printf("\n=== Run Complete ===\n")
		fmt.Printf("Epic: %s\n", result.EpicID)
		fmt.Printf("Iterations: %d\n", result.Iterations)
		fmt.Printf("Tokens: %d\n", result.TotalTokens)
		fmt.Printf("Cost: $%.4f\n", result.TotalCost)
		fmt.Printf("Duration: %v\n", result.Duration.Round(time.Second))
		fmt.Printf("Completed tasks: %d\n", len(result.CompletedTasks))
		fmt.Printf("Exit reason: %s\n", result.ExitReason)
		if result.Signal != engine.SignalNone {
			fmt.Printf("Signal: %s\n", result.Signal)
			if result.SignalReason != "" {
				fmt.Printf("Signal reason: %s\n", result.SignalReason)
			}
		}
	}
}
