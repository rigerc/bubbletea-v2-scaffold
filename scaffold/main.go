// scaffold is a minimal BubbleTea v2 skeleton.
// It wires up logging, an optional Cobra CLI, and then starts the TUI.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"

	"scaffold/cmd"
	"scaffold/config"
	applogger "scaffold/internal/logger"
	"scaffold/internal/ui"
)

func main() {
	// Execute the Cobra CLI. Subcommands (version, completion) set runUI=false
	// and exit early; the root command falls through to the TUI.
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Command execution failed: %v\n", err)
		os.Exit(1)
	}

	if !cmd.ShouldRunUI() {
		return
	}

	cfg, configPath := loadConfig()

	// In TUI mode the terminal is occupied, so all logging must go to a file
	// (debug mode) or be silenced entirely (normal mode).
	logOutput, cleanup, err := setupLogOutput(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
	if cleanup != nil {
		defer cleanup()
	}

	if err := initLogger(cfg, logOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}

	applogger.Info().Msg("Starting scaffold")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 8192)
			n := runtime.Stack(buf, false)
			applogger.Fatal().
				Str("panic", fmt.Sprintf("%v", r)).
				Str("stack", string(buf[:n])).
				Msg("panic: unrecovered")
			fmt.Fprintf(os.Stderr, "\n[scaffold] crashed — see debug.log for details\n")
			os.Exit(2)
		}
	}()

	firstRun := config.IsFirstRun(configPath)
	if err := ui.Run(ctx, ui.New(*cfg, configPath, firstRun)); err != nil {
		applogger.Fatal().Err(err).Msg("UI failed")
	}
}

// setupLogOutput returns the writer to use for logging and an optional cleanup
// function that must be deferred by the caller.
func setupLogOutput(cfg *config.Config) (io.Writer, func(), error) {
	if cfg.Debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return nil, nil, fmt.Errorf("opening debug log: %w", err)
		}
		return f, func() { f.Close() }, nil
	}
	return io.Discard, nil, nil
}

// initLogger initialises the global zerolog logger.
func initLogger(cfg *config.Config, output io.Writer) error {
	format := "console"
	if os.Getenv("ENV") == "production" {
		format = "json"
	}

	if err := applogger.Init(applogger.Config{
		Level:  applogger.LogLevel(cfg.GetEffectiveLogLevel()),
		Format: format,
		Output: output,
	}); err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	return nil
}

// loadConfig builds the effective config following priority order:
// defaults → config file → CLI flags (only when explicitly set).
// Returns the config and the path used (empty if no file was loaded).
func loadConfig() (*config.Config, string) {
	cfg := config.DefaultConfig()
	configPath := ""

	if path := cmd.GetConfigFile(); path != "" {
		fileCfg, err := config.Load(path)
		if err == nil {
			cfg = fileCfg
			configPath = path
		}
		// ErrConfigNotFound or parse error → silently fall back to defaults
	}

	// CLI flags override file/defaults only when explicitly passed.
	if cmd.IsDebugMode() {
		cfg.Debug = true
	}
	if cmd.WasLogLevelSet() {
		cfg.LogLevel = cmd.GetLogLevel()
	}

	return cfg, configPath
}
