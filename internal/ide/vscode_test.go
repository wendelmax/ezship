package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestVSCodeConfiguratorName(t *testing.T) {
	vsc := &ide.VSCodeConfigurator{}
	if vsc.Name() != "VS Code" {
		t.Errorf("expected name 'VS Code', got '%s'", vsc.Name())
	}
}
