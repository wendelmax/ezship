package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NativeProvider struct{}

func (n *NativeProvider) Name() string {
	return "native"
}

func (n *NativeProvider) IsEngineSupported(engine string) bool {
	return strings.ToLower(engine) == "docker"
}

func (n *NativeProvider) InstallEngine(engine string) error {
	if !n.IsEngineSupported(engine) {
		return fmt.Errorf("engine %s is not supported natively on Windows", engine)
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	binDir := filepath.Join(userHome, ".ezship", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	fmt.Printf("Native Docker installation directory: %s\n", binDir)
	return nil
}

func (n *NativeProvider) EnsureEngineRunning(engine string) error {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq dockerd.exe")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "dockerd.exe") {
		return nil
	}
	return nil
}

func (n *NativeProvider) StopEngine(engine string) error {
	cmd := exec.Command("taskkill", "/F", "/IM", "dockerd.exe")
	return cmd.Run()
}

func (n *NativeProvider) GetStatus(engine string) EngineStatus {
	running := false
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq dockerd.exe")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "dockerd.exe") {
		running = true
	}

	version := "Not Installed"
	if running {
		vCmd := exec.Command("docker", "--version")
		if vOut, vErr := vCmd.Output(); vErr == nil {
			version = strings.TrimSpace(string(vOut))
		}
	}

	return EngineStatus{
		Name:    "docker",
		Running: running,
		Version: version,
		Mode:    "Native",
	}
}

func (n *NativeProvider) RunProxyCommand(engine string, args []string) error {
	cmd := exec.Command(engine, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
