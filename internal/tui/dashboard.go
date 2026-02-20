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
}

func initialModel() model {
	return model{
		choices: []string{"Dashboard", "Engines", "Cleanup", "Settings", "About", "Exit"},
		engines: wsl.GetAllEnginesStatus(),
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
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.choices[m.cursor] == "Exit" {
				return m, tea.Quit
			}
			m.selected = m.choices[m.cursor]
		}
	}

	return m, nil
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
	if m.selected == "About" {
		content.WriteString(TitleStyle.Render("About ezship") + "\n\n")
		content.WriteString("  ezship - Multi-Engine Control Panel\n")
		content.WriteString("  Version: 0.1.0-alpha\n\n")
		content.WriteString(fmt.Sprintf("  Author: %s\n", "Jackson Wendel Santos Sá"))
		content.WriteString(fmt.Sprintf("  Email:  %s\n", "jacksonwendel@gmail.com"))
		content.WriteString(fmt.Sprintf("  Repo:   %s\n", "github.com/wendelmax/ezship"))
	} else {
		content.WriteString(TitleStyle.Render("Local Engines") + "\n\n")
		for _, e := range m.engines {
			statusColor := SecondaryColor
			statusText := "Running"
			if !e.Running {
				statusColor = ErrorColor
				statusText = "Stopped"
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

	// Footer
	footer := FooterStyle.Render(" ↑/↓: move • enter: select • q: quit ")
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
