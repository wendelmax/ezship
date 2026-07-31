package ide

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type IntelliJConfigurator struct{}

func (i *IntelliJConfigurator) Name() string {
	return "IntelliJ IDEA"
}

func (i *IntelliJConfigurator) IsInstalled() bool {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return false
	}
	jbDir := filepath.Join(appData, "JetBrains")
	_, err := os.Stat(jbDir)
	return err == nil
}

func (i *IntelliJConfigurator) Configure(dockerMode string) error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA environment variable not set")
	}

	jbDir := filepath.Join(appData, "JetBrains")
	entries, err := os.ReadDir(jbDir)
	if err != nil {
		return fmt.Errorf("no JetBrains directories found: %w", err)
	}

	socketURL := "npipe:////./pipe/docker_engine"
	if dockerMode == "wsl" {
		socketURL = "unix:///var/run/docker.sock"
	}

	configuredCount := 0
	for _, entry := range entries {
		if entry.IsDir() && (strings.HasPrefix(entry.Name(), "IntelliJIdea") || strings.HasPrefix(entry.Name(), "IdeaIC")) {
			optionsDir := filepath.Join(jbDir, entry.Name(), "options")
			if err := os.MkdirAll(optionsDir, 0755); err != nil {
				continue
			}

			// Generate remote-servers.xml for Docker
			xmlFile := filepath.Join(optionsDir, "remote-servers.xml")
			bakFile := filepath.Join(optionsDir, "remote-servers.xml.ezship.bak")

			// Backup if file exists
			if data, err := os.ReadFile(xmlFile); err == nil {
				_ = os.WriteFile(bakFile, data, 0644)
			}

			xmlContent := fmt.Sprintf(`<application>
  <component name="RemoteServers">
    <remote-server name="Docker" type="docker">
      <settings>
        <option name="table">
          <map>
            <entry key="url" value="%s" />
          </map>
        </option>
      </settings>
    </remote-server>
  </component>
</application>`, socketURL)

			if err := os.WriteFile(xmlFile, []byte(xmlContent), 0644); err == nil {
				configuredCount++
			}
		}
	}

	if configuredCount == 0 {
		return fmt.Errorf("no active IntelliJ IDEA options directory found to configure")
	}

	return nil
}
