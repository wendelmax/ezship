package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if m.selected != "Dashboard" {
		t.Errorf("Expected initial selection to be Dashboard, got %s", m.selected)
	}
	if len(m.choices) == 0 {
		t.Error("Expected choices to be populated")
	}
	// Verify Update option is present (added for parity)
	foundUpdate := false
	for _, c := range m.choices {
		if c == "Update" {
			foundUpdate = true
			break
		}
	}
	if !foundUpdate {
		t.Error("Expected 'Update' option in choices")
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

	if m.infoMsg != "Update completed" {
		t.Errorf("Expected infoMsg 'Update completed', got '%s'", m.infoMsg)
	}
}
