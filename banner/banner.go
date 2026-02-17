// Package banner provides a framework for creating ASCII art banners with
// customizable fonts, gradients, and styling.
package banner

import (
	"fmt"
	"io"
	"os"
	"strings"

	"banner/pkg/color"
	"banner/pkg/fonts"
)

// Builder provides a fluent interface for constructing banners.
type Builder struct {
	text          string
	font          string
	size          int
	gradient      color.Gradient
	tagline       string
	version       string
	padding       int
	width         int
	output        io.Writer
	centerText    bool
	showTimestamp bool
}

// New creates a new banner builder with the specified text.
func New(text string) *Builder {
	return &Builder{
		text:          text,
		font:          "block",
		size:          1,
		gradient:      color.GradientAurora,
		padding:       1,
		output:        os.Stdout,
		centerText:    true,
		showTimestamp: false,
	}
}

// Font sets the font style for the banner text.
// Available fonts: "block", "slant", "minimal", "small", "standard".
func (b *Builder) Font(font string) *Builder {
	b.font = font
	return b
}

// Size sets the font size multiplier (1 = normal, 2 = 2x, etc).
// Values less than 1 are treated as 1.
func (b *Builder) Size(size int) *Builder {
	if size < 1 {
		size = 1
	}
	b.size = size
	return b
}

// Gradient sets the gradient by name.
// Available gradients: "sunset", "ocean", "neon", "cyberpunk", "miami",
// "fire", "forest", "galaxy", "retro", "aurora", "mint", "peach",
// "lavender", "gold", "ice", "blood", "matrix", "vaporwave", "rainbow",
// "terminal", "rose", "sky".
func (b *Builder) Gradient(name string) *Builder {
	b.gradient = color.GetGradient(name)
	return b
}

// GradientRGB sets a custom two-color gradient.
func (b *Builder) GradientRGB(start, end color.RGB) *Builder {
	b.gradient = color.NewGradient(start, end)
	return b
}

// GradientHex sets a custom gradient from hex color strings.
func (b *Builder) GradientHex(hexColors ...string) *Builder {
	b.gradient = color.NewMultiGradient(hexColors...)
	return b
}

// GradientCustom sets a custom gradient.
func (b *Builder) GradientCustom(g color.Gradient) *Builder {
	b.gradient = g
	return b
}

// Tagline adds a tagline below the main banner text.
func (b *Builder) Tagline(tagline string) *Builder {
	b.tagline = tagline
	return b
}

// Version adds a version string below the banner.
func (b *Builder) Version(version string) *Builder {
	b.version = version
	return b
}

// Padding sets the internal padding (in lines) of the banner.
func (b *Builder) Padding(p int) *Builder {
	if p < 0 {
		p = 0
	}
	b.padding = p
	return b
}

// Width sets a minimum width for the banner.
func (b *Builder) Width(w int) *Builder {
	if w < 0 {
		w = 0
	}
	b.width = w
	return b
}

// Output sets the output writer (default is os.Stdout).
func (b *Builder) Output(w io.Writer) *Builder {
	b.output = w
	return b
}

// Center sets whether the text should be centered (default is true).
func (b *Builder) Center(center bool) *Builder {
	b.centerText = center
	return b
}

// Build generates the banner string without outputting it.
func (b *Builder) Build() string {
	// Generate ASCII art from text with scaling
	artLines := fonts.GenerateScaled(b.text, b.font, b.size)
	coloredLines := b.gradient.ApplyLines(artLines)

	// Calculate the maximum width needed
	maxWidth := 0
	for _, line := range artLines {
		lineLen := len([]rune(line))
		if lineLen > maxWidth {
			maxWidth = lineLen
		}
	}

	// Apply minimum width if specified
	if b.width > maxWidth {
		maxWidth = b.width
	}

	// Build the banner
	var result strings.Builder
	result.WriteString("\n")

	// Top padding (empty lines)
	for i := 0; i < b.padding; i++ {
		result.WriteString("\n")
	}

	// Main text lines
	for idx, line := range coloredLines {
		actualLen := len([]rune(artLines[idx]))
		var paddedLine string
		if b.centerText {
			paddedLine = centerText(line, maxWidth, actualLen)
		} else {
			paddedLine = line + strings.Repeat(" ", maxWidth-actualLen)
		}
		result.WriteString(paddedLine)
		result.WriteString("\n")
	}

	// Tagline
	if b.tagline != "" {
		result.WriteString("\n")
		taglineColored := b.gradient.Apply(b.tagline)
		var paddedTagline string
		if b.centerText {
			paddedTagline = centerText(taglineColored, maxWidth, len(b.tagline))
		} else {
			paddedTagline = taglineColored + strings.Repeat(" ", maxWidth-len(b.tagline))
		}
		result.WriteString(paddedTagline)
		result.WriteString("\n")
	}

	// Version
	if b.version != "" {
		result.WriteString("\n")
		versionColored := color.Colorize(b.version, color.Gray)
		var paddedVersion string
		if b.centerText {
			paddedVersion = centerText(versionColored, maxWidth, len(b.version))
		} else {
			paddedVersion = versionColored + strings.Repeat(" ", maxWidth-len(b.version))
		}
		result.WriteString(paddedVersion)
		result.WriteString("\n")
	}

	// Bottom padding (empty lines)
	for i := 0; i < b.padding; i++ {
		result.WriteString("\n")
	}

	return result.String()
}

// Render outputs the banner to the configured output writer.
func (b *Builder) Render() error {
	_, err := fmt.Fprint(b.output, b.Build())
	return err
}

// String returns the banner as a string.
func (b *Builder) String() string {
	return b.Build()
}

// centerText centers text within a given width, accounting for ANSI escape codes.
func centerText(text string, totalWidth int, actualLen int) string {
	if actualLen >= totalWidth {
		return text
	}
	padding := (totalWidth - actualLen) / 2
	rightPad := totalWidth - actualLen - padding
	if rightPad < 0 {
		rightPad = 0
	}
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", rightPad)
}

// Quick creates a simple banner with default settings.
func Quick(text string) string {
	return New(text).Build()
}

// QuickWithGradient creates a banner with the specified gradient.
func QuickWithGradient(text string, gradientName string) string {
	return New(text).Gradient(gradientName).Build()
}

// QuickWithFont creates a banner with the specified font.
func QuickWithFont(text string, fontName string) string {
	return New(text).Font(fontName).Build()
}

// QuickWithSize creates a banner with the specified size.
func QuickWithSize(text string, size int) string {
	return New(text).Size(size).Build()
}
