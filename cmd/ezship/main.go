package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wendelmax/ezship/internal/tui"
	"github.com/wendelmax/ezship/internal/wsl"
)

var rootCmd = &cobra.Command{
	Use:     "ezship",
	Version: wsl.Version,
	Short:   "ezship is a lightweight multi-engine container manager for Windows via WSL2",
	Long: `ezship simplifies container management on Windows by using WSL2 and Ubuntu.
It supports Docker, Podman, nerdctl, k3d, and Kubernetes (k3s) with a beautiful TUI dashboard.

Author: Jackson Wendel Santos Sá <jacksonwendel@gmail.com>
Repo: github.com/wendelmax/ezship`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments, start the TUI
		if len(args) == 0 {
			tui.Start()
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(pruneCmd)
	rootCmd.AddCommand(vacuumCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(updateCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all container engines",
	Run: func(cmd *cobra.Command, args []string) {
		engines := wsl.GetAllEnginesStatus()
		fmt.Printf("%-10s %-12s %s\n", "ENGINE", "STATUS", "VERSION")
		fmt.Println(strings.Repeat("-", 40))
		for _, e := range engines {
			status := "Stopped"
			if e.Running {
				status = "Running"
			}
			if e.Version == "Not Installed" {
				status = "Not Found"
			}
			fmt.Printf("%-10s %-12s %s\n", e.Name, status, e.Version)
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start [engine]",
	Short: "Start a container engine or a WSL distro",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := strings.ToLower(args[0])
		// Check if it's an engine
		engines := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}
		isEngine := false
		for _, e := range engines {
			if e == target {
				isEngine = true
				break
			}
		}

		if isEngine {
			if err := wsl.EnsureEngineRunning(target); err != nil {
				fmt.Printf("Error starting engine %s: %v\n", target, err)
				os.Exit(1)
			}
			fmt.Printf("Engine %s started.\n", target)
		} else {
			// Try to start as a distro
			if err := wsl.StartDistro(target); err != nil {
				fmt.Printf("Error starting distro %s: %v\n", target, err)
				os.Exit(1)
			}
			fmt.Printf("Distro %s started.\n", target)
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [engine]",
	Short: "Stop a container engine or a WSL distro",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := strings.ToLower(args[0])
		// Check if it's an engine
		engines := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}
		isEngine := false
		for _, e := range engines {
			if e == target {
				isEngine = true
				break
			}
		}

		if isEngine {
			if err := wsl.StopEngine(target); err != nil {
				fmt.Printf("Error stopping engine %s: %v\n", target, err)
				os.Exit(1)
			}
			fmt.Printf("Engine %s stopped.\n", target)
		} else {
			// Try to stop as a distro
			if err := wsl.StopDistro(target); err != nil {
				fmt.Printf("Error stopping distro %s: %v\n", target, err)
				os.Exit(1)
			}
			fmt.Printf("Distro %s stopped.\n", target)
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ezship to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := wsl.SelfUpdate(wsl.Version); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused containers and images from all engines",
	Run: func(cmd *cobra.Command, args []string) {
		wsl.PruneEngines()
	},
}

var vacuumCmd = &cobra.Command{
	Use:   "vacuum",
	Short: "Compact the WSL disk (vhdx) to recover space",
	Run: func(cmd *cobra.Command, args []string) {
		if err := wsl.Vacuum(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Unregister and delete the ezship WSL environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := wsl.ResetDistro(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup [engine]",
	Short: "Setup the ezship WSL distro and optionally install an engine (docker, podman, k3s, nerdctl, k3d)",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Setup Distro
		if err := wsl.SetupDistro(); err != nil {
			fmt.Printf("Error setting up distro: %v\n", err)
			os.Exit(1)
		}

		// 2. Install Engine if provided
		if len(args) > 0 {
			engine := strings.ToLower(args[0])
			if err := wsl.InstallEngine(engine); err != nil {
				fmt.Printf("Error installing engine %s: %v\n", engine, err)
				os.Exit(1)
			}
		}
	},
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the ezship TUI dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		tui.Start()
	},
}

func main() {
	// Transparent Proxy Detection
	// If the binary name is 'docker', 'podman', or 'nerdctl', proxy immediately
	exeName := strings.ToLower(filepath.Base(os.Args[0]))
	exeName = strings.TrimSuffix(exeName, ".exe")

	if exeName == "docker" || exeName == "podman" || exeName == "nerdctl" || exeName == "kubectl" || exeName == "k3d" {
		if err := wsl.RunProxyCommand(exeName, os.Args[1:]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
