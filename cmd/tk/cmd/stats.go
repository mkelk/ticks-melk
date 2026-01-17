package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/github"
	"github.com/pengelbrecht/ticks/internal/query"
	"github.com/pengelbrecht/ticks/internal/tick"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show repository statistics",
	Long: `Show repository statistics.

Displays summary statistics about ticks in the repository including
counts by status, priority, and type, as well as ready and blocked counts.

Examples:
  # Show stats for current user
  tk stats

  # Show stats for all owners
  tk stats --all

  # Output as JSON
  tk stats --json`,
	Args: cobra.NoArgs,
	RunE: runStats,
}

var (
	statsAll  bool
	statsJSON bool
)

func init() {
	statsCmd.Flags().BoolVarP(&statsAll, "all", "a", false, "all owners")
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "output as JSON")

	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repo root: %w", err)
	}

	owner, err := resolveOwner(statsAll, "")
	if err != nil {
		return fmt.Errorf("failed to detect owner: %w", err)
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	ticks, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list ticks: %w", err)
	}

	filtered := query.Apply(ticks, query.Filter{Owner: owner})

	statusCounts := make(map[string]int)
	priorityCounts := make(map[int]int)
	typeCounts := make(map[string]int)

	for _, t := range filtered {
		statusCounts[t.Status]++
		priorityCounts[t.Priority]++
		typeCounts[t.Type]++
	}

	ready := query.Ready(filtered, ticks)
	blocked := query.Blocked(filtered, ticks)

	if statsJSON {
		payload := map[string]any{
			"total":    len(filtered),
			"status":   statusCounts,
			"priority": priorityCounts,
			"type":     typeCounts,
			"ready":    len(ready),
			"blocked":  len(blocked),
		}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(payload); err != nil {
			return fmt.Errorf("failed to encode json: %w", err)
		}
		return nil
	}

	project, err := github.DetectProject(nil)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}
	fmt.Println(project)
	fmt.Printf("\n  Total: %d ticks\n", len(filtered))
	fmt.Printf("  Status: %s\n", formatStatusCounts(statusCounts))
	fmt.Printf("  Priority: %s\n", formatPriorityCounts(priorityCounts))
	fmt.Printf("  Types: %s\n", formatTypeCounts(typeCounts))
	fmt.Printf("\n  Ready: %d\n", len(ready))
	fmt.Printf("  Blocked: %d\n", len(blocked))
	return nil
}

func formatStatusCounts(counts map[string]int) string {
	return fmt.Sprintf("open %d · in progress %d · closed %d",
		counts[tick.StatusOpen], counts[tick.StatusInProgress], counts[tick.StatusClosed])
}

func formatPriorityCounts(counts map[int]int) string {
	return fmt.Sprintf("P0:%d · P1:%d · P2:%d · P3:%d · P4:%d",
		counts[0], counts[1], counts[2], counts[3], counts[4])
}

func formatTypeCounts(counts map[string]int) string {
	return fmt.Sprintf("bug:%d · feature:%d · task:%d · epic:%d · chore:%d",
		counts[tick.TypeBug], counts[tick.TypeFeature], counts[tick.TypeTask], counts[tick.TypeEpic], counts[tick.TypeChore])
}
