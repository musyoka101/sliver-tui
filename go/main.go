package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/musyoka101/sliver-graphs/internal/alerts"
	"github.com/musyoka101/sliver-graphs/internal/client"
	"github.com/musyoka101/sliver-graphs/internal/config"
	"github.com/musyoka101/sliver-graphs/internal/models"
	"github.com/musyoka101/sliver-graphs/internal/tracking"
	"github.com/musyoka101/sliver-graphs/internal/tree"
)

// Type aliases for tracking package types
type ActivitySample = tracking.ActivitySample
type ActivityTracker = tracking.ActivityTracker

// NewActivityTracker is provided by tracking package
var NewActivityTracker = tracking.NewActivityTracker

// sampleCurrentActivity samples the current agent state and adds to tracker
func (m *model) sampleCurrentActivity() {
	if m.activityTracker == nil {
		return
	}
	
	// Use the tracking package's SampleCurrentActivity method
	m.activityTracker.SampleCurrentActivity(m.agents, m.stats)
}

// updateSubnetOrder builds ordered list of subnets for numbered shortcuts
func (m *model) updateSubnetOrder() {
	// Clear existing order
	m.subnetOrder = []string{}
	
	// Build subnet map from active agents
	subnetMap := make(map[string]bool)
	for _, agent := range m.agents {
		if agent.IsDead {
			continue
		}
		
		// Extract IP from RemoteAddress (format: "ip:port")
		ip := agent.RemoteAddress
		if idx := strings.Index(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		
		// Extract subnet (x.x.x.0/24)
		octets := strings.Split(ip, ".")
		if len(octets) >= 3 {
			subnet := fmt.Sprintf("%s.%s.%s.0/24", octets[0], octets[1], octets[2])
			subnetMap[subnet] = true
		}
	}
	
	// Convert map to ordered slice
	for subnet := range subnetMap {
		m.subnetOrder = append(m.subnetOrder, subnet)
	}
	
	// Sort alphabetically for consistent ordering
	sort.Strings(m.subnetOrder)
}

// Agent is an alias to models.Agent
type Agent = models.Agent

// Stats is an alias to models.Stats
type Stats = models.Stats

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
	theme           config.Theme // Current theme
	viewIndex       int  // Current view index
	view            config.View // Current view
	activityTracker *ActivityTracker // Activity tracking over time
	expandedSubnets map[string]bool  // Track which subnets are expanded
	subnetOrder     []string         // Track subnet display order for numbered shortcuts
	numberBuffer    string           // Buffer for multi-digit subnet number input
	alertManager    *alerts.AlertManager // Alert/notification system
	previousAgents  map[string]Agent // Track previous agent state for change detection
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
		sampleActivityCmd, // Start activity sampling timer
		pulseTimerCmd,     // Start pulse animation timer for alerts
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
			m.view = config.GetView(m.viewIndex)
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// config.Theme switching
		case "t":
			m.themeIndex = (m.themeIndex + 1) % config.GetThemeCount()
			m.theme = config.GetTheme(m.themeIndex)
			// Update viewport content with new theme
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// config.View switching
		case "v":
			m.viewIndex = (m.viewIndex + 1) % config.GetViewCount()
			m.view = config.GetView(m.viewIndex)
			// Update viewport content with new view
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// Expand/collapse subnets in network topology (dashboard view only)
		case "e":
			if m.viewIndex == 2 { // Dashboard view
				// Toggle all subnets
				allExpanded := true
				for subnet := range m.expandedSubnets {
					if !m.expandedSubnets[subnet] {
						allExpanded = false
						break
					}
				}
				
				// If all expanded, collapse all; otherwise expand all
				for subnet := range m.expandedSubnets {
					m.expandedSubnets[subnet] = !allExpanded
				}
				
				// For new subnets, expand them
				if !allExpanded {
					// Get all current subnets
					for _, agent := range m.agents {
						if agent.IsDead {
							continue
						}
						ip := agent.RemoteAddress
						if idx := strings.Index(ip, ":"); idx != -1 {
							ip = ip[:idx]
						}
						octets := strings.Split(ip, ".")
						if len(octets) >= 3 {
							subnet := fmt.Sprintf("%s.%s.%s.0/24", octets[0], octets[1], octets[2])
							m.expandedSubnets[subnet] = true
						}
					}
				}
				
				// Update viewport
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		// Multi-digit subnet number input (accumulate digits in buffer)
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.viewIndex == 2 { // Dashboard view only
				// Append digit to buffer
				m.numberBuffer += msg.String()
				// Update viewport to show the buffer indicator
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		// Enter key - activate subnet selection from buffer
		case "enter":
			if m.viewIndex == 2 && len(m.numberBuffer) > 0 {
				// Convert buffer to integer
				subnetNum := 0
				for _, ch := range m.numberBuffer {
					subnetNum = subnetNum*10 + int(ch-'0')
				}
				subnetNum-- // Convert 1-based to 0-based index
				
				// Toggle subnet if valid
				if subnetNum >= 0 && subnetNum < len(m.subnetOrder) {
					subnet := m.subnetOrder[subnetNum]
					m.expandedSubnets[subnet] = !m.expandedSubnets[subnet]
				}
				
				// Clear buffer
				m.numberBuffer = ""
				
				// Update viewport
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		// Escape key - clear number buffer
		case "esc":
			m.numberBuffer = ""
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

	case tea.MouseMsg:
		// Handle mouse clicks for interactive elements
		if msg.Type == tea.MouseLeft {
			// Check if in dashboard view
			if m.viewIndex == 2 {
				// TODO: Implement click detection for subnet expansion
				// For now, we'll add a simple toggle mechanism with 'e' key
			}
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
		// Detect changes and generate alerts
		m.detectAgentChanges(msg.agents)
		
		m.agents = msg.agents
		m.stats = msg.stats
		m.loading = false
		m.lastUpdate = time.Now()
		m.err = nil
		
		// Sample activity immediately when agents are fetched
		m.sampleCurrentActivity()
		
		// Update subnet order for numbered shortcuts
		m.updateSubnetOrder()
		
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

	case pulseTimerMsg:
		// Update pulse animation state for critical alerts
		if m.alertManager != nil {
			m.alertManager.UpdatePulse()
			// Update viewport to show new pulse state
			if m.ready {
				m.updateViewportContent()
			}
		}
		// Schedule next pulse update
		cmds = append(cmds, pulseTimerCmd)

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

// detectAgentChanges compares current agents with previous state and generates alerts
func (m *model) detectAgentChanges(newAgents []Agent) {
	if m.alertManager == nil {
		return
	}

	// Create map of new agents for quick lookup
	newAgentMap := make(map[string]Agent)
	for _, agent := range newAgents {
		newAgentMap[agent.ID] = agent
	}

	// Detect new agents (connected)
	for _, agent := range newAgentMap {
		if _, exists := m.previousAgents[agent.ID]; !exists {
			// New agent connected
			alertType := alerts.AlertSuccess
			
			// Determine appropriate alert based on agent type and privilege
			if agent.IsSession {
				// Session-specific alerts
				if agent.IsPrivileged {
					m.alertManager.AddAlert(alertType, alerts.CategoryPrivilegedSessionOpened, "Privileged session opened", agent.Hostname)
				} else {
					m.alertManager.AddAlert(alertType, alerts.CategorySessionOpened, "Session opened", agent.Hostname)
				}
			} else {
				// Beacon-specific alerts
				if agent.IsPrivileged {
					m.alertManager.AddAlert(alertType, alerts.CategoryAgentConnected, "Privileged beacon connected", agent.Hostname)
				} else {
					m.alertManager.AddAlert(alertType, alerts.CategoryAgentConnected, "Beacon connected", agent.Hostname)
				}
			}
		}
	}

	// Detect lost agents (disconnected)
	for id, oldAgent := range m.previousAgents {
		if _, exists := newAgentMap[id]; !exists {
			// Agent disappeared
			m.alertManager.AddAlert(alerts.AlertCritical, alerts.CategoryAgentDisconnected, "Agent lost", oldAgent.Hostname)
		}
	}

	// Detect beacon late/missed check-ins
	for _, agent := range newAgentMap {
		if !agent.IsSession { // Only check beacons
			// Check if beacon is late (this logic should be in your Agent struct or tracking)
			if agent.IsDead {
				m.alertManager.AddAlert(alerts.AlertWarning, alerts.CategoryBeaconMissed, "Beacon missed check-in", agent.Hostname)
			}
		}
	}

	// Detect session events and privilege changes
	for id, newAgent := range newAgentMap {
		if oldAgent, exists := m.previousAgents[id]; exists {
			// Check if privilege escalated (wasn't privileged before, is now)
			if newAgent.IsPrivileged && !oldAgent.IsPrivileged {
				m.alertManager.AddAlert(alerts.AlertSuccess, alerts.CategoryPrivilegedAccess, "Privilege escalated", newAgent.Hostname)
			}
			
			// Check if session state changed (beacon converted to session)
			if newAgent.IsSession && !oldAgent.IsSession {
				m.alertManager.AddAlert(alerts.AlertInfo, alerts.CategorySessionOpened, "Beacon upgraded to session", newAgent.Hostname)
			} else if !newAgent.IsSession && oldAgent.IsSession {
				m.alertManager.AddAlert(alerts.AlertInfo, alerts.CategorySessionClosed, "Session closed", newAgent.Hostname)
			}
		}
	}

	// Detect beacon task changes (queued/completed)
	for id, newAgent := range newAgentMap {
		if !newAgent.IsSession { // Only check beacons
			if oldAgent, exists := m.previousAgents[id]; exists {
				// Detect new tasks queued
				if newAgent.TasksCount > oldAgent.TasksCount {
					pendingTasks := newAgent.TasksCount - newAgent.TasksCompleted
					oldPendingTasks := oldAgent.TasksCount - oldAgent.TasksCompleted
					details := fmt.Sprintf("(%d→%d pending)", oldPendingTasks, pendingTasks)
					m.alertManager.AddAlertWithDetails(alerts.AlertInfo, alerts.CategoryBeaconTaskQueued, 
						"Task queued", newAgent.Hostname, details)
				}
				
				// Detect tasks completed
				if newAgent.TasksCompleted > oldAgent.TasksCompleted {
					completedCount := newAgent.TasksCompleted
					totalCount := newAgent.TasksCount
					details := fmt.Sprintf("(%d/%d done)", completedCount, totalCount)
					m.alertManager.AddAlertWithDetails(alerts.AlertSuccess, alerts.CategoryBeaconTaskComplete, 
						"Task completed", newAgent.Hostname, details)
				}
			}
		}
	}

	// Update previous agents map
	m.previousAgents = newAgentMap
}

// renderAlertPanel renders the military-style alert/notification panel
func (m model) renderAlertPanel() string {
	if m.alertManager == nil {
		return ""
	}

	activeAlerts := m.alertManager.GetAlerts()
	if len(activeAlerts) == 0 {
		return "" // No alerts to show
	}

	// Get pulse state for critical alerts animation
	pulseState := m.alertManager.GetPulseState()
	hasCritical := m.alertManager.HasCriticalAlerts()

	// Border color based on alert severity and pulse state (uses theme colors)
	borderColor := m.theme.TacticalBorder // Default uses tactical border color
	if hasCritical {
		// Pulse animation for critical alerts (uses theme's DeadColor for red/critical)
		switch pulseState {
		case 0:
			borderColor = m.theme.DeadColor // Bright
		case 1:
			borderColor = m.theme.TacticalMuted // Medium
		case 2:
			borderColor = m.theme.SeparatorColor // Dark
		}
	}

	// Panel styling with military aesthetic - slightly wider than tactical panel
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderColor).
		Background(m.theme.TacticalPanelBg). // Use theme background
		Padding(0, 1).
		Width(45) // Wider for better alert display

	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder). // Use theme color
		Bold(true)

	// Status indicator
	statusIndicator := "◉ ACTIVE"
	statusColor := m.theme.SessionColor // Green from theme
	if hasCritical {
		statusIndicator = "◉ ALERT"
		statusColor = m.theme.DeadColor // Red from theme
	}

	// Compact header for 35-char width panel
	headerLine := fmt.Sprintf("⚠ ALERTS %s",
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusIndicator))

	var lines []string
	lines = append(lines, titleStyle.Render(headerLine))

	// Render each alert
	for _, alert := range activeAlerts {
		// Color based on alert type (uses theme colors)
		var textColor lipgloss.Color
		switch alert.Type {
		case alerts.AlertCritical:
			textColor = m.theme.DeadColor // Red/critical
		case alerts.AlertWarning:
			textColor = m.theme.BeaconColor // Orange/yellow warning
		case alerts.AlertSuccess:
			textColor = m.theme.SessionColor // Green success
		case alerts.AlertInfo:
			textColor = m.theme.TitleColor // Cyan/blue info
		case alerts.AlertNotice:
			textColor = m.theme.TacticalMuted // Gray/muted
		}

		// Format timestamp
		timeStr := alert.Timestamp.Format("15:04")

		// Build alert line with military styling
		icon := alert.GetIcon()
		label := alert.GetLabel()
		agentName := alert.AgentName
		if agentName == "" {
			agentName = alert.Message
		}
		
		// Truncate agent name if too long (max 15 chars for single line)
		if len(agentName) > 15 {
			agentName = agentName[:15]
		}

		// Single line format: ║█║ 19:45 PRIV ESCALATED m3dc
		// Or with details: ║█║ 19:45 TASK QUEUED m3dc (3→4)
		alertLine := fmt.Sprintf("%s %s %s %s",
			lipgloss.NewStyle().Foreground(textColor).Bold(true).Render(icon),
			lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(timeStr),
			lipgloss.NewStyle().Foreground(textColor).Bold(true).Render(label),
			lipgloss.NewStyle().Foreground(m.theme.TacticalValue).Render(agentName))
		
		// Add details if present (e.g., task counts)
		if alert.Details != "" {
			alertLine += lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(" " + alert.Details)
		}

		lines = append(lines, alertLine)
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
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
	
	lostCount := tracking.GetLostAgentsCount()
	
	if lostCount > 0 {
		lostLine := fmt.Sprintf("  ⚠️  Recently Lost: %d (displayed for %d min)",
			lostCount, int(tracking.GetLostAgentTimeout().Minutes()))
		footerLines = append(footerLines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff9900")).
			Italic(true).
			Padding(0, 1).
			Render(lostLine))
		footerLines = append(footerLines, "") // Add empty line for spacing
	}
	
	// Show number buffer indicator if user is typing a subnet number
	if len(m.numberBuffer) > 0 {
		bufferStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f1fa8c")). // Yellow
			Bold(true).
			Padding(0, 1)
		bufferText := fmt.Sprintf("> Subnet #%s_ (press Enter to toggle, Esc to cancel)", m.numberBuffer)
		footerLines = append(footerLines, bufferStyle.Render(bufferText))
		footerLines = append(footerLines, "") // Add empty line for spacing
	}
	
	helpStyle := lipgloss.NewStyle().Foreground(m.theme.HelpColor).Padding(0, 1)
	helpText := "Press 'r' to refresh  │  't' to change theme  │  'v' to change view  │  'd' for dashboard  │  Type subnet # + Enter to expand  │  'e' expand all  │  '↑↓' or 'j/k' to scroll  │  'q' to quit"
	footerLines = append(footerLines, helpStyle.Render(helpText))
	footerLines = append(footerLines, "")
	
	// Combine header + content + footer for left side
	leftContent := strings.Join(append(append(headerLines, contentLines...), footerLines...), "\n")
	
	// Add tactical panel as overlay on the right side
	tacticalPanel := m.renderTacticalPanel()
	
	if len(m.agents) > 0 && tacticalPanel != "" && m.termWidth > 100 {
		// Tactical Panel: Width(35) + Padding(1,2) + Border = 37 chars total
		tacticalPanelWidth := 37
		
		// Tactical panel at right edge
		tacticalPanelX := m.termWidth - tacticalPanelWidth
		if tacticalPanelX < 100 {
			tacticalPanelX = 100
		}
		
		// Split content into lines
		leftLines := strings.Split(leftContent, "\n")
		tacticalPanelLines := strings.Split(tacticalPanel, "\n")
		
		// Calculate how many header lines we have
		headerLineCount := len(headerLines)
		
		// Ensure we have enough lines for the panel
		totalLines := len(leftLines)
		if headerLineCount + len(tacticalPanelLines) > totalLines {
			totalLines = headerLineCount + len(tacticalPanelLines)
		}
		
		// Build output by overlaying tactical panel on right side
		var result []string
		for i := 0; i < totalLines; i++ {
			var line string
			
			// Add left content line (if exists)
			if i < len(leftLines) {
				line = leftLines[i]
			}
			
			// Calculate visual width
			currentWidth := lipgloss.Width(line)
			
			if i < headerLineCount {
				// Header line - don't truncate
			} else {
				// Content/footer line - pad or truncate to fit
				if currentWidth < tacticalPanelX {
					line += strings.Repeat(" ", tacticalPanelX-currentWidth)
				} else if currentWidth > tacticalPanelX {
					line = lipgloss.NewStyle().MaxWidth(tacticalPanelX).Render(line)
					currentWidth = lipgloss.Width(line)
					if currentWidth < tacticalPanelX {
						line += strings.Repeat(" ", tacticalPanelX-currentWidth)
					}
				}
				
				// Overlay tactical panel
				tacticalPanelLineIndex := i - headerLineCount
				if tacticalPanelLineIndex >= 0 && tacticalPanelLineIndex < len(tacticalPanelLines) {
					line += tacticalPanelLines[tacticalPanelLineIndex]
				}
			}
			
			result = append(result, line)
		}
		
		leftContent = strings.Join(result, "\n")
	}
	
	// Now add alert panel overlay in the footer area (bottom right)
	// Position slightly left of tactical panel to accommodate wider width
	alertPanel := m.renderAlertPanel()
	if alertPanel != "" {
		// Alert panel is wider (49 chars), position it about 10 chars left of tactical panel
		tacticalPanelWidth := 37
		alertPanelOffset := 10 // Offset from tactical panel position
		alertPanelX := m.termWidth - tacticalPanelWidth - alertPanelOffset
		if alertPanelX < 100 {
			alertPanelX = 100
		}
		
		leftLines := strings.Split(leftContent, "\n")
		alertPanelLines := strings.Split(alertPanel, "\n")
		
		// Position alert panel at the bottom (last N lines)
		// Start from bottom and work up
		totalLines := len(leftLines)
		alertStartLine := totalLines - len(alertPanelLines) - 2 // 2 lines from bottom for padding
		if alertStartLine < 0 {
			alertStartLine = 0
		}
		
		var result []string
		for i := 0; i < totalLines; i++ {
			line := ""
			if i < len(leftLines) {
				line = leftLines[i]
			}
			
			// Check if this line should have alert panel overlay
			alertLineIndex := i - alertStartLine
			if alertLineIndex >= 0 && alertLineIndex < len(alertPanelLines) {
				currentWidth := lipgloss.Width(line)
				
				// Pad to alert panel position
				if currentWidth < alertPanelX {
					line += strings.Repeat(" ", alertPanelX-currentWidth)
				} else if currentWidth > alertPanelX {
					line = lipgloss.NewStyle().MaxWidth(alertPanelX).Render(line)
					currentWidth = lipgloss.Width(line)
					if currentWidth < alertPanelX {
						line += strings.Repeat(" ", alertPanelX-currentWidth)
					}
				}
				
				// Add alert panel line
				line += alertPanelLines[alertLineIndex]
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
	
	// Create 2x3 grid layout for panels
	// Top row: C2 Infrastructure | OS & Privilege Matrix | Network Topology
	// Bottom row: Security Status | Activity Metrics
	
	c2Panel := m.renderC2InfrastructurePanel()
	archPanel := m.renderArchitecturePanel()
	networkPanel := m.renderNetworkTopologyPanel()
	securityPanel := m.renderSecurityStatusPanel()
	sparklinePanel := m.renderSparklinePanel()
	
	// Use lipgloss JoinHorizontal to place panels side by side
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, c2Panel, "  ", archPanel, "  ", networkPanel)
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
		Width(38).
		Height(18)
	
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

// renderArchitecturePanel shows OS/architecture distribution with privilege breakdown
func (m model) renderArchitecturePanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(38).
		Height(18)
	
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
	
	// Privilege color styles
	privStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50fa7b")) // Green for privileged
	
	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f1fa8c")) // Yellow for user-level
	
	var lines []string
	lines = append(lines, titleStyle.Render("💻 OS & PRIVILEGE MATRIX"))
	lines = append(lines, "")
	
	// Group by OS + Architecture
	type OSArch struct {
		OS   string
		Arch string
	}
	
	osArchData := make(map[OSArch]struct {
		privileged int
		user       int
	})
	
	totalAgents := len(m.agents)
	
	for _, agent := range m.agents {
		if agent.IsDead {
			continue // Skip dead agents
		}
		
		os := agent.OS
		if os == "" {
			os = "unknown"
		}
		arch := agent.Arch
		if arch == "" {
			arch = "unknown"
		}
		
		key := OSArch{OS: os, Arch: arch}
		data := osArchData[key]
		
		if agent.IsPrivileged {
			data.privileged++
		} else {
			data.user++
		}
		osArchData[key] = data
	}
	
	if totalAgents == 0 {
		lines = append(lines, mutedStyle.Render("No agents available"))
	} else {
		// Display each OS/Arch combination
		for osArch, data := range osArchData {
			
			// OS icon
			icon := "🖥️"
			osName := osArch.OS
			if strings.Contains(strings.ToLower(osName), "linux") {
				icon = "🐧"
			} else if strings.Contains(strings.ToLower(osName), "darwin") || strings.Contains(strings.ToLower(osName), "mac") {
				icon = "🍎"
			}
			
			// Show OS + Arch
			lines = append(lines, fmt.Sprintf("%s %s (%s)",
				icon,
				labelStyle.Render(osName),
				mutedStyle.Render(osArch.Arch)))
			
			// Show privilege breakdown with mini bars
			if data.privileged > 0 {
				privBar := strings.Repeat("█", min(data.privileged*2, 8))
				lines = append(lines, fmt.Sprintf("   %s %s",
					privStyle.Render(fmt.Sprintf("💎 %-8s", privBar)),
					valueStyle.Render(fmt.Sprintf("%d priv", data.privileged))))
			}
			
			if data.user > 0 {
				userBar := strings.Repeat("█", min(data.user*2, 8))
				lines = append(lines, fmt.Sprintf("   %s %s",
					userStyle.Render(fmt.Sprintf("👤 %-8s", userBar)),
					mutedStyle.Render(fmt.Sprintf("%d user", data.user))))
			}
			
			lines = append(lines, "")
		}
	}
	
	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderNetworkTopologyPanel shows subnet/IP-based location tracking
func (m model) renderNetworkTopologyPanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(38).
		Height(18)
	
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
	
	// Cyan bar style for subnet visualization
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")) // Dark turquoise
	
	// Clickable style
	clickableStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f1fa8c")). // Yellow
		Underline(true)
	
	var lines []string
	lines = append(lines, titleStyle.Render("🌍 NETWORK TOPOLOGY"))
	lines = append(lines, mutedStyle.Render("(type subnet # + Enter to expand)"))
	lines = append(lines, "")
	
	// Group agents by subnet (first 3 octets) with deduplication by hostname
	subnetHosts := make(map[string]map[string]Agent) // subnet -> hostname -> agent
	pivotCount := 0
	
	for _, agent := range m.agents {
		if agent.IsDead {
			continue
		}
		
		// Extract subnet from RemoteAddress (format: IP:Port)
		ip := agent.RemoteAddress
		if idx := strings.Index(ip, ":"); idx != -1 {
			ip = ip[:idx] // Remove port
		}
		
		// Get subnet (first 3 octets)
		octets := strings.Split(ip, ".")
		subnet := "unknown"
		if len(octets) >= 3 {
			subnet = fmt.Sprintf("%s.%s.%s.0/24", octets[0], octets[1], octets[2])
		}
		
		// Initialize subnet map if not exists
		if subnetHosts[subnet] == nil {
			subnetHosts[subnet] = make(map[string]Agent)
		}
		
		// Deduplicate by hostname (keep first occurrence)
		if _, exists := subnetHosts[subnet][agent.Hostname]; !exists {
			subnetHosts[subnet][agent.Hostname] = agent
		}
		
		// Count pivots (agents with parent)
		if agent.ParentID != "" {
			pivotCount++
		}
	}
	
	totalSubnets := len(subnetHosts)
	
	if totalSubnets == 0 {
		lines = append(lines, mutedStyle.Render("No network data available"))
	} else {
		// Use pre-built subnet order from model for numbered shortcuts
		// This ensures consistent numbering across renders
		
		// Show total subnets
		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Networks:"),
			valueStyle.Render(fmt.Sprintf("%d subnet(s)", totalSubnets))))
		lines = append(lines, "")
		
		// Show each subnet (limit to 5 for space, but can handle unlimited via multi-digit input)
		count := 0
		maxVisible := 5
		for subnetIdx, subnet := range m.subnetOrder {
			if count >= maxVisible {
				remaining := totalSubnets - maxVisible
				if remaining > 0 {
					lines = append(lines, mutedStyle.Render(fmt.Sprintf("... and %d more (type number to expand)", remaining)))
				}
				break
			}
			
			// Skip if this subnet doesn't have hosts (shouldn't happen, but safety check)
			hosts, exists := subnetHosts[subnet]
			if !exists {
				continue
			}
			
			// Convert map to slice for iteration
			var agents []Agent
			for _, agent := range hosts {
				agents = append(agents, agent)
			}
			
			// Count privileged agents in this subnet
			privCount := 0
			for _, agent := range agents {
				if agent.IsPrivileged {
					privCount++
				}
			}
			
			// Create mini bar for this subnet
			barLength := len(agents)
			if barLength > 10 {
				barLength = 10
			}
			bar := strings.Repeat("█", barLength)
			
			// Subnet icon
			icon := "📡"
			if strings.HasPrefix(subnet, "10.") || strings.HasPrefix(subnet, "192.168.") || strings.HasPrefix(subnet, "172.") {
				icon = "🏢" // Internal network
			}
			
			// Check if subnet is expanded
			isExpanded := m.expandedSubnets[subnet]
			expandIcon := "▶"
			if isExpanded {
				expandIcon = "▼"
			}
			
			// Show subnet header with number (clickable)
			subnetNum := subnetIdx + 1
			lines = append(lines, fmt.Sprintf("%s %s %s %s",
				clickableStyle.Render(fmt.Sprintf("[%d]", subnetNum)),
				clickableStyle.Render(expandIcon),
				icon,
				labelStyle.Render(subnet)))
			lines = append(lines, fmt.Sprintf("   %s %s",
				barStyle.Render(bar),
				valueStyle.Render(fmt.Sprintf("%d host(s)", len(agents)))))
			
			// Show individual hostnames based on expansion state
			if isExpanded {
				// Show all hosts when expanded
				for hostIdx, agent := range agents {
					// Host icon based on position
					hostIcon := "├─"
					if hostIdx == len(agents)-1 {
						hostIcon = "└─"
					}
					
					// Privilege indicator
					privIcon := "👤"
					if agent.IsPrivileged {
						privIcon = "💎"
					}
					
					// Truncate hostname if too long
					hostname := agent.Hostname
					if len(hostname) > 18 {
						hostname = hostname[:15] + "..."
					}
					
					lines = append(lines, fmt.Sprintf("      %s %s %s",
						mutedStyle.Render(hostIcon),
						privIcon,
						labelStyle.Render(hostname)))
				}
			} else {
				// Show limited hosts when collapsed (up to 3)
				hostCount := 0
				for _, agent := range agents {
					if hostCount >= 3 {
						remaining := len(agents) - 3
						if remaining > 0 {
							lines = append(lines, fmt.Sprintf("      %s",
								clickableStyle.Render(fmt.Sprintf("... click to see +%d more", remaining))))
						}
						break
					}
					
					// Host icon based on position
					hostIcon := "├─"
					if hostCount == len(agents)-1 || hostCount == 2 {
						hostIcon = "└─"
					}
					
					// Privilege indicator
					privIcon := "👤"
					if agent.IsPrivileged {
						privIcon = "💎"
					}
					
					// Truncate hostname if too long
					hostname := agent.Hostname
					if len(hostname) > 18 {
						hostname = hostname[:15] + "..."
					}
					
					lines = append(lines, fmt.Sprintf("      %s %s %s",
						mutedStyle.Render(hostIcon),
						privIcon,
						labelStyle.Render(hostname)))
					
					hostCount++
				}
			}
			
			lines = append(lines, "")
			count++
		}
		
		// Show pivot chains if any
		if pivotCount > 0 {
			lines = append(lines, fmt.Sprintf("%s %s",
				labelStyle.Render("Pivot Chains:"),
				valueStyle.Render(fmt.Sprintf("%d", pivotCount))))
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
		Width(38).
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
		Width(38).
		Height(18)
	
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
		Width(78). // Wider to span 2 columns
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
	if m.view.Type == config.ViewTypeDashboard {
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
	
	if m.view.Type == config.ViewTypeTree {
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

// renderAgentInView renders an agent based on the current view type
func (m model) renderAgentInView(agent Agent, viewType config.ViewType) []string {
	switch viewType {
	case config.ViewTypeBox:
		return m.renderAgentBox(agent)
	case config.ViewTypeTree:
		fallthrough
	default:
		return m.renderAgentLine(agent)
	}
}

// renderAgentBox renders an agent in a compact box style
func (m model) renderAgentBox(agent Agent) []string {
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

	// Privilege badge
	privBadge := ""
	if agent.IsPrivileged && !agent.IsDead {
		privBadge = " 💎"
	}

	// NEW badge
	newBadge := ""
	if agent.IsNew && !agent.IsDead {
		newBadge = " " + lipgloss.NewStyle().
			Foreground(m.theme.NewBadgeColor).
			Bold(true).
			Render("✨")
	}

	// Build box content
	// Line 1: status icon, OS icon, username@hostname, badges
	userInfo := fmt.Sprintf("%s %s  %s%s%s",
		lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
		osIcon,
		lipgloss.NewStyle().Foreground(usernameColor).Bold(true).Render(fmt.Sprintf("%s@%s", agent.Username, agent.Hostname)),
		privBadge,
		newBadge,
	)

	// Line 2: ID, IP, transport
	detailsInfo := fmt.Sprintf("%s | %s | %s",
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.RemoteAddress),
		lipgloss.NewStyle().Foreground(m.theme.TacticalValue).Render(agent.Transport),
	)

	// Combine both lines
	content := userInfo + "\n" + detailsInfo

	// Border color
	borderColor := m.theme.TacticalBorder
	if agent.IsDead {
		borderColor = m.theme.DeadColor
	}

	// Use lipgloss border style for proper continuous borders
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	// Render the box
	boxed := boxStyle.Render(content)
	
	// Split into lines for return
	lines = strings.Split(boxed, "\n")

	return lines
}

// renderAgentTreeWithView renders agent tree with view-specific formatting
func (m model) renderAgentTreeWithView(agent Agent, depth int, viewType config.ViewType) []string {
	return m.renderAgentTreeWithViewAndContext(agent, depth, viewType, false, false)
}

// renderAgentTreeWithViewAndContext renders agent tree with context about siblings
func (m model) renderAgentTreeWithViewAndContext(agent Agent, depth int, viewType config.ViewType, hasNextSibling bool, isLastChild bool) []string {
	var lines []string

	// Render current agent based on view type
	agentLines := m.renderAgentInView(agent, viewType)

	if viewType == config.ViewTypeBox {
		// Box view: use vertical connectors for parent-child relationships
		if depth > 0 {
			// Child agent - add connector from parent
			indent := strings.Repeat("   ", depth-1)
			connectorColor := m.theme.TacticalBorder
			
			// Add vertical line and L-shaped connector before the box
			lines = append(lines, indent+lipgloss.NewStyle().Foreground(connectorColor).Render("   │"))
			lines = append(lines, indent+lipgloss.NewStyle().Foreground(connectorColor).Render("   ╰──▶ ")+agentLines[0])
			
			// Indent remaining box lines
			for i := 1; i < len(agentLines); i++ {
				lines = append(lines, indent+"        "+agentLines[i])
			}
		} else {
			// Root level - boxes on the right, vertical line with arrows on left
			connectorColor := m.theme.TacticalBorder
			
			// Calculate which line is the middle of the box (for arrow placement)
			middleLine := 1 // Second line (0-indexed) for 4-line box
			
			for i, boxLine := range agentLines {
				if i == middleLine {
					// Middle line: add T-junction with arrow pointing right
					if !isLastChild {
						connector := lipgloss.NewStyle().Foreground(connectorColor).Render("├──────────────▶ ")
						lines = append(lines, connector+boxLine)
					} else {
						// Last box: use corner
						connector := lipgloss.NewStyle().Foreground(connectorColor).Render("╰──────────────▶ ")
						lines = append(lines, connector+boxLine)
					}
				} else {
					// Other lines: just vertical line or spaces
					if !isLastChild || i < middleLine {
						vline := lipgloss.NewStyle().Foreground(connectorColor).Render("│                ")
						lines = append(lines, vline+boxLine)
					} else {
						// After the corner on last box, use spaces
						lines = append(lines, "                 "+boxLine)
					}
				}
			}
		}
	} else {
		// Tree view: use existing tree connector logic
		indent := strings.Repeat("  ", depth)

		if depth > 0 {
			// Add tree connector for child agents
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
	}

	// Add spacing between agents at root level (only for Tree view)
	if depth == 0 && len(agent.Children) == 0 && viewType == config.ViewTypeTree {
		lines = append(lines, "")
	}

	// Recursively render children
	for i, child := range agent.Children {
		hasNext := i < len(agent.Children)-1
		childLines := m.renderAgentTreeWithViewAndContext(child, depth+1, viewType, hasNext, !hasNext)
		lines = append(lines, childLines...)
	}

	return lines
}

func (m model) renderAgents() []string {
	var lines []string

	// Build hierarchical tree
	tree := tree.BuildAgentTree(m.agents)

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

// Messages
type agentsMsg struct {
	agents []Agent
	stats  Stats
}

type refreshMsg struct{}

type activitySampleMsg struct{}

type pulseTimerMsg struct{}

type errMsg struct {
	err error
}

// Commands
func fetchAgentsCmd() tea.Msg {
	// Connect to Sliver and fetch real data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	agents, stats, err := client.FetchAgents(ctx)
	if err != nil {
		return errMsg{err: err}
	}

	// Track agent changes (NEW badges, lost agents)
	agents = tracking.TrackAgentChanges(agents)

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

// pulseTimerCmd triggers pulse animation updates for alert panel
func pulseTimerCmd() tea.Msg {
	time.Sleep(500 * time.Millisecond) // Pulse every 500ms
	return pulseTimerMsg{}
}

func main() {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	
	// Initialize with default theme (index 0)
	defaultTheme := config.GetTheme(0)
	s.Style = lipgloss.NewStyle().Foreground(defaultTheme.TitleColor)
	
	// Initialize with default view (index 0)
	defaultView := config.GetView(0)

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
		expandedSubnets: make(map[string]bool), // Initialize expanded subnets map
		alertManager:    alerts.NewAlertManager(5), // Max 5 visible alerts
		previousAgents:  make(map[string]Agent), // Initialize agent tracking map
	}

	// Create and run program with alt screen
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
