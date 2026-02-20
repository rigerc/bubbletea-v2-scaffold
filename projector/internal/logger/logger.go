// Package logger provides structured logging using zerolog.
// It supports both console and JSON output formats with configurable levels.
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	// globalLogger is the singleton logger instance.
	globalLogger *zerolog.Logger

	// globalLoggerMu protects globalLogger during initialization.
	globalLoggerMu sync.RWMutex

	// currentLevel stores the current log level.
	currentLevel zerolog.Level
)

// LogLevel represents the available logging levels.
type LogLevel string

// Available log levels.
const (
	LevelTrace LogLevel = "trace"
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Config holds the logger configuration.
type Config struct {
	// Level is the minimum log level to output.
	Level LogLevel

	// Format specifies the output format ("console" or "json").
	Format string

	// Output is the writer to send logs to.
	// If nil, os.Stderr is used.
	Output io.Writer

	// TimeFormat specifies the time format for log entries.
	// Use "zerolog.TimeFormatUnix" for Unix timestamps or
	// "zerolog.TimeFormatUnixMs" for millisecond timestamps.
	TimeFormat string

	// NoColor disables colored output in console mode.
	NoColor bool
}

// Init initializes the global logger with the given configuration.
// It should be called once at application startup.
// If the logger is already initialized, it will be replaced.
func Init(cfg Config) error {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()

	// Set error stack marshaling
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Set default output if not specified
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	// Parse log level
	level, err := parseLevel(string(cfg.Level))
	if err != nil {
		return fmt.Errorf("parsing log level: %w", err)
	}
	currentLevel = level

	// Set zerolog global level
	zerolog.SetGlobalLevel(level)

	// Configure time format
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = "2006-01-02 15:04:05"
	}

	// Create console or JSON writer
	var logger zerolog.Logger
	if cfg.Format == "json" {
		// JSON output for production
		logger = zerolog.New(cfg.Output).
			Level(level).
			With().
			Timestamp().
			Stack().
			Logger()
	} else {
		// Console output for development
		consoleWriter := zerolog.NewConsoleWriter(
			func(w *zerolog.ConsoleWriter) {
				w.Out = cfg.Output
				w.TimeFormat = cfg.TimeFormat
				w.NoColor = cfg.NoColor
			},
		)

		logger = zerolog.New(consoleWriter).
			Level(level).
			With().
			Timestamp().
			Stack().
			Logger()
	}

	globalLogger = &logger
	return nil
}

// InitWithLevel initializes the logger with a simple level string.
// This is a convenience function that uses default settings.
func InitWithLevel(level string) error {
	return Init(Config{
		Level:  LogLevel(level),
		Format: "console",
	})
}

// InitForDevelopment initializes the logger for development.
// It enables pretty console output with colors.
func InitForDevelopment() error {
	return Init(Config{
		Level:  LevelDebug,
		Format: "console",
		Output: os.Stderr,
	})
}

// InitForProduction initializes the logger for production.
// It enables JSON output without colors.
func InitForProduction() error {
	return Init(Config{
		Level:  LevelInfo,
		Format: "json",
		Output: os.Stderr,
	})
}

// Global returns the global logger instance.
// If the logger has not been initialized, it returns a no-op logger.
func Global() *zerolog.Logger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()

	if globalLogger != nil {
		return globalLogger
	}

	// Return a no-op logger if not initialized
	nop := zerolog.Nop()
	return &nop
}

// SetLevel changes the minimum log level dynamically.
func SetLevel(level LogLevel) error {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()

	parsed, err := parseLevel(string(level))
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(parsed)
	currentLevel = parsed
	return nil
}

// GetLevel returns the current log level.
func GetLevel() zerolog.Level {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return currentLevel
}

// parseLevel converts a string to a zerolog.Level.
func parseLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// Convenience functions that use the global logger.

// Trace logs a trace message.
func Trace() *zerolog.Event {
	return Global().Trace()
}

// Debug logs a debug message.
func Debug() *zerolog.Event {
	return Global().Debug()
}

// Info logs an info message.
func Info() *zerolog.Event {
	return Global().Info()
}

// Warn logs a warning message.
func Warn() *zerolog.Event {
	return Global().Warn()
}

// Error logs an error message.
func Error() *zerolog.Event {
	return Global().Error()
}

// Fatal logs a fatal message and exits.
func Fatal() *zerolog.Event {
	return Global().Fatal()
}

// With creates a logger with context fields.
func With() zerolog.Context {
	return Global().With()
}
