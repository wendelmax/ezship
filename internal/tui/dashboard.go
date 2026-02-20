package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wendelmax/ezship/internal/wsl"
)

type model struct {
	choices  []string
	cursor   int
	selected string
	engines  []wsl.EngineInfo
	distros  []wsl.DistroInfo
	dCursor  int
	eCursor  int
	cCursor  int
	sCursor  int
	config   wsl.Config
	infoMsg  string
}

type engineActionMsg struct {
	engine string
	err    error
	action string
}

type maintenanceMsg struct {
	task string
	err  error
}

type refreshMsg struct{}

func initialModel() model {
	distros, _ := wsl.ListDistros()
	return model{
		choices:  []string{"Dashboard", "Engines", "WSL Distros", "Cleanup", "Settings", "About", "Exit"},
		cursor:   0,
		engines:  wsl.GetAllEnginesStatus(),
		distros:  distros,
		dCursor:  0,
		eCursor:  0,
		cCursor:  0,
		sCursor:  0,
		config:   wsl.LoadConfig(),
		selected: "Dashboard",
		infoMsg:  "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(true)
		case "down", "j":
			m.moveCursor(false)
		case "enter", " ":
			return m.handleSelection()
		case "s": // Start
			return m, m.cmdStart()
		case "t": // Stop
			return m, m.cmdStop()
		case "i": // Install engine
			if m.selected == "Engines" && len(m.engines) > 0 {
				engine := m.engines[m.eCursor].Name
				m.infoMsg = "Installing " + engine + "..."
				return m, m.cmdInstall(engine)
			}
		case "esc", "backspace", "left", "h":
			m.selected = "Dashboard"
			m.infoMsg = ""
		}

	case engineActionMsg:
		if msg.err != nil {
			m.infoMsg = "Error: " + msg.err.Error()
		} else {
			m.infoMsg = fmt.Sprintf("%s: %s completed", msg.action, msg.engine)
		}
		m.engines = wsl.GetAllEnginesStatus()
		return m, nil

	case maintenanceMsg:
		if msg.err != nil {
			m.infoMsg = "Error: " + msg.err.Error()
		} else {
			m.infoMsg = msg.task + " completed"
		}
		return m, nil

	case refreshMsg:
		m.engines = wsl.GetAllEnginesStatus()
		m.distros, _ = wsl.ListDistros()
		return m, nil
	}

	return m, nil
}

// Helpers to reduce Update complexity

func (m *model) moveCursor(up bool) {
	delta := 1
	if up {
		delta = -1
	}

	switch m.selected {
	case "WSL Distros":
		m.applyMovement(&m.dCursor, delta, len(m.distros)-1)
	case "Engines":
		m.applyMovement(&m.eCursor, delta, len(m.engines)-1)
	case "Cleanup":
		m.applyMovement(&m.cCursor, delta, 1) // 2 tasks
	case "Settings":
		m.applyMovement(&m.sCursor, delta, 2) // 3 settings: Auto, Default, Reset
	default: // Main menu
		m.applyMovement(&m.cursor, delta, len(m.choices)-1)
	}
}

func (m *model) applyMovement(cursor *int, delta int, max int) {
	*cursor += delta
	if *cursor < 0 {
		*cursor = 0
	}
	if max < 0 {
		max = 0
	}
	if *cursor > max {
		*cursor = max
	}
}

func (m *model) handleSelection() (tea.Model, tea.Cmd) {
	// If in a sub-view, execute its action
	switch m.selected {
	case "Engines":
		if len(m.engines) > 0 {
			engine := m.engines[m.eCursor].Name
			m.infoMsg = "Installing " + engine + "..."
			return *m, m.cmdInstall(engine)
		}
		return *m, nil

	case "Cleanup":
		switch m.cCursor {
		case 0: // Prune
			m.infoMsg = "Pruning engines..."
			return *m, m.cmdPrune()
		case 1: // Vacuum
			m.infoMsg = "Vacuuming disk (this may take a while)..."
			return *m, m.cmdVacuum()
		}
		return *m, nil

	case "Settings":
		switch m.sCursor {
		case 0: // Auto-Start
			m.config.AutoStartDaemon = !m.config.AutoStartDaemon
			wsl.SaveConfig(m.config)
			m.infoMsg = "Auto-Start toggled"
		case 1: // Default Engine
			opts := []string{"docker", "podman", "k3s"}
			idx := 0
			for i, e := range opts {
				if e == m.config.DefaultEngine {
					idx = (i + 1) % len(opts)
					break
				}
			}
			m.config.DefaultEngine = opts[idx]
			wsl.SaveConfig(m.config)
			m.infoMsg = "Default engine: " + m.config.DefaultEngine
		case 2: // Reset
			m.infoMsg = "Reset requested. Use CLI: ezship reset"
		}
		return *m, nil
	}

	// If in main menu, select the item
	choice := m.choices[m.cursor]
	if choice == "Exit" {
		return *m, tea.Quit
	}
	m.selected = choice
	m.enterView()
	return *m, nil
}

func (m *model) enterView() {
	m.infoMsg = ""
	switch m.selected {
	case "WSL Distros":
		m.distros, _ = wsl.ListDistros()
		m.dCursor = 0
	case "Engines":
		m.engines = wsl.GetAllEnginesStatus()
		m.eCursor = 0
	case "Cleanup":
		m.cCursor = 0
	case "Settings":
		m.sCursor = 0
		m.config = wsl.LoadConfig()
	}
}

func (m *model) cmdStart() tea.Cmd {
	return func() tea.Msg {
		if m.selected == "WSL Distros" && len(m.distros) > 0 {
			err := wsl.StartDistro(m.distros[m.dCursor].Name)
			return maintenanceMsg{task: "Start Distro", err: err}
		} else if m.selected == "Engines" && len(m.engines) > 0 {
			engine := m.engines[m.eCursor].Name
			err := wsl.EnsureEngineRunning(engine)
			return engineActionMsg{engine: engine, action: "Start", err: err}
		}
		return nil
	}
}

func (m *model) cmdStop() tea.Cmd {
	return func() tea.Msg {
		if m.selected == "WSL Distros" && len(m.distros) > 0 {
			err := wsl.StopDistro(m.distros[m.dCursor].Name)
			return maintenanceMsg{task: "Stop Distro", err: err}
		} else if m.selected == "Engines" && len(m.engines) > 0 {
			engine := m.engines[m.eCursor].Name
			err := wsl.StopEngine(engine)
			return engineActionMsg{engine: engine, action: "Stop", err: err}
		}
		return nil
	}
}

func (m *model) cmdInstall(engine string) tea.Cmd {
	return func() tea.Msg {
		err := wsl.InstallEngine(engine)
		return engineActionMsg{engine: engine, action: "Install", err: err}
	}
}

func (m *model) cmdPrune() tea.Cmd {
	return func() tea.Msg {
		err := wsl.PruneEngines()
		return maintenanceMsg{task: "Prune", err: err}
	}
}

func (m *model) cmdVacuum() tea.Cmd {
	return func() tea.Msg {
		err := wsl.Vacuum()
		return maintenanceMsg{task: "Vacuum", err: err}
	}
}

func (m model) View() string {
	var b strings.Builder

	// Header
	header := HeaderStyle.Render(" ezship - Multi-Engine Control Panel ")
	b.WriteString(header + "\n\n")

	// Sidebar (Menu)
	var menu strings.Builder
	for i, choice := range m.choices {
		if m.cursor == i {
			menu.WriteString(SelectedStyle.Render(" > "+choice) + "\n")
		} else {
			menu.WriteString(NormalStyle.Render("   "+choice) + "\n")
		}
	}

	// Main Content
	var content strings.Builder
	switch m.selected {
	case "About":
		content.WriteString(TitleStyle.Render("About ezship") + "\n\n")
		content.WriteString("  ezship - Multi-Engine Control Panel\n")
		content.WriteString(fmt.Sprintf("  Version: %s\n\n", wsl.Version))
		content.WriteString(fmt.Sprintf("  Author: %s\n", "Jackson Wendel Santos Sá"))
		content.WriteString(fmt.Sprintf("  Email:  %s\n", "jacksonwendel@gmail.com"))
		content.WriteString(fmt.Sprintf("  Repo:   %s\n", "github.com/wendelmax/ezship"))

	case "Cleanup":
		content.WriteString(TitleStyle.Render("System Cleanup") + "\n\n")
		content.WriteString("  Controls: [Enter] Run Selected Task\n\n")

		tasks := []struct {
			name string
			desc string
		}{
			{"Prune Engines", "Remove unused containers/images"},
			{"Vacuum Disk", "Compact WSL disk (vhdx)"},
		}

		for i, t := range tasks {
			prefix := "  "
			if i == m.cCursor {
				prefix = "> "
			}
			taskStyle := lipgloss.NewStyle().Foreground(SecondaryColor)
			content.WriteString(fmt.Sprintf("%s%s\n", prefix, taskStyle.Render(t.name)))
			content.WriteString(fmt.Sprintf("    %s\n\n", t.desc))
		}

	case "Settings":
		content.WriteString(TitleStyle.Render("Settings") + "\n\n")
		content.WriteString("  Controls: [Enter] Change Value\n\n")

		settings := []struct {
			name  string
			value string
		}{
			{"Auto-Start", "ON"},
			{"Default Engine", m.config.DefaultEngine},
			{"Danger Zone", "Reset Environment"},
		}
		if !m.config.AutoStartDaemon {
			settings[0].value = "OFF"
		}

		for i, s := range settings {
			prefix := "  "
			if i == m.sCursor {
				prefix = "> "
			}
			valStyle := SecondaryColor
			if i == 2 {
				valStyle = ErrorColor
			}
			content.WriteString(fmt.Sprintf("%s%-15s [%s]\n", prefix, s.name, lipgloss.NewStyle().Foreground(valStyle).Render(s.value)))
		}

		content.WriteString("\n  " + NormalStyle.Render(fmt.Sprintf("Distro: %s | v%s", wsl.DistroName, wsl.Version)))

	case "Engines":
		content.WriteString(TitleStyle.Render("Engine Management") + "\n\n")
		content.WriteString("  Controls: [i] Install | [s] Start | [t] Stop\n\n")
		for i, e := range m.engines {
			prefix := "  "
			if i == m.eCursor {
				prefix = "> "
			}
			status := "Not Installed"
			if e.Version != "Not Installed" {
				status = "Installed"
				if e.Running {
					status = "Running"
				}
			}
			content.WriteString(fmt.Sprintf("%s%-10s [%s]\n", prefix, e.Name, status))
		}

	case "WSL Distros":
		content.WriteString(TitleStyle.Render("WSL Distributions") + "\n\n")
		content.WriteString("  Controls: [s] Start | [t] Stop | [X] Delete\n\n")
		for i, d := range m.distros {
			prefix := "  "
			if i == m.dCursor {
				prefix = "> "
			}
			stateColor := SecondaryColor
			if d.State == "Stopped" {
				stateColor = ErrorColor
			}
			state := lipgloss.NewStyle().Foreground(stateColor).Render(d.State)
			line := fmt.Sprintf("%s%-24s %-10s v%s", prefix, d.Name, state, d.Version)
			if d.IsDefault {
				line += " (Default)"
			}
			content.WriteString(line + "\n")
		}

	default: // Dashboard
		content.WriteString(TitleStyle.Render("Local Engines") + "\n\n")
		for _, e := range m.engines {
			statusColor := SecondaryColor
			statusText := "Running"
			if !e.Running {
				statusColor = ErrorColor
				statusText = "Stopped"
				if e.Version == "Not Installed" {
					statusText = "Not Found"
				}
			}
			statusItem := lipgloss.NewStyle().Foreground(statusColor).Render("● " + statusText)
			content.WriteString(fmt.Sprintf("  %-10s %-12s %s\n", e.Name, statusItem, e.Version))
		}
	}

	// Join Panels
	mainPanel := lipgloss.JoinHorizontal(lipgloss.Top,
		BoxStyle.Width(20).Render(menu.String()),
		BoxStyle.Width(50).Render(content.String()),
	)
	b.WriteString(mainPanel + "\n")

	// Info Message
	if m.infoMsg != "" {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render(m.infoMsg) + "\n")
	}

	// Footer
	footer := FooterStyle.Render(" ↑/↓: move • enter: select • esc: back • q: quit ")
	b.WriteString(footer)

	return b.String()
}

func Start() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
