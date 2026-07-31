package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestIntelliJConfiguratorName(t *testing.T) {
	ij := &ide.IntelliJConfigurator{}
	if ij.Name() != "IntelliJ IDEA" {
		t.Errorf("expected name 'IntelliJ IDEA', got '%s'", ij.Name())
	}
}
