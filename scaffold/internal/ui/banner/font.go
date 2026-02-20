package banner

import (
	"math/rand/v2"

	"github.com/lsferreira42/figlet-go/figlet"
)

// SafeFonts is a curated subset of the 145 embedded figlet-go fonts.
// These fonts render cleanly at typical terminal widths (80–120 chars)
// and produce readable output without wide Unicode artifacts.
var SafeFonts = []string{
	"slant",
	"big",
	"banner3",
	"doom",
	"epic",
	"isometric1",
	"larry3d",
	"lean",
	"ogre",
	"roman",
	"shadow",
	"small",
	"smslant",
	"standard",
	"straight",
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

// RandomSafeFont returns a randomly selected font from the SafeFonts list.
func RandomSafeFont() string {
	return SafeFonts[rand.IntN(len(SafeFonts))]
}

// RandomFontFrom returns a randomly selected font from the given pool.
// Falls back to RandomSafeFont if pool is empty.
func RandomFontFrom(pool []string) string {
	if len(pool) == 0 {
		return RandomSafeFont()
	}
	return pool[rand.IntN(len(pool))]
}
