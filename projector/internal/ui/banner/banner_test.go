package banner

import (
	"fmt"
	"strings"
	"testing"
)

func TestRenderBanner(t *testing.T) {
	cfg := BannerConfig{
		Text:       "TEST",
		Background: true,
		Gradient:   &GradientSunset,
	}
	res, err := RenderBanner(cfg, 40)
	if err != nil {
		t.Fatalf("RenderBanner failed: %v", err)
	}

	if res == "" {
		t.Error("RenderBanner returned empty string")
	}

	// Check for background color escape sequence
	// GradientSunset.BG is "1A0A0A" -> R:26, G:10, B:10
	fmt.Printf("Result hex (first 100): %x\n", res[:min(100, len(res))])
	if !strings.Contains(res, "\x1b[48;") {
		t.Errorf("Result should contain some background sequence, but it doesn't. Res: %q", res)
	}

	fmt.Printf("Result length: %d\n", len(res))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
