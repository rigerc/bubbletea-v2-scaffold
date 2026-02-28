// Package logger provides a simple debug logging utility.
// Logging is only enabled when debug mode is active via config or CLI flag.
package logger

import (
	"io"
	"log"
	"os"
	"sync"

	tea "charm.land/bubbletea/v2"
)

// Logger is the global logger instance. It writes to debug.log when enabled,
// or discards output when disabled.
var Logger *log.Logger

// fileHandle stores the log file handle for cleanup.
var fileHandle io.WriteCloser

var mu sync.Mutex

// NoOpWriter discards all writes.
type NoOpWriter struct{}

func (w *NoOpWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// syncWriter wraps a writer and syncs after each write for real-time output.
type syncWriter struct {
	w io.Writer
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	n, err := sw.w.Write(p)
	if syncer, ok := sw.w.(interface{ Sync() error }); ok {
		_ = syncer.Sync()
	}
	return n, err
}

// Setup initializes the global logger based on debug mode.
// When debug is true, logs are written to "debug.log" in the current directory.
// When debug is false, all log output is discarded.
func Setup(debug bool) {
	mu.Lock()
	defer mu.Unlock()

	// Close existing file handle if switching modes
	if fileHandle != nil {
		_ = fileHandle.Close()
		fileHandle = nil
	}

	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		fileHandle = f
		Logger = log.New(&syncWriter{w: f}, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		Logger = log.New(&NoOpWriter{}, "", 0)
	}
}

// SetupWithWriter initializes the logger with a custom writer.
// This is useful for testing or redirecting output elsewhere.
func SetupWithWriter(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if fileHandle != nil {
		_ = fileHandle.Close()
		fileHandle = nil
	}
	Logger = log.New(w, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Close closes the log file if one was opened.
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if fileHandle != nil {
		_ = fileHandle.Close()
		fileHandle = nil
	}
}

// Debug logs a message when debug mode is enabled.
func Debug(format string, v ...any) {
	if Logger != nil {
		Logger.Printf(format, v...)
	}
}

// Fatal logs a message and exits when debug mode is enabled.
func Fatal(format string, v ...any) {
	if Logger != nil {
		Logger.Fatalf(format, v...)
	}
	os.Exit(1)
}
