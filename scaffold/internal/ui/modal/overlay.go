package modal

import "charm.land/lipgloss/v2"

// Overlay centers popup over a w×h area using lipgloss.Place.
// The base content is not currently composited — popup is shown on a
// whitespace backdrop. Pass the rendered base if dimming is added later.
func Overlay(base, popup string, w, h int) string {
	_ = base // reserved for future dimming/compositing
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, popup)
}
