package theme

// ThemeAware provides reusable theme state holder for components.
// Embed this struct to get theme state storage and accessors.
type ThemeAware struct {
	themeState State
}

// ApplyThemeState updates the stored theme state.
// Note: This method has a pointer receiver. Types embedding ThemeAware
// should use pointer receivers for their ApplyTheme methods if they need
// to call this, OR handle state storage differently.
func (t *ThemeAware) ApplyThemeState(state State) {
	t.themeState = state
}

// ThemeState returns current theme state.
func (t *ThemeAware) ThemeState() State {
	return t.themeState
}

// Palette is a convenience accessor.
func (t *ThemeAware) Palette() Palette {
	return t.themeState.Palette
}

// IsDark is a convenience accessor.
func (t *ThemeAware) IsDark() bool {
	return t.themeState.IsDark
}

// ThemeName is a convenience accessor.
func (t *ThemeAware) ThemeName() string {
	return t.themeState.Name
}

// ThemeWidth is a convenience accessor.
func (t *ThemeAware) ThemeWidth() int {
	return t.themeState.Width
}

// SetThemeState allows setting theme state from a value receiver method.
func (t *ThemeAware) SetThemeState(state State) {
	t.themeState = state
}
