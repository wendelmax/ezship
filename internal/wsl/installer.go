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
		fmt.Println("Downloading Alpine Linux rootfs...")
		if err := downloadFile(AlpineURL, rootfsPath); err != nil {
			return fmt.Errorf("failed to download Alpine: %w", err)
		}
	}

	// Import distro
	fmt.Println("Importing ezship distro into WSL2...")
	cmd := exec.Command("wsl", "--import", DistroName, installDir, rootfsPath, "--version", "2")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to import distro: %s (%w)", string(output), err)
	}

	fmt.Println("WSL distro imported successfully.")
	return nil
}

// InstallEngine installs a specific container engine inside the Alpine distro
func InstallEngine(engine string) error {
	fmt.Printf("Installing %s inside WSL...\n", engine)

	var setupCmd string
	switch engine {
	case "docker":
		setupCmd = "apk add docker docker-cli-compose && addgroup root docker && rc-update add docker default"
	case "podman":
		setupCmd = "apk add podman"
	case "k3s":
		setupCmd = "curl -sfL https://get.k3s.io | sh -"
	default:
		return fmt.Errorf("unknown engine: %s", engine)
	}

	cmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "sh", "-c", setupCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install %s: %s (%w)", engine, string(output), err)
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
