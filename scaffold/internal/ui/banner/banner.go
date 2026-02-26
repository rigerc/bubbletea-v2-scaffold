// Package banner provides figlet-go ASCII art rendering with gradient color support.
// This is a simplified unified version without safefonts, spacing/kerning, or background features.
package banner

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"strings"

	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/lsferreira42/figlet-go/figlet"
)

// ansiColors maps lowercase ANSI color names to their figlet constants.
var ansiColors = map[string]figlet.Color{
	"black":   figlet.ColorBlack,
	"red":     figlet.ColorRed,
	"green":   figlet.ColorGreen,
	"yellow":  figlet.ColorYellow,
	"blue":    figlet.ColorBlue,
	"magenta": figlet.ColorMagenta,
	"cyan":    figlet.ColorCyan,
	"white":   figlet.ColorWhite,
}

// ansiColorNames is a stable slice used for random ANSI color selection.
var ansiColorNames = []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

// resolveColor converts a color string to a figlet.Color.
// It accepts ANSI color names (case-insensitive) or hex strings (with or without '#').
func resolveColor(s string) (figlet.Color, error) {
	if c, ok := ansiColors[strings.ToLower(s)]; ok {
		return c, nil
	}
	hex := strings.TrimPrefix(s, "#")
	tc, err := figlet.NewTrueColorFromHexString(hex)
	if err != nil {
		return nil, fmt.Errorf("banner: unrecognised color %q (use an ANSI name or hex value)", s)
	}
	return tc, nil
}

// Gradient holds a named set of hex color stops for figlet-go TrueColor rendering.
// Colors are hex strings without '#', e.g. "FF6B6B".
// figlet-go cycles through the stops across rendered characters; more stops
// produce smoother-looking transitions.
type Gradient struct {
	Name   string
	Colors []string
}

// Predefined gradients — each uses 6–7 stops for gradual color transitions.
var (
	GradientSunset = Gradient{Name: "sunset", Colors: []string{
		"FF4E50", "F9845B", "FC913A", "F5D063", "FECA57", "FFB3C6", "FF9FF3",
	}}

	GradientOcean = Gradient{Name: "ocean", Colors: []string{
		"023E8A", "0077B6", "0096C7", "00B4D8", "48CAE4", "90E0EF", "ADE8F4",
	}}

	GradientForest = Gradient{Name: "forest", Colors: []string{
		"0D3B2E", "134E5E", "1B6B3A", "2D8B4E", "3A9653", "5DB364", "71B280",
	}}

	GradientNeon = Gradient{Name: "neon", Colors: []string{
		"FF006E", "FF00CC", "FF00FF", "9900FF", "0066FF", "00CCFF", "00FFFF",
	}}

	GradientAurora = Gradient{Name: "aurora", Colors: []string{
		"00F5FF", "00C6FF", "0072FF", "4361EE", "7209B7", "B5179E", "F72585",
	}}

	GradientFire = Gradient{Name: "fire", Colors: []string{
		"7B0D1E", "C1121F", "F12711", "F5431A", "F76B1C", "F5AF19", "FFF176",
	}}

	GradientPastel = Gradient{Name: "pastel", Colors: []string{
		"FFB3BA", "FFCBA4", "FFDFBA", "FFFFBA", "BAFFC9", "BAE1FF", "C9B3FF",
	}}

	GradientMono = Gradient{Name: "monochrome", Colors: []string{
		"FFFFFF", "E0E0E0", "BBBBBB", "999999", "777777", "555555", "333333",
	}}

	GradientVaporwave = Gradient{Name: "vaporwave", Colors: []string{
		"FF71CE", "FF9DE2", "D4A5F5", "B967FF", "8B5CF6", "3ABFF8", "01CDFE",
	}}

	GradientMatrix = Gradient{Name: "matrix", Colors: []string{
		"001200", "002200", "003B00", "007300", "00C800", "00FF41", "7FFF7F",
	}}

	GradientMind = Gradient{Name: "mind", Colors: []string{
		"473B7B", "3D5A80", "3584A7", "2CA58D", "30D2BE", "5BE0CA", "7EEEE3",
	}}

	GradientRainbow = Gradient{Name: "rainbow", Colors: []string{
		"FF0000", "FF7F00", "FFFF00", "00FF00", "0000FF", "4B0082", "9400D3",
	}}

	GradientGalaxy = Gradient{Name: "galaxy", Colors: []string{
		"360033", "2A0040", "1F004D", "14005A", "090067", "0B8793", "10A99F",
	}}

	GradientLunar = Gradient{Name: "lunar", Colors: []string{
		"0F0C29", "1E1A4A", "302B63", "3D3168", "24243E", "2D2B52", "38385C",
	}}

	GradientPhoenix = Gradient{Name: "phoenix", Colors: []string{
		"F83600", "FA4E1A", "FC681D", "FD8620", "F9A423", "F9D423", "FCDF57",
	}}

	GradientSpirit = Gradient{Name: "spirit", Colors: []string{
		"B92B27", "A83236", "963D45", "5C4D7D", "1565C0", "1976D2", "42A5F5",
	}}

	GradientCherry = Gradient{Name: "cherry", Colors: []string{
		"EB3349", "DC2B42", "D0303B", "F45C43", "F86B4F", "FA8A75", "FFA99C",
	}}

	GradientWaves = Gradient{Name: "waves", Colors: []string{
		"667EEA", "5E72D9", "5561C9", "6B5AA1", "764BA2", "8559B3", "9469C4",
	}}

	GradientDreamy = Gradient{Name: "dreamy", Colors: []string{
		"FDA085", "FBB876", "F6D365", "7ED6DF", "4FACFE", "2CE0F5", "00F2FE",
	}}

	GradientMagic = Gradient{Name: "magic", Colors: []string{
		"59C173", "4DB062", "7B68EE", "A17FE0", "8B5CF6", "7C3AED", "5D26C1",
	}}

	GradientElectric = Gradient{Name: "electric", Colors: []string{
		"4776E6", "3D68D1", "6459BC", "8E54E9", "7E3BBF", "6C2795", "5A1A6B",
	}}

	GradientVenom = Gradient{Name: "venom", Colors: []string{
		"8360C3", "7351B0", "64429D", "2EBF91", "4CC98F", "6AD9A5", "8AE8B9",
	}}

	GradientMirage = Gradient{Name: "mirage", Colors: []string{
		"16222A", "1E3340", "264556", "2E566C", "3A6073", "4A7885", "5A8F97",
	}}

	GradientRebel = Gradient{Name: "rebel", Colors: []string{
		"F093FB", "E87FEC", "DC6ADD", "D057CE", "C246BF", "F5576C", "E84A5F",
	}}

	GradientDrift = Gradient{Name: "drift", Colors: []string{
		"00D2FF", "00B8E6", "009ECC", "3A7BD5", "2D6BB5", "205995", "134775",
	}}

	GradientBloom = Gradient{Name: "bloom", Colors: []string{
		"FFECD2", "FFDFB8", "FFC99F", "FFB386", "FC9D6D", "FCB69F", "FFA58C",
	}}

	GradientAtlas = Gradient{Name: "atlas", Colors: []string{
		"FEAC5E", "EFA04D", "D4883C", "C779D0", "9B5DB5", "4BC0C8", "2EC4C2",
	}}
)

var allGradients = []Gradient{
	GradientSunset, GradientOcean, GradientForest, GradientNeon,
	GradientAurora, GradientFire, GradientPastel, GradientMono,
	GradientVaporwave, GradientMatrix, GradientMind, GradientRainbow,
	GradientGalaxy, GradientLunar, GradientPhoenix, GradientSpirit,
	GradientCherry, GradientWaves, GradientDreamy, GradientMagic,
	GradientElectric, GradientVenom, GradientMirage, GradientRebel,
	GradientDrift, GradientBloom, GradientAtlas,
}

// RandomGradient returns a randomly selected predefined gradient.
func RandomGradient() Gradient {
	return allGradients[rand.IntN(len(allGradients))]
}

// RandomFont returns a randomly selected font from the full figlet-go list.
func RandomFont() string {
	fonts := figlet.ListFonts()
	return fonts[rand.IntN(len(fonts))]
}

// GradientThemed builds a *Gradient that flows from primary to secondary by
// blending them across 7 stops with lipgloss.Blend1D.
// Returns *Gradient so it can be assigned inline in banner.Config{Gradient: ...}.
// Pass palette.Primary and palette.Secondary to derive a theme-matched gradient.
func GradientThemed(primary, secondary color.Color) *Gradient {
	stops := lipgloss.Blend1D(7, primary, secondary)
	hexes := make([]string, len(stops))
	for i, c := range stops {
		cf, ok := colorful.MakeColor(c)
		if !ok {
			hexes[i] = "888888"
			continue
		}
		hexes[i] = cf.Hex()[1:] // strip leading '#'
	}
	return &Gradient{Name: "themed", Colors: hexes}
}

// Config defines parameters for rendering an ASCII banner.
type Config struct {
	// Text is the string to render as ASCII art. Required.
	Text string

	// Font is the figlet font name. Empty string selects a random font.
	Font string

	// FontDir sets a custom font directory for loading .flf files from disk.
	// Leave empty to use the 145 fonts embedded in figlet-go.
	FontDir string

	// Width sets the terminal width for rendering. Clamped to minimum 20.
	Width int

	// Justification controls horizontal alignment.
	//   -1  auto (font decides)
	//    0  left (default)
	//    1  center
	//    2  right
	Justification int

	// RightToLeft controls text direction.
	//   -1  auto (font decides)
	//    0  left-to-right (default)
	//    1  right-to-left
	RightToLeft int

	// Color is a single color applied uniformly to all characters.
	// Accepts ANSI names (black, red, green, yellow, blue, magenta, cyan, white)
	// or hex values with or without '#' (e.g. "FF0000", "#FF0000").
	// Mutually exclusive with Gradient, RandomGradient, and RandomColor.
	Color string

	// Gradient is a specific color gradient to apply across characters.
	// Mutually exclusive with Color, RandomGradient, and RandomColor.
	Gradient *Gradient

	// RandomGradient picks a random predefined gradient.
	// Mutually exclusive with Color, Gradient, and RandomColor.
	// When no color option is set, this is the default behaviour.
	RandomGradient bool

	// RandomColor picks a random ANSI color applied uniformly to all characters.
	// Mutually exclusive with Color, Gradient, and RandomGradient.
	RandomColor bool

	// Parser selects the output format. Valid values: "terminal-color" (default),
	// "terminal" (plain text, no ANSI), "html".
	Parser string
}

// Render renders ASCII art for the given config.
// Returns ANSI-colored (or plain/HTML) figlet output ready for display.
func Render(cfg Config) (string, error) {
	// Resolve font
	font := cfg.Font
	if font == "" {
		font = RandomFont()
	}

	// Resolve colors — exactly one color source may be set.
	colorSources := 0
	if cfg.Color != "" {
		colorSources++
	}
	if cfg.Gradient != nil {
		colorSources++
	}
	if cfg.RandomGradient {
		colorSources++
	}
	if cfg.RandomColor {
		colorSources++
	}
	if colorSources > 1 {
		return "", fmt.Errorf("banner: Color, Gradient, RandomGradient, and RandomColor are mutually exclusive")
	}

	var colors []figlet.Color
	switch {
	case cfg.Color != "":
		tc, err := resolveColor(cfg.Color)
		if err != nil {
			return "", err
		}
		colors = []figlet.Color{tc}
	case cfg.RandomColor:
		name := ansiColorNames[rand.IntN(len(ansiColorNames))]
		colors = []figlet.Color{ansiColors[name]}
	default:
		// cfg.Gradient set, cfg.RandomGradient set, or nothing set — all use a gradient.
		grad := cfg.Gradient
		if grad == nil {
			rg := RandomGradient()
			grad = &rg
		}
		colors = make([]figlet.Color, len(grad.Colors))
		for i, hex := range grad.Colors {
			tc, err := figlet.NewTrueColorFromHexString(hex)
			if err != nil {
				return "", fmt.Errorf("invalid hex %q in gradient %q: %w", hex, grad.Name, err)
			}
			colors[i] = tc
		}
	}

	// Resolve width
	width := cfg.Width
	if width < 20 {
		width = 80 // default terminal width
	}

	// Resolve parser
	parser := cfg.Parser
	if parser == "" {
		parser = "terminal-color"
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

	result, err := figlet.Render(cfg.Text, opts...)
	if err != nil {
		return cfg.Text, fmt.Errorf("figlet render failed (font=%q): %w", font, err)
	}

	return result, nil
}
