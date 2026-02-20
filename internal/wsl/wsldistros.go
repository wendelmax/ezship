package wsl

import (
	"fmt"
	"os/exec"
	"strings"
)

type DistroInfo struct {
	Name      string
	State     string
	Version   string
	IsDefault bool
}

// ListDistros parses 'wsl --list --verbose' to get all installed distros
func ListDistros() ([]DistroInfo, error) {
	cmd := exec.Command("wsl", "--list", "--verbose")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list distros: %w", err)
	}

	// Handle UTF-16LE or null bytes
	s := strings.ReplaceAll(string(output), "\x00", "")
	lines := strings.Split(s, "\n")

	var distros []DistroInfo
	for i, line := range lines {
		// Skip header and empty lines
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		isDefault := false
		nameIdx := 0
		if parts[0] == "*" {
			isDefault = true
			nameIdx = 1
		}

		// Parts structure can be: [*] [NAME] [STATE] [VERSION]
		if len(parts) >= nameIdx+3 {
			distros = append(distros, DistroInfo{
				Name:      parts[nameIdx],
				State:     parts[nameIdx+1],
				Version:   parts[nameIdx+2],
				IsDefault: isDefault,
			})
		}
	}

	return distros, nil
}

// StartDistro starts a WSL distro by running a simple command
func StartDistro(name string) error {
	cmd := exec.Command("wsl", "-d", name, "--", "whoami")
	return cmd.Run()
}

// StopDistro terminates a WSL distro
func StopDistro(name string) error {
	cmd := exec.Command("wsl", "--terminate", name)
	return cmd.Run()
}

// DeleteDistro unregisters a WSL distro
func DeleteDistro(name string) error {
	cmd := exec.Command("wsl", "--unregister", name)
	return cmd.Run()
}
