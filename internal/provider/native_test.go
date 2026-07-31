package provider_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/provider"
)

func TestNativeProviderName(t *testing.T) {
	p := &provider.NativeProvider{}
	if p.Name() != "native" {
		t.Errorf("expected provider name 'native', got '%s'", p.Name())
	}
}

func TestNativeProviderEngineSupport(t *testing.T) {
	p := &provider.NativeProvider{}
	if !p.IsEngineSupported("docker") {
		t.Errorf("expected docker to be supported natively")
	}
	if p.IsEngineSupported("podman") {
		t.Errorf("expected podman not to be supported natively in ezship")
	}
}
