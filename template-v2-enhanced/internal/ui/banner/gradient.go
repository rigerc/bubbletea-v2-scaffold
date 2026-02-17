// Package banner provides figlet-go ASCII art rendering with gradient color support.
package banner

import "math/rand/v2"

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
		"FF4E50", // warm red
		"F9845B", // orange-red
		"FC913A", // orange
		"F5D063", // yellow-orange
		"FECA57", // yellow
		"FFB3C6", // soft pink
		"FF9FF3", // pink-lavender
	}}

	GradientOcean = Gradient{Name: "ocean", Colors: []string{
		"023E8A", // dark navy
		"0077B6", // deep blue
		"0096C7", // ocean blue
		"00B4D8", // medium blue
		"48CAE4", // light blue
		"90E0EF", // pale blue
		"ADE8F4", // very light blue
	}}

	GradientForest = Gradient{Name: "forest", Colors: []string{
		"0D3B2E", // very dark teal
		"134E5E", // dark teal
		"1B6B3A", // forest green
		"3A9653", // medium green
		"71B280", // sage green
		"A8D8A8", // light green
	}}

	GradientNeon = Gradient{Name: "neon", Colors: []string{
		"FF006E", // hot pink
		"FF00CC", // neon pink
		"FF00FF", // magenta
		"9900FF", // purple
		"0066FF", // blue
		"00CCFF", // light cyan
		"00FFFF", // cyan
	}}

	GradientAurora = Gradient{Name: "aurora", Colors: []string{
		"00F5FF", // bright cyan
		"00C6FF", // sky blue
		"0072FF", // deep blue
		"4361EE", // indigo
		"7209B7", // deep violet
		"B5179E", // magenta
		"F72585", // hot pink
	}}

	GradientFire = Gradient{Name: "fire", Colors: []string{
		"7B0D1E", // dark crimson
		"C1121F", // deep red
		"F12711", // bright red
		"F5431A", // red-orange
		"F5AF19", // amber
		"FFF176", // warm yellow
	}}

	GradientPastel = Gradient{Name: "pastel", Colors: []string{
		"FFB3BA", // pastel pink
		"FFCBA4", // peach
		"FFDFBA", // light peach
		"FFFFBA", // lemon
		"BAFFC9", // mint
		"BAE1FF", // sky blue
		"C9B3FF", // lavender
	}}

	GradientMono = Gradient{Name: "monochrome", Colors: []string{
		"FFFFFF", // white
		"E0E0E0", // light grey
		"BBBBBB", // lighter mid
		"999999", // mid grey
		"777777", // darker mid
		"555555", // dark grey
		"333333", // near black
	}}

	GradientVaporwave = Gradient{Name: "vaporwave", Colors: []string{
		"FF71CE", // hot pink
		"FF9DE2", // light pink
		"D4A5F5", // lilac
		"B967FF", // purple
		"8B5CF6", // deep purple
		"3ABFF8", // sky blue
		"01CDFE", // cyan
	}}

	GradientMatrix = Gradient{Name: "matrix", Colors: []string{
		"001200", // near black
		"002200", // very dark green
		"003B00", // dark green
		"007300", // medium green
		"00C800", // green
		"00FF41", // bright green
		"7FFF7F", // light green
	}}
)

var allGradients = []Gradient{
	GradientSunset, GradientOcean, GradientForest, GradientNeon,
	GradientAurora, GradientFire, GradientPastel, GradientMono,
	GradientVaporwave, GradientMatrix,
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
