package wsl

import (
	"os/exec"
	"strings"
)

type EngineInfo struct {
	Name    string
	Running bool
	Version string
}

// GetEngineStatus checks the status of a specific engine in WSL
func GetEngineStatus(engine string) EngineInfo {
	info := EngineInfo{Name: engine, Running: false, Version: "Not Installed"}

	// Check if installed (binary exists)
	checkCmd := exec.Command("wsl", "-d", DistroName, "which", engine)
	if err := checkCmd.Run(); err != nil {
		return info
	}

	// Try to get version
	versionCmd := exec.Command("wsl", "-d", DistroName, engine, "--version")
	output, err := versionCmd.Output()
	if err == nil {
		info.Version = strings.TrimSpace(strings.Replace(string(output), engine+" version ", "", 1))
		// Handle potential multi-line or long version strings
		parts := strings.Split(info.Version, "\n")
		if len(parts) > 0 {
			info.Version = parts[0]
		}
	}

	// Check if running
	// For docker/podman, we can check for the socket or process
	var statusCmd *exec.Cmd
	switch engine {
	case "docker":
		statusCmd = exec.Command("wsl", "-d", DistroName, "docker", "info")
	case "podman":
		statusCmd = exec.Command("wsl", "-d", DistroName, "podman", "info")
	case "k3s":
		statusCmd = exec.Command("wsl", "-d", DistroName, "kubectl", "get", "nodes")
	case "nerdctl":
		statusCmd = exec.Command("wsl", "-d", DistroName, "nerdctl", "info")
	default:
		statusCmd = exec.Command("wsl", "-d", DistroName, "ps", "aux")
	}

	if err := statusCmd.Run(); err == nil {
		info.Running = true
	}

	return info
}

// GetAllEnginesStatus returns the status of all supported engines
func GetAllEnginesStatus() []EngineInfo {
	engines := []string{"docker", "podman", "k3s", "nerdctl"}
	results := make([]EngineInfo, len(engines))
	for i, name := range engines {
		results[i] = GetEngineStatus(name)
	}
	return results
}
