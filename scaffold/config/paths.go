package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Slugify converts a string to a lowercase, hyphen-separated slug.
// Used to derive config directory name from app name.
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace spaces and underscores with hyphens
	s = strings.NewReplacer(" ", "-", "_", "-").Replace(s)
	// Remove non-alphanumeric characters except hyphens
	re := regexp.MustCompile("[^a-z0-9-]")
	s = re.ReplaceAllString(s, "")
	// Collapse multiple hyphens
	re = regexp.MustCompile("-+")
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// DefaultConfigPath returns the XDG-compliant default config file location.
// Directory name is derived from the default config's App.Name field.
func DefaultConfigPath() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		cfgDir = filepath.Join(home, ".config")
	}
	// Get app name from defaults and slugify it
	appName := Slugify(DefaultConfig().App.Name)
	return filepath.Join(cfgDir, appName, "config.json")
}
