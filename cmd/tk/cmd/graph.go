package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/github"
	"github.com/pengelbrecht/ticks/internal/styles"
	"github.com/pengelbrecht/ticks/internal/tick"
)

var graphCmd = &cobra.Command{
	Use:   "graph <epic-id>",
	Short: "Show dependency graph and parallelization opportunities for an epic",
	Long: `Show the dependency structure of an epic's tasks.

Displays tasks organized into "waves" - groups that can be executed in parallel.
Wave 1 contains tasks with no blockers (ready now), Wave 2 contains tasks
that become ready after Wave 1 completes, and so on.

This helps agents understand:
- How many subagents can run in parallel at each stage
- The critical path through the epic (minimum sequential steps)
- Which tasks are blocking others

Examples:
  tk graph abc          # Show dependency graph for epic abc
  tk graph abc --all    # Include closed tasks`,
	Args: cobra.ExactArgs(1),
	RunE: runGraph,
}

var (
	graphAll  bool
	graphJSON bool
)

func init() {
	graphCmd.Flags().BoolVarP(&graphAll, "all", "a", false, "include closed tasks")
	graphCmd.Flags().BoolVar(&graphJSON, "json", false, "output as JSON (agent-optimized)")
	rootCmd.AddCommand(graphCmd)
}

// wave represents a group of tasks that can be executed in parallel.
type wave struct {
	level int
	ticks []tick.Tick
}

// graphOutput is the JSON output structure for agents.
type graphOutput struct {
	Epic         graphEpic   `json:"epic"`
	Stats        graphStats  `json:"stats"`
	Waves        []graphWave `json:"waves"`
	CriticalPath int         `json:"critical_path"`
}

type graphEpic struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type graphStats struct {
	TotalTasks     int `json:"total_tasks"`
	WaveCount      int `json:"wave_count"`
	MaxParallel    int `json:"max_parallel"`
	ReadyForAgent  int `json:"ready_for_agent"`
	AwaitingHuman  int `json:"awaiting_human"`
	Deferred       int `json:"deferred"`
}

type graphWave struct {
	Wave     int         `json:"wave"`
	Parallel int         `json:"parallel"`
	Ready    bool        `json:"ready"`
	Tasks    []graphTask `json:"tasks"`
}

type graphTask struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Priority     int      `json:"priority"`
	Status       string   `json:"status"`
	BlockedBy    []string `json:"blocked_by,omitempty"`
	Blocks       []string `json:"blocks,omitempty"`
	Awaiting     string   `json:"awaiting,omitempty"`
	DeferredUntil string  `json:"deferred_until,omitempty"`
	AgentReady   bool     `json:"agent_ready"`
}

func runGraph(cmd *cobra.Command, args []string) error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repo root: %w", err)
	}

	project, err := github.DetectProject(nil)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	epicID, err := github.NormalizeID(project, args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))

	// Read the epic
	epic, err := store.Read(epicID)
	if err != nil {
		return fmt.Errorf("failed to read epic: %w", err)
	}

	if epic.Type != tick.TypeEpic {
		return fmt.Errorf("%s is not an epic (type: %s)", epicID, epic.Type)
	}

	// Get all ticks
	allTicks, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list ticks: %w", err)
	}

	// Filter to tasks under this epic
	var tasks []tick.Tick
	tickMap := make(map[string]tick.Tick)
	for _, t := range allTicks {
		tickMap[t.ID] = t
		if t.Parent == epicID && t.Type != tick.TypeEpic {
			if graphAll || t.Status != tick.StatusClosed {
				tasks = append(tasks, t)
			}
		}
	}

	if len(tasks) == 0 {
		fmt.Printf("Epic %s has no tasks\n", epicID)
		return nil
	}

	// Build dependency graph
	// For each task, find which tasks in this epic block it
	blockedBy := make(map[string][]string)   // task -> tasks that block it
	blocks := make(map[string][]string)      // task -> tasks it blocks
	inDegree := make(map[string]int)         // number of open blockers

	taskSet := make(map[string]bool)
	for _, t := range tasks {
		taskSet[t.ID] = true
		inDegree[t.ID] = 0
	}

	for _, t := range tasks {
		for _, blockerID := range t.BlockedBy {
			// Only count blockers that are in this epic and not closed
			if taskSet[blockerID] {
				blocker, exists := tickMap[blockerID]
				if exists && blocker.Status != tick.StatusClosed {
					blockedBy[t.ID] = append(blockedBy[t.ID], blockerID)
					blocks[blockerID] = append(blocks[blockerID], t.ID)
					inDegree[t.ID]++
				}
			}
		}
	}

	// Compute waves using Kahn's algorithm (topological sort by levels)
	var waves []wave
	remaining := make(map[string]bool)
	for _, t := range tasks {
		remaining[t.ID] = true
	}

	waveNum := 1
	for len(remaining) > 0 {
		// Find all tasks with no remaining blockers
		var ready []tick.Tick
		for _, t := range tasks {
			if remaining[t.ID] && inDegree[t.ID] == 0 {
				ready = append(ready, t)
			}
		}

		if len(ready) == 0 {
			// Cycle detected - remaining tasks have circular dependencies
			var cycleIDs []string
			for id := range remaining {
				cycleIDs = append(cycleIDs, id)
			}
			sort.Strings(cycleIDs)
			fmt.Printf("\n%s Circular dependency detected among: %s\n",
				styles.StatusBlockedStyle.Render("!"),
				strings.Join(cycleIDs, ", "))
			break
		}

		// Sort by priority within wave
		sort.Slice(ready, func(i, j int) bool {
			if ready[i].Priority != ready[j].Priority {
				return ready[i].Priority < ready[j].Priority
			}
			return ready[i].ID < ready[j].ID
		})

		waves = append(waves, wave{level: waveNum, ticks: ready})

		// Remove ready tasks and update inDegree
		for _, t := range ready {
			delete(remaining, t.ID)
			for _, dependentID := range blocks[t.ID] {
				if remaining[dependentID] {
					inDegree[dependentID]--
				}
			}
		}
		waveNum++
	}

	// Calculate stats
	maxParallel := 0
	for _, w := range waves {
		if len(w.ticks) > maxParallel {
			maxParallel = len(w.ticks)
		}
	}

	// Count workflow states
	readyForAgent := 0
	awaitingHuman := 0
	deferred := 0
	now := time.Now()

	for _, t := range tasks {
		isDeferred := t.DeferUntil != nil && t.DeferUntil.After(now)
		isAwaiting := t.IsAwaitingHuman()
		isBlocked := inDegree[t.ID] > 0
		isClosed := t.Status == tick.StatusClosed

		if isDeferred {
			deferred++
		} else if isAwaiting {
			awaitingHuman++
		} else if !isBlocked && !isClosed {
			readyForAgent++
		}
	}

	// JSON output for agents
	if graphJSON {
		output := graphOutput{
			Epic: graphEpic{
				ID:    epic.ID,
				Title: epic.Title,
			},
			Stats: graphStats{
				TotalTasks:    len(tasks),
				WaveCount:     len(waves),
				MaxParallel:   maxParallel,
				ReadyForAgent: readyForAgent,
				AwaitingHuman: awaitingHuman,
				Deferred:      deferred,
			},
			CriticalPath: len(waves),
		}

		for _, w := range waves {
			gw := graphWave{
				Wave:     w.level,
				Parallel: len(w.ticks),
				Ready:    w.level == 1,
			}
			for _, t := range w.ticks {
				isDeferred := t.DeferUntil != nil && t.DeferUntil.After(now)
				isAwaiting := t.IsAwaitingHuman()
				isBlocked := inDegree[t.ID] > 0
				isClosed := t.Status == tick.StatusClosed
				agentReady := !isDeferred && !isAwaiting && !isBlocked && !isClosed

				gt := graphTask{
					ID:         t.ID,
					Title:      t.Title,
					Priority:   t.Priority,
					Status:     t.Status,
					BlockedBy:  blockedBy[t.ID],
					Blocks:     blocks[t.ID],
					AgentReady: agentReady,
				}
				if t.Awaiting != nil {
					gt.Awaiting = *t.Awaiting
				}
				if t.DeferUntil != nil {
					gt.DeferredUntil = t.DeferUntil.Format("2006-01-02")
				}
				gw.Tasks = append(gw.Tasks, gt)
			}
			output.Waves = append(output.Waves, gw)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Human-readable output
	fmt.Printf("%s %s\n", styles.TypeEpicStyle.Render("Epic:"), epic.Title)
	fmt.Printf("%s %d tasks, %d waves, max %d parallel\n",
		styles.DimStyle.Render("Stats:"),
		len(tasks), len(waves), maxParallel)

	// Show workflow breakdown if there are awaiting/deferred tasks
	if awaitingHuman > 0 || deferred > 0 {
		parts := []string{fmt.Sprintf("%d agent-ready", readyForAgent)}
		if awaitingHuman > 0 {
			parts = append(parts, fmt.Sprintf("%d awaiting human", awaitingHuman))
		}
		if deferred > 0 {
			parts = append(parts, fmt.Sprintf("%d deferred", deferred))
		}
		fmt.Printf("%s %s\n", styles.DimStyle.Render("       "), strings.Join(parts, ", "))
	}
	fmt.Println()

	for _, w := range waves {
		// Count truly agent-ready tasks in this wave
		agentReadyInWave := 0
		for _, t := range w.ticks {
			isDeferred := t.DeferUntil != nil && t.DeferUntil.After(now)
			isAwaiting := t.IsAwaitingHuman()
			isClosed := t.Status == tick.StatusClosed
			if !isDeferred && !isAwaiting && !isClosed {
				agentReadyInWave++
			}
		}

		parallelHint := ""
		if agentReadyInWave > 1 {
			parallelHint = styles.DimStyle.Render(fmt.Sprintf(" (%d parallel)", agentReadyInWave))
		} else if agentReadyInWave == 0 && len(w.ticks) > 0 {
			parallelHint = styles.DimStyle.Render(" (none agent-ready)")
		}

		if w.level == 1 {
			if agentReadyInWave > 0 {
				fmt.Printf("%s%s\n", styles.StatusInProgressStyle.Render("Wave 1 (ready now)"), parallelHint)
			} else {
				fmt.Printf("%s%s\n", styles.DimStyle.Render("Wave 1"), parallelHint)
			}
		} else {
			fmt.Printf("%s%s\n", styles.DimStyle.Render(fmt.Sprintf("Wave %d", w.level)), parallelHint)
		}

		for _, t := range w.ticks {
			statusIcon := renderTaskStatus(t, tickMap, taskSet, now)
			blockerInfo := ""
			if len(blockedBy[t.ID]) > 0 {
				blockerInfo = styles.DimStyle.Render(" ← " + strings.Join(blockedBy[t.ID], ", "))
			}
			// Show deferred date if applicable
			if t.DeferUntil != nil && t.DeferUntil.After(now) {
				blockerInfo += styles.DimStyle.Render(fmt.Sprintf(" [deferred until %s]", t.DeferUntil.Format("Jan 2")))
			}
			fmt.Printf("  %s %s %s %s%s\n",
				statusIcon,
				t.ID,
				styles.RenderPriority(t.Priority),
				t.Title,
				blockerInfo)
		}
		fmt.Println()
	}

	// Critical path info
	fmt.Printf("%s %d waves (minimum sequential steps to complete epic)\n",
		styles.DimStyle.Render("Critical path:"), len(waves))

	return nil
}

// renderTaskStatus returns a status icon for a task in the graph context.
func renderTaskStatus(t tick.Tick, tickMap map[string]tick.Tick, taskSet map[string]bool, now time.Time) string {
	// Deferred takes precedence (shown as pending/clock)
	if t.DeferUntil != nil && t.DeferUntil.After(now) {
		return styles.DimStyle.Render(styles.IconPending)
	}

	// Awaiting human
	if t.IsAwaitingHuman() {
		return styles.StatusAwaitingStyle.Render(styles.IconAwaiting)
	}

	// Check if blocked by any open task in the epic
	for _, blockerID := range t.BlockedBy {
		if taskSet[blockerID] {
			blocker, exists := tickMap[blockerID]
			if exists && blocker.Status != tick.StatusClosed {
				return styles.StatusBlockedStyle.Render(styles.IconBlocked)
			}
		}
	}

	return styles.RenderStatus(t.Status)
}
