package wsl

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultEngine   string `json:"default_engine"`
	AutoStartDaemon bool   `json:"auto_start_daemon"`
	Theme           string `json:"theme"`
}

func GetConfigPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "ezship", "config.json")
}

func LoadConfig() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		// Default config
		return Config{
			DefaultEngine:   "docker",
			AutoStartDaemon: true,
			Theme:           "dark",
		}
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{DefaultEngine: "docker", AutoStartDaemon: true}
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
