package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wendelmax/ezship/internal/wsl"
)

const (
	maxLogs   = 50
	logPanelH = 6
	mainWidth = 73 // sidebar(20) + content(50) + gap(3)
)

type model struct {
	choices   []string
	cursor    int
	selected  string
	engines   []wsl.EngineInfo
	distros   []wsl.DistroInfo
	dCursor   int
	eCursor   int
	cCursor   int
	sCursor   int
	config    wsl.Config
	logs      []string // rolling log lines
	logScroll int      // scroll offset (from bottom)
	width     int
	height    int
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
type enginesLoadedMsg []wsl.EngineInfo
type distrosLoadedMsg []wsl.DistroInfo

// --- Log helpers ---

func (m *model) addLog(line string) {
	ts := time.Now().Format("15:04:05")
	m.logs = append(m.logs, fmt.Sprintf("[%s] %s", ts, line))
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
	m.logScroll = 0 // auto-scroll to bottom on new log
}

// --- Init ---

func initialModel() model {
	return model{
		choices:  []string{"Dashboard", "Engines", "WSL Distros", "Cleanup", "Settings", "About", "Update", "Exit"},
		cursor:   0,
		engines:  nil,
		distros:  nil,
		config:   wsl.LoadConfig(),
		selected: "Dashboard",
		logs:     []string{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return refreshMsg{} },
		m.tickCmd(),
	)
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(true)
		case "down", "j":
			m.moveCursor(false)
		case "pgup":
			m.logScroll++
		case "pgdn":
			if m.logScroll > 0 {
				m.logScroll--
			}
		case "enter", " ":
			return m.handleSelection()
		case "s": // Start
			return m, m.cmdStart()
		case "t": // Stop
			return m, m.cmdStop()
		case "i": // Install engine
			if m.selected == "Engines" && len(m.engines) > 0 {
				engine := m.engines[m.eCursor].Name
				m.addLog("Queued install for " + engine + "...")
				return m, m.cmdInstall(engine)
			}
		case "r": // Manual refresh
			m.addLog("Manual refresh triggered")
			return m, func() tea.Msg { return refreshMsg{} }
		case "esc", "backspace", "left", "h":
			m.selected = "Dashboard"
		}

	case engineActionMsg:
		if msg.err != nil {
			m.addLog(fmt.Sprintf("ERROR [%s %s]: %s", msg.action, msg.engine, msg.err.Error()))
		} else {
			m.addLog(fmt.Sprintf("OK [%s]: %s completed", msg.action, msg.engine))
		}
		// Immediately refresh engine states after an action
		return m, func() tea.Msg { return enginesLoadedMsg(wsl.GetAllEnginesStatus()) }

	case maintenanceMsg:
		if msg.err != nil {
			m.addLog(fmt.Sprintf("ERROR [%s]: %s", msg.task, msg.err.Error()))
		} else {
			m.addLog(fmt.Sprintf("OK [%s]: completed", msg.task))
		}
		return m, nil

	case refreshMsg:
		m.addLog("Refreshing status...")
		return m, tea.Batch(
			func() tea.Msg { return enginesLoadedMsg(wsl.GetAllEnginesStatus()) },
			func() tea.Msg {
				distros, _ := wsl.ListDistros()
				return distrosLoadedMsg(distros)
			},
			m.tickCmd(),
		)

	case enginesLoadedMsg:
		m.engines = msg
		m.addLog(fmt.Sprintf("Engines updated (%d found)", len(m.engines)))
		return m, nil

	case distrosLoadedMsg:
		m.distros = msg
		return m, nil
	}

	return m, nil
}

// --- Cursor helpers ---

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
		m.applyMovement(&m.cCursor, delta, 1)
	case "Settings":
		m.applyMovement(&m.sCursor, delta, 2)
	default:
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

// --- Selection ---

func (m *model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.selected {
	case "Engines":
		if len(m.engines) > 0 {
			engine := m.engines[m.eCursor].Name
			m.addLog("Installing " + engine + "...")
			return *m, m.cmdInstall(engine)
		}
		return *m, nil

	case "Cleanup":
		switch m.cCursor {
		case 0:
			m.addLog("Pruning engines...")
			return *m, m.cmdPrune()
		case 1:
			m.addLog("Vacuuming disk (this may take a while)...")
			return *m, m.cmdVacuum()
		}
		return *m, nil

	case "Settings":
		switch m.sCursor {
		case 0:
			m.config.AutoStartDaemon = !m.config.AutoStartDaemon
			wsl.SaveConfig(m.config)
			m.addLog(fmt.Sprintf("Auto-Start set to %v", m.config.AutoStartDaemon))
		case 1:
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
			m.addLog("Default engine: " + m.config.DefaultEngine)
		case 2:
			m.addLog("Reset: use CLI 'ezship reset' to proceed")
		}
		return *m, nil

	case "Update":
		m.addLog("Starting self-update...")
		return *m, m.cmdUpdate()
	}

	// Main menu navigation
	choice := m.choices[m.cursor]
	if choice == "Exit" {
		return *m, tea.Quit
	}
	m.selected = choice
	return *m, m.enterView()
}

func (m *model) enterView() tea.Cmd {
	switch m.selected {
	case "WSL Distros":
		m.dCursor = 0
		return func() tea.Msg {
			distros, _ := wsl.ListDistros()
			return distrosLoadedMsg(distros)
		}
	case "Engines":
		m.eCursor = 0
		return func() tea.Msg {
			return enginesLoadedMsg(wsl.GetAllEnginesStatus())
		}
	case "Cleanup":
		m.cCursor = 0
	case "Settings":
		m.sCursor = 0
		m.config = wsl.LoadConfig()
	}
	return nil
}

// --- Commands ---

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

func (m *model) cmdUpdate() tea.Cmd {
	return func() tea.Msg {
		err := wsl.SelfUpdate(wsl.Version)
		return maintenanceMsg{task: "Update", err: err}
	}
}

// --- View ---

func (m model) View() string {
	var b strings.Builder

	// Header
	header := HeaderStyle.Render(" ezship - Multi-Engine Control Panel ")
	b.WriteString(header + "\n\n")

	// Sidebar (Menu)
	var menu strings.Builder
	for i, choice := range m.choices {
		if m.selected != "Dashboard" && choice == m.selected {
			menu.WriteString(SelectedStyle.Render(" > "+choice) + "\n")
		} else if m.selected == "Dashboard" && m.cursor == i {
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
		tasks := []struct{ name, desc string }{
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
		settings := []struct{ name, value string }{
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
		content.WriteString("  Controls: [i] Install | [Enter] Install | [s] Start | [t] Stop | [r] Refresh\n\n")
		if len(m.engines) == 0 {
			content.WriteString("  Loading...\n")
		}
		for i, e := range m.engines {
			prefix := "  "
			if i == m.eCursor {
				prefix = "> "
			}
			statusText := "Not Installed"
			statusColor := ErrorColor
			if e.Version != "Not Installed" {
				statusText = "Installed"
				statusColor = SecondaryColor
				if e.Running {
					statusText = "Running"
					statusColor = SuccessColor
				}
			}
			statusStr := lipgloss.NewStyle().Foreground(statusColor).Render(statusText)
			content.WriteString(fmt.Sprintf("%s%-10s [%s]  %s\n", prefix, e.Name, statusStr, e.Version))
		}

	case "WSL Distros":
		content.WriteString(TitleStyle.Render("WSL Distributions") + "\n\n")
		content.WriteString("  Controls: [s] Start | [t] Stop | [r] Refresh\n\n")
		if len(m.distros) == 0 {
			content.WriteString("  Loading...\n")
		}
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
		if len(m.engines) == 0 {
			content.WriteString("  Loading engine status...\n")
		}
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
			if e.Running {
				statusColor = SuccessColor
			}
			statusItem := lipgloss.NewStyle().Foreground(statusColor).Render("● " + statusText)
			content.WriteString(fmt.Sprintf("  %-10s %-20s %s\n", e.Name, statusItem, e.Version))
		}
	}

	// Join main panels
	mainPanel := lipgloss.JoinHorizontal(lipgloss.Top,
		BoxStyle.Width(20).Render(menu.String()),
		BoxStyle.Width(50).Render(content.String()),
	)
	b.WriteString(mainPanel + "\n")

	// --- Scrollable Log Panel ---
	logLines := m.visibleLogs()
	var logBuf strings.Builder
	logBuf.WriteString(lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render("  Logs") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  [PgUp/PgDn scroll]") + "\n")
	for _, line := range logLines {
		logBuf.WriteString("  " + line + "\n")
	}
	logPanel := LogPanelStyle.Width(mainWidth).Render(logBuf.String())
	b.WriteString(logPanel + "\n")

	// Footer
	footer := FooterStyle.Render(" ↑/↓: move • enter: select • s: start • t: stop • r: refresh • esc: back • q: quit ")
	b.WriteString(footer)

	return b.String()
}

// visibleLogs returns the last N log lines respecting logScroll offset
func (m model) visibleLogs() []string {
	total := len(m.logs)
	if total == 0 {
		return []string{lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No logs yet...")}
	}

	// show logPanelH-1 lines (one for header)
	visible := logPanelH - 1
	end := total - m.logScroll
	if end <= 0 {
		end = 1
	}
	start := end - visible
	if start < 0 {
		start = 0
	}

	lines := m.logs[start:end]
	styled := make([]string, len(lines))
	for i, l := range lines {
		c := lipgloss.Color("250")
		if strings.Contains(l, "ERROR") {
			c = ErrorColor
		} else if strings.Contains(l, "OK") {
			c = SuccessColor
		}
		styled[i] = lipgloss.NewStyle().Foreground(c).Render(l)
	}
	return styled
}

func Start() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
