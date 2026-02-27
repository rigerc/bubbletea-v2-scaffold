package config

import "os"

// IsFirstRun returns true when no config file exists at the given path.
// A first run means the app has never written its config to disk.
func IsFirstRun(configPath string) bool {
	if configPath == "" {
		return false // no config file expected
	}
	_, err := os.Stat(configPath)
	return os.IsNotExist(err)
}

// NeedsUpgrade returns true when the loaded config's version is behind
// the current schema version. Callers should migrate and re-save.
func NeedsUpgrade(cfg *Config) bool {
	return cfg.ConfigVersion < CurrentConfigVersion
}
