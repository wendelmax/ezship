package provider_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/provider"
)

func TestWSLProviderName(t *testing.T) {
	p := &provider.WSLProvider{}
	if p.Name() != "wsl" {
		t.Errorf("expected provider name 'wsl', got '%s'", p.Name())
	}
}

func TestWSLProviderEngineSupport(t *testing.T) {
	p := &provider.WSLProvider{}
	engines := []string{"docker", "podman", "k3s", "nerdctl", "k3d"}
	for _, e := range engines {
		if !p.IsEngineSupported(e) {
			t.Errorf("expected engine %s to be supported in WSLProvider", e)
		}
	}
	if p.IsEngineSupported("unsupported-engine") {
		t.Errorf("expected unsupported-engine to return false")
	}
}
