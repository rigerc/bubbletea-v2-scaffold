// scaffold is a minimal BubbleTea v2 skeleton.
// It wires up logging, an optional Cobra CLI, and then starts the TUI.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"scaffold/cmd"
	"scaffold/config"
	"scaffold/internal/logger"
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

	// Initialize logger early based on CLI flag (config may override later)
	logger.Setup(cmd.IsDebugMode())
	defer logger.Close()

	cfg, configPath := loadConfig()

	// Re-initialize if config debug setting differs from CLI flag
	if cfg.Debug {
		logger.Setup(true)
	}

	logger.Debug("starting scaffold (debug mode enabled)")
	logger.Debug("config path: %s", configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 8192)
			n := runtime.Stack(buf, false)
			logger.Debug("panic recovered: %v\n%s", r, string(buf[:n]))
			fmt.Fprintf(os.Stderr, "\n[scaffold] crashed\npanic: %v\nstack: %s\n", r, string(buf[:n]))
			os.Exit(2)
		}
	}()

	firstRun := config.IsFirstRun(configPath)
	logger.Debug("first run: %v", firstRun)
	logger.Debug("starting UI")

	if err := ui.Run(ctx, ui.New(ctx, cancel, *cfg, configPath, firstRun)); err != nil {
		logger.Debug("Program exited: %v", err)
		os.Exit(1)
	}
}

// loadConfig builds the effective config following priority order:
// defaults → config file → CLI flags (only when explicitly set).
// Returns the config and the path to use (default path even if file doesn't exist yet).
func loadConfig() (*config.Config, string) {
	cfg := config.DefaultConfig()
	configPath := cmd.GetConfigFile() // Get default or explicit path

	if configPath != "" {
		fileCfg, err := config.Load(configPath)
		if err == nil {
			cfg = fileCfg
			logger.Debug("loaded config from: %s", configPath)
		} else {
			logger.Debug("config load failed, using defaults: %v", err)
		}
		// ErrConfigNotFound or parse error → silently fall back to defaults
		// but keep configPath so first-run detection and saving work
	}

	// CLI flags override file/defaults only when explicitly passed.
	if cmd.IsDebugMode() {
		cfg.Debug = true
	}

	return cfg, configPath
}
