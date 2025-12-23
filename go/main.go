package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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
	agents       []Agent
	stats        Stats
	spinner      spinner.Model
	viewport     viewport.Model // Scrollable viewport for agent list
	loading      bool
	err          error
	lastUpdate   time.Time
	termWidth    int // Terminal width for responsive layout
	termHeight   int // Terminal height
	ready        bool // Viewport initialized
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, fetchAgentsCmd
		
		// Viewport scrolling controls
		case "up", "k":
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		case "down", "j":
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		case "pgup", "b", "ctrl+u":
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		case "pgdown", "f", "ctrl+d":
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		// Capture terminal dimensions for responsive layout
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		
		if !m.ready {
			// Initialize viewport on first window size message
			// Reserve space for header (5 lines) + footer (5 lines) + tactical panel
			headerFooterHeight := 10
			m.viewport = viewport.New(msg.Width, msg.Height-headerFooterHeight)
			m.viewport.YPosition = 5 // Start after header
			m.ready = true
		} else {
			// Update viewport dimensions on resize
			headerFooterHeight := 10
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerFooterHeight
		}
		
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case agentsMsg:
		m.agents = msg.agents
		m.stats = msg.stats
		m.loading = false
		m.lastUpdate = time.Now()
		m.err = nil
		
		// Update viewport content with new agent list
		if m.ready {
			m.updateViewportContent()
		}
		
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return refreshMsg{}
		}))

	case refreshMsg:
		m.loading = true
		cmds = append(cmds, fetchAgentsCmd)

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Create main content (left side)
	mainContent := m.renderMainContent()
	
	// Create tactical panel (right side)
	tacticalPanel := m.renderTacticalPanel()
	
	// If no terminal width yet, return just main content
	if m.termWidth == 0 {
		return mainContent
	}
	
	// If no panel (no agents), return just main content
	if tacticalPanel == "" {
		return mainContent
	}
	
	// Calculate dimensions
	// Tactical panel is 35 chars wide + 4 for border/padding = ~39 actual width
	panelWidth := 39
	
	// Get main content width (approximate based on longest line)
	mainLines := strings.Split(mainContent, "\n")
	mainWidth := 0
	for _, line := range mainLines {
		// Use lipgloss.Width to account for ANSI codes
		lineWidth := lipgloss.Width(line)
		if lineWidth > mainWidth {
			mainWidth = lineWidth
		}
	}
	
	// Calculate spacing needed to push panel to far right
	// Formula: termWidth - mainWidth - panelWidth - 2 (for safety margin)
	spacingWidth := m.termWidth - mainWidth - panelWidth - 2
	if spacingWidth < 2 {
		spacingWidth = 2 // Minimum spacing
	}
	
	// Create spacing string
	spacing := strings.Repeat(" ", spacingWidth)
	
	// Join horizontally with calculated spacing
	return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, spacing, tacticalPanel)
}

func (m model) renderMainContent() string {
	var lines []string

	// Title
	title := titleStyle.Render("🎯 Sliver C2 Network Topology")
	lines = append(lines, title)

	// Status with scroll indicator
	statusText := fmt.Sprintf("Last Update: %s",
		m.lastUpdate.Format("15:04:05"))
	
	// Add scroll position indicator if viewport is active
	if m.ready && len(m.agents) > 0 {
		scrollPercent := int(m.viewport.ScrollPercent() * 100)
		statusText += fmt.Sprintf("  │  Scroll: %d%%", scrollPercent)
	}
	lines = append(lines, statusStyle.Render(statusText))

	lines = append(lines, "")

	// Show agents
	if len(m.agents) == 0 {
		lines = append(lines, "  No agents connected")
		lines = append(lines, "")
	} else if m.ready {
		// Use viewport for scrolling if initialized
		lines = append(lines, m.viewport.View())
	} else {
		// Show agents without viewport until it's ready
		agentLines := m.renderAgents()
		logo := []string{
			"   🎯 C2    ",
			"  ▄████▄   ",
			"  ████████  ",
			"  ▀██████▀  ",
			"    ▀██▀    ",
		}
		
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
	
	statsLine := fmt.Sprintf("🟢 Sessions: %d  │  🟡 Beacons: %d  │  🔵 Total: %d",
		m.stats.Sessions, m.stats.Beacons, m.stats.Compromised)
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
	
	// Enhanced help text with scroll controls
	helpText := "  Press 'r' to refresh  │  '↑↓' or 'j/k' to scroll  │  'q' to quit"
	lines = append(lines, helpStyle.Render(helpText))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

func (m model) renderTacticalPanel() string {
	if len(m.agents) == 0 {
		return ""
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00d7ff")).
		Padding(1, 2).
		Width(35)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00d7ff")).
		Bold(true).
		Underline(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f1fa8c")).
		Bold(true).
		MarginTop(1)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50fa7b"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4"))

	var lines []string

	// Header
	lines = append(lines, headerStyle.Render("📊 TACTICAL INTELLIGENCE"))
	lines = append(lines, "")

	// Analyze data
	subnets := make(map[string]int)
	domains := make(map[string]int)
	osCount := make(map[string]int)
	transports := make(map[string]int)
	privilegedCount := 0
	pivotCount := 0
	newCount := 0

	for _, agent := range m.agents {
		// Extract subnet (first 3 octets)
		if agent.RemoteAddress != "" {
			parts := strings.Split(agent.RemoteAddress, ".")
			if len(parts) >= 3 {
				subnet := strings.Join(parts[:3], ".") + ".0/24"
				subnets[subnet]++
			}
		}

		// Extract domain
		if strings.Contains(agent.Username, "\\") {
			domain := strings.Split(agent.Username, "\\")[0]
			domains[domain]++
		}

		// Count OS
		if agent.OS != "" {
			osType := "Unknown"
			if strings.Contains(strings.ToLower(agent.OS), "windows") {
				osType = "Windows"
			} else if strings.Contains(strings.ToLower(agent.OS), "linux") {
				osType = "Linux"
			} else if strings.Contains(strings.ToLower(agent.OS), "darwin") {
				osType = "macOS"
			}
			osCount[osType]++
		}

		// Count transports
		transports[agent.Transport]++

		if agent.IsPrivileged {
			privilegedCount++
		}
		if agent.ParentID != "" {
			pivotCount++
		}
		if agent.IsNew {
			newCount++
		}
	}

	// Compromised Subnets
	lines = append(lines, sectionStyle.Render("🌐 Compromised Subnets"))
	if len(subnets) > 0 {
		for subnet, count := range subnets {
			lines = append(lines, fmt.Sprintf("  %s %s",
				valueStyle.Render(subnet),
				mutedStyle.Render(fmt.Sprintf("(%d hosts)", count))))
		}
	} else {
		lines = append(lines, mutedStyle.Render("  No subnet data"))
	}

	// Domains/Workgroups
	lines = append(lines, "")
	lines = append(lines, sectionStyle.Render("🏢 Domains Discovered"))
	if len(domains) > 0 {
		for domain, count := range domains {
			lines = append(lines, fmt.Sprintf("  %s %s",
				valueStyle.Render(domain),
				mutedStyle.Render(fmt.Sprintf("(%d users)", count))))
		}
	} else {
		lines = append(lines, mutedStyle.Render("  No domain data"))
	}

	// OS Distribution
	lines = append(lines, "")
	lines = append(lines, sectionStyle.Render("💻 OS Distribution"))
	if len(osCount) > 0 {
		for os, count := range osCount {
			icon := "💻"
			if os == "Windows" {
				icon = "🖥️"
			} else if os == "Linux" {
				icon = "🐧"
			} else if os == "macOS" {
				icon = "🍎"
			}
			lines = append(lines, fmt.Sprintf("  %s %s: %s",
				icon,
				os,
				valueStyle.Render(fmt.Sprintf("%d", count))))
		}
	} else {
		lines = append(lines, mutedStyle.Render("  No OS data"))
	}

	// Access Level
	lines = append(lines, "")
	lines = append(lines, sectionStyle.Render("💎 Access Level"))
	lines = append(lines, fmt.Sprintf("  Privileged: %s / %d",
		valueStyle.Render(fmt.Sprintf("%d", privilegedCount)),
		len(m.agents)))
	lines = append(lines, fmt.Sprintf("  Standard: %s",
		mutedStyle.Render(fmt.Sprintf("%d", len(m.agents)-privilegedCount))))

	// Transports
	lines = append(lines, "")
	lines = append(lines, sectionStyle.Render("🔐 Transports"))
	if len(transports) > 0 {
		for transport, count := range transports {
			lines = append(lines, fmt.Sprintf("  %s: %s",
				transport,
				valueStyle.Render(fmt.Sprintf("%d", count))))
		}
	}

	// Pivots
	if pivotCount > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionStyle.Render("🔗 Active Pivots"))
		lines = append(lines, fmt.Sprintf("  Pivoted agents: %s",
			valueStyle.Render(fmt.Sprintf("%d", pivotCount))))
	}

	// Activity
	if newCount > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionStyle.Render("⚡ Recent Activity"))
		lines = append(lines, fmt.Sprintf("  New (< 5min): %s",
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f1fa8c")).
				Bold(true).
				Render(fmt.Sprintf("✨ %d", newCount))))
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

// updateViewportContent updates the viewport with the current agent list
func (m *model) updateViewportContent() {
	// Render agents to string
	agentLines := m.renderAgents()
	
	// Add logo integration
	logo := []string{
		"   🎯 C2    ",
		"  ▄████▄   ",
		"  ████████  ",
		"  ▀██████▀  ",
		"    ▀██▀    ",
	}
	
	var contentLines []string
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
		contentLines = append(contentLines, "  "+logoLine+"    "+agentLine)
	}
	
	// Set viewport content
	m.viewport.SetContent(strings.Join(contentLines, "\n"))
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
