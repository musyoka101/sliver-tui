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
	termWidth    int  // Terminal width for responsive layout
	termHeight   int  // Terminal height
	ready        bool // Viewport initialized
	themeIndex   int  // Current theme index
	theme        Theme // Current theme
	viewIndex    int  // Current view index
	view         View // Current view
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
		
		// Theme switching
		case "t":
			m.themeIndex = (m.themeIndex + 1) % GetThemeCount()
			m.theme = GetTheme(m.themeIndex)
			// Update viewport content with new theme
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// View switching
		case "v":
			m.viewIndex = (m.viewIndex + 1) % GetViewCount()
			m.view = GetView(m.viewIndex)
			// Update viewport content with new view
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
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
			// Header: title(1 line, no border) + status(1) + empty(1) = 3 lines
			// Footer: empty(1) + separator(1) + empty(1) + stats(1) + lost?(0-1) + empty(1) + help(1) + empty(1) = ~7 lines
			headerFooterHeight := 10 // Reserve 3 for header + 7 for footer
			m.viewport = viewport.New(msg.Width, msg.Height-headerFooterHeight)
			m.viewport.YPosition = 3 // Start after header (3 lines)
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
	// Build header (title + status) - this is FIXED at top, not scrollable
	var headerLines []string
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.TitleColor)
	title := titleStyle.Render("🎯 Sliver C2 Network Topology")
	headerLines = append(headerLines, title)
	
	statusStyle := lipgloss.NewStyle().Foreground(m.theme.StatusColor).Italic(true).MarginBottom(1)
	statusText := fmt.Sprintf("Last Update: %s", m.lastUpdate.Format("15:04:05"))
	if m.ready && len(m.agents) > 0 {
		scrollPercent := int(m.viewport.ScrollPercent() * 100)
		statusText += fmt.Sprintf("  │  Scroll: %d%%", scrollPercent)
	}
	if m.termWidth > 0 && m.termHeight > 0 {
		statusText += fmt.Sprintf("  │  Term: %dx%d", m.termWidth, m.termHeight)
	}
	statusText += fmt.Sprintf("  │  Theme: %s  │  View: %s", m.theme.Name, m.view.Name)
	headerLines = append(headerLines, statusStyle.Render(statusText))
	headerLines = append(headerLines, "")
	
	// Build scrollable content area (agents)
	var contentLines []string
	if len(m.agents) == 0 {
		contentLines = append(contentLines, "  No agents connected")
		contentLines = append(contentLines, "")
	} else if m.ready {
		// Use viewport for scrolling
		contentLines = append(contentLines, m.viewport.View())
	} else {
		// Initial render before viewport ready
		agentLines := m.renderAgents()
		logo := []string{
			"   🎯 C2    ",
			"  ▄████▄   ",
			"  ████████  ",
			"  ▀██████▀  ",
			"    ▀██▀    ",
		}
		logoStyle := lipgloss.NewStyle().Foreground(m.theme.LogoColor).Bold(true)
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
	}
	
	// Build footer (stats + help)
	var footerLines []string
	footerLines = append(footerLines, "")
	separatorStyle := lipgloss.NewStyle().Foreground(m.theme.SeparatorColor)
	footerLines = append(footerLines, separatorStyle.Render(strings.Repeat("─", 90)))
	footerLines = append(footerLines, "")
	
	statsStyle := lipgloss.NewStyle().Foreground(m.theme.StatsColor).Bold(true).Padding(0, 1)
	statsLine := fmt.Sprintf("🟢 Sessions: %d  │  🟡 Beacons: %d  │  🔵 Total: %d",
		m.stats.Sessions, m.stats.Beacons, m.stats.Compromised)
	footerLines = append(footerLines, statsStyle.Render(statsLine))
	
	trackerMutex.RLock()
	lostCount := len(lostAgents)
	trackerMutex.RUnlock()
	
	if lostCount > 0 {
		lostLine := fmt.Sprintf("⚠️  Recently Lost: %d (displayed for %d min)",
			lostCount, int(lostAgentTimeout.Minutes()))
		footerLines = append(footerLines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff9900")).
			Italic(true).
			Render(lostLine))
	}
	
	footerLines = append(footerLines, "")
	helpStyle := lipgloss.NewStyle().Foreground(m.theme.HelpColor)
	helpText := "  Press 'r' to refresh  │  't' to change theme  │  'v' to change view  │  '↑↓' or 'j/k' to scroll  │  'q' to quit"
	footerLines = append(footerLines, helpStyle.Render(helpText))
	footerLines = append(footerLines, "")
	
	// Combine header + content + footer for left side
	leftContent := strings.Join(append(append(headerLines, contentLines...), footerLines...), "\n")
	
	// Add tactical panel if we have agents - position it as an absolute overlay
	tacticalPanel := m.renderTacticalPanel()
	if len(m.agents) > 0 && tacticalPanel != "" && m.termWidth > 100 {
		// DEBUG: Write view composition details
		if os.Getenv("DEBUG_VIEW") == "1" {
			debugInfo := fmt.Sprintf("=== VIEW DEBUG ===\n"+
				"Header lines: %d\n"+
				"Content lines: %d\n"+
				"Footer lines: %d\n"+
				"Left total: %d\n"+
				"Panel lines: %d\n"+
				"Term: %dx%d\n"+
				"Panel X pos: %d\n",
				len(headerLines),
				len(contentLines),
				len(footerLines),
				len(strings.Split(leftContent, "\n")),
				len(strings.Split(tacticalPanel, "\n")),
				m.termWidth, m.termHeight,
				m.termWidth-37)
			os.WriteFile("/tmp/view_debug.txt", []byte(debugInfo), 0644)
		}
		
		// Calculate panel position (right edge minus panel width)
		// Panel is Width(35) + Padding(1,2) + Border = 37 chars total
		panelWidth := 37
		panelX := m.termWidth - panelWidth
		if panelX < 100 {
			panelX = 100
		}
		
		// Split content into lines
		leftLines := strings.Split(leftContent, "\n")
		panelLines := strings.Split(tacticalPanel, "\n")
		
		// Calculate how many header lines we have (these should not be truncated)
		headerLineCount := len(headerLines)
		
		// Ensure we have enough lines for the full panel
		totalLines := len(leftLines)
		if len(panelLines) > totalLines {
			totalLines = len(panelLines)
		}
		
		// Build output by overlaying panel on right side at fixed position
		var result []string
		for i := 0; i < totalLines; i++ {
			var line string
			
			// Add left content line (if exists)
			if i < len(leftLines) {
				line = leftLines[i]
			}
			
			// Calculate visual width (handles ANSI codes correctly)
			currentWidth := lipgloss.Width(line)
			
			// For header lines (title, status), don't truncate - let them extend fully
			// Only truncate content/footer lines that might overlap with panel
			if i < headerLineCount {
				// Header line - just keep as is, don't pad or truncate
				// Panel won't be overlaid on these lines
			} else {
				// Content/footer line - pad or truncate to fit
				if currentWidth < panelX {
					line += strings.Repeat(" ", panelX-currentWidth)
				} else if currentWidth > panelX {
					// Line is too long - truncate it to make room for panel
					// Use lipgloss MaxWidth to preserve ANSI codes
					line = lipgloss.NewStyle().MaxWidth(panelX).Render(line)
					// Ensure we're exactly at panelX width
					currentWidth = lipgloss.Width(line)
					if currentWidth < panelX {
						line += strings.Repeat(" ", panelX-currentWidth)
					}
				}
				
				// Overlay panel line (adjusted for header offset)
				panelLineIndex := i - headerLineCount
				if panelLineIndex >= 0 && panelLineIndex < len(panelLines) {
					line += panelLines[panelLineIndex]
				}
			}
			
			result = append(result, line)
		}
		
		return strings.Join(result, "\n")
	}
	
	return leftContent
}

func (m model) renderTacticalPanel() string {
	if len(m.agents) == 0 {
		return ""
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(35)

	headerStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection).
		Bold(true)
		// Removed MarginTop(1) - we'll add empty lines manually instead

	valueStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalValue)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)

	var lines []string

	// Header
	headerText := "📊 TACTICAL INTELLIGENCE"
	// Add panel height indicator for debugging
	if m.termHeight > 0 {
		// Calculate approx panel lines (will be rendered with border/padding)
		headerText += fmt.Sprintf(" (H:%d)", m.termHeight)
	}
	lines = append(lines, headerStyle.Render(headerText))
	lines = append(lines, "")

	// Analyze data
	subnetHosts := make(map[string]map[string]bool) // subnet -> unique hostnames
	domains := make(map[string]int)
	osHosts := make(map[string]map[string]bool) // OS type -> unique hostnames
	transports := make(map[string]int)
	privilegedCount := 0
	pivotCount := 0
	newCount := 0

	for _, agent := range m.agents {
		// Extract subnet (first 3 octets) and track unique hostnames
		if agent.RemoteAddress != "" {
			parts := strings.Split(agent.RemoteAddress, ".")
			if len(parts) >= 3 {
				subnet := strings.Join(parts[:3], ".") + ".0/24"
				if subnetHosts[subnet] == nil {
					subnetHosts[subnet] = make(map[string]bool)
				}
				subnetHosts[subnet][agent.Hostname] = true // Track unique hostnames
			}
		}

		// Extract domain
		if strings.Contains(agent.Username, "\\") {
			domain := strings.Split(agent.Username, "\\")[0]
			domains[domain]++
		}

		// Count OS by unique hostnames
		if agent.OS != "" {
			osType := "Unknown"
			if strings.Contains(strings.ToLower(agent.OS), "windows") {
				osType = "Windows"
			} else if strings.Contains(strings.ToLower(agent.OS), "linux") {
				osType = "Linux"
			} else if strings.Contains(strings.ToLower(agent.OS), "darwin") {
				osType = "macOS"
			}
			if osHosts[osType] == nil {
				osHosts[osType] = make(map[string]bool)
			}
			osHosts[osType][agent.Hostname] = true // Track unique hostnames per OS
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
	if len(subnetHosts) > 0 {
		for subnet, hosts := range subnetHosts {
			hostCount := len(hosts)
			lines = append(lines, fmt.Sprintf("  %s %s",
				valueStyle.Render(subnet),
				mutedStyle.Render(fmt.Sprintf("(%d hosts)", hostCount))))
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
	if len(osHosts) > 0 {
		for os, hosts := range osHosts {
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
				valueStyle.Render(fmt.Sprintf("%d", len(hosts)))))
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

	rendered := panelStyle.Render(strings.Join(lines, "\n"))
	
	// DEBUG: Write panel content to file for inspection
	if os.Getenv("DEBUG_PANEL") == "1" {
		panelDebug := fmt.Sprintf("=== PANEL DEBUG (Lines: %d) ===\n%s\n=== RAW LINES ===\n%s\n",
			len(strings.Split(rendered, "\n")),
			rendered,
			strings.Join(lines, "\n"))
		os.WriteFile("/tmp/tactical_panel_debug.txt", []byte(panelDebug), 0644)
	}
	
	return rendered
}

// updateViewportContent updates the viewport with the current agent list
func (m *model) updateViewportContent() {
	// Render agents to string
	agentLines := m.renderAgents()
	
	// Logo definition
	logo := []string{
		"   🎯 C2    ",
		"  ▄████▄   ",
		"  ████████  ",
		"  ▀██████▀  ",
		"    ▀██▀    ",
	}
	
	logoStyle := lipgloss.NewStyle().Foreground(m.theme.LogoColor).Bold(true)
	
	var contentLines []string
	
	if m.view.Type == ViewTypeTree {
		// Tree view: Logo overlaid on left side (original behavior)
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
	} else {
		// Box view: Logo at top, then vertical line connecting to boxes
		connectorColor := m.theme.TacticalBorder
		
		// Render logo with proper indentation
		for _, logoLine := range logo {
			contentLines = append(contentLines, logoStyle.Render(logoLine))
		}
		
		// Add vertical line from logo down to first box
		// The vertical line should align with the left edge where boxes connect
		contentLines = append(contentLines, lipgloss.NewStyle().Foreground(connectorColor).Render("│"))
		contentLines = append(contentLines, lipgloss.NewStyle().Foreground(connectorColor).Render("│"))
		
		// Add boxes with arrows pointing from the vertical line
		for _, agentLine := range agentLines {
			contentLines = append(contentLines, agentLine)
		}
	}
	
	// Set viewport content
	m.viewport.SetContent(strings.Join(contentLines, "\n"))
}

func (m model) renderAgents() []string {
	var lines []string

	// Build hierarchical tree
	tree := buildAgentTree(m.agents)

	// Render tree with indentation using current view
	for i, agent := range tree {
		hasNext := i < len(tree)-1
		lines = append(lines, m.renderAgentTreeWithViewAndContext(agent, 0, m.view.Type, hasNext, !hasNext)...)
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
		statusColor = m.theme.DeadColor
	} else if agent.IsSession {
		statusIcon = "◆"
		statusColor = m.theme.SessionColor
	} else {
		statusIcon = "◇"  
		statusColor = m.theme.BeaconColor
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

	// Username color (dead overrides all)
	var usernameColor lipgloss.Color
	if agent.IsDead {
		usernameColor = m.theme.DeadColor
	} else if agent.IsPrivileged {
		usernameColor = m.theme.PrivilegedUser
	} else {
		usernameColor = m.theme.NormalUser
	}

	// Protocol color based on transport type
	var protocolColor lipgloss.Color
	if agent.IsDead {
		protocolColor = m.theme.DeadColor
	} else {
		transportLower := strings.ToLower(agent.Transport)
		switch {
		case strings.Contains(transportLower, "mtls"):
			protocolColor = m.theme.ProtocolMTLS
		case strings.Contains(transportLower, "http"):
			protocolColor = m.theme.ProtocolHTTP
		case strings.Contains(transportLower, "dns"):
			protocolColor = m.theme.ProtocolDNS
		case strings.Contains(transportLower, "tcp"):
			protocolColor = m.theme.ProtocolTCP
		default:
			protocolColor = m.theme.ProtocolDefault
		}
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
			Foreground(m.theme.NewBadgeColor).
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
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.RemoteAddress),
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
	
	// Initialize with default theme (index 0)
	defaultTheme := GetTheme(0)
	s.Style = lipgloss.NewStyle().Foreground(defaultTheme.TitleColor)
	
	// Initialize with default view (index 0)
	defaultView := GetView(0)

	// Initialize model with default terminal size as fallback
	m := model{
		agents:     []Agent{},
		spinner:    s,
		loading:    true,
		termWidth:  180, // Default fallback width
		termHeight: 40,  // Default fallback height
		themeIndex: 0,   // Start with default theme
		theme:      defaultTheme,
		viewIndex:  0,   // Start with default view
		view:       defaultView,
	}

	// Create and run program with alt screen
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
