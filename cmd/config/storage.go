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

// GetLogDir returns the log directory path based on OS
func GetLogDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "log"), nil
}

// GetLogFilePath returns the full path to a log file with the given name
func GetLogFilePath(filename string) (string, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return "", err
	}

	// Create log directory if it doesn't exist
	err = os.MkdirAll(logDir, 0700)
	if err != nil {
		return "", err
	}

	return filepath.Join(logDir, filename), nil
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
		return SavedConfig{Commands: true, Progress: true}
	}

	var config SavedConfig
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return SavedConfig{Commands: true, Progress: true}
	}

	// Set defaults for new fields if not explicitly set
	// If loading an old config, ensure Progress defaults to true
	if !config.Commands && !config.Debug && !config.Progress {
		config.Commands = true
		config.Progress = true
	}

	return config
}
