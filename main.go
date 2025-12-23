package main

import (
	"context"
	"fmt"
	"os"
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
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#00d7ff")).
			Padding(0, 1)

	agentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			MarginTop(1)

	sessionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true)

	beaconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00")).
			Bold(true)

	privilegedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Bold(true)

	ipStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00d7ff"))

	statsStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 2).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

// Agent represents a Sliver agent (session or beacon)
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

// Stats holds statistics about all agents
type Stats struct {
	TotalSessions int
	TotalBeacons  int
	UniqueHosts   int
	Privileged    int
	Standard      int
	Windows       int
	Linux         int
	NewAgents     int
}

// Model represents the Bubble Tea application state
type model struct {
	agents       []Agent
	stats        Stats
	spinner      spinner.Model
	loading      bool
	err          error
	lastUpdate   time.Time
	refreshTimer *time.Timer
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
	)
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Manual refresh
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
		// Schedule next auto-refresh in 5 seconds
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

// View renders the UI
func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit, 'r' to retry", m.err)
	}

	var s string

	// Title
	title := titleStyle.Render("🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION")
	s += title + "\n\n"

	// Status bar
	statusText := fmt.Sprintf("⏰ Last Update: %s", m.lastUpdate.Format("15:04:05"))
	if m.loading {
		statusText += fmt.Sprintf("  %s Refreshing...", m.spinner.View())
	}
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(statusText) + "\n"

	// Stats summary
	if m.stats.TotalSessions > 0 || m.stats.TotalBeacons > 0 {
		statsText := fmt.Sprintf(
			"🟢 Sessions: %d  🟡 Beacons: %d  🔵 Hosts: %d  🔴 Privileged: %d  🟢 Standard: %d",
			m.stats.TotalSessions,
			m.stats.TotalBeacons,
			m.stats.UniqueHosts,
			m.stats.Privileged,
			m.stats.Standard,
		)
		
		if m.stats.Windows > 0 {
			statsText += fmt.Sprintf("  💻 OS: Windows(%d)", m.stats.Windows)
		}
		if m.stats.Linux > 0 {
			statsText += fmt.Sprintf(" Linux(%d)", m.stats.Linux)
		}

		s += statsStyle.Render(statsText) + "\n"
	}

	// Agents list
	if len(m.agents) == 0 {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("No agents connected") + "\n"
	} else {
		for _, agent := range m.agents {
			s += m.renderAgent(agent, 0) + "\n"
		}
	}

	// Help text
	help := helpStyle.Render("Press 'r' to refresh, 'q' to quit")
	s += "\n" + help

	return s
}

// renderAgent renders a single agent with proper styling
func (m model) renderAgent(agent Agent, indent int) string {
	var content string

	// Determine icon and style based on agent type
	var statusIcon, typeLabel string
	var userStyle lipgloss.Style

	if agent.IsDead {
		statusIcon = "💀"
		typeLabel = "[DEAD]"
		userStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	} else if agent.IsSession {
		statusIcon = "◆"
		typeLabel = "session"
		userStyle = sessionStyle
	} else {
		statusIcon = "◇"
		typeLabel = "beacon"
		userStyle = beaconStyle
	}

	// OS icon
	osIcon := "💻"
	if agent.OS == "windows" {
		if agent.IsSession {
			osIcon = "🖥️"
		} else {
			osIcon = "💻"
		}
	} else if agent.OS == "linux" {
		osIcon = "🐧"
	}

	// Privilege badge
	privBadge := ""
	if agent.IsPrivileged {
		privBadge = " " + privilegedStyle.Render("💎")
	}

	// Build connector
	connector := "╰────────"
	if indent > 0 {
		connector = "    ├───"
	}

	// Line 1: Username@Hostname
	line1 := fmt.Sprintf("%s[ %s ]── %s %s %s%s",
		connector,
		agent.Transport,
		statusIcon,
		osIcon,
		userStyle.Render(fmt.Sprintf("%s@%s", agent.Username, agent.Hostname)),
		privBadge,
	)

	// Line 2: ID
	line2 := fmt.Sprintf("            └─ ID: %s (%s)",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(agent.ID[:8]),
		typeLabel,
	)

	// Line 3: IP
	line3 := fmt.Sprintf("            └─ IP: %s",
		ipStyle.Render(agent.RemoteAddress),
	)

	content = line1 + "\n" + line2 + "\n" + line3

	// Render children
	if len(agent.Children) > 0 {
		for _, child := range agent.Children {
			content += "\n" + m.renderAgent(child, indent+1)
		}
	}

	return content
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
	// For now, return mock data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate network delay
	time.Sleep(500 * time.Millisecond)

	// Mock data for testing
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
			IsDead:        false,
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
			IsDead:        false,
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
			IsDead:        false,
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
			IsDead:        false,
		},
	}

	stats := Stats{
		TotalSessions: 1,
		TotalBeacons:  3,
		UniqueHosts:   3,
		Privileged:    2,
		Standard:      2,
		Windows:       4,
		Linux:         0,
		NewAgents:     0,
	}

	_ = ctx // Use context (will be used when implementing real Sliver connection)

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

	// Create program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
