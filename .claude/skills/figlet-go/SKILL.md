---
name: figlet-go
description: >-
  Guide for using the figlet-go library to generate ASCII art text in Go applications.
  Use when working with the figlet-go package, creating ASCII art output, rendering
  animated ASCII art, or integrating FIGlet functionality into Go CLI tools.
  Covers rendering, fonts, colors, animations, and configuration.
---

# figlet-go Skill

Guide for using the figlet-go library (`github.com/lsferreira42/figlet-go/figlet`) to generate ASCII art text in Go applications.

## Installation

```bash
go get github.com/lsferreira42/figlet-go/figlet
```

## Quick Start

### Simple Rendering

```go
import (
    "fmt"
    "log"
    "github.com/lsferreira42/figlet-go/figlet"
)

// Basic usage with default font
result, err := figlet.Render("Hello!")
if err != nil {
    log.Fatal(err)
}
fmt.Print(result)
```

### With Specific Font

```go
// Using a specific font (146 fonts embedded)
result, err := figlet.RenderWithFont("Go!", "slant")
if err != nil {
    log.Fatal(err)
}
fmt.Print(result)
```

## API Patterns

### Functional Options Pattern

Use option functions for configuration:

```go
result, err := figlet.Render("Text",
    figlet.WithFont("big"),                    // Font name
    figlet.WithWidth(60),                      // Output width
    figlet.WithJustification(1),               // 0=left, 1=center, 2=right
    figlet.WithParser("terminal-color"),       // Output format
    figlet.WithColors(figlet.ColorRed, figlet.ColorGreen),
)
```

### Direct Config Usage

For advanced control, use the Config struct directly:

```go
cfg := figlet.New()
cfg.Fontname = "banner"
cfg.Outputwidth = 100

if err := cfg.LoadFont(); err != nil {
    log.Fatal(err)
}

result := cfg.RenderString("Config")
```

## Available Options

| Option | Description |
|--------|-------------|
| `WithFont(name)` | Set font (146 available, see ListFonts()) |
| `WithWidth(width)` | Set output width in characters |
| `WithJustification(n)` | 0=left, 1=center, 2=right |
| `WithParser(type)` | `terminal`, `terminal-color`, `html` |
| `WithColors(colors...)` | ANSI colors or TrueColor |

## Colors

### ANSI Colors

```go
result, err := figlet.Render("Colors!",
    figlet.WithColors(figlet.ColorRed, figlet.ColorGreen, figlet.ColorBlue),
)
```

### TrueColor (24-bit RGB)

```go
tcRed, _ := figlet.NewTrueColorFromHexString("FF0000")
tcGreen, _ := figlet.NewTrueColorFromHexString("00FF00")
result, err := figlet.Render("TrueColor",
    figlet.WithColors(tcRed, tcGreen),
)
```

## Utility Functions

```go
// List all available fonts
fonts := figlet.ListFonts()

// Get version info
version := figlet.GetVersion()      // String version
versionInt := figlet.GetVersionInt() // Integer version
```

## Popular Fonts

- `standard` - Default font
- `banner` - Large banner style
- `big` - Large block letters
- `slant` - Slanted text
- `shadow` - Shadow effect
- `script` - Script style
- `doom` - DOOM game style
- `starwars` - Star Wars style

## Output Formats

- `terminal` - Plain text (default)
- `terminal-color` - ANSI colored output
- `html` - HTML formatted output

## Animations

FIGlet-Go supports generating animated ASCII art with several animation types:

```go
import (
    "fmt"
    "time"
    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    cfg := figlet.New()
    cfg.Fontname = "slant"
    
    animator := figlet.NewAnimator(cfg)
    
    // Generate animation frames (delay between frames)
    frames, err := animator.GenerateAnimation("GO!", "reveal", 50*time.Millisecond)
    if err != nil {
        panic(err)
    }
    
    // Play the animation in terminal or generate HTML player
    figlet.PlayAnimation(frames)
}
```

### Animation Types

| Type | Description |
|------|-------------|
| `reveal` | Reveals text character by character from left to right |
| `scroll` | Scrolls text from the right margin to final position |
| `rain` | Characters "fall" into place from the top |
| `wave` | Sinusoidal wave effect that settles over time |
| `explosion` | Text explodes into particles and reforms |

### HTML Animation Player

When using the `html` parser, `PlayAnimation` automatically generates a **standalone HTML animation player** with:
- Professional terminal aesthetic
- Optimized monospaced fonts
- High-performance JavaScript for fluid playback
- Stable color mapping (colors stay pinned to characters)

```go
// Generate HTML animation
frames, _ := animator.GenerateAnimation("GO!", "explode", 50*time.Millisecond)

// With HTML parser, automatically creates interactive HTML player
figlet.PlayAnimation(frames) // Detects HTML format and generates player
```

### List Available Animations

```go
animations := figlet.ListAnimations()
for _, anim := range animations {
    fmt.Println(anim)
}
```

## Advanced Configuration

### Config Struct Fields

For fine-grained control, use the Config struct directly:

```go
cfg := figlet.New()
cfg.Fontname = "banner"
cfg.Fontdirname = "/path/to/fonts"  // Custom font directory
cfg.Outputwidth = 120              // Max output width
cfg.Justification = 1              // 0=left, 1=center, 2=right
cfg.Right2left = 0                  // -1=auto, 0=LTR, 1=RTL
cfg.Smushmode = 0                   // Smushing mode flags
cfg.Smushoverride = 0               // Override font's smush mode
cfg.Paragraphflag = false           // Paragraph mode
cfg.Deutschflag = false             // German character translation

// Add control file for character translation
cfg.AddControlFile("utf8")

if err := cfg.LoadFont(); err != nil {
    log.Fatal(err)
}

result := cfg.RenderString("Hello")
```

### Smushing Modes

Control how characters overlap:

```go
figlet.WithFullWidth()   // No smushing, characters don't touch
figlet.WithKerning()     // Characters touch but don't overlap
figlet.WithSmushing()    // Characters overlap (default)
figlet.WithOverlapping() // Enable overlapping mode
```

## Version Info

```go
fmt.Println("Version:", figlet.GetVersion())      // "2.2.5"
fmt.Println("Version Int:", figlet.GetVersionInt()) // 20205
fmt.Println("Columns:", figlet.GetColumns())      // Terminal width
```

## Example

See complete working examples:
- [references/_examples/example.go](references/_examples/example.go) - Basic rendering, fonts, colors, HTML output
- [references/_examples/animation.go](references/_examples/animation.go) - Animated ASCII art

## Resources

- GitHub: https://github.com/lsferreira42/figlet-go
- Documentation: https://lsferreira42.github.io/figlet-go/
- Library Docs: https://github.com/lsferreira42/figlet-go/blob/main/lib.md
