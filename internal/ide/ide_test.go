package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestDetectIDEs(t *testing.T) {
	ides := ide.DetectIDEs()
	if ides == nil {
		t.Errorf("expected DetectIDEs to return a non-nil slice")
	}
}
