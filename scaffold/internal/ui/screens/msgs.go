package screens

import (
	"time"

	"scaffold/config"
)

// BackMsg signals that the current screen wants to go back.
type BackMsg struct{}

// SettingsSavedMsg carries the updated config after the user submits the form.
type SettingsSavedMsg struct {
	Cfg config.Config
}

// detailTickMsg is sent every second while the detail screen is loading,
// demonstrating the canonical tea.Tick periodic-task pattern (§7C).
type detailTickMsg time.Time
