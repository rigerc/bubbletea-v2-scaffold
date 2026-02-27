package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Load ---

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/does-not-exist.json")
	assert.ErrorIs(t, err, ErrConfigNotFound)
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeJSON(t, `{"logLevel":"info","configVersion":1}`)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	path := writeJSON(t, `{"logLevel":"invalid","configVersion":1}`)

	_, err := Load(path)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestLoad_RoundTrip(t *testing.T) {
	original := DefaultConfig()
	original.UI.ThemeName = "catppuccin"
	original.LogLevel = "warn"

	data, err := original.ToJSON()
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))

	loaded, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, original.UI.ThemeName, loaded.UI.ThemeName)
	assert.Equal(t, original.LogLevel, loaded.LogLevel)
	assert.Equal(t, original.Debug, loaded.Debug)
}

// TestLoad_MissingFieldsGetDefaults verifies that old config files (missing new fields)
// get default values for those fields instead of zero values.
func TestLoad_MissingFieldsGetDefaults(t *testing.T) {
	// Simulate an old config that only has logLevel (missing ui, app, debug, etc.)
	path := writeJSON(t, `{"logLevel":"debug"}`)

	cfg, err := Load(path)
	require.NoError(t, err)

	// Build expected config: defaults with only logLevel overridden
	expected := DefaultConfig()
	expected.LogLevel = "debug"

	// Compare entire structs - this catches any field that doesn't match defaults
	assert.Equal(t, expected, cfg, "missing fields should get default values")
}

// TestLoad_UserValuesOverrideDefaults verifies that user-specified values
// override the defaults.
func TestLoad_UserValuesOverrideDefaults(t *testing.T) {
	path := writeJSON(t, `{
		"logLevel": "warn",
		"debug": true,
		"ui": {
			"mouseEnabled": false,
			"themeName": "nightfly",
			"showBanner": false
		}
	}`)

	cfg, err := Load(path)
	require.NoError(t, err)

	// User values override defaults
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.True(t, cfg.Debug)
	assert.False(t, cfg.UI.MouseEnabled)
	assert.Equal(t, "nightfly", cfg.UI.ThemeName)
	assert.False(t, cfg.UI.ShowBanner)
}

// --- LoadFromBytes ---

func TestLoadFromBytes_ValidJSON(t *testing.T) {
	cfg, err := LoadFromBytes([]byte(`{"logLevel":"debug","configVersion":1}`))
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadFromBytes_ValidatesConfig(t *testing.T) {
	// After the fix, LoadFromBytes must call Validate() just like Load() does.
	_, err := LoadFromBytes([]byte(`{"logLevel":"invalid","configVersion":1}`))
	assert.ErrorIs(t, err, ErrInvalidConfig,
		"LoadFromBytes should reject an invalid logLevel")
}

// TestLoadFromBytes_MissingFieldsGetDefaults verifies that partial configs
// get default values for missing fields.
func TestLoadFromBytes_MissingFieldsGetDefaults(t *testing.T) {
	// Partial config with only logLevel
	cfg, err := LoadFromBytes([]byte(`{"logLevel":"debug"}`))
	require.NoError(t, err)

	// Build expected config: defaults with only logLevel overridden
	expected := DefaultConfig()
	expected.LogLevel = "debug"

	// Compare entire structs - this catches any field that doesn't match defaults
	assert.Equal(t, expected, cfg, "missing fields should get default values")
}

// TestLoadFromBytes_UserValuesOverrideDefaults verifies that user-specified values
// override the defaults.
func TestLoadFromBytes_UserValuesOverrideDefaults(t *testing.T) {
	cfg, err := LoadFromBytes([]byte(`{
		"logLevel": "error",
		"ui": {"themeName": "catppuccin"}
	}`))
	require.NoError(t, err)

	assert.Equal(t, "error", cfg.LogLevel)
	assert.Equal(t, "catppuccin", cfg.UI.ThemeName)
}

// --- DefaultConfig ---

func TestDefaultConfig_DebugFalse(t *testing.T) {
	cfg := DefaultConfig()
	assert.False(t, cfg.Debug, "DefaultConfig should ship with Debug disabled")
}

func TestDefaultConfig_ValidLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.Validate(), "DefaultConfig must be self-consistent")
}

// --- Validate ---

func TestValidate_ValidLogLevels(t *testing.T) {
	levels := []string{"trace", "debug", "info", "warn", "error", "fatal"}
	for _, level := range levels {
		cfg := &Config{LogLevel: level}
		assert.NoError(t, cfg.Validate(), "level %q should be valid", level)
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{LogLevel: "verbose"}
	err := cfg.Validate()
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

// --- GetEffectiveLogLevel ---

func TestGetEffectiveLogLevel_DebugOverride(t *testing.T) {
	cfg := &Config{LogLevel: "info", Debug: true}
	assert.Equal(t, "trace", cfg.GetEffectiveLogLevel())
}

func TestGetEffectiveLogLevel_Normal(t *testing.T) {
	cfg := &Config{LogLevel: "warn", Debug: false}
	assert.Equal(t, "warn", cfg.GetEffectiveLogLevel())
}

// --- Slugify ---

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My App", "my-app"},
		{"some_cool_name", "some-cool-name"},
		{"App!!Name", "appname"}, // special chars are removed, not replaced
		{"scaffold", "scaffold"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"UPPERCASE", "uppercase"},
		{"already-hyphenated", "already-hyphenated"},
		{"---leading-trailing---", "leading-trailing"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, Slugify(tt.input))
		})
	}
}

// --- DefaultConfigPath ---

func TestDefaultConfigPath_NotEmpty(t *testing.T) {
	path := DefaultConfigPath()
	assert.NotEmpty(t, path, "DefaultConfigPath should return a non-empty path")
}

func TestDefaultConfigPath_ContainsAppName(t *testing.T) {
	path := DefaultConfigPath()
	appName := Slugify(DefaultConfig().App.Name)
	assert.Contains(t, path, appName, "DefaultConfigPath should contain slugified app name")
	assert.Contains(t, path, "config.json", "DefaultConfigPath should end with config.json")
}

func TestDefaultConfigPath_RespectsXDGConfigHome(t *testing.T) {
	// Set custom XDG_CONFIG_HOME
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	path := DefaultConfigPath()
	appName := Slugify(DefaultConfig().App.Name)
	expectedSuffix := filepath.Join(tmpDir, appName, "config.json")
	assert.Equal(t, expectedSuffix, path)
}

// --- helpers ---

// writeJSON writes content to a temp file and returns its path.
func writeJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
