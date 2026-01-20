package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pengelbrecht/ticks/internal/migrate"
	"github.com/spf13/cobra"
)

var (
	migrateDryRun bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run data migrations",
	Long: `Run data migrations to upgrade .tick data to the latest format.

Currently supports:
  - run-records: Migrate run records from tick JSON files to .tick/logs/records/

Use --dry-run to preview changes without modifying any files.`,
	RunE: runMigrate,
}

func init() {
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "preview changes without modifying files")
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	root, err := repoRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repo root: %w", err)
	}

	tickDir := filepath.Join(root, ".tick")
	if _, err := os.Stat(tickDir); os.IsNotExist(err) {
		return fmt.Errorf("no .tick directory found - run 'tk init' first")
	}

	// Check if migration is needed
	needsMigration, err := migrate.NeedsMigration(tickDir)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needsMigration {
		fmt.Println("No migrations needed - all data is up to date.")
		return nil
	}

	// Run the run record migration
	fmt.Println("Migrating run records from tick JSON to .tick/logs/records/...")
	if migrateDryRun {
		fmt.Println("(dry-run mode - no files will be modified)")
	}

	m := migrate.NewRunRecordMigration(tickDir)
	m.SetDryRun(migrateDryRun)

	result, err := m.Run()
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Report results
	if migrateDryRun {
		fmt.Printf("\nDry run complete:\n")
		fmt.Printf("  Would migrate: %d run records\n", result.Migrated)
		fmt.Printf("  Would skip: %d ticks (no run record)\n", result.Skipped)
	} else {
		fmt.Printf("\nMigration complete:\n")
		fmt.Printf("  Migrated: %d run records\n", result.Migrated)
		fmt.Printf("  Skipped: %d ticks (no run record)\n", result.Skipped)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("    - %s\n", e)
		}
	}

	return nil
}
