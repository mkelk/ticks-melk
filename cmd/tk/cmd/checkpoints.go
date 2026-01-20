package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/checkpoint"
)

var checkpointsCmd = &cobra.Command{
	Use:   "checkpoints [epic-id]",
	Short: "List available checkpoints",
	Long: `List available checkpoints for resuming agent runs.

If an epic-id is provided, only checkpoints for that epic are shown.
Otherwise, all checkpoints are listed.

Checkpoints are sorted by timestamp (newest first).

Examples:
  tk checkpoints                    # List all checkpoints
  tk checkpoints abc123             # List checkpoints for epic abc123
  tk checkpoints --json             # Output as JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheckpoints,
}

var (
	checkpointsJSON bool
)

func init() {
	checkpointsCmd.Flags().BoolVar(&checkpointsJSON, "json", false, "output as JSON")

	rootCmd.AddCommand(checkpointsCmd)
}

// checkpointsOutput wraps the output for JSON formatting.
type checkpointsOutput struct {
	Checkpoints []checkpointInfo `json:"checkpoints"`
	EpicFilter  string           `json:"epic_filter,omitempty"`
}

// checkpointInfo is a subset of checkpoint fields for display.
type checkpointInfo struct {
	ID             string   `json:"id"`
	EpicID         string   `json:"epic_id"`
	Iteration      int      `json:"iteration"`
	Timestamp      string   `json:"timestamp"`
	TotalTokens    int      `json:"total_tokens"`
	TotalCost      float64  `json:"total_cost"`
	CompletedTasks int      `json:"completed_tasks"`
	TaskIDs        []string `json:"task_ids,omitempty"`
	GitCommit      string   `json:"git_commit,omitempty"`
	Worktree       bool     `json:"worktree"`
}

func runCheckpoints(cmd *cobra.Command, args []string) error {
	mgr := checkpoint.NewManager()

	var checkpoints []checkpoint.Checkpoint
	var epicFilter string
	var err error

	if len(args) > 0 {
		epicFilter = args[0]
		checkpoints, err = mgr.ListForEpic(epicFilter)
	} else {
		checkpoints, err = mgr.List()
	}

	if err != nil {
		return NewExitError(ExitGeneric, "failed to list checkpoints: %v", err)
	}

	if len(checkpoints) == 0 {
		if checkpointsJSON {
			output := checkpointsOutput{
				Checkpoints: []checkpointInfo{},
				EpicFilter:  epicFilter,
			}
			enc := json.NewEncoder(os.Stdout)
			return enc.Encode(output)
		}
		if epicFilter != "" {
			fmt.Printf("No checkpoints for epic %s\n", epicFilter)
		} else {
			fmt.Println("No checkpoints")
		}
		return nil
	}

	// Convert to display format
	infos := make([]checkpointInfo, len(checkpoints))
	for i, cp := range checkpoints {
		infos[i] = checkpointInfo{
			ID:             cp.ID,
			EpicID:         cp.EpicID,
			Iteration:      cp.Iteration,
			Timestamp:      cp.Timestamp.Format("2006-01-02 15:04:05"),
			TotalTokens:    cp.TotalTokens,
			TotalCost:      cp.TotalCost,
			CompletedTasks: len(cp.CompletedTasks),
			TaskIDs:        cp.CompletedTasks,
			GitCommit:      cp.GitCommit,
			Worktree:       cp.WorktreePath != "",
		}
	}

	if checkpointsJSON {
		output := checkpointsOutput{
			Checkpoints: infos,
			EpicFilter:  epicFilter,
		}
		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(output)
	}

	// Table output
	fmt.Println(" ID                 EPIC    ITER  TASKS  COST      TIMESTAMP            WORKTREE")
	for _, info := range infos {
		worktree := ""
		if info.Worktree {
			worktree = "yes"
		}
		fmt.Printf(" %-18s %-6s %4d  %5d  $%-7.4f %s  %s\n",
			info.ID,
			info.EpicID,
			info.Iteration,
			info.CompletedTasks,
			info.TotalCost,
			info.Timestamp,
			worktree,
		)
	}
	fmt.Printf("\n%d checkpoint(s)\n", len(infos))

	return nil
}
