package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Global tracking for agent changes
var (
	agentTracker     = make(map[string]time.Time) // ID -> first seen time
	trackerMutex     sync.RWMutex
	newAgentTimeout  = 5 * time.Minute // Mark as NEW if seen < 5 minutes ago
	lostAgents       = make(map[string]Agent)
	lostAgentTimeout = 5 * time.Minute
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00d7ff")).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#00d7ff")).
			Padding(0, 2).
			MarginBottom(1)

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d75fff")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00d7ff")).
			Bold(true).
			Padding(0, 1)

	agentLineStyle = lipgloss.NewStyle().
			PaddingLeft(2)
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
	IsNew         bool      // Newly discovered (< 5 min)
	FirstSeen     time.Time // When first discovered
	ProxyURL      string    // Non-empty if pivoted through another agent
	ParentID      string    // ID of parent agent (if pivoted)
	Children      []Agent   // Child agents (pivoted through this one)
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
		return fmt.Sprintf("\n  ❌ Error: %v\n\n  Press 'q' to quit, 'r' to retry\n", m.err)
	}

	var lines []string

	// Title with better spacing
	title := titleStyle.Render("🎯  SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION")
	lines = append(lines, "")
	lines = append(lines, title)

	// Status with softer styling
	statusText := fmt.Sprintf("Last Update: %s  │  Press Ctrl+C to exit",
		m.lastUpdate.Format("15:04:05"))
	lines = append(lines, statusStyle.Render(statusText))

	// Logo
	logo := []string{
		"   🎯 C2    ",
		"  ▄████▄   ",
		"  ████████  ",
		"  ▀██████▀  ",
		"    ▀██▀    ",
	}

	// Agents with logo on left and better spacing
	if len(m.agents) == 0 {
		lines = append(lines, "")
		lines = append(lines, "  No agents connected")
		lines = append(lines, "")
	} else {
		lines = append(lines, "")
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
			lines = append(lines, "  "+logoLine+"    "+agentLine)
		}
	}

	// Stats footer with better separator
	lines = append(lines, "")
	lines = append(lines, separatorStyle.Render(strings.Repeat("─", 90)))
	lines = append(lines, "")
	
	statsLine := fmt.Sprintf("🟢 Sessions: %d  │  🟡 Beacons: %d  │  🔵 Total: %d  │  🖥️  Hosts: %d",
		m.stats.Sessions, m.stats.Beacons, m.stats.Compromised, m.stats.Hosts)
	lines = append(lines, statsStyle.Render(statsLine))

	// Show lost agents if any
	trackerMutex.RLock()
	lostCount := len(lostAgents)
	trackerMutex.RUnlock()
	
	if lostCount > 0 {
		lostLine := fmt.Sprintf("⚠️  Recently Lost: %d (displayed for %d min)",
			lostCount, int(lostAgentTimeout.Minutes()))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff9900")).
			Italic(true).
			Render(lostLine))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Press 'r' to refresh manually  │  'q' to quit"))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

func (m model) renderAgents() []string {
	var lines []string

	// Build hierarchical tree
	tree := buildAgentTree(m.agents)

	// Render tree with indentation
	for _, agent := range tree {
		lines = append(lines, m.renderAgentTree(agent, 0)...)
	}

	return lines
}

func (m model) renderAgentTree(agent Agent, depth int) []string {
	var lines []string
	
	// Render current agent (returns 2 lines now)
	indent := strings.Repeat("  ", depth)
	agentLines := m.renderAgentLine(agent)
	
	if depth > 0 {
		// Add tree connector for child agents with better styling
		connector := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Render("  ╰─")
		
		// First line gets the connector
		agentLines[0] = indent + connector + agentLines[0]
		// Second line gets matching indentation
		agentLines[1] = indent + "    " + agentLines[1]
	}
	
	// Add both lines
	lines = append(lines, agentLines...)
	
	// Add spacing between agents at root level
	if depth == 0 && len(agent.Children) == 0 {
		lines = append(lines, "")
	}
	
	// Recursively render children
	for _, child := range agent.Children {
		childLines := m.renderAgentTree(child, depth+1)
		lines = append(lines, childLines...)
	}
	
	return lines
}

func (m model) renderAgentLine(agent Agent) []string {
	var lines []string
	
	// Status icon
	var statusIcon string
	var statusColor lipgloss.Color
	
	if agent.IsDead {
		statusIcon = "💀"
		statusColor = lipgloss.Color("#626262") // Gray for dead
	} else if agent.IsSession {
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

	// Username color (gray if dead)
	var usernameColor lipgloss.Color
	if agent.IsDead {
		usernameColor = lipgloss.Color("#626262") // Gray
	} else if agent.IsPrivileged {
		usernameColor = lipgloss.Color("#ff5555") // Softer red
	} else {
		usernameColor = lipgloss.Color("#50fa7b") // Softer cyan/green
	}

	// Protocol color
	protocolColor := lipgloss.Color("#8be9fd") // Softer cyan for MTLS
	if agent.IsDead {
		protocolColor = lipgloss.Color("#626262") // Gray for dead
	}

	// Privilege badge
	privBadge := ""
	if agent.IsPrivileged && !agent.IsDead {
		privBadge = "  💎"
	}

	// NEW badge
	newBadge := ""
	if agent.IsNew && !agent.IsDead {
		newBadge = "  " + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f1fa8c")).
			Bold(true).
			Render("✨ NEW")
	}

	// Type label
	typeLabel := "beacon"
	if agent.IsSession {
		typeLabel = "session"
	} else if agent.IsDead {
		typeLabel = "dead"
	}

	// Build first line - main agent info
	line1 := fmt.Sprintf("──────────[ %s ]──────────▶  %s %s  %s%s%s",
		lipgloss.NewStyle().Foreground(protocolColor).Render(agent.Transport),
		lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
		osIcon,
		lipgloss.NewStyle().Foreground(usernameColor).Bold(true).Render(fmt.Sprintf("%s@%s", agent.Username, agent.Hostname)),
		privBadge,
		newBadge,
	)

	// Build second line - details (ID, IP, type)
	line2 := fmt.Sprintf("                                          %s  │  %s  │  (%s)",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")).Render(agent.RemoteAddress),
		lipgloss.NewStyle().Foreground(statusColor).Render(typeLabel),
	)

	lines = append(lines, line1)
	lines = append(lines, line2)

	return lines
}

// buildAgentTree organizes agents into a hierarchical tree based on pivot relationships
func buildAgentTree(agents []Agent) []Agent {
	// Create maps for quick lookup
	agentMap := make(map[string]*Agent)
	for i := range agents {
		agentMap[agents[i].ID] = &agents[i]
	}

	// Identify parent-child relationships
	// ProxyURL format can be like "socks5://agent-id" or contain agent ID
	var rootAgents []Agent
	
	for i := range agents {
		if agents[i].ProxyURL == "" {
			// This is a root agent (directly connected to C2)
			rootAgents = append(rootAgents, agents[i])
		} else {
			// This agent is pivoted through another
			// Try to extract parent ID from ProxyURL
			agents[i].ParentID = extractParentID(agents[i].ProxyURL, agentMap)
		}
	}

	// Build tree structure
	for i := range rootAgents {
		rootAgents[i].Children = findChildren(rootAgents[i].ID, agents)
	}

	// If no root agents found, return all agents as roots
	if len(rootAgents) == 0 {
		return agents
	}

	return rootAgents
}

// extractParentID tries to extract parent agent ID from ProxyURL
func extractParentID(proxyURL string, agentMap map[string]*Agent) string {
	// ProxyURL might be in format like: "socks5://127.0.0.1:9050"
	// For now, we'll look for agents with matching connection info
	// This is a simplified approach - real implementation might need protocol-specific parsing
	for id := range agentMap {
		if strings.Contains(proxyURL, id) {
			return id
		}
	}
	return ""
}

// findChildren recursively finds all children of a given parent agent
func findChildren(parentID string, allAgents []Agent) []Agent {
	var children []Agent
	
	for _, agent := range allAgents {
		if agent.ParentID == parentID {
			// This agent is a child of parentID
			child := agent
			// Recursively find this child's children
			child.Children = findChildren(child.ID, allAgents)
			children = append(children, child)
		}
	}
	
	return children
}

// trackAgentChanges updates the tracking maps for new/lost agents
func trackAgentChanges(agents []Agent) []Agent {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()

	now := time.Now()
	currentAgentIDs := make(map[string]bool)

	// Process current agents
	for i := range agents {
		agentID := agents[i].ID
		currentAgentIDs[agentID] = true

		// Check if this is a new agent
		if firstSeen, exists := agentTracker[agentID]; exists {
			agents[i].FirstSeen = firstSeen
			agents[i].IsNew = now.Sub(firstSeen) < newAgentTimeout
		} else {
			// First time seeing this agent
			agentTracker[agentID] = now
			agents[i].FirstSeen = now
			agents[i].IsNew = true
		}
	}

	// Find lost agents (previously tracked but not in current list)
	for agentID, firstSeen := range agentTracker {
		if !currentAgentIDs[agentID] {
			// This agent is missing, add to lost agents if not already there
			if _, exists := lostAgents[agentID]; !exists {
				// Create a lost agent entry (we don't have full data)
				lostAgents[agentID] = Agent{
					ID:        agentID,
					FirstSeen: firstSeen,
					IsDead:    true,
				}
			}
		}
	}

	// Clean up old lost agents
	for agentID, agent := range lostAgents {
		if now.Sub(agent.FirstSeen) > lostAgentTimeout {
			delete(lostAgents, agentID)
			delete(agentTracker, agentID)
		}
	}

	return agents
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
	// Connect to Sliver and fetch real data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	agents, stats, err := FetchAgents(ctx)
	if err != nil {
		return errMsg{err: err}
	}

	// Track agent changes (NEW badges, lost agents)
	agents = trackAgentChanges(agents)

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
