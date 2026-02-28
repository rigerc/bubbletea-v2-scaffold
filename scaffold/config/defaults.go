// Package config provides configuration management for the application.
package config

// DefaultConfig returns a configuration with sensible default values.
// These defaults can be overridden by loading a configuration file or
// setting environment variables.
func DefaultConfig() *Config {
	return &Config{
		ConfigVersion: CurrentConfigVersion,
		LogLevel:      "info",
		Debug:         false,
		UI: UIConfig{
			MouseEnabled:    true,
			CompactMode:     false,
			OutputFormat:    "text",
			DateFormat:      "2006-01-02",
			ThemeName:       "ember",
			ShowBanner:      true,
			AnimationSpeed:  "normal",
			ShowHelpBar:     true,
			Language:        "en",
		},
		Editor: EditorConfig{
			EditorCommand:     "vim",
			TabWidth:          4,
			ExpandTabs:        true,
			AutoSave:          false,
			AutoSaveInterval:  30,
			ShowLineNumbers:   true,
		},
		Network: NetworkConfig{
			APIEndpoint: "https://api.example.com",
			Timeout:     30,
			RetryCount:  3,
			ProxyURL:    "",
			VerifySSL:   true,
		},
		Notifications: NotificationsConfig{
			EnableNotifications: true,
			SoundEnabled:        true,
			NotifyOnError:       true,
			NotifyOnComplete:    true,
			QuietHoursStart:     "22:00",
			QuietHoursEnd:       "07:00",
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
