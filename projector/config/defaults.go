// Package config provides configuration management for the application.
package config

// DefaultConfig returns a configuration with sensible default values.
// These defaults can be overridden by loading a configuration file or
// setting environment variables.
func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Debug:    false,
		UI: UIConfig{
			AltScreen:     false,
			MouseEnabled:  true,
			ThemeName:     "default",
		},
		App: AppConfig{
			Name:    "projector",
			Version: "1.0.0",
			Title:   "projector",
		},
	}
}

// DefaultConfigJSON returns the default configuration as a JSON byte slice.
// This can be used to create a default configuration file or as a fallback
// when no configuration file is found.
func DefaultConfigJSON() ([]byte, error) {
	return DefaultConfig().ToJSON()
}
