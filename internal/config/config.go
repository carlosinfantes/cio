// Package config handles configuration file management for the CTO Advisory Board.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

const (
	// AdvisoryDir is the configuration directory name.
	AdvisoryDir = ".cto-advisory"
	// ConfigFile is the configuration file name.
	ConfigFile = "config.yaml"
	// ContextDir is the context subdirectory name.
	ContextDir = "context"
	// DecisionsDir is the decisions subdirectory name.
	DecisionsDir = "decisions"
	// DiscoveryDir is the discovery sessions subdirectory name.
	DiscoveryDir = "discovery"
	// PluginsDir is the plugins subdirectory name.
	PluginsDir = "plugins"
	// PluginsInstalledDir is the installed plugins subdirectory.
	PluginsInstalledDir = "installed"
	// PluginsCustomDir is the custom plugins subdirectory.
	PluginsCustomDir = "custom"
)

// GetProjectRoot returns the current working directory.
func GetProjectRoot() (string, error) {
	return os.Getwd()
}

// GetAdvisoryDir returns the path to the .cto-advisory directory.
func GetAdvisoryDir() (string, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, AdvisoryDir), nil
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	dir, err := GetAdvisoryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// GetContextDir returns the path to the context directory.
func GetContextDir() (string, error) {
	dir, err := GetAdvisoryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ContextDir), nil
}

// GetDecisionsDir returns the path to the decisions directory.
func GetDecisionsDir() (string, error) {
	dir, err := GetAdvisoryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DecisionsDir), nil
}

// GetDiscoveryDir returns the path to the discovery sessions directory.
func GetDiscoveryDir() (string, error) {
	dir, err := GetAdvisoryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DiscoveryDir), nil
}

// GetPluginsDir returns the path to the plugins directory.
func GetPluginsDir() (string, error) {
	dir, err := GetAdvisoryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PluginsDir), nil
}

// GetInstalledPluginsDir returns the path to installed plugins directory.
func GetInstalledPluginsDir() (string, error) {
	dir, err := GetPluginsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PluginsInstalledDir), nil
}

// GetCustomPluginsDir returns the path to custom plugins directory.
func GetCustomPluginsDir() (string, error) {
	dir, err := GetPluginsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PluginsCustomDir), nil
}

// GetContextFilePath returns the path to a specific context file.
func GetContextFilePath(name string) (string, error) {
	dir, err := GetContextDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".md"), nil
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureAdvisoryDir creates the .cto-advisory directory structure.
func EnsureAdvisoryDir() error {
	advisoryDir, err := GetAdvisoryDir()
	if err != nil {
		return err
	}

	if err := EnsureDir(advisoryDir); err != nil {
		return err
	}

	contextDir, err := GetContextDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(contextDir); err != nil {
		return err
	}

	decisionsDir, err := GetDecisionsDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(decisionsDir); err != nil {
		return err
	}

	discoveryDir, err := GetDiscoveryDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(discoveryDir); err != nil {
		return err
	}

	// Create plugins directories
	installedPluginsDir, err := GetInstalledPluginsDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(installedPluginsDir); err != nil {
		return err
	}

	customPluginsDir, err := GetCustomPluginsDir()
	if err != nil {
		return err
	}
	return EnsureDir(customPluginsDir)
}

// IsInitialized checks if the project has been initialized.
func IsInitialized() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// Load reads the configuration from disk.
func Load() (types.Config, error) {
	cfg := types.DefaultConfig()

	configPath, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if file doesn't exist
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return types.DefaultConfig(), fmt.Errorf("parsing config: %w", err)
	}

	// Merge with defaults to ensure all fields have values
	defaults := types.DefaultConfig()
	if cfg.Model == "" {
		cfg.Model = defaults.Model
	}
	if cfg.DefaultMode == "" {
		cfg.DefaultMode = defaults.DefaultMode
	}
	if len(cfg.DefaultAdvisors) == 0 {
		cfg.DefaultAdvisors = defaults.DefaultAdvisors
	}
	if cfg.ContextRefreshDays == 0 {
		cfg.ContextRefreshDays = defaults.ContextRefreshDays
	}
	if cfg.MaxAdvisors == 0 {
		cfg.MaxAdvisors = defaults.MaxAdvisors
	}

	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg types.Config) error {
	if err := EnsureAdvisoryDir(); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Update modifies specific fields in the configuration.
func Update(updates func(*types.Config)) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	updates(&cfg)

	return Save(cfg)
}

// MaskAPIKey returns a masked version of the API key for display.
func MaskAPIKey(apiKey string) string {
	if len(apiKey) < 14 {
		return "****"
	}
	return apiKey[:10] + "..." + apiKey[len(apiKey)-4:]
}
