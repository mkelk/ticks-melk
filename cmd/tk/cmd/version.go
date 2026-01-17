package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of tk",
	Long:  `Print the version number of tk.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tk %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
