package wsl

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	AlpineURL = "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.1-x86_64.tar.gz"
)

// SetupDistro downloads the Alpine rootfs and imports it into WSL
func SetupDistro() error {
	appData := os.Getenv("APPDATA")
	installDir := filepath.Join(appData, "ezship")
	rootfsPath := filepath.Join(installDir, "alpine-rootfs.tar.gz")

	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Download Alpine rootfs if not exists
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		if err := downloadFile(AlpineURL, rootfsPath); err != nil {
			return fmt.Errorf("failed to download Alpine: %w", err)
		}
	}

	// Check if already installed
	installed, err := IsDistroInstalled()
	if err == nil && installed {
		fmt.Println("ezship distro already imported. Skipping.")
		return nil
	}

	// Import distro
	cmd := exec.Command("wsl", "--import", DistroName, installDir, rootfsPath, "--version", "2")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to import distro: %s (%w)", string(output), err)
	}

	return nil
}

// InstallEngine installs a specific container engine inside the Alpine distro
func InstallEngine(engine string) error {

	var setupCmd string
	switch engine {
	case "docker":
		setupCmd = "apk add docker docker-cli-compose openrc && (addgroup root docker || true) && (rc-update add docker default || true) && (rc-update add cgroups default || true)"
	case "podman":
		setupCmd = "apk add podman openrc && (rc-update add podman default || true) && (rc-update add cgroups default || true)"
	case "k3s":
		setupCmd = "apk add curl openrc && curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_ENABLE=true sh - && (rc-update add k3s default || true)"
	default:
		return fmt.Errorf("unknown engine: %s", engine)
	}

	cmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "sh", "-c", setupCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install %s: %s (%w)", engine, string(output), err)
	}

	// Start engine automatically
	if err := EnsureEngineRunning(engine); err != nil {
		return fmt.Errorf("installed but failed to start %s: %w", engine, err)
	}

	// Create global alias (proxy binary)
	if err := CreateProxyBinary(engine); err != nil {
		fmt.Printf("Warning: failed to create global alias for %s: %v\n", engine, err)
	}

	return nil
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// StopEngine stops an engine's daemon inside WSL
func StopEngine(engine string) error {
	cmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "service", engine, "stop")
	return cmd.Run()
}
