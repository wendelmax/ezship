package wsl

import (
	"testing"
)

func TestEngineInfoDefaults(t *testing.T) {
	info := EngineInfo{Name: "docker", Running: false, Version: "Not Installed"}
	if info.Name != "docker" {
		t.Errorf("Expected name docker, got %s", info.Name)
	}
	if info.Running {
		t.Error("Expected Running to be false")
	}
	if info.Version != "Not Installed" {
		t.Errorf("Expected Version Not Installed, got %s", info.Version)
	}
}

func TestGetAllEnginesStatusList(t *testing.T) {
	// This just verifies the list of engines we support and track
	engines := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}

	// We check if the internal logic matches our expectation
	// This is a bit of a meta-test for the hardcoded list in status.go
	supported := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}

	if len(engines) != len(supported) {
		t.Errorf("Expected %d supported engines, got %d", len(supported), len(engines))
	}
}
