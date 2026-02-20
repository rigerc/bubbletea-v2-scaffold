// Package banner provides figlet-go ASCII art rendering with gradient color support.
package banner

import "math/rand/v2"

// Gradient holds a named set of hex color stops for figlet-go TrueColor rendering.
// Colors are hex strings without '#', e.g. "FF6B6B".
// figlet-go cycles through the stops across rendered characters; more stops
// produce smoother-looking transitions.
// BG is the background color for the rendered text (hex without '#').
type Gradient struct {
	Name   string
	Colors []string
	BG     string
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
	}, BG: "1A0A0A"}

	GradientOcean = Gradient{Name: "ocean", Colors: []string{
		"023E8A", // dark navy
		"0077B6", // deep blue
		"0096C7", // ocean blue
		"00B4D8", // medium blue
		"48CAE4", // light blue
		"90E0EF", // pale blue
		"ADE8F4", // very light blue
	}, BG: "0A1428"}

	GradientForest = Gradient{Name: "forest", Colors: []string{
		"0D3B2E", // very dark teal
		"134E5E", // dark teal
		"1B6B3A", // forest green
		"2D8B4E", // medium green
		"3A9653", // green
		"5DB364", // light green
		"71B280", // sage green
	}, BG: "050F0A"}

	GradientNeon = Gradient{Name: "neon", Colors: []string{
		"FF006E", // hot pink
		"FF00CC", // neon pink
		"FF00FF", // magenta
		"9900FF", // purple
		"0066FF", // blue
		"00CCFF", // light cyan
		"00FFFF", // cyan
	}, BG: "0A0A14"}

	GradientAurora = Gradient{Name: "aurora", Colors: []string{
		"00F5FF", // bright cyan
		"00C6FF", // sky blue
		"0072FF", // deep blue
		"4361EE", // indigo
		"7209B7", // deep violet
		"B5179E", // magenta
		"F72585", // hot pink
	}, BG: "0A0A1A"}

	GradientFire = Gradient{Name: "fire", Colors: []string{
		"7B0D1E", // dark crimson
		"C1121F", // deep red
		"F12711", // bright red
		"F5431A", // red-orange
		"F76B1C", // orange
		"F5AF19", // amber
		"FFF176", // warm yellow
	}, BG: "1A0505"}

	GradientPastel = Gradient{Name: "pastel", Colors: []string{
		"FFB3BA", // pastel pink
		"FFCBA4", // peach
		"FFDFBA", // light peach
		"FFFFBA", // lemon
		"BAFFC9", // mint
		"BAE1FF", // sky blue
		"C9B3FF", // lavender
	}, BG: "F5F5F5"}

	GradientMono = Gradient{Name: "monochrome", Colors: []string{
		"FFFFFF", // white
		"E0E0E0", // light grey
		"BBBBBB", // lighter mid
		"999999", // mid grey
		"777777", // darker mid
		"555555", // dark grey
		"333333", // near black
	}, BG: "0A0A0A"}

	GradientVaporwave = Gradient{Name: "vaporwave", Colors: []string{
		"FF71CE", // hot pink
		"FF9DE2", // light pink
		"D4A5F5", // lilac
		"B967FF", // purple
		"8B5CF6", // deep purple
		"3ABFF8", // sky blue
		"01CDFE", // cyan
	}, BG: "1A0A28"}

	GradientMatrix = Gradient{Name: "matrix", Colors: []string{
		"001200", // near black
		"002200", // very dark green
		"003B00", // dark green
		"007300", // medium green
		"00C800", // green
		"00FF41", // bright green
		"7FFF7F", // light green
	}, BG: "000800"}

	GradientMind = Gradient{Name: "mind", Colors: []string{
		"473B7B", // dark purple
		"3D5A80", // slate blue
		"3584A7", // medium blue
		"2CA58D", // teal
		"30D2BE", // light teal
		"5BE0CA", // mint
		"7EEEE3", // pale teal
	}, BG: "14141A"}

	GradientRainbow = Gradient{Name: "rainbow", Colors: []string{
		"FF0000", // red
		"FF7F00", // orange
		"FFFF00", // yellow
		"00FF00", // green
		"0000FF", // blue
		"4B0082", // indigo
		"9400D3", // violet
	}, BG: "141414"}

	GradientGalaxy = Gradient{Name: "galaxy", Colors: []string{
		"360033", // dark purple
		"2A0040", // deep violet
		"1F004D", // violet
		"14005A", // purple
		"090067", // blue-purple
		"0B8793", // teal
		"10A99F", // light teal
	}, BG: "0A0014"}

	GradientLunar = Gradient{Name: "lunar", Colors: []string{
		"0F0C29", // very dark blue
		"1E1A4A", // dark purple-blue
		"302B63", // dark purple
		"3D3168", // purple
		"24243E", // dark blue-purple
		"2D2B52", // muted purple
		"38385C", // grey-blue
	}, BG: "050510"}

	GradientPhoenix = Gradient{Name: "phoenix", Colors: []string{
		"F83600", // bright red-orange
		"FA4E1A", // red-orange
		"FC681D", // orange
		"FD8620", // orange-yellow
		"F9A423", // yellow-orange
		"F9D423", // yellow
		"FCDF57", // bright yellow
	}, BG: "1A0F05"}

	GradientSpirit = Gradient{Name: "spirit", Colors: []string{
		"B92B27", // deep red
		"A83236", // red
		"963D45", // red-brown
		"5C4D7D", // purple-brown
		"1565C0", // blue
		"1976D2", // medium blue
		"42A5F5", // light blue
	}, BG: "1A1014"}

	GradientCherry = Gradient{Name: "cherry", Colors: []string{
		"EB3349", // red
		"DC2B42", // medium red
		"D0303B", // red
		"F45C43", // orange-red
		"F86B4F", // coral
		"FA8A75", // light coral
		"FFA99C", // pale coral
	}, BG: "1A0A0F"}

	GradientWaves = Gradient{Name: "waves", Colors: []string{
		"667EEA", // purple-blue
		"5E72D9", // medium purple-blue
		"5561C9", // blue-purple
		"6B5AA1", // purple
		"764BA2", // deep purple
		"8559B3", // violet
		"9469C4", // light purple
	}, BG: "14101A"}

	GradientDreamy = Gradient{Name: "dreamy", Colors: []string{
		"FDA085", // peach
		"FBB876", // light orange
		"F6D365", // yellow
		"7ED6DF", // light blue
		"4FACFE", // blue
		"2CE0F5", // cyan
		"00F2FE", // bright cyan
	}, BG: "1A1A1A"}

	GradientMagic = Gradient{Name: "magic", Colors: []string{
		"59C173", // green
		"4DB062", // medium green
		"7B68EE", // medium purple
		"A17FE0", // purple
		"8B5CF6", // deep purple
		"7C3AED", // violet
		"5D26C1", // deep violet
	}, BG: "0F1A14"}

	GradientElectric = Gradient{Name: "electric", Colors: []string{
		"4776E6", // blue
		"3D68D1", // medium blue
		"6459BC", // blue-purple
		"8E54E9", // purple
		"7E3BBF", // deep purple
		"6C2795", // violet
		"5A1A6B", // dark violet
	}, BG: "0F0F1A"}

	GradientVenom = Gradient{Name: "venom", Colors: []string{
		"8360C3", // purple
		"7351B0", // medium purple
		"64429D", // deep purple
		"2EBF91", // green
		"4CC98F", // light green
		"6AD9A5", // mint
		"8AE8B9", // pale mint
	}, BG: "14101A"}

	GradientMirage = Gradient{Name: "mirage", Colors: []string{
		"16222A", // dark blue
		"1E3340", // dark teal
		"264556", // teal
		"2E566C", // medium teal
		"3A6073", // teal-grey
		"4A7885", // light teal
		"5A8F97", // pale teal
	}, BG: "050A0F"}

	GradientRebel = Gradient{Name: "rebel", Colors: []string{
		"F093FB", // pink
		"E87FEC", // medium pink
		"DC6ADD", // pink-magenta
		"D057CE", // magenta
		"C246BF", // deep magenta
		"F5576C", // red
		"E84A5F", // medium red
	}, BG: "1A0F14"}

	GradientDrift = Gradient{Name: "drift", Colors: []string{
		"00D2FF", // cyan
		"00B8E6", // medium cyan
		"009ECC", // dark cyan
		"3A7BD5", // blue
		"2D6BB5", // medium blue
		"205995", // dark blue
		"134775", // deep blue
	}, BG: "051019"}

	GradientBloom = Gradient{Name: "bloom", Colors: []string{
		"FFECD2", // light peach
		"FFDFB8", // pale peach
		"FFC99F", // peach
		"FFB386", // orange-peach
		"FC9D6D", // light orange
		"FCB69F", // coral
		"FFA58C", // light coral
	}, BG: "F5EDE5"}

	GradientAtlas = Gradient{Name: "atlas", Colors: []string{
		"FEAC5E", // orange
		"EFA04D", // orange-yellow
		"D4883C", // gold
		"C779D0", // purple
		"9B5DB5", // violet
		"4BC0C8", // cyan
		"2EC4C2", // teal
	}, BG: "1A1A1A"}
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
