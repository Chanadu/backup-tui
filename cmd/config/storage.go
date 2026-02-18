package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// SavedConfig represents the persisted parameters (excluding password)
type SavedConfig struct {
	User       string `json:"user"`
	Server     string `json:"server"`
	BackupPath string `json:"backupPath"`
	Debug      bool   `json:"debug"`
	Commands   bool   `json:"commands"`
	Progress   bool   `json:"progress"`
}

// GetConfigDir returns the config directory path based on OS
func GetConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		// Use standard AppData location on Windows
		appDataDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(appDataDir, "backup-tui"), nil
	}

	// On Linux/Mac, use ~/.config/backup-tui
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "backup-tui"), nil
}

// GetConfigFilePath returns the full path to the config file
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// SaveConfig saves the parameters (except password) to disk
func SaveConfig(user, server, backupPath string, debug, commands, progress bool) error {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return err
	}

	config := SavedConfig{
		User:       user,
		Server:     server,
		BackupPath: backupPath,
		Debug:      debug,
		Commands:   commands,
		Progress:   progress,
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, jsonData, 0600)
	if err != nil {
		return err
	}

	return nil
}

// LoadConfig loads the saved parameters from disk
// Returns zero values if file doesn't exist or can't be read
func LoadConfig() SavedConfig {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return SavedConfig{}
	}

	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist or can't be read, return defaults
		return SavedConfig{Commands: true} // Commands is true by default
	}

	var config SavedConfig
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return SavedConfig{Commands: true}
	}

	// Ensure Commands defaults to true if not set
	if !config.Commands && !config.Debug && !config.Progress {
		config.Commands = true
	}

	return config
}
