package screens

import (
	"os"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	appkeys "template-v2-enhanced/internal/ui/keys"
	"template-v2-enhanced/internal/ui/nav"
)

// filePickerHelpKeys combines the filepicker and global key maps for help display.
type filePickerHelpKeys struct {
	fp  filepicker.KeyMap
	app appkeys.GlobalKeyMap
}

func (k filePickerHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.fp.Up, k.fp.Down, k.fp.Open, k.app.Back}
}

func (k filePickerHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.fp.Up, k.fp.Down, k.fp.GoToTop, k.fp.GoToLast},
		{k.fp.Back, k.fp.Open, k.fp.Select, k.app.Back},
	}
}

// fileReadMsg carries the result of reading a selected file from disk.
type fileReadMsg struct {
	path    string
	content string
	err     error
}

func readFileContent(path string) tea.Cmd {
	return func() tea.Msg {
		b, err := os.ReadFile(path)
		if err != nil {
			return fileReadMsg{path: path, err: err}
		}
		return fileReadMsg{path: path, content: string(b)}
	}
}

// FilePickerScreen lets the user browse the filesystem and open text files
// in a DetailScreen. It implements nav.Screen and nav.Themeable.
type FilePickerScreen struct {
	ScreenBase
	fp        filepicker.Model
	statusMsg string
}

// NewFilePickerScreen creates a new FilePickerScreen rooted at startDir.
func NewFilePickerScreen(startDir string, isDark bool, appName string) *FilePickerScreen {
	fp := filepicker.New()
	fp.CurrentDirectory = startDir
	fp.AutoHeight = false

	// Remove ESC from the filepicker's own Back binding so ESC navigates
	// the nav stack instead of the directory tree.
	fp.KeyMap.Back.SetKeys("h", "backspace", "left")
	fp.KeyMap.Back.SetHelp("h", "back")

	return &FilePickerScreen{
		ScreenBase: NewBase(isDark, appName),
		fp:         fp,
	}
}

// Init reads the starting directory.
func (s *FilePickerScreen) Init() tea.Cmd {
	return s.fp.Init()
}

// Update handles messages and returns an updated screen and command.
func (s *FilePickerScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height
		s.updateSize()

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.Keys.Help):
			s.Help.ShowAll = !s.Help.ShowAll
			return s, nil
		case key.Matches(msg, s.Keys.Back):
			return s, nav.Pop()
		}

	case fileReadMsg:
		if msg.err != nil {
			s.statusMsg = "Error: " + msg.err.Error()
			return s, nil
		}
		return s, nav.Push(NewDetailScreen(msg.path, msg.content, s.IsDark, s.AppName))
	}

	var cmd tea.Cmd
	s.fp, cmd = s.fp.Update(msg)

	if didSelect, path := s.fp.DidSelectFile(msg); didSelect {
		s.statusMsg = "Opening " + path + "…"
		return s, tea.Batch(cmd, readFileContent(path))
	}
	if didSelect, path := s.fp.DidSelectDisabledFile(msg); didSelect {
		s.statusMsg = path + " is not a valid selection"
		return s, cmd
	}

	return s, cmd
}

// View renders the filepicker screen.
func (s *FilePickerScreen) View() string {
	helpKeys := filePickerHelpKeys{fp: s.fp.KeyMap, app: s.Keys}
	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			s.statusView(),
			s.fp.View(),
			s.RenderHelp(helpKeys),
		),
	)
}

// SetTheme updates styles for the current terminal background.
// Implements nav.Themeable.
func (s *FilePickerScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
}

// statusView renders a one-line bar showing the current directory and any status message.
func (s *FilePickerScreen) statusView() string {
	dir := s.Theme.Subtle.Render("  " + s.fp.CurrentDirectory)
	if s.statusMsg != "" {
		msg := s.Theme.StatusMessage.Render("  " + s.statusMsg)
		return lipgloss.JoinVertical(lipgloss.Left, dir, msg)
	}
	return dir
}

// updateSize recalculates the filepicker height from the remaining space after
// the header and status bar.
func (s *FilePickerScreen) updateSize() {
	if !s.IsSized() {
		return
	}
	_, frameV := s.Theme.App.GetFrameSize()
	headerH := lipgloss.Height(s.HeaderView())
	statusH := lipgloss.Height(s.statusView())

	fpH := s.Height - frameV - headerH - statusH
	if cap := s.Height / 3; fpH > cap {
		fpH = cap
	}
	if fpH < 4 {
		fpH = 4
	}
	s.fp.SetHeight(fpH)
}
