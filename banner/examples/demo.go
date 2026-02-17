package main

import (
	"fmt"
	"os"

	"banner"
	"banner/pkg/color"
	"banner/pkg/fonts"
)

func main() {
	fmt.Println("=== ASCII Banner Framework Demo ===\n")

	// Example 1: Simple banner with defaults
	fmt.Println("Example 1: Simple banner with defaults")
	banner.New("HELLO").Render()

	// Example 2: Banner with tagline and version
	fmt.Println("\nExample 2: With tagline and version")
	banner.New("MyApp").
		Tagline("A powerful CLI tool").
		Version("v1.0.0").
		Render()

	// Example 3: Different gradients
	fmt.Println("\nExample 3: Different gradients")

	fmt.Println("Fire gradient:")
	banner.New("FIRE").
		Gradient("fire").
		Font("slant").
		Render()

	fmt.Println("\nOcean gradient:")
	banner.New("OCEAN").
		Gradient("ocean").
		Font("block").
		Render()

	fmt.Println("\nRainbow gradient:")
	banner.New("RAINBOW").
		Gradient("rainbow").
		Font("minimal").
		Render()

	// Example 4: Different fonts
	fmt.Println("\nExample 4: Different fonts")

	for _, fontName := range []string{"block", "slant", "minimal", "small", "standard"} {
		fmt.Printf("\nFont: %s\n", fontName)
		banner.New("DEMO").
			Font(fontName).
			Gradient("neon").
			Render()
	}

	// Example 5: Different sizes (NEW!)
	fmt.Println("\nExample 5: Different font sizes")

	for _, size := range []int{1, 2, 3} {
		fmt.Printf("\nSize: %dx\n", size)
		banner.New("SIZE").
			Size(size).
			Font("small").
			Gradient("gold").
			Render()
	}

	// Example 6: Custom gradient with hex colors
	fmt.Println("\nExample 6: Custom hex gradient")
	banner.New("CUSTOM").
		GradientHex("#FF6B6B", "#4ECDC4", "#45B7D1").
		Font("block").
		Tagline("Using custom colors").
		Render()

	// Example 7: Custom gradient with RGB
	fmt.Println("\nExample 7: Custom RGB gradient")
	banner.New("RGB").
		GradientRGB(color.RGB{255, 0, 128}, color.RGB{0, 255, 128}).
		Font("slant").
		Render()

	// Example 8: Quick functions
	fmt.Println("\nExample 8: Quick functions")
	fmt.Println(banner.Quick("QUICK"))
	fmt.Println(banner.QuickWithGradient("FIRE", "fire"))
	fmt.Println(banner.QuickWithFont("SLANT", "slant"))
	fmt.Println(banner.QuickWithSize("BIG", 2))

	// Example 9: List available options
	fmt.Println("\n=== Available Options ===")

	fmt.Println("\nAvailable fonts:")
	for _, name := range fonts.List() {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\nAvailable gradients:")
	for _, name := range color.ListGradients() {
		fmt.Printf("  - %s\n", name)
	}

	// Example 10: Write to file
	fmt.Println("\nExample 10: Write banner to file")
	file, err := os.CreateTemp("", "banner-*.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer os.Remove(file.Name())

	banner.New("FILE").
		Gradient("matrix").
		Tagline("Written to file").
		Output(file).
		Render()
	file.Close()

	content, _ := os.ReadFile(file.Name())
	fmt.Printf("Banner written to: %s\n", file.Name())
	fmt.Printf("Content preview:\n%s\n", string(content))

	fmt.Println("\n=== Demo Complete ===")
}
