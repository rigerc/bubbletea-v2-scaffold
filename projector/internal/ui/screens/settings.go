// Package screens provides the individual screen implementations for the application.
package screens

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	tea "charm.land/bubbletea/v2"

	"projector/internal/ui/nav"
)

// SettingsData holds form data for the settings screen.
type SettingsData struct {
	Category       string  // "basic", "advanced", or "developer"
	Username       string
	Email          string
	LogLevel       string
	EnableDebug    bool
	ApiEndpoint    string
	MaxConnections int
}

// SettingsAppliedMsg is emitted when settings are successfully submitted.
type SettingsAppliedMsg struct {
	Data SettingsData
}

// SettingsScreen displays a dynamic settings form that reacts to user choices.
// It demonstrates Huh's dynamic form capabilities using Func variants.
type SettingsScreen struct {
	*FormScreen
	data SettingsData
}

// NewSettingsScreen creates a settings form with dynamic fields
// that react to earlier answers using Func variants.
func NewSettingsScreen(isDark bool, appName string) *SettingsScreen {
	data := SettingsData{
		Category:    "basic",
		LogLevel:    "info",
	}

	// formBuilder is a function that can rebuild the form when needed
	formBuilder := func() *huh.Form {
		return huh.NewForm(
			// ===== PAGE 1: Introduction =====
			huh.NewGroup(
				huh.NewNote().
					Title("Settings").
					Description("Configure your application preferences.\nChoose a category to see relevant options.").
					Next(true).
					NextLabel("Begin"),
			),

			// ===== PAGE 2: Category Selection =====
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Category").
					Description("Choose settings category").
					Options(
						huh.NewOption("Basic", "basic"),
						huh.NewOption("Advanced", "advanced"),
						huh.NewOption("Developer", "developer"),
					).
					Value(&data.Category),
			),

			// ===== PAGE 3: Basic Settings (shown for all) =====
			huh.NewGroup(
				// Dynamic title based on category
				huh.NewInput().
					TitleFunc(func() string {
						if data.Category == "developer" {
							return "Developer Username"
						}
						return "Username"
					}, &data.Category).
					DescriptionFunc(func() string {
						if data.Category == "developer" {
							return "Your developer handle (3+ chars)"
						}
						return "Your display name"
					}, &data.Category).
					Validate(func(s string) error {
						if len(s) < 3 {
							return fmt.Errorf("must be at least 3 characters")
						}
						return nil
					}).
					Value(&data.Username),

				// Email field - only for basic/advanced categories
				huh.NewInput().
					Title("Email").
					Placeholder("user@example.com").
					Validate(func(s string) error {
						if s != "" && !strings.Contains(s, "@") {
							return fmt.Errorf("invalid email format")
						}
						return nil
					}).
					Value(&data.Email),
			).WithHideFunc(func() bool {
				// Hide this group for developer category
				return data.Category == "developer"
			}),

			// ===== PAGE 4: Category-Specific Settings =====
			// This group's content changes dynamically based on category
			huh.NewGroup(
				// Log level - always shown but options change
				huh.NewSelect[string]().
					TitleFunc(func() string {
						if data.Category == "developer" {
							return "Verbosity"
						}
						return "Log Level"
					}, &data.Category).
					OptionsFunc(func() []huh.Option[string] {
						if data.Category == "developer" {
							// More options for developers
							return []huh.Option[string]{
								huh.NewOption("Trace", "trace"),
								huh.NewOption("Debug", "debug"),
								huh.NewOption("Info", "info").Selected(true),
								huh.NewOption("Warning", "warn"),
								huh.NewOption("Error", "error"),
							}
						}
						// Simplified options for users
						return []huh.Option[string]{
							huh.NewOption("Normal", "info").Selected(true),
							huh.NewOption("Errors Only", "error"),
						}
					}, &data.Category).
					Value(&data.LogLevel),

				// Debug/diagnostics toggle - description adapts
				huh.NewConfirm().
					TitleFunc(func() string {
						if data.Category == "developer" {
							return "Enable Debug Mode?"
						}
						return "Show Diagnostics?"
					}, &data.Category).
					DescriptionFunc(func() string {
						if data.Category == "developer" {
							return "This will show detailed stack traces and timing info"
						}
						return "Display additional diagnostic information"
					}, &data.Category).
					Value(&data.EnableDebug),
			),

			// ===== PAGE 5: Advanced/Developer Settings =====
			// This entire group is hidden for "basic" category
			huh.NewGroup(
				huh.NewInput().
					Title("API Endpoint").
					Description("Custom API server URL").
					Placeholder("https://api.example.com").
					Value(&data.ApiEndpoint),

				huh.NewSelect[int]().
					Title("Max Connections").
					Options(
						huh.NewOption("Low (5)", 5),
						huh.NewOption("Medium (10)", 10).Selected(true),
						huh.NewOption("High (20)", 20),
						huh.NewOption("Unlimited (0)", 0),
					).
					Value(&data.MaxConnections),
			).WithHideFunc(func() bool {
				// Hide this group for basic category
				return data.Category == "basic"
			}),

			// ===== PAGE 6: Confirmation =====
			huh.NewGroup(
				huh.NewNote().
					TitleFunc(func() string {
						return strings.ToUpper(data.Category) + " Settings Ready"
					}, &data.Category).
					DescriptionFunc(func() string {
						var parts []string
						parts = append(parts, fmt.Sprintf("* Username: %s", data.Username))
						if data.Category != "developer" && data.Email != "" {
							parts = append(parts, fmt.Sprintf("* Email: %s", data.Email))
						}
						parts = append(parts, fmt.Sprintf("* Log Level: %s", data.LogLevel))
						if data.EnableDebug {
							parts = append(parts, "* Debug: enabled")
						}
						if data.Category != "basic" {
							parts = append(parts, fmt.Sprintf("* API: %s", data.ApiEndpoint))
							parts = append(parts, fmt.Sprintf("* Connections: %d", data.MaxConnections))
						}
						return strings.Join(parts, "\n")
					}, &data.Category).
					Next(true).
					NextLabel("Apply Settings"),
			),
		).WithShowHelp(true).WithShowErrors(true)
	}

	onSubmit := func() tea.Cmd {
		return func() tea.Msg {
			return SettingsAppliedMsg{Data: data}
		}
	}

	onAbort := func() tea.Cmd {
		return nav.Pop()
	}

	fs := newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)

	return &SettingsScreen{
		FormScreen: fs,
		data:       data,
	}
}

// Update handles incoming messages and returns an updated screen and command.
func (s *SettingsScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	// Delegate to FormScreen
	screen, cmd := s.FormScreen.Update(msg)

	// Update our reference if the FormScreen changed
	if fs, ok := screen.(*FormScreen); ok {
		s.FormScreen = fs
	}

	return s, cmd
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *SettingsScreen) SetTheme(isDark bool) {
	s.FormScreen.SetTheme(isDark)
}
