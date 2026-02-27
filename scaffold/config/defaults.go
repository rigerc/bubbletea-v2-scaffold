// Package config provides configuration management for the application.
package config

// DefaultConfig returns a configuration with sensible default values.
// These defaults can be overridden by loading a configuration file or
// setting environment variables.
func DefaultConfig() *Config {
	return &Config{
		ConfigVersion: CurrentConfigVersion,
		LogLevel:      "info",
		Debug:    false,
		UI: UIConfig{
			MouseEnabled: true,
			ThemeName:    "ember",
			ShowBanner:   true,
		},
		App: AppConfig{
			Name:        "scaffold",
			Description: "A scaffold application",
			Version:     "1.0.0",
		},
	}
}

// DefaultConfigJSON returns the default configuration as a JSON byte slice.
// This can be used to create a default configuration file or as a fallback
// when no configuration file is found.
func DefaultConfigJSON() ([]byte, error) {
	return DefaultConfig().ToJSON()
}
