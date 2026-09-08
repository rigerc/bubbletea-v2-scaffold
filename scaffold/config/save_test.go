package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSave_HappyPath verifies that a valid config is written and can be read back.
func TestSave_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.LogLevel = "warn"
	cfg.UI.ThemeName = "ocean"

	require.NoError(t, Save(cfg, path))

	// File must exist
	_, err := os.Stat(path)
	require.NoError(t, err, "saved file must exist")

	// File must be parseable and round-trip correctly
	loaded, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, cfg.LogLevel, loaded.LogLevel)
	assert.Equal(t, cfg.UI.ThemeName, loaded.UI.ThemeName)
}

// TestSave_DirectoryAutoCreation verifies that Save creates intermediate
// directories when they don't exist.
func TestSave_DirectoryAutoCreation(t *testing.T) {
	base := t.TempDir()
	// Nested directory that does not yet exist
	path := filepath.Join(base, "nested", "deep", "config.json")

	cfg := DefaultConfig()
	require.NoError(t, Save(cfg, path))

	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist after directory auto-creation")
}

// TestSave_InvalidConfigRejected verifies that Save returns an error for an
// invalid Config and does not create any file.
func TestSave_InvalidConfigRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.LogLevel = "not-a-valid-level"

	err := Save(cfg, path)
	assert.ErrorIs(t, err, ErrInvalidConfig)

	// No file should have been written
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "no file should be created for invalid config")
}

// TestSave_NoTempFileLeftOnSuccess verifies that the .tmp staging file is
// cleaned up after a successful save.
func TestSave_NoTempFileLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	require.NoError(t, Save(DefaultConfig(), path))

	tmp := path + ".tmp"
	_, err := os.Stat(tmp)
	assert.True(t, os.IsNotExist(err), "temp file must be gone after successful save")
}
