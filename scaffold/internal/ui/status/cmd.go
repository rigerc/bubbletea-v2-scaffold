package status

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Default durations for each status kind.
var (
	DefaultInfoDuration    = 3 * time.Second
	DefaultSuccessDuration = 3 * time.Second
	DefaultWarningDuration = 5 * time.Second
	DefaultErrorDuration   = 5 * time.Second
)

// Set returns a command that sets a status message with explicit kind and duration.
// Duration of 0 means the message persists until cleared.
func Set(text string, kind Kind, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return Msg{Text: text, Kind: kind, Duration: duration}
	}
}

// SetWithClear sets a status message and schedules automatic clearing.
func SetWithClear(text string, kind Kind, duration time.Duration) tea.Cmd {
	return tea.Batch(
		Set(text, kind, duration),
		tea.Tick(duration, func(time.Time) tea.Msg { return ClearMsg{} }),
	)
}

// SetInfo sets an informational status message.
// If duration is 0, uses DefaultInfoDuration.
func SetInfo(text string, duration time.Duration) tea.Cmd {
	if duration == 0 {
		duration = DefaultInfoDuration
	}
	return SetWithClear(text, KindInfo, duration)
}

// SetSuccess sets a success status message.
// If duration is 0, uses DefaultSuccessDuration.
func SetSuccess(text string, duration time.Duration) tea.Cmd {
	if duration == 0 {
		duration = DefaultSuccessDuration
	}
	return SetWithClear(text, KindSuccess, duration)
}

// SetWarning sets a warning status message.
// If duration is 0, uses DefaultWarningDuration.
func SetWarning(text string, duration time.Duration) tea.Cmd {
	if duration == 0 {
		duration = DefaultWarningDuration
	}
	return SetWithClear(text, KindWarning, duration)
}

// SetError sets an error status message.
// If duration is 0, uses DefaultErrorDuration.
func SetError(text string, duration time.Duration) tea.Cmd {
	if duration == 0 {
		duration = DefaultErrorDuration
	}
	return SetWithClear(text, KindError, duration)
}

// Clear returns a command that clears the status.
func Clear() tea.Cmd {
	return func() tea.Msg { return ClearMsg{} }
}

// Persistent returns a command that sets a persistent status (no auto-clear).
func Persistent(text string, kind Kind) tea.Cmd {
	return Set(text, kind, 0)
}
