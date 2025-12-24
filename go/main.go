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

// ActivitySample represents a single data point in time
type ActivitySample struct {
	Timestamp       time.Time
	SessionsCount   int
	BeaconsCount    int
	NewCount        int
	PrivilegedCount int
}

// ActivityTracker tracks activity over time (12-hour rolling window)
type ActivityTracker struct {
	StartTime       time.Time
	Samples         []ActivitySample
	SampleInterval  time.Duration // 10 minutes
	MaxSamples      int           // 72 samples (12 hours)
	mutex           sync.RWMutex
}

// NewActivityTracker creates a new activity tracker
func NewActivityTracker() *ActivityTracker {
	return &ActivityTracker{
		StartTime:      time.Now(),
		Samples:        []ActivitySample{},
		SampleInterval: 10 * time.Minute,
		MaxSamples:     72, // 12 hours at 10-minute intervals
	}
}

// AddSample adds a new activity sample (rolling window)
func (at *ActivityTracker) AddSample(sessions, beacons, newAgents, privileged int) {
	at.mutex.Lock()
	defer at.mutex.Unlock()
	
	sample := ActivitySample{
		Timestamp:       time.Now(),
		SessionsCount:   sessions,
		BeaconsCount:    beacons,
		NewCount:        newAgents,
		PrivilegedCount: privileged,
	}
	
	at.Samples = append(at.Samples, sample)
	
	// Keep only last MaxSamples (rolling window)
	if len(at.Samples) > at.MaxSamples {
		at.Samples = at.Samples[len(at.Samples)-at.MaxSamples:]
	}
}

// GetSamples returns a copy of all samples (thread-safe)
func (at *ActivityTracker) GetSamples() []ActivitySample {
	at.mutex.RLock()
	defer at.mutex.RUnlock()
	
	samplesCopy := make([]ActivitySample, len(at.Samples))
	copy(samplesCopy, at.Samples)
	return samplesCopy
}

// GetSessionDuration returns how long the tracker has been running
func (at *ActivityTracker) GetSessionDuration() time.Duration {
	return time.Since(at.StartTime)
}

// sampleCurrentActivity samples the current agent state and adds to tracker
func (m *model) sampleCurrentActivity() {
	if m.activityTracker == nil {
		return
	}
	
	// Count metrics from current agents
	newCount := 0
	privilegedCount := 0
	
	for _, agent := range m.agents {
		if agent.IsNew {
			newCount++
		}
		if agent.IsPrivileged {
			privilegedCount++
		}
	}
	
	// Add sample to tracker
	m.activityTracker.AddSample(
		m.stats.Sessions,
		m.stats.Beacons,
		newCount,
		privilegedCount,
	)
}

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
	
	// Additional fields from protobuf
	PID           int32     // Process ID
	Filename      string    // Process filename/path (Argv[0])
	Arch          string    // Architecture (x64, x86, arm64, etc.)
	Version       string    // Implant version
	ActiveC2      string    // Active C2 server URL
	Interval      int64     // Beacon check-in interval (nanoseconds)
	Jitter        int64     // Beacon jitter
	NextCheckin   int64     // Next beacon check-in time (unix timestamp)
	TasksCount    int64     // Total tasks queued
	TasksCompleted int64    // Tasks completed
	LastCheckin   int64     // Last check-in time (unix timestamp)
	Evasion       bool      // Evasion mode enabled
	Burned        bool      // Marked as compromised
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
	agents          []Agent
	stats           Stats
	spinner         spinner.Model
	viewport        viewport.Model // Scrollable viewport for agent list
	loading         bool
	err             error
	lastUpdate      time.Time
	termWidth       int  // Terminal width for responsive layout
	termHeight      int  // Terminal height
	ready           bool // Viewport initialized
	themeIndex      int  // Current theme index
	theme           Theme // Current theme
	viewIndex       int  // Current view index
	view            View // Current view
	activityTracker *ActivityTracker // Activity tracking over time
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
		sampleActivityCmd, // Start activity sampling timer
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
		
		// Dashboard keybind
		case "d":
			// Toggle to dashboard view directly
			m.viewIndex = 2 // Dashboard is index 2
			m.view = GetView(m.viewIndex)
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
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
		
		// Sample activity immediately when agents are fetched
		m.sampleCurrentActivity()
		
		// Update viewport content with new agent list
		if m.ready {
			m.updateViewportContent()
		}
		
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return refreshMsg{}
		}))

	case activitySampleMsg:
		// Sample activity when timer triggers
		m.sampleCurrentActivity()
		// Schedule next sample
		cmds = append(cmds, sampleActivityCmd)

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
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleColor).
		Background(m.theme.HeaderBg).
		Padding(0, 1)
	title := titleStyle.Render("🎯 Sliver C2 Network Topology")
	headerLines = append(headerLines, title)
	
	statusStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusColor).
		Background(m.theme.HeaderBg).
		Italic(true).
		MarginBottom(1).
		Padding(0, 1)
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
			"    🔥🔥     ",
			"  ▄▄▄▄▄▄▄   ",
			"  █ C2  █   ",
			"  █▓▓▓▓▓█   ",
			"  ▀▀▀▀▀▀▀   ",
		}
		logoStyle := lipgloss.NewStyle().Foreground(m.theme.LogoColor).Bold(true)
		
		// Logo on left, agents on right with connectors from logo area
		logoStart := 0 // Start logo from the top
		if len(agentLines) > len(logo) {
			// If there are more agent lines than logo lines, center the logo
			logoStart = (len(agentLines) - len(logo)) / 2
		}
		
		// Build lines with logo on left and agents on right
		maxLines := len(agentLines)
		if len(logo) > maxLines {
			maxLines = len(logo)
		}
		
		for i := 0; i < maxLines; i++ {
			var logoLine string
			if i >= logoStart && i < logoStart+len(logo) {
				logoLine = logoStyle.Render(logo[i-logoStart])
			} else {
				logoLine = strings.Repeat(" ", 12)
			}
			
			var agentLine string
			if i < len(agentLines) {
				agentLine = agentLines[i]
			}
			
			contentLines = append(contentLines, " "+logoLine+"  "+agentLine)
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
	helpText := "  Press 'r' to refresh  │  't' to change theme  │  'v' to change view  │  'd' for dashboard  │  '↑↓' or 'j/k' to scroll  │  'q' to quit"
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
		Background(m.theme.TacticalPanelBg).
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

// renderDashboard renders the dashboard view with analytics panels
func (m model) renderDashboard() string {
	var content strings.Builder
	
	// Dashboard header
	headerStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true).
		Underline(true).
		MarginBottom(1)
	
	content.WriteString(headerStyle.Render("📊 DASHBOARD - OPERATIONAL ANALYTICS"))
	content.WriteString("\n\n")
	
	// Create 2x2 grid layout for panels
	// Top row: C2 Infrastructure | Architecture Distribution
	// Bottom row: Security Status | Activity Metrics
	
	c2Panel := m.renderC2InfrastructurePanel()
	archPanel := m.renderArchitecturePanel()
	securityPanel := m.renderSecurityStatusPanel()
	sparklinePanel := m.renderSparklinePanel()
	
	// Use lipgloss JoinHorizontal to place panels side by side
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, c2Panel, "  ", archPanel)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, securityPanel, "  ", sparklinePanel)
	
	content.WriteString(topRow)
	content.WriteString("\n\n")
	content.WriteString(bottomRow)
	
	return content.String()
}

// renderC2InfrastructurePanel shows active C2 servers with agent counts
func (m model) renderC2InfrastructurePanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(50).
		Height(15)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection).
		Bold(true)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalValue)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	var lines []string
	lines = append(lines, titleStyle.Render("🌐 C2 INFRASTRUCTURE MAP"))
	lines = append(lines, "")
	
	// Group agents by C2 server
	c2Servers := make(map[string][]Agent)
	for _, agent := range m.agents {
		c2 := agent.ActiveC2
		if c2 == "" {
			c2 = "Unknown"
		}
		c2Servers[c2] = append(c2Servers[c2], agent)
	}
	
	if len(c2Servers) == 0 {
		lines = append(lines, mutedStyle.Render("No C2 data available"))
	} else {
		for server, agents := range c2Servers {
			lines = append(lines, labelStyle.Render(fmt.Sprintf("🌐 %s", server)))
			
			// Count protocols
			protocols := make(map[string]int)
			for _, agent := range agents {
				protocols[agent.Transport]++
			}
			
			// Show agent count
			lines = append(lines, fmt.Sprintf("   %s %s",
				valueStyle.Render(fmt.Sprintf("%d agents", len(agents))),
				mutedStyle.Render("")))
			
			// Show protocol breakdown
			var protoList []string
			for proto, count := range protocols {
				protoList = append(protoList, fmt.Sprintf("%s:%d", proto, count))
			}
			lines = append(lines, fmt.Sprintf("   └─ %s",
				mutedStyle.Render(strings.Join(protoList, ", "))))
			lines = append(lines, "")
		}
	}
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderArchitecturePanel shows architecture distribution
func (m model) renderArchitecturePanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(50).
		Height(15)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalValue).
		Bold(true)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	// Cyan bar style for architecture
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")) // Dark turquoise
	
	var lines []string
	lines = append(lines, titleStyle.Render("🔹 ARCHITECTURE DISTRIBUTION"))
	lines = append(lines, "")
	
	// Count architectures
	archCount := make(map[string]int)
	totalAgents := len(m.agents)
	
	for _, agent := range m.agents {
		arch := agent.Arch
		if arch == "" {
			arch = "unknown"
		}
		archCount[arch]++
	}
	
	if totalAgents == 0 {
		lines = append(lines, mutedStyle.Render("No agents available"))
	} else {
		// Sort by count (simple display)
		for arch, count := range archCount {
			percentage := float64(count) / float64(totalAgents) * 100
			
			// Create bar graph
			barLength := int(percentage / 5) // 5% per block
			if barLength > 20 {
				barLength = 20
			}
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
			
			// Icon based on arch
			icon := "🔹"
			if arch == "x86" || arch == "386" {
				icon = "🔸"
			} else if strings.Contains(arch, "arm") {
				icon = "🔶"
			}
			
			lines = append(lines, fmt.Sprintf("%s %s",
				icon,
				labelStyle.Render(fmt.Sprintf("%-10s", arch))))
			lines = append(lines, fmt.Sprintf("  %s %s",
				barStyle.Render(bar),
				valueStyle.Render(fmt.Sprintf("%3.0f%% (%d)", percentage, count))))
			lines = append(lines, "")
		}
	}
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderTaskQueuePanel shows beacon task queue status
func (m model) renderTaskQueuePanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(50).
		Height(15)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalValue)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	// Cyan bar style for task progress
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")) // Dark turquoise
	
	var lines []string
	lines = append(lines, titleStyle.Render("📋 TASK QUEUE MONITOR"))
	lines = append(lines, "")
	
	// Find beacons with tasks
	beaconsWithTasks := 0
	for _, agent := range m.agents {
		if !agent.IsSession && agent.TasksCount > 0 {
			beaconsWithTasks++
			
			// Show task progress
			percentage := float64(0)
			if agent.TasksCount > 0 {
				percentage = float64(agent.TasksCompleted) / float64(agent.TasksCount) * 100
			}
			
			barLength := int(percentage / 10) // 10% per block
			if barLength > 10 {
				barLength = 10
			}
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)
			
			// Status icon
			statusIcon := "📋"
			if percentage == 100 {
				statusIcon = "✅"
			} else if percentage < 30 {
				statusIcon = "⚠️"
			}
			
			lines = append(lines, fmt.Sprintf("%s %s",
				statusIcon,
				labelStyle.Render(fmt.Sprintf("%-15s", agent.Hostname[:min(15, len(agent.Hostname))]))))
			lines = append(lines, fmt.Sprintf("  %s %s",
				barStyle.Render(bar),
				valueStyle.Render(fmt.Sprintf("%d/%d", agent.TasksCompleted, agent.TasksCount))))
			
			if beaconsWithTasks >= 5 {
				break // Limit to 5 beacons for space
			}
		}
	}
	
	if beaconsWithTasks == 0 {
		lines = append(lines, mutedStyle.Render("No active beacon tasks"))
		lines = append(lines, "")
		lines = append(lines, mutedStyle.Render("All beacons are idle 💤"))
	}
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderSecurityStatusPanel shows agent security states
func (m model) renderSecurityStatusPanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(50).
		Height(15)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection)
	
	stealthStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")). // Medium purple
		Bold(true)
	
	burnedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4500")). // Orange red
		Bold(true)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	var lines []string
	lines = append(lines, titleStyle.Render("🔒 SECURITY STATUS"))
	lines = append(lines, "")
	
	// Count agents by security state
	stealthAgents := []Agent{}
	burnedAgents := []Agent{}
	normalAgents := 0
	
	for _, agent := range m.agents {
		if agent.IsDead {
			continue // Skip dead agents
		}
		
		if agent.Evasion {
			stealthAgents = append(stealthAgents, agent)
		} else if agent.Burned {
			burnedAgents = append(burnedAgents, agent)
		} else {
			normalAgents++
		}
	}
	
	// Show stealth agents
	if len(stealthAgents) > 0 {
		lines = append(lines, stealthStyle.Render("🕵️  STEALTH MODE"))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("   %d agent(s) in evasion mode", len(stealthAgents))))
		lines = append(lines, "")
		for i, agent := range stealthAgents {
			if i >= 3 {
				lines = append(lines, mutedStyle.Render(fmt.Sprintf("   ... and %d more", len(stealthAgents)-3)))
				break
			}
			lines = append(lines, labelStyle.Render(fmt.Sprintf("   • %s", agent.Hostname)))
		}
		lines = append(lines, "")
	}
	
	// Show burned agents
	if len(burnedAgents) > 0 {
		lines = append(lines, burnedStyle.Render("🔥 COMPROMISED"))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("   %d agent(s) burned/detected", len(burnedAgents))))
		lines = append(lines, "")
		for i, agent := range burnedAgents {
			if i >= 3 {
				lines = append(lines, mutedStyle.Render(fmt.Sprintf("   ... and %d more", len(burnedAgents)-3)))
				break
			}
			lines = append(lines, labelStyle.Render(fmt.Sprintf("   • %s", agent.Hostname)))
		}
		lines = append(lines, "")
	}
	
	// Show normal status if no special states
	if len(stealthAgents) == 0 && len(burnedAgents) == 0 {
		lines = append(lines, mutedStyle.Render("All agents operating normally"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render(fmt.Sprintf("✓ %d agents in standard mode", normalAgents)))
	}
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderSparklinePanel shows activity over time
func (m model) renderSparklinePanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(60). // Wider for sparklines + time axis
		Height(18) // Taller for time axis
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true).
		Underline(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	// Green style for sparklines
	sparklineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")) // Bright green
	
	var lines []string
	
	// Get activity samples
	samples := m.activityTracker.GetSamples()
	sessionDuration := m.activityTracker.GetSessionDuration()
	
	// Title with session duration
	durationStr := formatDuration(sessionDuration)
	lines = append(lines, titleStyle.Render("ACTIVITY METRICS (Last 12 Hours)"))
	lines = append(lines, mutedStyle.Render(fmt.Sprintf("Session: %s | Samples: %d/72", 
		durationStr, len(samples))))
	lines = append(lines, "")
	
	sparklineWidth := 48 // Fixed width for all sparklines
	
	if len(samples) == 0 {
		lines = append(lines, mutedStyle.Render("Collecting data... (first sample in 10min)"))
		return panelStyle.Render(strings.Join(lines, "\n"))
	}
	
	// Calculate statistics
	stats := calculateActivityStats(samples)
	
	// Generate sparklines with proper time mapping
	sessionsSparkline := generateHistoricalSparkline(samples, "sessions", sparklineWidth)
	beaconsSparkline := generateHistoricalSparkline(samples, "beacons", sparklineWidth)
	newSparkline := generateHistoricalSparkline(samples, "new", sparklineWidth)
	privilegedSparkline := generateHistoricalSparkline(samples, "privileged", sparklineWidth)
	
	// Sessions
	lines = append(lines, fmt.Sprintf("%s  %s  Peak: %-2d  Now: %-2d",
		labelStyle.Render("Sessions    "),
		sparklineStyle.Render(sessionsSparkline),
		stats.SessionsPeak,
		stats.SessionsCurrent))
	
	// Beacons
	lines = append(lines, fmt.Sprintf("%s  %s  Peak: %-2d  Now: %-2d",
		labelStyle.Render("Beacons     "),
		sparklineStyle.Render(beaconsSparkline),
		stats.BeaconsPeak,
		stats.BeaconsCurrent))
	
	// New Agents
	lines = append(lines, fmt.Sprintf("%s  %s  Peak: %-2d  Now: %-2d",
		labelStyle.Render("New Agents  "),
		sparklineStyle.Render(newSparkline),
		stats.NewPeak,
		stats.NewCurrent))
	
	// Privileged
	lines = append(lines, fmt.Sprintf("%s  %s  Peak: %-2d  Now: %-2d",
		labelStyle.Render("Privileged  "),
		sparklineStyle.Render(privilegedSparkline),
		stats.PrivilegedPeak,
		stats.PrivilegedCurrent))
	
	lines = append(lines, "")
	
	// Time axis (aligned with sparkline)
	timeAxis := generateTimeAxis(samples, sparklineWidth, m.activityTracker.StartTime)
	lines = append(lines, mutedStyle.Render(timeAxis))
	
	lines = append(lines, "")
	
	// Session statistics
	lines = append(lines, labelStyle.Render("Averages"))
	lines = append(lines, mutedStyle.Render(fmt.Sprintf(
		"  Sess: %.1f | Beacons: %.1f | New: %.1f | Priv: %.1f",
		stats.SessionsAvg, stats.BeaconsAvg, stats.NewAvg, stats.PrivilegedAvg)))
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// ActivityStats holds statistical data
type ActivityStats struct {
	SessionsPeak     int
	SessionsCurrent  int
	SessionsAvg      float64
	BeaconsPeak      int
	BeaconsCurrent   int
	BeaconsAvg       float64
	NewPeak          int
	NewCurrent       int
	NewAvg           float64
	PrivilegedPeak   int
	PrivilegedCurrent int
	PrivilegedAvg    float64
}

// calculateActivityStats calculates statistics from samples
func calculateActivityStats(samples []ActivitySample) ActivityStats {
	if len(samples) == 0 {
		return ActivityStats{}
	}
	
	stats := ActivityStats{}
	var sessionsSum, beaconsSum, newSum, privilegedSum int
	
	for i, sample := range samples {
		// Track peaks
		if sample.SessionsCount > stats.SessionsPeak {
			stats.SessionsPeak = sample.SessionsCount
		}
		if sample.BeaconsCount > stats.BeaconsPeak {
			stats.BeaconsPeak = sample.BeaconsCount
		}
		if sample.NewCount > stats.NewPeak {
			stats.NewPeak = sample.NewCount
		}
		if sample.PrivilegedCount > stats.PrivilegedPeak {
			stats.PrivilegedPeak = sample.PrivilegedCount
		}
		
		// Sum for averages
		sessionsSum += sample.SessionsCount
		beaconsSum += sample.BeaconsCount
		newSum += sample.NewCount
		privilegedSum += sample.PrivilegedCount
		
		// Current (last sample)
		if i == len(samples)-1 {
			stats.SessionsCurrent = sample.SessionsCount
			stats.BeaconsCurrent = sample.BeaconsCount
			stats.NewCurrent = sample.NewCount
			stats.PrivilegedCurrent = sample.PrivilegedCount
		}
	}
	
	count := float64(len(samples))
	stats.SessionsAvg = float64(sessionsSum) / count
	stats.BeaconsAvg = float64(beaconsSum) / count
	stats.NewAvg = float64(newSum) / count
	stats.PrivilegedAvg = float64(privilegedSum) / count
	
	return stats
}

// generateHistoricalSparkline generates sparkline from historical samples
func generateHistoricalSparkline(samples []ActivitySample, metric string, width int) string {
	if len(samples) == 0 {
		return strings.Repeat("░", width)
	}
	
	// Extract values for the specified metric
	values := make([]int, len(samples))
	maxValue := 0
	
	for i, sample := range samples {
		var value int
		switch metric {
		case "sessions":
			value = sample.SessionsCount
		case "beacons":
			value = sample.BeaconsCount
		case "new":
			value = sample.NewCount
		case "privileged":
			value = sample.PrivilegedCount
		}
		values[i] = value
		if value > maxValue {
			maxValue = value
		}
	}
	
	if maxValue == 0 {
		return strings.Repeat("░", width)
	}
	
	// Map samples to sparkline width (interpolation if needed)
	var sparkline strings.Builder
	
	if len(samples) <= width {
		// Fewer samples than width - pad with empty space on left
		padding := width - len(samples)
		sparkline.WriteString(strings.Repeat("░", padding))
		
		// Render each sample as a character
		for _, value := range values {
			sparkline.WriteString(heightToChar(value, maxValue))
		}
	} else {
		// More samples than width - downsample
		samplesPerChar := float64(len(samples)) / float64(width)
		
		for i := 0; i < width; i++ {
			startIdx := int(float64(i) * samplesPerChar)
			endIdx := int(float64(i+1) * samplesPerChar)
			if endIdx > len(values) {
				endIdx = len(values)
			}
			
			// Average values in this bucket
			sum := 0
			count := 0
			for j := startIdx; j < endIdx; j++ {
				sum += values[j]
				count++
			}
			avg := 0
			if count > 0 {
				avg = sum / count
			}
			
			sparkline.WriteString(heightToChar(avg, maxValue))
		}
	}
	
	return sparkline.String()
}

// heightToChar converts a value to a block character based on height
func heightToChar(value, maxValue int) string {
	if maxValue == 0 {
		return "░"
	}
	
	percentage := float64(value) / float64(maxValue)
	
	// Use block characters for 8 levels
	if percentage == 0 {
		return "░"
	} else if percentage < 0.125 {
		return "▁"
	} else if percentage < 0.25 {
		return "▂"
	} else if percentage < 0.375 {
		return "▃"
	} else if percentage < 0.5 {
		return "▄"
	} else if percentage < 0.625 {
		return "▅"
	} else if percentage < 0.75 {
		return "▆"
	} else if percentage < 0.875 {
		return "▇"
	} else {
		return "█"
	}
}

// generateTimeAxis generates time labels aligned with sparkline
func generateTimeAxis(samples []ActivitySample, width int, startTime time.Time) string {
	if len(samples) == 0 {
		return strings.Repeat(" ", width)
	}
	
	// Generate hour markers for 12-hour window
	var axis strings.Builder
	
	// Calculate which hour labels to show
	now := time.Now()
	sessionDuration := now.Sub(startTime)
	
	if sessionDuration < 12*time.Hour {
		// Partial session - show from start hour to now
		hoursElapsed := int(sessionDuration.Hours()) + 1
		if hoursElapsed > 12 {
			hoursElapsed = 12
		}
		
		charsPerHour := width / hoursElapsed
		if charsPerHour < 1 {
			charsPerHour = 1
		}
		
		for i := 0; i < hoursElapsed; i++ {
			hourTime := startTime.Add(time.Duration(i) * time.Hour)
			label := hourTime.Format("15:04")
			
			if i == hoursElapsed-1 {
				// Last label - add "Now"
				axis.WriteString(label)
				remaining := width - axis.Len()
				if remaining > 4 {
					axis.WriteString(strings.Repeat(" ", remaining-3))
					axis.WriteString("Now")
				}
			} else {
				axis.WriteString(label)
				// Pad to next hour marker
				padding := charsPerHour - len(label)
				if padding > 0 {
					axis.WriteString(strings.Repeat(" ", padding))
				}
			}
		}
	} else {
		// Full 12-hour window
		hoursToShow := 6 // Show every 2 hours for clarity
		charsPerLabel := width / hoursToShow
		
		for i := 0; i < hoursToShow; i++ {
			hourOffset := i * 2 // Every 2 hours
			hourTime := now.Add(-time.Duration(12-hourOffset) * time.Hour)
			label := hourTime.Format("15:04")
			
			if i == hoursToShow-1 {
				axis.WriteString(strings.Repeat(" ", width-axis.Len()-3))
				axis.WriteString("Now")
			} else {
				axis.WriteString(label)
				padding := charsPerLabel - len(label)
				if padding > 0 && i < hoursToShow-1 {
					axis.WriteString(strings.Repeat(" ", padding))
				}
			}
		}
	}
	
	// Ensure exact width
	result := axis.String()
	if len(result) < width {
		result += strings.Repeat(" ", width-len(result))
	} else if len(result) > width {
		result = result[:width]
	}
	
	return result
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// generateSparkline generates a simple bar graph
func generateSparkline(value, maxValue, width int) string {
	if maxValue == 0 {
		return strings.Repeat("░", width)
	}
	
	filled := int(float64(value) / float64(maxValue) * float64(width))
	if filled > width {
		filled = width
	}
	
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// updateViewportContent updates the viewport with the current agent list
func (m *model) updateViewportContent() {
	// Check if we're in dashboard view
	if m.view.Type == ViewTypeDashboard {
		m.viewport.SetContent(m.renderDashboard())
		return
	}
	
	// Render agents to string
	agentLines := m.renderAgents()
	
	// Logo definition
	logo := []string{
		"    🔥🔥     ",
		"  ▄▄▄▄▄▄▄   ",
		"  █ C2  █   ",
		"  █▓▓▓▓▓█   ",
		"  ▀▀▀▀▀▀▀   ",
	}
	
	logoStyle := lipgloss.NewStyle().Foreground(m.theme.LogoColor).Bold(true)
	
	var contentLines []string
	
	if m.view.Type == ViewTypeTree {
		// Tree view: Logo on left, agents on right with connectors from logo area
		logoStart := 0 // Start logo from the top
		if len(agentLines) > len(logo) {
			// If there are more agent lines than logo lines, center the logo
			logoStart = (len(agentLines) - len(logo)) / 2
		}
		
		// Build lines with logo on left and agents on right
		maxLines := len(agentLines)
		if len(logo) > maxLines {
			maxLines = len(logo)
		}
		
		for i := 0; i < maxLines; i++ {
			var logoLine string
			if i >= logoStart && i < logoStart+len(logo) {
				logoLine = logoStyle.Render(logo[i-logoStart])
			} else {
				logoLine = strings.Repeat(" ", 12)
			}
			
			var agentLine string
			if i < len(agentLines) {
				agentLine = agentLines[i]
			}
			
			contentLines = append(contentLines, " "+logoLine+"  "+agentLine)
		}
	} else {
		// Box view: Logo at top with vertical line starting from bottom
		connectorColor := m.theme.TacticalBorder
		
		// Render logo with padding
		for _, logoLine := range logo {
			contentLines = append(contentLines, "      "+logoStyle.Render(logoLine))
		}
		
		// Start vertical line from bottom of logo
		vlinePrefix := "            " // 12 spaces to position line under logo
		contentLines = append(contentLines, vlinePrefix+lipgloss.NewStyle().Foreground(connectorColor).Render("│"))
		contentLines = append(contentLines, vlinePrefix+lipgloss.NewStyle().Foreground(connectorColor).Render("│"))
		
		// Add boxes with arrows pointing from the vertical line
		// Need to add prefix to boxes to align with the vertical line
		for _, agentLine := range agentLines {
			contentLines = append(contentLines, vlinePrefix+agentLine)
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
	
	// Render current agent (returns 3 lines now)
	indent := strings.Repeat("  ", depth)
	agentLines := m.renderAgentLine(agent)
	
	if depth > 0 {
		// Add tree connector for child agents with better styling
		connector := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Render("  ╰─")
		
		// First line gets the connector
		agentLines[0] = indent + connector + agentLines[0]
		// Second and third lines get matching indentation
		agentLines[1] = indent + "    " + agentLines[1]
		agentLines[2] = indent + "    " + agentLines[2]
	}
	
	// Add all lines
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
	
	// Status icon (only for live agents)
	var statusIcon string
	var statusColor lipgloss.Color
	
	if agent.IsDead {
		// For dead agents, use beacon icon but with dead color
		statusIcon = "◇"
		statusColor = m.theme.DeadColor
	} else if agent.IsSession {
		statusIcon = "◆"
		statusColor = m.theme.SessionColor
	} else {
		statusIcon = "◇"  
		statusColor = m.theme.BeaconColor
	}

	// OS icon - differentiate session vs beacon for Windows
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
	var connectorColor lipgloss.Color = m.theme.TacticalBorder
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

	// NEW badge
	newBadge := ""
	if agent.IsNew && !agent.IsDead {
		newBadge = " " + lipgloss.NewStyle().
			Foreground(m.theme.NewBadgeColor).
			Bold(true).
			Render("✨ NEW!")
	}

	// Privilege badge
	privBadge := ""
	if agent.IsPrivileged && !agent.IsDead {
		privBadge = " 💎"
	}

	// Dead badge (shown after hostname)
	deadBadge := ""
	if agent.IsDead {
		deadBadge = " 💀"
	}

	// Type label
	typeLabel := "beacon"
	if agent.IsSession {
		typeLabel = "session"
	} else if agent.IsDead {
		typeLabel = "dead"
	}

	// Build first line - connector from left with protocol box integrated
	connectorStyle := lipgloss.NewStyle().Foreground(connectorColor)
	
	// Protocol box with background color
	protocolBoxStyle := lipgloss.NewStyle().
		Foreground(protocolColor).
		Background(m.theme.ProtocolBg).
		Bold(true).
		Padding(0, 1)
	
	protocolBox := protocolBoxStyle.Render(strings.ToUpper(agent.Transport))
	
	line1 := fmt.Sprintf("%s%s%s▶ %s %s  %s%s%s",
		connectorStyle.Render("╰────────"),
		protocolBox,
		connectorStyle.Render("────────"),
		lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
		osIcon,
		lipgloss.NewStyle().Foreground(usernameColor).Bold(true).Render(fmt.Sprintf("%s@%s", agent.Username, agent.Hostname)),
		deadBadge,
		privBadge,
	)

	// Calculate indent for ID/IP lines - should align where the hostname starts
	// Protocol box [ MTLS ] is ~8 chars, connector ────────── is 10, arrow ▶ is 1, 
	// status icon ◆ is 1, space is 1, OS icon 🖥️ is 1, two spaces is 2
	// Total before hostname starts: 8 + 10 + 1 + 1 + 1 + 1 + 2 = 24 chars (approximately)
	// But we need to account for visual width of emojis which may render wider
	idIpIndent := 28
	
	// Build second line - ID with connector (aligned where hostname starts)
	line2 := fmt.Sprintf("%s└─ ID: %s (%s)%s",
		strings.Repeat(" ", idIpIndent),
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(statusColor).Render(typeLabel),
		newBadge,
	)
	
	// Build third line - IP with connector (aligned where hostname starts)
	line3 := fmt.Sprintf("%s└─ IP: %s",
		strings.Repeat(" ", idIpIndent),
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.RemoteAddress),
	)

	lines = append(lines, line1)
	lines = append(lines, line2)
	lines = append(lines, line3)

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

type activitySampleMsg struct{}

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

// sampleActivityCmd waits for the sample interval then triggers a sample
func sampleActivityCmd() tea.Msg {
	time.Sleep(10 * time.Minute) // Sample every 10 minutes
	return activitySampleMsg{}
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
		agents:          []Agent{},
		spinner:         s,
		loading:         true,
		termWidth:       180, // Default fallback width
		termHeight:      40,  // Default fallback height
		themeIndex:      0,   // Start with default theme
		theme:           defaultTheme,
		viewIndex:       0,   // Start with default view
		view:            defaultView,
		activityTracker: NewActivityTracker(), // Initialize activity tracker
	}

	// Create and run program with alt screen
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
