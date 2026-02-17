# ASCII Banner Framework

A reusable Go framework for creating beautiful ASCII art banners with customizable fonts, font sizes, gradients, and styling.

## Features

- **Multiple Fonts**: Block, Slant, Minimal, Small, Standard
- **Font Scaling**: Scale fonts to 2x, 3x, or any size
- **Rich Gradients**: 22 preset gradients + custom RGB/Hex gradients
- **Fluent API**: Builder pattern for easy configuration
- **Color Utilities**: RGB colors with ANSI support
- **Flexible Output**: Write to stdout, files, or any io.Writer

## Installation

```bash
go mod init your-project
go mod edit -replace banner=./banner
```

Or copy the `banner/` directory into your project.

## Quick Start

```go
package main

import (
    "banner"
)

func main() {
    // Simple banner with defaults
    banner.New("HELLO").Render()
    
    // Quick one-liners
    fmt.Println(banner.Quick("WORLD"))
}
```

## Usage Examples

### Basic Banner

```go
banner.New("MyApp").Render()
```

### With Tagline and Version

```go
banner.New("MyApp").
    Tagline("A powerful CLI tool").
    Version("v1.0.0").
    Render()
```

### Font Sizes (NEW!)

```go
// Double size
banner.New("BIG").Size(2).Render()

// Triple size
banner.New("HUGE").Size(3).Render()

// Quick size function
fmt.Println(banner.QuickWithSize("TEXT", 2))
```

### Different Gradients

```go
banner.New("FIRE").Gradient("fire").Render()
banner.New("OCEAN").Gradient("ocean").Render()
banner.New("MATRIX").Gradient("matrix").Render()
```

### Different Fonts

```go
banner.New("BLOCK").Font("block").Render()
banner.New("SLANT").Font("slant").Render()
banner.New("MINIMAL").Font("minimal").Render()
```

### Custom Gradients

```go
// From hex colors
banner.New("CUSTOM").GradientHex("#FF6B6B", "#4ECDC4", "#45B7D1").Render()

// From RGB values
banner.New("RGB").GradientRGB(
    color.RGB{255, 0, 128},
    color.RGB{0, 255, 128},
).Render()

// Full custom gradient
gradient := color.NewGradientFromRGBs(
    color.RGB{255, 0, 0},
    color.RGB{0, 255, 0},
    color.RGB{0, 0, 255},
)
banner.New("RAINBOW").GradientCustom(gradient).Render()
```

### Custom Output

```go
file, _ := os.Create("banner.txt")
banner.New("TO FILE").Output(file).Render()
```

## Available Fonts

- `block` - Large decorative block letters (6 lines)
- `slant` - Slanted/italic style (5 lines)
- `minimal` - Compact 3-line letters
- `small` - Box-drawing characters (4 lines)
- `standard` - Hash-based letters (5 lines)

## Available Gradients

- `sunset` - Warm red to orange to pink
- `ocean` - Cool blues
- `neon` - Magenta to cyan
- `cyberpunk` - Purple to blue
- `miami` - Pink to cyan
- `fire` - Red to orange to yellow
- `forest` - Natural greens
- `galaxy` - Deep purples
- `retro` - Pink and blue
- `aurora` - Cyan to purple to pink (default)
- `mint` - Fresh green
- `peach` - Soft peach
- `lavender` - Soft purple
- `gold` - Rich gold
- `ice` - Cool blues
- `blood` - Dark reds
- `matrix` - Terminal green
- `vaporwave` - Pink, cyan, purple
- `rainbow` - Full spectrum
- `terminal` - Classic green
- `rose` - Romantic pinks
- `sky` - Sky blues

## API Reference

### Builder Methods

| Method | Description |
|--------|-------------|
| `New(text)` | Create a new banner builder |
| `Font(name)` | Set font style |
| `Size(n)` | Set font size multiplier (1, 2, 3...) |
| `Gradient(name)` | Set gradient by name |
| `GradientRGB(start, end)` | Set 2-color RGB gradient |
| `GradientHex(colors...)` | Set gradient from hex strings |
| `GradientCustom(g)` | Set custom gradient object |
| `Tagline(text)` | Add tagline below main text |
| `Version(text)` | Add version string |
| `Padding(n)` | Set internal padding (lines) |
| `Width(n)` | Set minimum width |
| `Output(w)` | Set output writer |
| `Build()` | Generate banner string |
| `Render()` | Output banner |
| `String()` | Alias for Build() |

### Quick Functions

- `Quick(text)` - Simple banner with defaults
- `QuickWithGradient(text, gradient)` - Banner with specific gradient
- `QuickWithFont(text, font)` - Banner with specific font
- `QuickWithSize(text, size)` - Banner with specific size

### Color Utilities

```go
import "banner/pkg/color"

// Standard colors
color.Red, color.Green, color.Blue, etc.

// RGB colors
c := color.RGB{R: 255, G: 128, B: 0}
color.Colorize("text", c)

// From hex
c := color.FromHex("#FF8000")

// Gradients
g := color.NewGradient(color.Red, color.Blue)
colored := g.Apply("Hello World")
```

### Font Utilities

```go
import "banner/pkg/fonts"

// List available fonts
fonts.List() // ["block", "slant", "minimal", "small", "standard"]

// Generate ASCII art
lines := fonts.Generate("TEXT", "block")

// Generate scaled ASCII art
lines := fonts.GenerateScaled("TEXT", "block", 2)

// Get font
font := fonts.Get("slant")

// Register custom font
fonts.Register(myCustomFont)
```

## Project Structure

```
banner/
├── go.mod
├── banner.go              # Main builder API
├── pkg/
│   ├── color/
│   │   ├── color.go       # RGB colors and utilities
│   │   ├── gradient.go    # Gradient functionality
│   │   └── preset.go      # Preset gradients
│   └── fonts/
│       └── fonts.go       # Font definitions and generation
└── examples/
    └── demo.go            # Usage examples
```

## License

MIT
