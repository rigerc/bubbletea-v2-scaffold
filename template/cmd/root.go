package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "test-app",
	Short: "A BubbleTea v2 TUI application template",
	Long: `test-app is a template application demonstrating BubbleTea v2
with cobra CLI integration for building modern terminal UIs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
