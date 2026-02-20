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
	// A lighter check: look for the daemon process (e.g., dockerd, podman)
	daemonName := engine + "d"
	socketPath := "/var/run/docker.sock"

	switch engine {
	case "podman":
		daemonName = "podman"
		socketPath = "/run/podman/podman.sock"
	case "k3s":
		daemonName = "k3s"
		socketPath = "/run/k3s/containerd/containerd.sock"
	case "nerdctl":
		daemonName = "containerd"
		socketPath = "/run/containerd/containerd.sock"
	}

	statusCmd := exec.Command("wsl", "-d", DistroName, "pgrep", "-x", daemonName)
	if err := statusCmd.Run(); err == nil {
		// Also check if socket exists for better accuracy
		socketCheck := exec.Command("wsl", "-d", DistroName, "ls", socketPath)
		if socketErr := socketCheck.Run(); socketErr == nil {
			info.Running = true
		}
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
