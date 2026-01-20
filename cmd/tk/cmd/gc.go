package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pengelbrecht/ticks/internal/gc"
	"github.com/spf13/cobra"
)

var (
	gcDryRun bool
	gcMaxAge string
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Clean up old log files",
	Long: `Clean up old log files from .tick/logs/ and trim old activity entries.

Targets:
  - .tick/logs/records/*.json  (run records)
  - .tick/logs/runs/*.jsonl    (run logs)
  - .tick/logs/checkpoints/*.json
  - .tick/logs/context/*.md
  - .tick/activity/activity.jsonl (trims old entries)

Live files (.live.json) are never deleted.

Use --dry-run to preview what would be deleted without making changes.
Use --max-age to specify how old files must be to be deleted (default: 30d).`,
	Args: cobra.NoArgs,
	RunE: runGC,
}

func init() {
	gcCmd.Flags().BoolVar(&gcDryRun, "dry-run", false, "preview changes without deleting files")
	gcCmd.Flags().StringVar(&gcMaxAge, "max-age", "30d", "maximum age of files to keep (e.g., 7d, 2w, 1m)")
	rootCmd.AddCommand(gcCmd)
}

func runGC(cmd *cobra.Command, args []string) error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repo root: %w", err)
	}

	tickDir := filepath.Join(root, ".tick")
	if _, err := os.Stat(tickDir); os.IsNotExist(err) {
		return fmt.Errorf("no .tick directory found - run 'tk init' first")
	}

	// Parse max-age duration
	maxAge, err := parseDuration(gcMaxAge)
	if err != nil {
		return fmt.Errorf("invalid --max-age: %w", err)
	}

	// Run cleanup
	cleaner := gc.NewCleaner(root).
		WithMaxAge(maxAge).
		WithDryRun(gcDryRun)

	if gcDryRun {
		fmt.Println("Dry run - no files will be deleted")
		fmt.Println()
	}

	result, err := cleaner.Cleanup()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Report results
	if result.FilesDeleted == 0 && result.EntriesTrimmed == 0 {
		fmt.Println("Nothing to clean up.")
		return nil
	}

	if gcDryRun {
		fmt.Println("Would delete:")
	} else {
		fmt.Println("Deleted:")
	}

	if result.FilesDeleted > 0 {
		fmt.Printf("  %d files (%s)\n", result.FilesDeleted, formatBytes(result.BytesFreed))
	}

	if result.EntriesTrimmed > 0 {
		if gcDryRun {
			fmt.Printf("  %d activity log entries would be trimmed\n", result.EntriesTrimmed)
		} else {
			fmt.Printf("  %d activity log entries trimmed\n", result.EntriesTrimmed)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered: %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

// parseDuration parses a human-friendly duration string like "7d", "2w", "1m".
// Supports: d (days), w (weeks), m (months, 30 days).
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("duration too short: %q", s)
	}

	unit := s[len(s)-1]
	valueStr := s[:len(s)-1]

	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return 0, fmt.Errorf("invalid number: %q", valueStr)
	}

	if value <= 0 {
		return 0, fmt.Errorf("duration must be positive: %d", value)
	}

	switch unit {
	case 'd':
		return time.Duration(value) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case 'm':
		return time.Duration(value) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit %q (use d, w, or m)", string(unit))
	}
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
