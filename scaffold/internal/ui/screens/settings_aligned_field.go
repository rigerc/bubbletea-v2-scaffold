package screens

import (
	"io"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// fieldAlignment holds shared column-alignment data for inline settings fields.
// Both alignedField and inlineSelect embed this to render title, description,
// and control in fixed-width columns.
type fieldAlignment struct {
	label  string // display label (without ":")
	desc   string // description text
	titleW int    // column width for label text
	descW  int    // column width for description
}

// columnGap is the number of space characters between alignment columns.
const columnGap = 4

// renderAligned joins title, description, and control content into a
// horizontally aligned row with fixed-width columns and spacing gaps.
func (a *fieldAlignment) renderAligned(styles *huh.FieldStyles, content string) string {
	title := styles.Title.Width(a.titleW).MarginRight(columnGap).Render(a.label)
	desc := styles.Description.Width(a.descW).MarginRight(columnGap).Render(a.desc)
	return lipgloss.JoinHorizontal(lipgloss.Left, title, desc, content)
}

// alignmentOverhead returns the total horizontal space consumed by the
// title column, description column, and inter-column gaps.
func (a *fieldAlignment) alignmentOverhead() int {
	return a.titleW + a.descW + columnGap*2
}

// alignedField wraps a huh.Field (Input, Confirm, Note) to render with
// column-aligned title, description, and control. The inner field is created
// with empty Title/Description so it only renders its control widget.
type alignedField struct {
	inner     huh.Field
	alignment fieldAlignment
	width     int
	focused   bool
	theme     huh.Theme
	hasDarkBg bool
}

// newAlignedField creates an aligned wrapper around an inner huh.Field.
func newAlignedField(label, desc string, titleW, descW int, inner huh.Field) *alignedField {
	return &alignedField{
		inner: inner,
		alignment: fieldAlignment{
			label:  label,
			desc:   desc,
			titleW: titleW,
			descW:  descW,
		},
	}
}

func (f *alignedField) Init() tea.Cmd {
	return f.inner.Init()
}

func (f *alignedField) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	if bgMsg, ok := msg.(tea.BackgroundColorMsg); ok {
		f.hasDarkBg = bgMsg.IsDark()
	}
	m, cmd := f.inner.Update(msg)
	if field, ok := m.(huh.Field); ok {
		f.inner = field
	}
	return f, cmd
}

func (f *alignedField) View() string {
	styles := f.activeStyles()
	controlView := f.inner.View()
	aligned := f.alignment.renderAligned(styles, controlView)
	return styles.Base.Width(f.width).Render(aligned)
}

func (f *alignedField) Focus() tea.Cmd {
	f.focused = true
	return f.inner.Focus()
}

func (f *alignedField) Blur() tea.Cmd {
	f.focused = false
	return f.inner.Blur()
}

func (f *alignedField) KeyBinds() []key.Binding {
	return f.inner.KeyBinds()
}

func (f *alignedField) Error() error {
	return f.inner.Error()
}

func (f *alignedField) Skip() bool {
	return f.inner.Skip()
}

func (f *alignedField) Zoom() bool {
	return f.inner.Zoom()
}

func (f *alignedField) WithTheme(theme huh.Theme) huh.Field {
	f.theme = theme
	// Create a bare theme for the inner field: strip Base/Card so the inner
	// field renders only its control widget without borders or background.
	// The alignedField applies the real Base around the entire aligned row.
	bare := huh.ThemeFunc(func(isDark bool) *huh.Styles {
		s := theme.Theme(isDark)
		s.Focused.Base = lipgloss.NewStyle()
		s.Focused.Card = lipgloss.NewStyle()
		s.Blurred.Base = lipgloss.NewStyle()
		s.Blurred.Card = lipgloss.NewStyle()
		return s
	})
	f.inner = f.inner.WithTheme(bare)
	return f
}

func (f *alignedField) WithKeyMap(k *huh.KeyMap) huh.Field {
	f.inner = f.inner.WithKeyMap(k)
	return f
}

func (f *alignedField) WithWidth(width int) huh.Field {
	f.width = width
	styles := f.activeStyles()
	baseFrame := styles.Base.GetHorizontalFrameSize()
	controlWidth := width - baseFrame - f.alignment.alignmentOverhead()
	if controlWidth < 10 {
		controlWidth = 10
	}
	f.inner = f.inner.WithWidth(controlWidth)
	return f
}

func (f *alignedField) WithHeight(height int) huh.Field {
	f.inner = f.inner.WithHeight(height)
	return f
}

func (f *alignedField) WithPosition(p huh.FieldPosition) huh.Field {
	f.inner = f.inner.WithPosition(p)
	return f
}

func (f *alignedField) GetKey() string {
	return f.inner.GetKey()
}

func (f *alignedField) GetValue() any {
	return f.inner.GetValue()
}

func (f *alignedField) Run() error {
	return f.inner.Run()
}

func (f *alignedField) RunAccessible(w io.Writer, r io.Reader) error {
	return f.inner.RunAccessible(w, r)
}

func (f *alignedField) activeStyles() *huh.FieldStyles {
	theme := f.theme
	if theme == nil {
		theme = huh.ThemeFunc(huh.ThemeCharm)
	}
	if f.focused {
		return &theme.Theme(f.hasDarkBg).Focused
	}
	return &theme.Theme(f.hasDarkBg).Blurred
}

var _ huh.Field = (*alignedField)(nil)
