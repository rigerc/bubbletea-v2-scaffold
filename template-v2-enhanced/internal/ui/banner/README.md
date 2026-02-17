# banner

ASCII art banner rendering via [figlet-go](https://github.com/lsferreira42/figlet-go), with gradient color support.

## Usage

```go
// Random font + random gradient, centered
cfg := banner.RandomBanner("MyApp")
output, err := banner.RenderBanner(cfg, terminalWidth)

// Named font + named gradient
cfg := banner.NamedBanner("MyApp", "slant", "aurora")
output, err := banner.RenderBanner(cfg, terminalWidth)

// Full manual config
grad, _ := banner.GradientByName("fire")
cfg := banner.BannerConfig{
    Text:          "MyApp",
    Font:          "larry3d",
    Gradient:      &grad,
    Width:         60,
    Justification: 1,               // center
    Spacing:       banner.SpacingFullWidth,
    Parser:        "terminal-color",
}
output, err := banner.RenderBanner(cfg, terminalWidth)
```

`output` contains ANSI-colored (or plain/HTML) text ready for `fmt.Print` or lipgloss.

## BannerConfig fields

| Field           | Type          | Default              | Description                                                   |
|-----------------|---------------|----------------------|---------------------------------------------------------------|
| `Text`          | `string`      | —                    | Text to render. Required.                                     |
| `Font`          | `string`      | random from pool     | figlet font name. Empty = random.                             |
| `FontPool`      | `[]string`    | `SafeFonts`          | Pool for random font selection when `Font` is empty.          |
| `FontDir`       | `string`      | embedded fonts       | Custom `.flf` font directory. Empty = use embedded 145 fonts. |
| `Width`         | `int`         | `width` arg          | Override output width. `0` uses the passed-in width.          |
| `Justification` | `int`         | `0` (left)           | `-1` auto · `0` left · `1` center · `2` right                |
| `RightToLeft`   | `int`         | `0` (LTR)            | `-1` auto · `0` LTR · `1` RTL                                |
| `Spacing`       | `SpacingMode` | `SpacingDefault`     | See SpacingMode constants below.                              |
| `Gradient`      | `*Gradient`   | random gradient      | Color stops. Nil = random.                                    |
| `Parser`        | `string`      | `"terminal-color"`   | `"terminal-color"` · `"terminal"` · `"html"`                 |

### SpacingMode constants

```go
banner.SpacingDefault     // font's built-in smushing (zero value)
banner.SpacingKerning     // letters touch but do not overlap
banner.SpacingFullWidth   // no smushing — full character width
banner.SpacingSmushing    // force smushing regardless of font setting
banner.SpacingOverlapping // overlapping mode
```

## Fonts

15 safe fonts are pre-curated from figlet-go's 145 embedded fonts:

```
slant  big  banner3  doom  epic  isometric1  larry3d
lean   ogre roman    shadow  small  smslant  standard  straight
```

```go
banner.SafeFonts              // []string — curated list
banner.RandomSafeFont()       // string — from SafeFonts
banner.RandomFont()           // string — from all 145
banner.RandomFontFrom(pool)   // string — from a custom pool; falls back to SafeFonts if empty
banner.AllFonts()             // []string — full list
```

### Scoped random font

```go
cfg := banner.BannerConfig{
    Text:     "MyApp",
    FontPool: []string{"doom", "slant", "epic"}, // random picks from these three only
}
```

## Gradients

Each gradient uses 6–7 color stops for gradual transitions. figlet-go cycles
through stops character-by-character.

| Name         | Palette                                         |
|--------------|-------------------------------------------------|
| `sunset`     | warm red → orange → yellow → pink-lavender      |
| `ocean`      | dark navy → ocean blue → pale blue              |
| `forest`     | dark teal → forest green → light green          |
| `neon`       | hot pink → magenta → purple → cyan              |
| `aurora`     | cyan → indigo → violet → hot pink               |
| `fire`       | dark crimson → bright red → amber → warm yellow |
| `pastel`     | pink → peach → lemon → mint → lavender          |
| `monochrome` | white → mid-grey → near black                   |
| `vaporwave`  | hot pink → lilac → purple → cyan                |
| `matrix`     | near black → dark green → bright green          |

```go
banner.AllGradients()          // []Gradient — all ten
banner.GradientByName("fire")  // Gradient, bool
banner.RandomGradient()        // Gradient
```

### Custom gradient

```go
custom := banner.Gradient{
    Name:   "candy",
    Colors: []string{"FF0080", "FF4DA6", "FF99CC", "FFCCEE"},
}
cfg := banner.BannerConfig{Text: "Hello", Gradient: &custom}
```
