// Package screens provides the individual screen implementations for the application.
package screens

import (
	"os"

	"charm.land/huh/v2"
	tea "charm.land/bubbletea/v2"

	"projector/internal/ui/nav"
)

// HuhFilePickerScreen uses Huh's FilePicker field for filesystem browsing.
// It implements nav.Screen and nav.Themeable.
type HuhFilePickerScreen struct {
	*FormScreen
}

// fileReadMsg carries the result of reading a selected file from disk.
type huhFileReadMsg struct {
	path    string
	content string
	err     error
}

func huhReadFileContent(path string) tea.Cmd {
	return func() tea.Msg {
		b, err := os.ReadFile(path)
		if err != nil {
			return huhFileReadMsg{path: path, err: err}
		}
		return huhFileReadMsg{path: path, content: string(b)}
	}
}

// NewHuhFilePickerScreen creates a file picker using Huh's FilePicker field.
// The isDark parameter should be false initially; the correct value will be
// set via SetTheme when the screen is pushed onto the stack.
func NewHuhFilePickerScreen(startDir string, isDark bool, appName string) *HuhFilePickerScreen {
	selectedPath := new(string) // Use pointer to share value across form rebuilds

	// formBuilder is a function that can rebuild the form when needed
	formBuilder := func() *huh.Form {
		return huh.NewForm(
			huh.NewGroup(
				huh.NewFilePicker().
					Title("Select a file").
					CurrentDirectory(startDir).
					Value(selectedPath),
			),
		).WithShowHelp(true).WithShowErrors(true)
	}

	onSubmit := func() tea.Cmd {
		if *selectedPath != "" {
			return huhReadFileContent(*selectedPath)
		}
		return nil
	}

	onAbort := func() tea.Cmd {
		return nav.Pop() // ESC goes back to previous screen
	}

	fs := newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)

	return &HuhFilePickerScreen{
		FormScreen: fs,
	}
}

// Update handles incoming messages and returns an updated screen and command.
// It handles file reading and pushes a DetailScreen when a file is selected.
func (s *HuhFilePickerScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case huhFileReadMsg:
		if msg.err != nil {
			// Stay on picker on error - the form will display the error
			return s, nil
		}
		// Push DetailScreen with the file content
		return s, nav.Push(NewDetailScreen(msg.path, msg.content, s.IsDark, s.AppName))
	}

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
func (s *HuhFilePickerScreen) SetTheme(isDark bool) {
	s.FormScreen.SetTheme(isDark)
}
