package theme

// State represents the complete theme state.
type State struct {
	Name    string  // theme name (e.g., "ocean", "forest")
	IsDark  bool    // dark/light mode
	Palette Palette // cached palette (computed once)
	Width   int     // for width-dependent styles
}

// Themeable is implemented by components that need theme updates.
type Themeable interface {
	ApplyTheme(state State)
}

// Sizable is for components that need dimension updates.
type Sizable interface {
	SetSize(width, height int)
}
