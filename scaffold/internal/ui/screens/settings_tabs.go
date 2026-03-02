package screens

import (
	"strings"

	"scaffold/config"

	"charm.land/lipgloss/v2"
)

// tabStyles holds lipgloss styles for the group tab bar.
type tabStyles struct {
	active   lipgloss.Style
	inactive lipgloss.Style
	tabBar   lipgloss.Style
}

// initTabStyles creates tab styles from the current theme palette.
func (s *Settings) initTabStyles() {
	p := s.Palette()
	s.tabStyles = tabStyles{
		active: lipgloss.NewStyle().
			Foreground(p.OnPrimary).
			Background(p.Primary).
			Bold(true).
			Padding(0, 1),
		inactive: lipgloss.NewStyle().
			Foreground(p.ForegroundSubtle).
			Padding(0, 1),
		tabBar: lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1),
	}
}

// renderTabBar renders the horizontal tab bar with group labels.
func (s *Settings) renderTabBar() string {
	if len(s.groups) <= 1 {
		return ""
	}
	// Sync currentGroup with form's actual state by checking focused field
	s.syncCurrentGroup()

	var tabs []string
	for i, g := range s.groups {
		if i == s.currentGroup {
			tabs = append(tabs, s.tabStyles.active.Render(g.Label))
		} else {
			tabs = append(tabs, s.tabStyles.inactive.Render(g.Label))
		}
	}
	return s.tabStyles.tabBar.Render(strings.Join(tabs, " "))
}

// syncCurrentGroup updates currentGroup to match the form's actual focused group.
func (s *Settings) syncCurrentGroup() {
	field := s.form.GetFocusedField()
	if field == nil {
		return
	}

	// Try to get the field's key
	keyer, ok := field.(interface{ GetKey() string })
	if !ok {
		return
	}
	key := keyer.GetKey()

	// Find which group contains this key
	for i, g := range s.groups {
		for _, f := range g.Fields {
			if f.Key == key {
				if s.currentGroup != i {
					s.currentGroup = i
				}
				return
			}
		}
	}
}

// requiredHeight calculates the minimum height needed to display the form.
func requiredHeight(groups []config.GroupMeta) int {
	const (
		fieldHeight  = 1
		groupHeader  = 2
		submitHeight = 2
		tabBarHeight = 2
	)

	total := submitHeight
	if len(groups) > 1 {
		total += tabBarHeight
	}
	for _, g := range groups {
		total += groupHeader
		total += len(g.Fields) * fieldHeight
	}
	return total
}
