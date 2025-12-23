package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00d7ff")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00d7ff")).
			Padding(0, 1)

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d75fff")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00d7ff"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// Agent represents a Sliver agent
type Agent struct {
	ID            string
	Hostname      string
	Username      string
	OS            string
	Transport     string
	RemoteAddress string
	IsSession     bool
	IsPrivileged  bool
	IsDead        bool
	Children      []Agent
}

// Stats holds statistics
type Stats struct {
	Sessions    int
	Beacons     int
	Hosts       int
	Compromised int
}

// Model represents the application state
type model struct {
	agents     []Agent
	stats      Stats
	spinner    spinner.Model
	loading    bool
	err        error
	lastUpdate time.Time
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, fetchAgentsCmd
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case agentsMsg:
		m.agents = msg.agents
		m.stats = msg.stats
		m.loading = false
		m.lastUpdate = time.Now()
		m.err = nil
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return refreshMsg{}
		})

	case refreshMsg:
		m.loading = true
		return m, fetchAgentsCmd

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit, 'r' to retry", m.err)
	}

	var lines []string

	// Title
	title := titleStyle.Render("🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION")
	lines = append(lines, title)
	lines = append(lines, "")

	// Status
	statusText := fmt.Sprintf("⏰ Last Update: %s  |  Press Ctrl+C to exit",
		m.lastUpdate.Format("2006-01-02 15:04:05"))
	lines = append(lines, statusStyle.Render(statusText))
	lines = append(lines, "")

	// Logo
	logo := []string{
		"   🎯 C2    ",
		"  ▄████▄   ",
		"  ████████  ",
		"  ▀██████▀  ",
		"    ▀██▀    ",
	}

	// Agents with logo on left
	if len(m.agents) == 0 {
		lines = append(lines, "No agents connected")
	} else {
		agentLines := m.renderAgents()
		logoStart := len(agentLines)/2 - len(logo)/2
		if logoStart < 0 {
			logoStart = 0
		}

		for i, agentLine := range agentLines {
			var logoLine string
			if i >= logoStart && i < logoStart+len(logo) {
				logoLine = logoStyle.Render(logo[i-logoStart])
			} else {
				logoLine = strings.Repeat(" ", 12)
			}
			lines = append(lines, logoLine+"      "+agentLine)
		}
	}

	// Stats footer
	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", 80))
	
	statsLine := fmt.Sprintf("🟢 Active Sessions: %d  🟡 Active Beacons: %d  🔵 Total Compromised: %d",
		m.stats.Sessions, m.stats.Beacons, m.stats.Compromised)
	lines = append(lines, statsLine)

	return strings.Join(lines, "\n")
}

func (m model) renderAgents() []string {
	var lines []string

	for _, agent := range m.agents {
		lines = append(lines, m.renderAgentLine(agent))
	}

	return lines
}

func (m model) renderAgentLine(agent Agent) string {
	// Status icon
	var statusIcon string
	var statusColor lipgloss.Color
	
	if agent.IsSession {
		statusIcon = "◆"
		statusColor = lipgloss.Color("#00ff00") // Green
	} else {
		statusIcon = "◇"  
		statusColor = lipgloss.Color("#ffff00") // Yellow
	}

	// OS icon
	osIcon := "💻"
	if strings.Contains(strings.ToLower(agent.OS), "windows") {
		if agent.IsSession {
			osIcon = "🖥️"
		} else {
			osIcon = "💻"
		}
	} else if strings.Contains(strings.ToLower(agent.OS), "linux") {
		osIcon = "🐧"
	}

	// Username color
	var usernameColor lipgloss.Color
	if agent.IsPrivileged {
		usernameColor = lipgloss.Color("#ff0000") // Red
	} else {
		usernameColor = lipgloss.Color("#00d7ff") // Cyan
	}

	// Protocol color
	protocolColor := lipgloss.Color("#00d7ff") // Cyan for MTLS

	// Privilege badge
	privBadge := ""
	if agent.IsPrivileged {
		privBadge = " 💎"
	}

	// Build the line
	line := fmt.Sprintf("——[ %s ]——▶ %s %s  %s%s  %s (%s)",
		lipgloss.NewStyle().Foreground(protocolColor).Render(agent.Transport),
		lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
		osIcon,
		lipgloss.NewStyle().Foreground(usernameColor).Bold(true).Render(fmt.Sprintf("%s@%s", agent.Username, agent.Hostname)),
		privBadge,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(statusColor).Render(func() string {
			if agent.IsSession {
				return "session"
			}
			return "beacon"
		}()),
	)

	return line
}

// Messages
type agentsMsg struct {
	agents []Agent
	stats  Stats
}

type refreshMsg struct{}

type errMsg struct {
	err error
}

// Commands
func fetchAgentsCmd() tea.Msg {
	// TODO: Connect to Sliver and fetch real data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate network delay
	time.Sleep(300 * time.Millisecond)

	// Mock data
	agents := []Agent{
		{
			ID:            "22bf4a82-abc123",
			Hostname:      "cywebdw",
			Username:      "NT AUTHORITY\\NETWORK SERVICE",
			OS:            "windows",
			Transport:     "MTLS",
			RemoteAddress: "10.10.110.10:50199",
			IsSession:     true,
			IsPrivileged:  false,
		},
		{
			ID:            "4370d26a-def456",
			Hostname:      "m3dc",
			Username:      "M3C\\Administrator",
			OS:            "windows",
			Transport:     "MTLS",
			RemoteAddress: "10.10.110.250:63805",
			IsSession:     false,
			IsPrivileged:  true,
		},
		{
			ID:            "b773f522-ghi789",
			Hostname:      "m3webaw",
			Username:      "M3C\\Administrator",
			OS:            "windows",
			Transport:     "MTLS",
			RemoteAddress: "10.10.110.250:21392",
			IsSession:     false,
			IsPrivileged:  true,
		},
		{
			ID:            "263f4501-jkl012",
			Hostname:      "cywebdw",
			Username:      "NT AUTHORITY\\NETWORK SERVICE",
			OS:            "windows",
			Transport:     "MTLS",
			RemoteAddress: "10.10.110.10:49908",
			IsSession:     false,
			IsPrivileged:  false,
		},
	}

	stats := Stats{
		Sessions:    1,
		Beacons:     3,
		Hosts:       3,
		Compromised: 4,
	}

	_ = ctx

	return agentsMsg{
		agents: agents,
		stats:  stats,
	}
}

func main() {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d7ff"))

	// Initialize model
	m := model{
		agents:  []Agent{},
		spinner: s,
		loading: true,
	}

	// Create and run program
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
