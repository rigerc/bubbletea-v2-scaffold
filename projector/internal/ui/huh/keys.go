// Package huh provides integration adapters for Huh-v2 forms.
package huh

import (
	"charm.land/huh/v2"
	"projector/internal/ui/keys"
)

// KeyMap creates a Huh keymap that preserves global bindings.
// It starts with Huh's default keymap and overrides the Quit binding
// to match the application's global quit key (Ctrl+C).
func KeyMap(global keys.GlobalKeyMap) *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = global.Quit
	return km
}
