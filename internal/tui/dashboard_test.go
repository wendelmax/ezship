package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wendelmax/ezship/internal/wsl"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if m.selected != "Dashboard" {
		t.Errorf("Expected initial selection to be Dashboard, got %s", m.selected)
	}
	if len(m.choices) == 0 {
		t.Error("Expected choices to be populated")
	}
	// With async optimization, these are nil at start
	if m.engines != nil {
		t.Error("Expected engines to be nil at start")
	}
}

func TestNavigation(t *testing.T) {
	m := initialModel()
	initialCursor := m.cursor

	// Move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newM, _ := m.Update(msg)
	m = newM.(model)

	if m.cursor != initialCursor+1 {
		t.Errorf("Expected cursor to be %d, got %d", initialCursor+1, m.cursor)
	}

	// Move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newM, _ = m.Update(msg)
	m = newM.(model)

	if m.cursor != initialCursor {
		t.Errorf("Expected cursor to return to %d, got %d", initialCursor, m.cursor)
	}
}

func TestViewSwitching(t *testing.T) {
	m := initialModel()

	// Navigate to "Engines" (usually index 1)
	m.cursor = 1
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newM, _ := m.Update(msg)
	m = newM.(model)

	if m.selected != "Engines" {
		t.Errorf("Expected selection to be Engines, got %s", m.selected)
	}

	// Go back to Dashboard
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newM, _ = m.Update(msg)
	m = newM.(model)

	if m.selected != "Dashboard" {
		t.Errorf("Expected selection to return to Dashboard, got %s", m.selected)
	}
}

func TestEngineActions(t *testing.T) {
	m := initialModel()
	m.selected = "Engines"

	// Test Start Key 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("Expected command for starting engine")
	}

	// Test Stop Key 't'
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	_, cmd = m.Update(msg)
	if cmd == nil {
		t.Error("Expected command for stopping engine")
	}
}

func TestMaintenanceMessages(t *testing.T) {
	m := initialModel()

	// Simulate a maintenance completion message
	msg := maintenanceMsg{task: "Update", err: nil}
	newM, _ := m.Update(msg)
	m = newM.(model)

	// infoMsg replaced by rolling logs; last log should contain task name
	if len(m.logs) == 0 {
		t.Fatal("Expected at least one log entry after maintenance message")
	}
	lastLog := m.logs[len(m.logs)-1]
	if !contains(lastLog, "Update") {
		t.Errorf("Expected log to contain 'Update', got '%s'", lastLog)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRefreshMessages(t *testing.T) {
	m := initialModel()

	// Test enginesLoadedMsg
	engines := []wsl.EngineInfo{{Name: "docker", Running: true}}
	newM, _ := m.Update(enginesLoadedMsg(engines))
	m = newM.(model)
	if len(m.engines) != 1 {
		t.Error("Expected engines to be loaded")
	}

	// Test distrosLoadedMsg
	distros := []wsl.DistroInfo{{Name: "ezship", State: "Running"}}
	newM, _ = m.Update(distrosLoadedMsg(distros))
	m = newM.(model)
	if len(m.distros) != 1 {
		t.Error("Expected distros to be loaded")
	}
}
