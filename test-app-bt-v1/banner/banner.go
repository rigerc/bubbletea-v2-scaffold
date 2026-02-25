// Package banner provides figlet-go ASCII art rendering with gradient color support.
// This is a simplified unified version without safefonts, spacing/kerning, or background features.
package banner

import (
	"fmt"
	"math/rand/v2"

	"github.com/lsferreira42/figlet-go/figlet"
)

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

var gradientIndex = func() map[string]Gradient {
	m := make(map[string]Gradient, len(allGradients))
	for _, g := range allGradients {
		m[g.Name] = g
	}
	return m
}()

// AllGradients returns a copy of all predefined gradients.
func AllGradients() []Gradient {
	return append([]Gradient(nil), allGradients...)
}

// GradientByName returns the gradient for the given name.
// The second return value reports whether the name was found.
func GradientByName(name string) (Gradient, bool) {
	g, ok := gradientIndex[name]
	return g, ok
}

// RandomGradient returns a randomly selected predefined gradient.
func RandomGradient() Gradient {
	return allGradients[rand.IntN(len(allGradients))]
}

// AllFonts returns the full list of fonts embedded in figlet-go.
func AllFonts() []string {
	return figlet.ListFonts()
}

// RandomFont returns a randomly selected font from the full figlet-go list.
func RandomFont() string {
	fonts := figlet.ListFonts()
	return fonts[rand.IntN(len(fonts))]
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

	// Gradient is the color gradient to apply. Nil selects a random gradient.
	Gradient *Gradient

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

	// Resolve gradient
	grad := cfg.Gradient
	if grad == nil {
		rg := RandomGradient()
		grad = &rg
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

	// Build TrueColor values from gradient hex stops
	colors := make([]figlet.Color, len(grad.Colors))
	for i, hex := range grad.Colors {
		tc, err := figlet.NewTrueColorFromHexString(hex)
		if err != nil {
			return "", fmt.Errorf("invalid hex %q in gradient %q: %w", hex, grad.Name, err)
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

	result, err := figlet.Render(cfg.Text, opts...)
	if err != nil {
		return cfg.Text, fmt.Errorf("figlet render failed (font=%q): %w", font, err)
	}

	return result, nil
}

// RandomBanner returns a Config with random font and random gradient, centered.
func RandomBanner(text string) Config {
	rg := RandomGradient()
	return Config{
		Text:          text,
		Gradient:      &rg,
		Justification: 1,
	}
}

// NamedBanner returns a Config with explicit font and gradient names, centered.
// Unknown names fall back to random selections.
func NamedBanner(text, fontName, gradientName string) Config {
	grad, ok := GradientByName(gradientName)
	if !ok {
		grad = RandomGradient()
	}
	font := fontName
	if font == "" {
		font = RandomFont()
	}
	return Config{
		Text:          text,
		Font:          font,
		Gradient:      &grad,
		Justification: 1,
	}
}
