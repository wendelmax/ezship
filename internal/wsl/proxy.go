package wsl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const DistroName = "ezship"

var Version = "dev"

// RunProxyCommand executes a command inside the ezship WSL distro
func RunProxyCommand(engine string, args []string) error {
	// Ensure engine is running before executing command
	if err := EnsureEngineRunning(engine); err != nil {
		return fmt.Errorf("failed to start engine %s: %w", engine, err)
	}

	translatedArgs := TranslateArgs(args)

	// Build the WSL command: wsl -d ezship -e <engine> <args>
	wslArgs := []string{"-d", DistroName, "-e", engine}
	wslArgs = append(wslArgs, translatedArgs...)

	cmd := exec.Command("wsl", wslArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run %s in WSL: %w", engine, err)
	}

	return nil
}

func EnsureEngineRunning(engine string) error {
	daemonName := engine + "d"
	if engine == "podman" || engine == "k3s" {
		daemonName = engine
	}

	// Check if daemon is running using pgrep
	checkCmd := exec.Command("wsl", "-d", DistroName, "pgrep", "-x", daemonName)
	if err := checkCmd.Run(); err == nil {
		return nil // Already running
	}

	fmt.Printf("Starting %s daemon...\n", engine)

	// Startup sequence:
	// 1. Ensure /run/openrc exists (required for OpenRC on Alpine)
	// 2. Start cgroups (required for Docker/Podman)
	// 3. Attempt to start via OpenRC service
	// 4. Fallback to starting daemon manually in background
	startupCmd := fmt.Sprintf(
		"mkdir -p /run/openrc && touch /run/openrc/softlevel && "+
			"(rc-service cgroups start || true) && "+
			"(rc-service %s start || (nohup %s > /var/log/%s.log 2>&1 &))",
		engine, daemonName, engine)

	startCmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "sh", "-c", startupCmd)
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute startup command: %w", err)
	}

	// Wait for socket to be ready (up to 5 seconds)
	socketPath := "/var/run/docker.sock"
	if engine == "podman" {
		socketPath = "/run/podman/podman.sock"
	}

	for i := 0; i < 10; i++ {
		checkSocket := exec.Command("wsl", "-d", DistroName, "ls", socketPath)
		if err := checkSocket.Run(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s socket at %s", engine, socketPath)
}

// IsDistroInstalled checks if the ezship distro is registered in WSL
func IsDistroInstalled() (bool, error) {
	cmd := exec.Command("wsl", "--list", "--quiet")
	output, err := cmd.Output()
	if err != nil {
		// If wsl --list fails, we assume it's not installed or WSL is broken
		return false, nil
	}

	// Output is usually UTF-16LE on Windows
	// A simple way to handle this is to remove null bytes and check the string
	s := strings.ReplaceAll(string(output), "\x00", "")
	return strings.Contains(s, DistroName), nil
}

// CreateProxyBinary creates a copy of the current executable with a different name in the same directory
func CreateProxyBinary(alias string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	installDir := filepath.Dir(exePath)
	proxyPath := filepath.Join(installDir, alias+".exe")

	// If it already exists, don't overwrite it (could be a custom user binary?)
	// But in our case, we probably want to ensure it's OUR proxy.
	// For now, let's just create it if it doesn't exist.
	if _, err := os.Stat(proxyPath); err == nil {
		return nil
	}

	// Use io.Copy to create a physical copy
	source, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(proxyPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to create proxy binary %s: %w", alias, err)
	}

	fmt.Printf("Created global alias: %s\n", alias)
	return nil
}
