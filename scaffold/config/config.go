// Package config provides configuration management for the application.
// It supports loading from JSON files, environment variables, and embedded defaults.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// CurrentConfigVersion is the schema version written by this build.
// Increment this whenever a breaking change is made to the Config struct.
const CurrentConfigVersion = 1

var (
	// ErrInvalidConfig is returned when the configuration validation fails.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrConfigNotFound is returned when no configuration file is found.
	ErrConfigNotFound = errors.New("configuration file not found")
)

// Config holds the application configuration.
// All fields are exported to support JSON marshaling and environment variable binding.
type Config struct {
	// ConfigVersion tracks the schema version. Used by NeedsUpgrade to detect
	// configs written by older builds. Not shown in the settings UI (cfg_exclude).
	ConfigVersion int `json:"configVersion" koanf:"configVersion" cfg_exclude:"true"`

	// LogLevel specifies the logging verbosity level.
	// Valid values: trace, debug, info, warn, error, fatal
	LogLevel string `json:"logLevel" mapstructure:"logLevel" koanf:"logLevel" cfg_label:"Log Level" cfg_desc:"Logging verbosity (effective level shown in footer)" cfg_options:"trace,debug,info,warn,error,fatal"`

	// Debug enables debug mode which sets log level to trace
	// and enables additional debugging features.
	Debug bool `json:"debug" mapstructure:"debug" koanf:"debug" cfg_label:"Debug Mode" cfg_desc:"Forces log level to trace; writes debug.log"`

	// UI contains user interface specific configuration.
	UI UIConfig `json:"ui" mapstructure:"ui" koanf:"ui" cfg_label:"UI Settings"`

	// App contains general application configuration.
	App AppConfig `json:"app" mapstructure:"app" koanf:"app" cfg_label:"Application" cfg_exclude:"true"`
}

// UIConfig contains configuration specific to the user interface.
type UIConfig struct {
	// MouseEnabled enables mouse support in the TUI.
	MouseEnabled bool `json:"mouseEnabled" mapstructure:"mouseEnabled" koanf:"mouseEnabled" cfg_label:"Mouse Support" cfg_desc:"Enable mouse click and scroll events"`

	// ThemeName specifies the color theme to use.
	ThemeName string `json:"themeName" mapstructure:"themeName" koanf:"themeName" cfg_label:"Color Theme" cfg_desc:"Visual theme for the application" cfg_options:"_themes"`

	// ShowBanner controls whether the ASCII art banner is shown in the header.
	// When false, a styled plain-text title is rendered instead.
	ShowBanner bool `json:"showBanner" mapstructure:"showBanner" koanf:"showBanner" cfg_label:"ASCII Banner" cfg_desc:"Show ASCII art banner in header"`
}

// AppConfig contains general application configuration.
type AppConfig struct {
	// Name is the application name.
	Name string `json:"name" mapstructure:"name" koanf:"name" cfg_label:"App Name" cfg_desc:"Displayed in the banner"`

	// Description is the application description.
	Description string `json:"description" mapstructure:"description" koanf:"description" cfg_label:"App Description" cfg_desc:"Displayed in the banner"`

	// Version is the application version.
	Version string `json:"version" mapstructure:"version" koanf:"version" cfg_label:"Version" cfg_readonly:"true"`
}

// Load reads configuration from the specified file path.
// If the file does not exist, it returns ErrConfigNotFound.
// If the file exists but cannot be parsed, it returns an error.
// Defaults are loaded first, then user config merges on top - this ensures
// new fields added to Config get their default values when user has old config files.
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	// Create koanf instance
	k := koanf.New(".")

	// 1. Load defaults first
	defaults := DefaultConfig()
	if err := k.Load(confmap.Provider(map[string]any{
		"configVersion": defaults.ConfigVersion,
		"logLevel":      defaults.LogLevel,
		"debug":         defaults.Debug,
		"ui": map[string]any{
			"mouseEnabled": defaults.UI.MouseEnabled,
			"themeName":    defaults.UI.ThemeName,
			"showBanner":   defaults.UI.ShowBanner,
		},
		"app": map[string]any{
			"name":        defaults.App.Name,
			"description": defaults.App.Description,
			"version":     defaults.App.Version,
		},
	}, "."), nil); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// 2. Load user config (merges, overrides defaults for set fields)
	if err := k.Load(file.Provider(path), koanfjson.Parser()); err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}

	// 3. Unmarshal merged result
	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("parsing configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromBytes loads configuration from a byte slice.
// This is useful for loading embedded default configurations.
// Defaults are loaded first, then provided config merges on top - this ensures
// new fields added to Config get their default values when loading partial configs.
func LoadFromBytes(data []byte) (*Config, error) {
	// Create koanf instance
	k := koanf.New(".")

	// 1. Load defaults first
	defaults := DefaultConfig()
	if err := k.Load(confmap.Provider(map[string]any{
		"configVersion": defaults.ConfigVersion,
		"logLevel":      defaults.LogLevel,
		"debug":         defaults.Debug,
		"ui": map[string]any{
			"mouseEnabled": defaults.UI.MouseEnabled,
			"themeName":    defaults.UI.ThemeName,
			"showBanner":   defaults.UI.ShowBanner,
		},
		"app": map[string]any{
			"name":        defaults.App.Name,
			"description": defaults.App.Description,
			"version":     defaults.App.Version,
		},
	}, "."), nil); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// 2. Load from bytes (merges, overrides defaults for set fields)
	if err := k.Load(rawbytes.Provider(data), koanfjson.Parser()); err != nil {
		return nil, fmt.Errorf("loading config from bytes: %w", err)
	}

	// 3. Unmarshal merged result
	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("parsing configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	// Validate log level
	validLogLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("%w: invalid log level '%s'", ErrInvalidConfig, c.LogLevel)
	}

	return nil
}

// ToJSON converts the configuration to a JSON byte slice.
// This is useful for writing the configuration to a file.
func (c *Config) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding configuration to JSON: %w", err)
	}
	return data, nil
}

// GetEffectiveLogLevel returns the effective log level.
// If debug mode is enabled, it returns "trace" regardless of the configured level.
func (c *Config) GetEffectiveLogLevel() string {
	if c.Debug {
		return "trace"
	}
	return c.LogLevel
}
