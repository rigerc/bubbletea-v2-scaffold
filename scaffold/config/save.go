package config

import (
	"fmt"
	"os"
	"path/filepath"

	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// Save persists cfg to path using koanf as the write pipeline.
// Atomic: writes to a temp file, then renames.
func Save(cfg *Config, path string) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: save validation: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("config: creating config directory: %w", err)
	}

	raw, err := cfg.ToJSON()
	if err != nil {
		return fmt.Errorf("config: encoding for save: %w", err)
	}

	k := koanf.New(".")
	if err := k.Load(rawbytes.Provider(raw), koanfjson.Parser()); err != nil {
		return fmt.Errorf("config: koanf parse during save: %w", err)
	}

	out, err := k.Marshal(koanfjson.Parser())
	if err != nil {
		return fmt.Errorf("config: koanf marshal during save: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return fmt.Errorf("config: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("config: atomic rename: %w", err)
	}
	return nil
}
