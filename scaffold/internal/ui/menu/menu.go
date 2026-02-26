// Package menu provides a menu list component using bubbles-v2 list.
package menu

import (
	"scaffold/internal/ui/theme"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// Item represents a menu item.
type Item struct {
	title       string
	description string
	screenID    string // identifier for navigation
}

// NewItem creates a new menu item.
func NewItem(title, description, screenID string) Item {
	return Item{
		title:       title,
		description: description,
		screenID:    screenID,
	}
}

// FilterValue implements list.Item.
func (i Item) FilterValue() string { return i.title }

// Title implements list.DefaultItem.
func (i Item) Title() string { return i.title }

// Description implements list.DefaultItem.
func (i Item) Description() string { return i.description }

// ScreenID returns the screen identifier for navigation.
func (i Item) ScreenID() string { return i.screenID }

// keyMap defines keybindings for the menu.
type keyMap struct {
	Select key.Binding
	Up     key.Binding
	Down   key.Binding
}

// defaultKeyMap returns the default key bindings.
func defaultKeyMap() keyMap {
	return keyMap{
		Select: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter/l", "select"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}
}

// ShortHelp implements help.KeyMap.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select}
}

// FullHelp implements help.KeyMap.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Select}}
}

// SelectionMsg is emitted when a menu item is selected.
type SelectionMsg struct {
	Item Item
}

// Model is the menu component.
type Model struct {
	list     list.Model
	delegate list.DefaultDelegate
	keys     keyMap
	ready    bool
	width    int
	height   int
	isDark   bool
}

// New creates a new menu model.
func New() Model {
	return Model{
		keys: defaultKeyMap(),
	}
}

// SetSize sets the menu dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	if m.ready {
		m.list.SetSize(width, height)
	}
	return m
}

// SetItems sets the menu items.
func (m Model) SetItems(items []Item) Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	if m.ready {
		m.list.SetItems(listItems)
	} else {
		// Create delegate with theme styles if isDark is set
		m.delegate = list.NewDefaultDelegate()
		p := theme.NewPalette(m.isDark)
		m.delegate.Styles = theme.ListItemStyles(p)

		m.list = list.New(listItems, m.delegate, m.width, m.height)
		m.list.Title = "Menu"
		m.list.Styles = theme.ListStyles(p)
		m.list.SetShowStatusBar(false)
		m.list.SetShowPagination(false)
		m.list.SetShowFilter(false)
		m.list.SetShowTitle(false)
		m.list.SetShowHelp(false) // Hide list's help, use global help
		m.list.DisableQuitKeybindings()
		m.ready = true
	}
	return m
}

// SetStyles sets the menu styles based on dark/light mode.
func (m Model) SetStyles(isDark bool) Model {
	m.isDark = isDark
	if m.ready {
		p := theme.NewPalette(isDark)
		m.list.Styles = theme.ListStyles(p)
		m.delegate.Styles = theme.ListItemStyles(p)
	}
	return m
}

// Init initializes the menu.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the menu.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Handle selection
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(keyMsg, m.keys.Select) {
			if item, ok := m.list.SelectedItem().(Item); ok {
				return m, func() tea.Msg {
					return SelectionMsg{Item: item}
				}
			}
		}
	}

	return m, cmd
}

// View renders the menu.
func (m Model) View() string {
	if !m.ready {
		return ""
	}
	return m.list.View()
}

// KeyBindings returns the key bindings for help display.
func (m Model) KeyBindings() []key.Binding {
	return m.keys.ShortHelp()
}

// Keys returns the keyMap for integration with global help.
func (m Model) Keys() keyMap {
	return m.keys
}

// Items returns the current menu items.
func (m Model) Items() []Item {
	if !m.ready {
		return nil
	}
	items := m.list.Items()
	result := make([]Item, len(items))
	for i, item := range items {
		if it, ok := item.(Item); ok {
			result[i] = it
		}
	}
	return result
}

// ItemCount returns the number of items in the menu.
func (m Model) ItemCount() int {
	if !m.ready {
		return 0
	}
	return len(m.list.Items())
}

// RequiredHeight calculates the height needed to display all items.
// Uses the delegate's Height() and Spacing() for accurate calculation.
func (m Model) RequiredHeight() int {
	if !m.ready {
		return 0
	}
	count := len(m.list.Items())
	if count == 0 {
		return 0
	}

	// Get height and spacing from the delegate
	itemHeight := m.delegate.Height()
	spacing := m.delegate.Spacing()

	// Title area: title (1) + blank line (1)
	const titleHeight = 0

	// Calculate total: title + (items * height) + spacing between items
	itemsHeight := count * itemHeight
	totalSpacing := (count - 1) * spacing
	if totalSpacing < 0 {
		totalSpacing = 0
	}

	return titleHeight + itemsHeight + totalSpacing + 1
}
