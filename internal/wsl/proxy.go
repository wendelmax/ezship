package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const DistroName = "ezship"

// RunProxyCommand executes a command inside the ezship WSL distro
func RunProxyCommand(engine string, args []string) error {
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

// IsDistroInstalled checks if the ezship distro is registered in WSL
func IsDistroInstalled() (bool, error) {
	cmd := exec.Command("wsl", "--list", "--quiet")
	output, err := cmd.Output()
	if err != nil {
		// If wsl --list fails, we assume it's not installed or WSL is broken
		return false, nil
	}

	// Output is usually UTF-16 on Windows, but let's do a simple check
	// Distro names are often separated by null bytes or newlines
	return strings.Contains(string(output), DistroName), nil
}
