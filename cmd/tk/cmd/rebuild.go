package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pengelbrecht/ticks/internal/query"
	"github.com/pengelbrecht/ticks/internal/tick"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the .tick index",
	Long: `Rebuild the .tick index file.

Regenerates the .index.json file from all tick files in the .tick/issues directory.
Use this if the index becomes out of sync with the actual tick files.`,
	Args: cobra.NoArgs,
	RunE: runRebuild,
}

var rebuildJSON bool

func init() {
	rebuildCmd.Flags().BoolVar(&rebuildJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(rebuildCmd)
}

func runRebuild(cmd *cobra.Command, args []string) error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repo root: %w", err)
	}

	store := tick.NewStore(filepath.Join(root, ".tick"))
	ticks, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list ticks: %w", err)
	}

	indexPath := filepath.Join(root, ".tick", ".index.json")
	if err := query.SaveIndex(indexPath, ticks); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	if rebuildJSON {
		payload := map[string]any{"count": len(ticks)}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(payload); err != nil {
			return fmt.Errorf("failed to encode json: %w", err)
		}
		return nil
	}

	fmt.Printf("Rebuilt index with %d ticks\n", len(ticks))
	return nil
}
