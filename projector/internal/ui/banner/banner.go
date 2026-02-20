package banner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lsferreira42/figlet-go/figlet"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func measureString(s string) (int, int) {
	lines := strings.Split(s, "\n")
	height := len(lines)
	if height > 0 && lines[height-1] == "" {
		height--
	}
	width := 0
	for _, line := range lines {
		w := len(stripANSI(line))
		if w > width {
			width = w
		}
	}
	return width, height
}

func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// SpacingMode controls how figlet renders character spacing.
// The modes are mutually exclusive; SpacingDefault uses the font's built-in setting.
type SpacingMode int

const (
	SpacingDefault     SpacingMode = iota // font's built-in smushing (zero value)
	SpacingKerning                        // letters touch but do not overlap
	SpacingFullWidth                      // no smushing — full character width
	SpacingSmushing                       // force smushing regardless of font setting
	SpacingOverlapping                    // overlapping mode
)

// BannerConfig defines all parameters for rendering an ASCII banner.
// Zero values produce sensible defaults: random safe font, random gradient,
// left-justified, terminal-color parser, font's default spacing.
type BannerConfig struct {
	// Text is the string to render as ASCII art. Required.
	Text string

	// Font is the figlet font name. Empty string selects a random font.
	// The pool used for random selection is FontPool (if set) or SafeFonts.
	Font string

	// FontPool is the set of font names to draw from when Font is empty.
	// Nil or empty falls back to SafeFonts.
	FontPool []string

	// FontDir sets a custom font directory for loading .flf files from disk.
	// Leave empty to use the 145 fonts embedded in figlet-go.
	FontDir string

	// Width overrides the terminal width passed to RenderBanner when > 0.
	// Use this to produce a banner narrower than the full terminal, e.g. Width: 60.
	// Clamped to a minimum of 20.
	Width int

	// Justification controls horizontal alignment within the effective width.
	//   -1  auto (font decides)
	//    0  left (default)
	//    1  center
	//    2  right
	Justification int

	// RightToLeft controls text direction.
	//   -1  auto (font decides)
	//    0  left-to-right (default, not passed to figlet)
	//    1  right-to-left
	RightToLeft int

	// Spacing controls character spacing / smushing mode.
	// See the SpacingMode constants. SpacingDefault (0) uses the font's setting.
	Spacing SpacingMode

	// Gradient is the color gradient to apply. Nil selects a random gradient.
	Gradient *Gradient

	// Background enables the gradient's background color.
	// When true, lipgloss is used to render the figlet output with a matching
	// background, ensuring text is always readable.
	Background bool

	// Parser selects the output format. Valid values: "terminal-color" (default),
	// "terminal" (plain text, no ANSI), "html".
	Parser string
}

// RenderBanner renders ASCII art for the given config at the specified terminal width.
// If cfg.Width > 0 it overrides width. The effective width is clamped to ≥ 20.
// Returns ANSI-colored (or plain/HTML) figlet output ready for display.
func RenderBanner(cfg BannerConfig, width int) (string, error) {
	// Resolve font
	font := cfg.Font
	if font == "" {
		font = RandomFontFrom(cfg.FontPool)
	}

	// Resolve gradient
	grad := cfg.Gradient
	if grad == nil {
		rg := RandomGradient()
		grad = &rg
	}

	// Resolve width
	if cfg.Width > 0 {
		width = cfg.Width
	}
	if width < 20 {
		width = 20
	}

	// Resolve parser
	parser := cfg.Parser
	if parser == "" {
		parser = "terminal-color"
	}

	// Build TrueColor values from gradient hex stops
	colors := make([]figlet.Color, len(grad.Colors))
	for i, hex := range grad.Colors {
		tc, err := figlet.NewTrueColorFromHexString(hex)
		if err != nil {
			return "", fmt.Errorf("banner: invalid hex %q in gradient %q: %w", hex, grad.Name, err)
		}
		colors[i] = tc
	}

	opts := []figlet.Option{
		figlet.WithFont(font),
		figlet.WithParser(parser),
		figlet.WithWidth(width),
		figlet.WithColors(colors...),
		figlet.WithJustification(cfg.Justification),
	}

	if cfg.RightToLeft != 0 {
		opts = append(opts, figlet.WithRightToLeft(cfg.RightToLeft))
	}

	if cfg.FontDir != "" {
		opts = append(opts, figlet.WithFontDir(cfg.FontDir))
	}

	switch cfg.Spacing {
	case SpacingKerning:
		opts = append(opts, figlet.WithKerning())
	case SpacingFullWidth:
		opts = append(opts, figlet.WithFullWidth())
	case SpacingSmushing:
		opts = append(opts, figlet.WithSmushing())
	case SpacingOverlapping:
		opts = append(opts, figlet.WithOverlapping())
	}

	result, err := figlet.Render(cfg.Text, opts...)
	if err != nil {
		return cfg.Text, fmt.Errorf("banner: figlet render failed (font=%q): %w", font, err)
	}

	if cfg.Background && grad.BG != "" {
		r, g, b := hexToRGB(grad.BG)
		bgSeq := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
		resetSeq := "\x1b[49m"

		lines := strings.Split(result, "\n")
		figletWidth, figletHeight := measureString(result)

		if figletWidth > 0 && figletHeight > 0 {
			var newLines []string
			for i, line := range lines {
				if i == len(lines)-1 && line == "" {
					continue
				}

				stripped := stripANSI(line)
				padding := ""
				if len(stripped) < figletWidth {
					padding = strings.Repeat(" ", figletWidth-len(stripped))
				}

				// Replace resets with reset + re-apply background
				safeLine := strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSeq)
				safeLine = strings.ReplaceAll(safeLine, "\x1b[m", "\x1b[m"+bgSeq)

				newLines = append(newLines, bgSeq+safeLine+padding+resetSeq)
			}
			result = strings.Join(newLines, "\n") + "\n"
		}
	}

	return result, nil
}

// RandomBanner returns a BannerConfig with a random safe font and random gradient,
// centered (Justification: 1).
func RandomBanner(text string) BannerConfig {
	rg := RandomGradient()
	return BannerConfig{
		Text:          text,
		Font:          RandomSafeFont(),
		Gradient:      &rg,
		Justification: 1,
	}
}

// NamedBanner returns a BannerConfig with explicit font and gradient names,
// centered. Unknown names fall back to random selections.
func NamedBanner(text, fontName, gradientName string) BannerConfig {
	grad, ok := GradientByName(gradientName)
	if !ok {
		grad = RandomGradient()
	}
	font := fontName
	if font == "" {
		font = RandomSafeFont()
	}
	return BannerConfig{
		Text:          text,
		Font:          font,
		Gradient:      &grad,
		Justification: 1,
	}
}
