package wsl

import (
	"os/exec"
	"strings"
	"sync"
)

type EngineInfo struct {
	Name    string
	Running bool
	Version string
}

// GetEngineStatus checks the status of a specific engine in WSL
func GetEngineStatus(engine string) EngineInfo {
	info := EngineInfo{Name: engine, Running: false, Version: "Not Installed"}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 1. Check if binary exists
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkCmd := exec.Command("wsl", "-d", DistroName, "which", engine)
		if err := checkCmd.Run(); err != nil {
			return
		}

		// 2. Try to get version (only if it exists)
		versionCmd := exec.Command("wsl", "-d", DistroName, engine, "--version")
		output, err := versionCmd.Output()
		if err == nil {
			mu.Lock()
			info.Version = strings.TrimSpace(strings.Replace(string(output), engine+" version ", "", 1))
			parts := strings.Split(info.Version, "\n")
			if len(parts) > 0 {
				info.Version = parts[0]
			}
			mu.Unlock()
		}
	}()

	// 3. Check if running
	wg.Add(1)
	go func() {
		defer wg.Done()
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
		case "k3d":
			daemonName = "dockerd"
			socketPath = "/run/docker.sock"
		}

		statusCmd := exec.Command("wsl", "-d", DistroName, "pgrep", "-x", daemonName)
		if err := statusCmd.Run(); err == nil {
			socketCheck := exec.Command("wsl", "-d", DistroName, "ls", socketPath)
			if socketErr := socketCheck.Run(); socketErr == nil {
				mu.Lock()
				info.Running = true
				mu.Unlock()
			}
		}
	}()

	wg.Wait()
	return info
}

// GetAllEnginesStatus returns the status of all supported engines
func GetAllEnginesStatus() []EngineInfo {
	engines := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}
	results := make([]EngineInfo, len(engines))
	var wg sync.WaitGroup

	for i, name := range engines {
		wg.Add(1)
		go func(idx int, engineName string) {
			defer wg.Done()
			results[idx] = GetEngineStatus(engineName)
		}(i, name)
	}

	wg.Wait()
	return results
}
