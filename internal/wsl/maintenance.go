package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PruneEngines runs 'prune' on all running engines supported by ezship
func PruneEngines() error {
	engines := []string{"docker", "podman"}
	errs := []string{}
	for _, engine := range engines {
		// Only prune if engine is installed and running
		status := GetEngineStatus(engine)
		if !status.Running {
			continue
		}

		cmd := exec.Command("wsl", "-d", DistroName, "-e", engine, "system", "prune", "-a", "-f", "--volumes")
		if output, err := cmd.CombinedOutput(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", engine, string(output)))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("prune errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ResetDistro unregisters the ezship distro, effectively deleting it
func ResetDistro() error {
	cmd := exec.Command("wsl", "--unregister", DistroName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unregister distro: %s (%w)", string(output), err)
	}
	return nil
}

// Vacuum compacts the WSL vhdx file to recover disk space
func Vacuum() error {
	appData := os.Getenv("APPDATA")
	vhdxPath := filepath.Join(appData, "ezship", "ext4.vhdx")

	if _, err := os.Stat(vhdxPath); os.IsNotExist(err) {
		return fmt.Errorf("vhdx file not found at %s", vhdxPath)
	}

	exec.Command("wsl", "--terminate", DistroName).Run()

	// Create diskpart script
	scriptPath := filepath.Join(os.TempDir(), "ezship_compact.txt")
	scriptContent := fmt.Sprintf("select vdisk file=\"%s\"\ncompact vdisk\ndetach vdisk\n", vhdxPath)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create diskpart script: %w", err)
	}
	defer os.Remove(scriptPath)

	cmd := exec.Command("diskpart", "/s", scriptPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("diskpart failed: %s (%w)", string(output), err)
	}

	return nil
}
