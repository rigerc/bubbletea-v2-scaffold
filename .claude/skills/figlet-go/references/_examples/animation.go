// Example demonstrating figlet-go animations
package main

import (
	"fmt"
	"time"

	"github.com/lsferreira42/figlet-go/figlet"
)

func main() {
	// Create config
	cfg := figlet.New()
	cfg.Fontname = "slant"

	// Create animator
	animator := figlet.NewAnimator(cfg)

	// Animation types: reveal, scroll, rain, wave, explosion
	animationTypes := []string{"reveal", "scroll", "rain", "wave", "explosion"}

	for _, animType := range animationTypes {
		fmt.Printf("=== Animation: %s ===\n", animType)

		frames, err := animator.GenerateAnimation("GO!", animType, 50*time.Millisecond)
		if err != nil {
			fmt.Printf("Error generating %s: %v\n", animType, err)
			continue
		}

		// Play animation in terminal
		figlet.PlayAnimation(frames)
		fmt.Println()
	}
}
