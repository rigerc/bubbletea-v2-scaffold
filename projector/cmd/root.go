// Package cmd provides the CLI commands for the application using Cobra.
// This is the root command that all subcommands are attached to.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// cfgFile holds the path to the configuration file.
	cfgFile string

	// debugMode indicates if debug mode is enabled.
	debugMode bool

	// logLevel sets the logging verbosity.
	logLevel string

	// runUI indicates whether to run the TUI after command execution.
	// This is set to false when running subcommands like version or completion.
	runUI = true
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "projector",
	Short: "A production-ready BubbleTea v2 template",
	Long: `projector is a comprehensive scaffold for building terminal
user interface applications using BubbleTea v2, Bubbles v2, and Lip Gloss v2.

This template includes:
- Cobra CLI framework with flag support
- Zerolog structured logging
- JSON configuration with environment variable overrides
- Embedded filesystem support via koanfs
- Debug mode for development
- Shell completions (bash/zsh/fish)`,
	Example: `  # Run with default settings
  projector

  # Run with custom config file
  projector --config /path/to/config.json

  # Run with debug logging
  projector --debug --log-level trace

  # Show version information
  projector version`,
	Version: "1.0.0",
	// Run executes the root command.
	RunE: func(cmd *cobra.Command, args []string) error {
		// The actual TUI application will be run from the main package
		// after the Cobra command is executed.
		return nil
	},
}

// Execute runs the root command. This is called from main.go.
// It returns an error if the command fails.
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCmd returns the root Cobra command.
// This allows the main package to access command configuration
// before executing the TUI application.
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// IsDebugMode returns whether debug mode is enabled.
// This can be checked anywhere in the codebase to enable
// additional logging or debugging features.
func IsDebugMode() bool {
	return debugMode
}

// ShouldRunUI returns whether the TUI should be run after command execution.
// This is false when running subcommands like version or completion.
func ShouldRunUI() bool {
	return runUI
}

// init initializes the root command with flags and configuration.
func init() {
	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Path to configuration file (default: $HOME/.projector.json)")

	// Debug mode flag
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false,
		"Enable debug mode with trace logging")

	// Log level flag
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Set logging level (trace, debug, info, warn, error, fatal)")
}

// GetConfigFile returns the path to the configuration file.
func GetConfigFile() string {
	return cfgFile
}

// GetLogLevel returns the configured log level.
func GetLogLevel() string {
	return logLevel
}

// WasLogLevelSet reports whether --log-level was explicitly passed on the command line.
// Use this to distinguish an explicit flag from Cobra's default value.
func WasLogLevelSet() bool {
	return rootCmd.PersistentFlags().Changed("log-level")
}
