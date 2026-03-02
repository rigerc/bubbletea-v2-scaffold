// Package config provides configuration management for the application.
// It supports loading from JSON files, environment variables, and embedded defaults.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	koanfjson "github.com/knadh/koanf/parsers/json"
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
	ConfigVersion int `json:"configVersion" koanf:"configVersion" cfg_default:"1" cfg_exclude:"true"`

	// LogLevel specifies the logging verbosity level.
	// Valid values: trace, debug, info, warn, error, fatal
	LogLevel string `json:"logLevel" mapstructure:"logLevel" koanf:"logLevel" cfg_default:"info" cfg_label:"Log Level" cfg_desc:"Logging verbosity (effective level shown in footer)" cfg_options:"trace,debug,info,warn,error,fatal"`

	// Debug enables debug mode which sets log level to trace
	// and enables additional debugging features.
	Debug bool `json:"debug" mapstructure:"debug" koanf:"debug" cfg_label:"Debug Mode" cfg_desc:"Forces log level to trace; writes debug.log"`

	// UI contains user interface specific configuration.
	UI UIConfig `json:"ui" mapstructure:"ui" koanf:"ui" cfg_label:"UI Settings"`

	// Editor contains editor-related configuration.
	Editor EditorConfig `json:"editor" mapstructure:"editor" koanf:"editor" cfg_label:"Editor"`

	// Network contains network-related configuration.
	Network NetworkConfig `json:"network" mapstructure:"network" koanf:"network" cfg_label:"Network"`

	// Notifications contains notification preferences.
	Notifications NotificationsConfig `json:"notifications" mapstructure:"notifications" koanf:"notifications" cfg_label:"Notifications"`

	// App contains general application configuration.
	App AppConfig `json:"app" mapstructure:"app" koanf:"app" cfg_label:"Application" cfg_exclude:"true"`
}

// UIConfig contains configuration specific to the user interface.
type UIConfig struct {
	// MouseEnabled enables mouse support in the TUI.
	MouseEnabled bool `json:"mouseEnabled" mapstructure:"mouseEnabled" koanf:"mouseEnabled" cfg_default:"true" cfg_label:"Mouse Support" cfg_desc:"Enable mouse click and scroll events"`

	// CompactMode reduces vertical spacing throughout the UI.
	CompactMode bool `json:"compactMode" mapstructure:"compactMode" koanf:"compactMode" cfg_label:"Compact Mode" cfg_desc:"Reduce vertical spacing in lists and menus"`

	// OutputFormat controls how structured output is rendered.
	OutputFormat string `json:"outputFormat" mapstructure:"outputFormat" koanf:"outputFormat" cfg_default:"text" cfg_label:"Output Format" cfg_desc:"Format for structured output" cfg_options:"text,json,table"`

	// DateFormat is the Go time layout used when displaying dates.
	DateFormat string `json:"dateFormat" mapstructure:"dateFormat" koanf:"dateFormat" cfg_default:"2006-01-02" cfg_label:"Date Format" cfg_desc:"Go time layout, e.g. 2006-01-02"`

	// ThemeName specifies the color theme to use.
	ThemeName string `json:"themeName" mapstructure:"themeName" koanf:"themeName" cfg_default:"ember" cfg_label:"Color Theme" cfg_desc:"Visual theme for the application" cfg_options:"_themes"`

	// ShowBanner controls whether the ASCII art banner is shown in the header.
	// When false, a styled plain-text title is rendered instead.
	ShowBanner bool `json:"showBanner" mapstructure:"showBanner" koanf:"showBanner" cfg_default:"true" cfg_label:"ASCII Banner" cfg_desc:"Show ASCII art banner in header"`

	// AnimationSpeed controls the speed of UI animations.
	AnimationSpeed string `json:"animationSpeed" mapstructure:"animationSpeed" koanf:"animationSpeed" cfg_default:"normal" cfg_label:"Animation Speed" cfg_desc:"Speed of transitions and animations" cfg_options:"slow,normal,fast,none"`

	// ShowHelpBar controls whether the persistent help bar is shown.
	ShowHelpBar bool `json:"showHelpBar" mapstructure:"showHelpBar" koanf:"showHelpBar" cfg_default:"true" cfg_label:"Show Help Bar" cfg_desc:"Display keybinding hints at the bottom"`

	// Language sets the interface language.
	Language string `json:"language" mapstructure:"language" koanf:"language" cfg_default:"en" cfg_label:"Language" cfg_desc:"Interface language" cfg_options:"en,es,fr,de,ja,zh"`
}

// EditorConfig contains editor-related configuration.
type EditorConfig struct {
	// EditorCommand is the command to launch the external editor.
	EditorCommand string `json:"editorCommand" mapstructure:"editorCommand" koanf:"editorCommand" cfg_default:"vim" cfg_label:"Editor Command" cfg_desc:"External editor command (e.g., vim, nano, code)"`

	// TabWidth is the number of spaces per tab.
	TabWidth int `json:"tabWidth" mapstructure:"tabWidth" koanf:"tabWidth" cfg_default:"4" cfg_label:"Tab Width" cfg_desc:"Number of spaces per tab stop"`

	// ExpandTabs converts tabs to spaces.
	ExpandTabs bool `json:"expandTabs" mapstructure:"expandTabs" koanf:"expandTabs" cfg_default:"true" cfg_label:"Expand Tabs" cfg_desc:"Convert tabs to spaces"`

	// AutoSave enables automatic saving of changes.
	AutoSave bool `json:"autoSave" mapstructure:"autoSave" koanf:"autoSave" cfg_label:"Auto Save" cfg_desc:"Automatically save changes"`

	// AutoSaveInterval is the interval in seconds between auto-saves.
	AutoSaveInterval int `json:"autoSaveInterval" mapstructure:"autoSaveInterval" koanf:"autoSaveInterval" cfg_default:"30" cfg_label:"Auto Save Interval" cfg_desc:"Seconds between auto-saves (if enabled)"`

	// ShowLineNumbers displays line numbers in editors.
	ShowLineNumbers bool `json:"showLineNumbers" mapstructure:"showLineNumbers" koanf:"showLineNumbers" cfg_default:"true" cfg_label:"Line Numbers" cfg_desc:"Show line numbers in text editors"`
}

// NetworkConfig contains network-related configuration.
type NetworkConfig struct {
	// APIEndpoint is the base URL for API requests.
	APIEndpoint string `json:"apiEndpoint" mapstructure:"apiEndpoint" koanf:"apiEndpoint" cfg_default:"https://api.example.com" cfg_label:"API Endpoint" cfg_desc:"Base URL for API requests"`

	// Timeout is the request timeout in seconds.
	Timeout int `json:"timeout" mapstructure:"timeout" koanf:"timeout" cfg_default:"30" cfg_label:"Request Timeout" cfg_desc:"HTTP request timeout in seconds"`

	// RetryCount is the number of times to retry failed requests.
	RetryCount int `json:"retryCount" mapstructure:"retryCount" koanf:"retryCount" cfg_default:"3" cfg_label:"Retry Count" cfg_desc:"Number of retry attempts for failed requests"`

	// ProxyURL is the HTTP proxy URL (optional).
	ProxyURL string `json:"proxyUrl" mapstructure:"proxyUrl" koanf:"proxyUrl" cfg_label:"Proxy URL" cfg_desc:"HTTP proxy URL (leave empty for direct connection)"`

	// VerifySSL enables SSL certificate verification.
	VerifySSL bool `json:"verifySSL" mapstructure:"verifySSL" koanf:"verifySSL" cfg_default:"true" cfg_label:"Verify SSL" cfg_desc:"Verify SSL certificates (disable for self-signed)"`
}

// NotificationsConfig contains notification preferences.
type NotificationsConfig struct {
	// EnableNotifications controls whether notifications are shown.
	EnableNotifications bool `json:"enableNotifications" mapstructure:"enableNotifications" koanf:"enableNotifications" cfg_default:"true" cfg_label:"Enable Notifications" cfg_desc:"Show desktop notifications"`

	// SoundEnabled controls notification sounds.
	SoundEnabled bool `json:"soundEnabled" mapstructure:"soundEnabled" koanf:"soundEnabled" cfg_default:"true" cfg_label:"Notification Sound" cfg_desc:"Play sound with notifications"`

	// NotifyOnError sends notifications on errors.
	NotifyOnError bool `json:"notifyOnError" mapstructure:"notifyOnError" koanf:"notifyOnError" cfg_default:"true" cfg_label:"Error Notifications" cfg_desc:"Notify when errors occur"`

	// NotifyOnComplete sends notifications when tasks complete.
	NotifyOnComplete bool `json:"notifyOnComplete" mapstructure:"notifyOnComplete" koanf:"notifyOnComplete" cfg_default:"true" cfg_label:"Completion Notifications" cfg_desc:"Notify when long tasks finish"`

	// QuietHoursStart is the start of quiet hours (24h format, e.g., "22:00").
	QuietHoursStart string `json:"quietHoursStart" mapstructure:"quietHoursStart" koanf:"quietHoursStart" cfg_default:"22:00" cfg_label:"Quiet Hours Start" cfg_desc:"Start time for quiet hours (HH:MM format)"`

	// QuietHoursEnd is the end of quiet hours (24h format, e.g., "07:00").
	QuietHoursEnd string `json:"quietHoursEnd" mapstructure:"quietHoursEnd" koanf:"quietHoursEnd" cfg_default:"07:00" cfg_label:"Quiet Hours End" cfg_desc:"End time for quiet hours (HH:MM format)"`
}

// AppConfig contains general application configuration.
// This struct is excluded from the settings UI (cfg_exclude:"true" on the parent field).
type AppConfig struct {
	// Name is the application name.
	Name string `json:"name" mapstructure:"name" koanf:"name" cfg_default:"A Scaffold"`

	// Description is the application description.
	Description string `json:"description" mapstructure:"description" koanf:"description" cfg_default:"A scaffold application"`

	// Version is the application version.
	Version string `json:"version" mapstructure:"version" koanf:"version" cfg_default:"1.0.0"`
}

// loadDefaults populates k with values from DefaultConfig.
// Called first by both Load and LoadFromBytes so that new Config fields
// always have a valid baseline before user data is merged on top.
func loadDefaults(k *koanf.Koanf) error {
	data, err := DefaultConfig().ToJSON()
	if err != nil {
		return fmt.Errorf("marshaling defaults: %w", err)
	}
	return k.Load(rawbytes.Provider(data), koanfjson.Parser())
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
	if err := loadDefaults(k); err != nil {
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
	if err := loadDefaults(k); err != nil {
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

// DefaultConfig returns a configuration with sensible default values
// derived from cfg_default struct tags. These defaults can be overridden
// by loading a configuration file or setting environment variables.
func DefaultConfig() *Config {
	cfg := &Config{}
	applyStructDefaults(reflect.ValueOf(cfg).Elem())
	return cfg
}

// DefaultConfigJSON returns the default configuration as a JSON byte slice.
// This can be used to create a default configuration file or as a fallback
// when no configuration file is found.
func DefaultConfigJSON() ([]byte, error) {
	return DefaultConfig().ToJSON()
}

func applyStructDefaults(rv reflect.Value) {
	rt := rv.Type()
	for i := range rt.NumField() {
		sf := rt.Field(i)
		fv := rv.Field(i)
		if fv.Kind() == reflect.Struct {
			applyStructDefaults(fv)
			continue
		}
		def := sf.Tag.Get("cfg_default")
		if def == "" {
			continue
		}
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(def)
		case reflect.Bool:
			fv.SetBool(def == "true")
		case reflect.Int, reflect.Int64:
			if n, err := strconv.Atoi(def); err == nil {
				fv.SetInt(int64(n))
			}
		}
	}
}
