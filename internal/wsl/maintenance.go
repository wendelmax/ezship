package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PruneEngines runs 'prune' on all engines supported by ezship
func PruneEngines() error {
	engines := []string{"docker", "podman"}
	for _, engine := range engines {
		fmt.Printf("Pruning %s resources...\n", engine)
		cmd := exec.Command("wsl", "-d", DistroName, "-e", engine, "system", "prune", "-a", "-f", "--volumes")
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("Warning: Failed to prune %s: %s\n", engine, string(output))
		}
	}
	return nil
}

// ResetDistro unregisters the ezship distro, effectively deleting it
func ResetDistro() error {
	fmt.Printf("Resetting ezship environment (unregistering distro)... ")
	cmd := exec.Command("wsl", "--unregister", DistroName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unregister distro: %s (%w)", string(output), err)
	}
	fmt.Println("Done.")
	return nil
}

// Vacuum compacts the WSL vhdx file to recover disk space
func Vacuum() error {
	appData := os.Getenv("APPDATA")
	vhdxPath := filepath.Join(appData, "ezship", "ext4.vhdx")

	if _, err := os.Stat(vhdxPath); os.IsNotExist(err) {
		return fmt.Errorf("vhdx file not found at %s", vhdxPath)
	}

	fmt.Println("Stopping ezship distro before compaction...")
	exec.Command("wsl", "--terminate", DistroName).Run()

	// Create diskpart script
	scriptPath := filepath.Join(os.TempDir(), "ezship_compact.txt")
	scriptContent := fmt.Sprintf("select vdisk file=\"%s\"\ncompact vdisk\ndetach vdisk\n", vhdxPath)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create diskpart script: %w", err)
	}
	defer os.Remove(scriptPath)

	fmt.Println("Running diskpart compaction (may require admin privileges)...")
	cmd := exec.Command("diskpart", "/s", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("diskpart failed: %w", err)
	}

	fmt.Println("Vacuum completed successfully!")
	return nil
}
