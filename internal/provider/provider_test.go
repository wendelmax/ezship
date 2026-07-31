package provider_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/provider"
)

func TestEngineStatus(t *testing.T) {
	status := provider.EngineStatus{
		Name:    "docker",
		Running: true,
		Version: "27.5.1",
		Mode:    "Native",
	}

	if status.Name != "docker" {
		t.Errorf("expected Name to be 'docker', got '%s'", status.Name)
	}
	if !status.Running {
		t.Errorf("expected Running to be true, got false")
	}
	if status.Version != "27.5.1" {
		t.Errorf("expected Version to be '27.5.1', got '%s'", status.Version)
	}
	if status.Mode != "Native" {
		t.Errorf("expected Mode to be 'Native', got '%s'", status.Mode)
	}
}
