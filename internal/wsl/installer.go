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
	UbuntuURL = "https://cloud-images.ubuntu.com/minimal/releases/noble/release/ubuntu-24.04-minimal-cloudimg-amd64-root.tar.xz"
)

// SetupDistro downloads the Ubuntu rootfs and imports it into WSL
func SetupDistro() error {
	appData := os.Getenv("APPDATA")
	installDir := filepath.Join(appData, "ezship")
	rootfsPath := filepath.Join(installDir, "ubuntu-rootfs.tar.xz")

	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Download Ubuntu rootfs if not exists
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		if err := downloadFile(UbuntuURL, rootfsPath); err != nil {
			return fmt.Errorf("failed to download Ubuntu: %w", err)
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

// InstallEngine installs a specific container engine inside the Ubuntu distro
func InstallEngine(engine string) error {
	// 0. Ensure distro exists
	if err := SetupDistro(); err != nil {
		return err
	}

	var setupCmd string
	switch engine {
	case "docker":
		setupCmd = "apt-get update && apt-get install -y docker.io"
	case "podman":
		setupCmd = "apt-get update && apt-get install -y podman"
	case "k3s":
		setupCmd = "apt-get update && apt-get install -y curl && curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_ENABLE=true sh -"
	case "nerdctl":
		setupCmd = "apt-get update && apt-get install -y containerd wget tar && wget -q https://github.com/containerd/nerdctl/releases/download/v1.7.3/nerdctl-1.7.3-linux-amd64.tar.gz -O /tmp/nerdctl.tar.gz && tar -C /usr/local/bin -xzf /tmp/nerdctl.tar.gz"
	case "k3d":
		setupCmd = "curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"
	default:
		return fmt.Errorf("unknown engine: %s", engine)
	}

	cmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "sh", "-c", setupCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install %s: %s (%w)", engine, string(output), err)
	}

	// Create global alias (proxy binary)
	if err := CreateProxyBinary(engine); err != nil {
		fmt.Printf("Warning: failed to create global alias for %s: %v\n", engine, err)
	}

	// Extra aliases
	if engine == "k3s" {
		CreateProxyBinary("kubectl")
	}

	// Start engine automatically
	if err := EnsureEngineRunning(engine); err != nil {
		return fmt.Errorf("installed but failed to start %s: %w", engine, err)
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
