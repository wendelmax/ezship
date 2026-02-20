package wsl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const DistroName = "ezship"

var Version = "dev"

// RunProxyCommand executes a command inside the ezship WSL distro
func RunProxyCommand(engine string, args []string) error {
	// Ensure engine is running before executing command
	EnsureEngineRunning(engine)

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

// EnsureEngineRunning checks if the engine daemon is running and starts it if necessary
func EnsureEngineRunning(engine string) {
	// Check if daemon is running (simple ps check)
	checkCmd := exec.Command("wsl", "-d", DistroName, "ps", "-A")
	output, err := checkCmd.Output()
	if err != nil {
		return
	}

	running := strings.Contains(string(output), engine+"d") || strings.Contains(string(output), engine)

	if !running {
		fmt.Printf("Starting %s daemon...\n", engine)
		// Start daemon using OpenRC service or manual start
		startCmd := exec.Command("wsl", "-d", DistroName, "-u", "root", "sh", "-c",
			"mkdir -p /run/openrc && touch /run/openrc/softlevel && service "+engine+" start || ("+engine+"d &)")
		startCmd.Run()
	}
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
