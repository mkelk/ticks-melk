package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/github"
	"github.com/pengelbrecht/ticks/internal/query"
	"github.com/pengelbrecht/ticks/internal/styles"
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

	// Build content lines
	var lines []string
	lines = append(lines, styles.HeaderStyle.Render(project))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s %d ticks", styles.RenderLabel("Total:"), len(filtered)))
	lines = append(lines, "")
	lines = append(lines, styles.RenderLabel("Status:")+"  "+formatStatusCounts(statusCounts))
	lines = append(lines, styles.RenderLabel("Priority:")+"  "+formatPriorityCounts(priorityCounts))
	lines = append(lines, styles.RenderLabel("Types:")+"  "+formatTypeCounts(typeCounts))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s %s",
		styles.RenderLabel("Ready:"),
		styles.StatusInProgressStyle.Render(fmt.Sprintf("%d", len(ready)))))
	lines = append(lines, fmt.Sprintf("%s %s",
		styles.RenderLabel("Blocked:"),
		styles.StatusBlockedStyle.Render(fmt.Sprintf("%d", len(blocked)))))

	// Render in box
	content := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorGray).
		Padding(0, 1).
		Render(content)

	fmt.Println(box)
	return nil
}

func formatStatusCounts(counts map[string]int) string {
	open := styles.StatusOpenStyle.Render(fmt.Sprintf("%s %d", styles.IconOpen, counts[tick.StatusOpen]))
	inProgress := styles.StatusInProgressStyle.Render(fmt.Sprintf("%s %d", styles.IconInProgress, counts[tick.StatusInProgress]))
	closed := styles.StatusClosedStyle.Render(fmt.Sprintf("%s %d", styles.IconClosed, counts[tick.StatusClosed]))
	return fmt.Sprintf("%s · %s · %s", open, inProgress, closed)
}

func formatPriorityCounts(counts map[int]int) string {
	var parts []string
	for i := 0; i <= 4; i++ {
		label := fmt.Sprintf("P%d:%d", i, counts[i])
		switch i {
		case 0:
			parts = append(parts, styles.PriorityP0Style.Render(label))
		case 1:
			parts = append(parts, styles.PriorityP1Style.Render(label))
		case 2:
			parts = append(parts, styles.PriorityP2Style.Render(label))
		case 3:
			parts = append(parts, styles.PriorityP3Style.Render(label))
		default:
			parts = append(parts, styles.PriorityP4Style.Render(label))
		}
	}
	return strings.Join(parts, " · ")
}

func formatTypeCounts(counts map[string]int) string {
	bug := styles.TypeBugStyle.Render(fmt.Sprintf("bug:%d", counts[tick.TypeBug]))
	feature := styles.TypeFeatureStyle.Render(fmt.Sprintf("feature:%d", counts[tick.TypeFeature]))
	task := styles.TypeTaskStyle.Render(fmt.Sprintf("task:%d", counts[tick.TypeTask]))
	epic := styles.TypeEpicStyle.Render(fmt.Sprintf("epic:%d", counts[tick.TypeEpic]))
	chore := styles.TypeChoreStyle.Render(fmt.Sprintf("chore:%d", counts[tick.TypeChore]))
	return fmt.Sprintf("%s · %s · %s · %s · %s", bug, feature, task, epic, chore)
}
