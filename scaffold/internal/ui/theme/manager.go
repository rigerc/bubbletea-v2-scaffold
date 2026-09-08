package theme

import (
	"sync"

	tea "charm.land/bubbletea/v2"
)

var (
	managerOnce sync.Once
	manager     *Manager
)

// GetManager returns the singleton theme manager.
func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			paletteCache: make(map[string]map[bool]Palette),
		}
	})
	return manager
}

// Manager holds theme state and provides cached palette access.
type Manager struct {
	mu           sync.RWMutex
	state        State
	paletteCache map[string]map[bool]Palette // name -> isDark -> Palette
}

// Init initializes the manager and returns initial theme command.
// If width is 0, no command is returned (will be triggered by first WindowSizeMsg).
func (m *Manager) Init(name string, isDark bool, width int) tea.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = State{
		Name:    name,
		IsDark:  isDark,
		Palette: m.getCachedPalette(name, isDark),
		Width:   width,
	}

	// Don't fire theme update until we have a valid width
	if width > 0 {
		return RequestThemeUpdate(m.state)
	}
	return nil
}

// getCachedPalette returns cached palette or creates and caches one.
func (m *Manager) getCachedPalette(name string, isDark bool) Palette {
	if m.paletteCache[name] == nil {
		m.paletteCache[name] = make(map[bool]Palette)
	}
	if p, ok := m.paletteCache[name][isDark]; ok {
		return p
	}
	p := NewPalette(name, isDark)
	m.paletteCache[name][isDark] = p
	return p
}

// SetDarkMode updates dark mode and returns command if changed.
func (m *Manager) SetDarkMode(isDark bool) tea.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.IsDark == isDark {
		return nil
	}
	m.state.IsDark = isDark
	m.state.Palette = m.getCachedPalette(m.state.Name, isDark)
	return RequestThemeUpdate(m.state)
}

// SetThemeName updates theme name and returns command if changed.
func (m *Manager) SetThemeName(name string) tea.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Name == name {
		return nil
	}
	m.state.Name = name
	m.state.Palette = m.getCachedPalette(name, m.state.IsDark)
	return RequestThemeUpdate(m.state)
}

// SetWidth updates width and returns command if changed.
func (m *Manager) SetWidth(width int) tea.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Width == width {
		return nil
	}
	m.state.Width = width
	return RequestThemeUpdate(m.state)
}

// State returns current theme state (read-only).
func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}
