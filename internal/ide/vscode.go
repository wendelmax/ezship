package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type VSCodeConfigurator struct{}

func (v *VSCodeConfigurator) Name() string {
	return "VS Code"
}

func (v *VSCodeConfigurator) IsInstalled() bool {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return false
	}
	codeDir := filepath.Join(appData, "Code", "User")
	_, err := os.Stat(codeDir)
	return err == nil
}

func (v *VSCodeConfigurator) Configure(dockerMode string) error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA environment variable not set")
	}

	userDir := filepath.Join(appData, "Code", "User")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return err
	}

	settingsFile := filepath.Join(userDir, "settings.json")
	bakFile := filepath.Join(userDir, "settings.json.ezship.bak")

	var settings map[string]interface{}

	if data, err := os.ReadFile(settingsFile); err == nil {
		_ = os.WriteFile(bakFile, data, 0644)
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	host := "npipe:////./pipe/docker_engine"
	if dockerMode == "wsl" {
		host = "unix:///var/run/docker.sock"
	}
	settings["docker.host"] = host

	outData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VS Code settings: %w", err)
	}

	return os.WriteFile(settingsFile, outData, 0644)
}
