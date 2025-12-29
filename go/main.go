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

// extractSubnet extracts subnet from IP address (e.g., "192.168.1.100" -> "192.168.1.0/24")
func extractSubnet(remoteAddress string) string {
	// Extract IP from RemoteAddress (format: "ip:port")
	ip := remoteAddress
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	
	// Extract subnet (x.x.x.0/24)
	octets := strings.Split(ip, ".")
	if len(octets) >= 3 {
		return fmt.Sprintf("%s.%s.%s.0/24", octets[0], octets[1], octets[2])
	}
	return ""
}

// updateSubnetOrder updates the list of subnets from active agents
func (m *model) updateSubnetOrder() {
	// Build subnet map from active agents
	subnetMap := make(map[string]bool)
	for _, agent := range m.agents {
		if agent.IsDead {
			continue
		}
		
		subnet := extractSubnet(agent.RemoteAddress)
		if subnet != "" {
			subnetMap[subnet] = true
		}
	}
	
	// Convert map to ordered slice
	m.subnetOrder = make([]string, 0, len(subnetMap))
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

// SparklineCache stores pre-rendered sparklines
type SparklineCache struct {
	sessionSparkline    string
	beaconSparkline     string
	newAgentsSparkline  string
	privilegedSparkline string
	timeAxis            string
	lastSampleCount     int
	lastUpdate          time.Time
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
	theme           config.Theme // Current theme
	viewIndex       int  // Current view index
	view            config.View // Current view
	dashboardPage   int  // Current dashboard page (0-4)
	activityTracker *ActivityTracker // Activity tracking over time
	expandedSubnets map[string]bool  // Track which subnets are expanded
	subnetOrder     []string         // Track subnet display order for numbered shortcuts
	numberBuffer    string           // Buffer for multi-digit subnet number input
	alertManager    *alerts.AlertManager // Alert/notification system
	previousAgents  map[string]Agent // Track previous agent state for change detection
	animationFrame  int              // Frame counter for animations (arrows, etc.)
	dnsCache        map[string]string // Cache for DNS lookups (IP -> domain)
	domainCache     map[string]string // Cache for agent domains (sessionID -> domain)
	
	// Performance optimization: content caching
	cachedContent   string // Last rendered content
	contentDirty    bool   // Flag to force re-render
	sparklineCache  SparklineCache // Cache for sparkline rendering
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchAgentsCmd,
		sampleActivityCmd, // Start activity sampling timer
		pulseTimerCmd,     // Start pulse animation timer for alerts
		animationTickCmd,  // Start animation frame timer for flowing arrows
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
			m.contentDirty = true
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// config.Theme switching
		case "t":
			m.themeIndex = (m.themeIndex + 1) % config.GetThemeCount()
			m.theme = config.GetTheme(m.themeIndex)
			m.contentDirty = true
			// Update viewport content with new theme
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// config.View switching
		case "v":
			m.viewIndex = (m.viewIndex + 1) % config.GetViewCount()
			m.view = config.GetView(m.viewIndex)
			m.contentDirty = true
			// Update viewport content with new view
			if m.ready {
				m.updateViewportContent()
			}
			return m, nil
		
		// Dashboard page navigation (when in dashboard view)
		case "tab":
			if m.viewIndex == 2 { // Dashboard view only
				m.dashboardPage = (m.dashboardPage + 1) % 5 // 5 pages: 0-4
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "shift+tab":
			if m.viewIndex == 2 { // Dashboard view only
				m.dashboardPage = (m.dashboardPage - 1 + 5) % 5 // 5 pages: 0-4
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "f1":
			if m.viewIndex == 2 {
				m.dashboardPage = 0
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "f2":
			if m.viewIndex == 2 {
				m.dashboardPage = 1
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "f3":
			if m.viewIndex == 2 {
				m.dashboardPage = 2
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "f4":
			if m.viewIndex == 2 {
				m.dashboardPage = 3
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		case "f5":
			if m.viewIndex == 2 {
				m.dashboardPage = 4
				m.contentDirty = true
				if m.ready {
					m.updateViewportContent()
				}
			}
			return m, nil
		
		// Expand/collapse subnets in network topology (dashboard and network map views)
		case "e":
			if m.view.Type == config.ViewTypeDashboard || m.view.Type == config.ViewTypeNetworkMap {
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
						subnet := extractSubnet(agent.RemoteAddress)
						if subnet != "" {
							m.expandedSubnets[subnet] = true
						}
					}
				}
				
				// Mark content as dirty and update viewport
				m.contentDirty = true
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
		m.contentDirty = true // Mark content as needing re-render
		
		// Sample activity immediately when agents are fetched
		m.sampleCurrentActivity()
		
		// Update subnet order for numbered shortcuts
		m.updateSubnetOrder()
		
		// Update viewport content with new agent list
		if m.ready {
			m.updateViewportContent()
		}
		
		// Trigger background domain queries for all sessions (non-blocking)
		for _, agent := range msg.agents {
			if agent.IsSession && !agent.IsDead {
				// Check if we already have this domain cached
				if _, exists := m.domainCache[agent.ID]; !exists {
					// Launch background query
					cmds = append(cmds, queryDomainCmd(agent.ID))
				}
			}
		}
		
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return refreshMsg{}
		}))

	case domainQueryMsg:
		// Domain query completed in background
		if msg.domain != "" {
			// Cache the domain result
			m.domainCache[msg.sessionID] = msg.domain
			// Mark content dirty to trigger re-render with new domain info
			m.contentDirty = true
			if m.ready {
				m.updateViewportContent()
			}
		}

	case activitySampleMsg:
		// Sample activity when timer triggers
		m.sampleCurrentActivity()
		m.contentDirty = true // Mark for dashboard refresh
		// Schedule next sample
		cmds = append(cmds, sampleActivityCmd)

	case pulseTimerMsg:
		// Update pulse animation state for critical alerts
		if m.alertManager != nil {
			m.alertManager.UpdatePulse()
			m.contentDirty = true // Mark for alert panel refresh
			// Update viewport to show new pulse state
			if m.ready {
				m.updateViewportContent()
			}
		}
		// Schedule next pulse update
		cmds = append(cmds, pulseTimerCmd)

	case animationTickMsg:
		// Update animation frame for flowing arrows
		m.animationFrame++
		if m.animationFrame > 3 {
			m.animationFrame = 0
		}
		// Only mark dirty and update if we're on Network Map, Box, or Tree view
		if m.view.Type == config.ViewTypeNetworkMap || m.view.Type == config.ViewTypeBox || m.view.Type == config.ViewTypeTree {
			m.contentDirty = true
			if m.ready {
				m.updateViewportContent()
			}
		}
		// Schedule next animation tick
		cmds = append(cmds, animationTickCmd)

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
			
			// Count current active agents for context
			activeCount := len(newAgentMap)
			details := fmt.Sprintf("(%d active)", activeCount)
			
			// Determine appropriate alert based on agent type and privilege
			if agent.IsSession {
				// Session-specific alerts
				if agent.IsPrivileged {
					m.alertManager.AddAlertWithDetails(alertType, alerts.CategoryPrivilegedSessionAcquired, 
						"Privileged session connected", agent.Hostname, details)
				} else {
					m.alertManager.AddAlertWithDetails(alertType, alerts.CategorySessionAcquired, 
						"Session connected", agent.Hostname, details)
				}
			} else {
				// Beacon-specific alerts
				if agent.IsPrivileged {
					m.alertManager.AddAlertWithDetails(alertType, alerts.CategoryPrivilegedBeaconAcquired, 
						"Privileged beacon connected", agent.Hostname, details)
				} else {
					m.alertManager.AddAlertWithDetails(alertType, alerts.CategoryBeaconAcquired, 
						"Beacon connected", agent.Hostname, details)
				}
			}
		}
	}

	// Detect lost agents (disconnected)
	for id, oldAgent := range m.previousAgents {
		if _, exists := newAgentMap[id]; !exists {
			// Agent disappeared - differentiate between session and beacon
			if oldAgent.IsSession {
				m.alertManager.AddAlert(alerts.AlertCritical, alerts.CategorySessionDisconnected, "Session lost", oldAgent.Hostname)
			} else {
				m.alertManager.AddAlert(alerts.AlertCritical, alerts.CategoryBeaconDisconnected, "Beacon lost", oldAgent.Hostname)
			}
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
				agentType := "beacon"
				if newAgent.IsSession {
					agentType = "session"
				}
				details := fmt.Sprintf("(%s)", agentType)
				m.alertManager.AddAlertWithDetails(alerts.AlertSuccess, alerts.CategoryPrivilegedAccess, 
					"Privilege escalated", newAgent.Hostname, details)
			}
			
			// Check if session state changed (beacon converted to session)
			if newAgent.IsSession && !oldAgent.IsSession {
				details := "(beacon→session)"
				if newAgent.IsPrivileged {
					m.alertManager.AddAlertWithDetails(alerts.AlertInfo, alerts.CategoryPrivilegedSessionOpened, 
						"Beacon upgraded to privileged session", newAgent.Hostname, details)
				} else {
					m.alertManager.AddAlertWithDetails(alerts.AlertInfo, alerts.CategorySessionOpened, 
						"Beacon upgraded to session", newAgent.Hostname, details)
				}
			} else if !newAgent.IsSession && oldAgent.IsSession {
				details := "(session→beacon)"
				m.alertManager.AddAlertWithDetails(alerts.AlertInfo, alerts.CategorySessionClosed, 
					"Session closed", newAgent.Hostname, details)
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

	// Status indicator
	statusIndicator := "◉ ACTIVE"
	statusColor := m.theme.SessionColor // Green from theme
	if hasCritical {
		statusIndicator = "◉ ALERT"
		statusColor = m.theme.DeadColor // Red from theme
	}

	// Build title that will cross the border
	titleText := fmt.Sprintf(" ⚠ ALERTS %s ",
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusIndicator))
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalBorder).
		Bold(true)
	
	styledTitle := titleStyle.Render(titleText)
	titleWidth := lipgloss.Width(titleText) // Actual visible width
	
	// Panel width (70 chars content + 2 for padding)
	panelWidth := 70
	contentWidth := panelWidth - 2 // Account for left/right padding
	
	// Create top border with title crossing through it
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	leftBorderChars := 3  // "┌──"
	rightBorderNeeded := contentWidth - titleWidth - leftBorderChars + 2
	if rightBorderNeeded < 1 {
		rightBorderNeeded = 1
	}
	
	topBorder := borderStyle.Render("┌" + strings.Repeat("─", leftBorderChars)) +
		styledTitle +
		borderStyle.Render(strings.Repeat("─", rightBorderNeeded) + "┐")
	
	// Build alert lines
	var lines []string
	lines = append(lines, topBorder)
	
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
		
		// Truncate agent name if too long (max 25 chars with expanded panel)
		if len(agentName) > 25 {
			agentName = agentName[:25]
		}

		// Pad label to fixed width (28 chars) so hostnames align in a column
		labelWidth := 28
		paddedLabel := fmt.Sprintf("%-*s", labelWidth, label) // Left-align, pad to width
		
		alertContent := fmt.Sprintf("%s %s %s %s",
			lipgloss.NewStyle().Foreground(textColor).Bold(true).Render(icon),
			lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(timeStr),
			lipgloss.NewStyle().Foreground(textColor).Bold(true).Render(paddedLabel),
			lipgloss.NewStyle().Foreground(m.theme.TacticalValue).Render(agentName))
		
		// Add details if present (e.g., task counts)
		if alert.Details != "" {
			alertContent += lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(" " + alert.Details)
		}
		
		// Pad content to fit panel width and add side borders
		contentLen := lipgloss.Width(alertContent)
		padding := contentWidth - contentLen
		if padding < 0 {
			padding = 0
		}
		
		alertLine := borderStyle.Render("│") + " " + alertContent + strings.Repeat(" ", padding) + " " + borderStyle.Render("│")
		lines = append(lines, alertLine)
	}
	
	// Bottom border
	bottomBorder := borderStyle.Render("└" + strings.Repeat("─", contentWidth+2) + "┘")
	lines = append(lines, bottomBorder)
	
	// Add background color
	result := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Background(m.theme.TacticalPanelBg).
		Render(result)
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
		// Alert panel is 72 chars wide (70 content + 2 border), position it left of right edge
		alertPanelWidth := 72 // Actual rendered width with border
		
		// Calculate position: from right edge, move left by alert panel width + some spacing
		// This ensures the panel fits on screen and extends leftward as intended
		alertPanelX := m.termWidth - alertPanelWidth // 0 padding - flush with right edge
		if alertPanelX < 60 { // Minimum left position to avoid overlap with tree
			alertPanelX = 60
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

		// Extract domain using multiple methods (priority order)
		var domain string
		
		// Method 1 (HIGHEST PRIORITY): Use domain from background query cache
		// This is populated asynchronously by querying USERDNSDOMAIN from sessions
		if cachedDomain, exists := m.domainCache[agent.ID]; exists {
			domain = cachedDomain
		}
		
		// Method 2: Use domain field if already populated (legacy/fallback)
		if domain == "" && agent.Domain != "" {
			domain = agent.Domain
		}
		
		// Method 3: Extract from hostname if it contains FQDN (e.g., "m3sqlw.m3c.local")
		if domain == "" {
			domain = config.ExtractDomainFromHostname(agent.Hostname)
		}
		
		// Method 4: Perform reverse DNS lookup on IP address to get FQDN (with caching)
		if domain == "" && agent.RemoteAddress != "" {
			// Check cache first
			if cachedDomain, exists := m.dnsCache[agent.RemoteAddress]; exists {
				domain = cachedDomain
			} else {
				// Perform DNS lookup (this may be slow, so we cache it)
				resolvedDomain := config.ResolveDomainFromIP(agent.RemoteAddress)
				m.dnsCache[agent.RemoteAddress] = resolvedDomain // Cache result (even if empty)
				domain = resolvedDomain
			}
		}
		
		// Method 5: Fallback to NetBIOS domain from username (Windows: DOMAIN\username)
		// Only use this as last resort if all DNS methods failed
		if domain == "" && strings.Contains(agent.Username, "\\") {
			netbiosDomain := strings.Split(agent.Username, "\\")[0]
			
			// Filter out system accounts that aren't real domains
			domainUpper := strings.ToUpper(netbiosDomain)
			if domainUpper != "NT AUTHORITY" && 
			   domainUpper != "BUILTIN" && 
			   domainUpper != "WORKGROUP" &&
			   netbiosDomain != "" {
				// Use NetBIOS name as fallback (not ideal but better than nothing)
				domain = netbiosDomain
			}
		}
		
		// Normalize domain to lowercase for consistent counting
		if domain != "" {
			domain = strings.ToLower(domain)
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
		// Deduplicate: Remove NetBIOS names if FQDN exists
		// e.g., if both "m3c" and "m3c.local" exist, only show "m3c.local"
		deduplicatedDomains := make(map[string]int)
		
		for domain, count := range domains {
			// Check if this is a potential NetBIOS name (no dots)
			if !strings.Contains(domain, ".") {
				// Look for a matching FQDN that starts with this NetBIOS name
				foundFQDN := false
				for otherDomain := range domains {
					if strings.Contains(otherDomain, ".") && 
					   strings.HasPrefix(strings.ToLower(otherDomain), strings.ToLower(domain)+".") {
						// Found matching FQDN, merge counts into the FQDN
						deduplicatedDomains[otherDomain] += count
						foundFQDN = true
						break
					}
				}
				// If no FQDN found, keep the NetBIOS name
				if !foundFQDN {
					deduplicatedDomains[domain] += count
				}
			} else {
				// This is an FQDN, always keep it
				deduplicatedDomains[domain] += count
			}
		}
		
		for domain, count := range deduplicatedDomains {
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

// SubnetGroup represents a group of agents in the same subnet
type SubnetGroup struct {
	Subnet      string
	Agents      []Agent
	HasPivots   bool
	PivotParent string
}

// getAnimatedArrow returns an animated arrow character based on frame
func (m model) getAnimatedArrow() string {
	arrows := []string{"▼", "▽", "▿", "˅"}
	return arrows[m.animationFrame%len(arrows)]
}

// getAnimatedHorizontalArrow returns an animated horizontal arrow based on frame
func (m model) getAnimatedHorizontalArrow() string {
	arrows := []string{"▶", "▷", "▹", "›"}
	return arrows[m.animationFrame%len(arrows)]
}

// renderNetworkMapView renders the network topology map with subnet-based layout
func (m model) renderNetworkMapView() string {
	var content strings.Builder
	
	// Calculate left padding for centering (approximate content width ~80 chars)
	leftPadding := ""
	if m.termWidth > 90 {
		padding := (m.termWidth - 80) / 2
		if padding > 0 {
			leftPadding = strings.Repeat(" ", padding)
		}
	}
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	// Group agents by subnet
	subnetGroups := make(map[string]*SubnetGroup)
	var subnets []string
	
	// Get C2 servers
	c2Servers := make(map[string]int)
	for _, agent := range m.agents {
		if agent.IsDead {
			continue
		}
		c2URL := agent.ActiveC2
		if c2URL == "" {
			c2URL = "Unknown C2"
		}
		c2Servers[c2URL]++
	}
	
	// Group agents by subnet
	for _, agent := range m.agents {
		subnet := extractSubnet(agent.RemoteAddress)
		if subnet == "" {
			subnet = "Unknown"
		}
		
		if _, exists := subnetGroups[subnet]; !exists {
			subnetGroups[subnet] = &SubnetGroup{
				Subnet: subnet,
				Agents: []Agent{},
			}
			subnets = append(subnets, subnet)
		}
		
		subnetGroups[subnet].Agents = append(subnetGroups[subnet].Agents, agent)
		
		// Check if any agent in this subnet is pivoted
		if agent.ProxyURL != "" {
			subnetGroups[subnet].HasPivots = true
		}
	}
	
	sort.Strings(subnets)
	
	// Initialize expandedSubnets map for all subnets (default to collapsed)
	for _, subnet := range subnets {
		if _, exists := m.expandedSubnets[subnet]; !exists {
			m.expandedSubnets[subnet] = false // Default to collapsed
		}
	}
	
	// Render title
	content.WriteString(leftPadding)
	content.WriteString(headerStyle.Render("🗺️  NETWORK TOPOLOGY MAP"))
	content.WriteString("  ")
	content.WriteString(mutedStyle.Render(fmt.Sprintf("%d Subnets | %d Agents", 
		len(subnets), len(m.agents))))
	content.WriteString("\n\n")
	
	// Render C2 infrastructure box (with padding)
	c2Box := m.renderC2Box(c2Servers)
	for _, line := range strings.Split(c2Box, "\n") {
		content.WriteString(leftPadding + line + "\n")
	}
	
	// Render connection lines from C2 to subnets
	if len(subnets) > 0 {
		// Vertical line down from C2
		centerLine := "                         │"
		content.WriteString(leftPadding)
		content.WriteString(mutedStyle.Render(centerLine))
		content.WriteString("\n")
		
		// Calculate number of branches (max 3 per row)
		numBranches := len(subnets)
		if numBranches > 3 {
			numBranches = 3
		}
		
		if numBranches == 1 {
			// Single subnet - straight line down with arrow
			content.WriteString(leftPadding)
			content.WriteString(mutedStyle.Render("                         │"))
			content.WriteString("\n")
			content.WriteString(leftPadding)
			content.WriteString(mutedStyle.Render("                         " + m.getAnimatedArrow()))
			content.WriteString("\n")
		} else {
			// Multiple subnets - branch out
			// Branch line with T-junctions
			branchLine := "       ┌─────────┴─────────"
			
			for i := 1; i < numBranches-1; i++ {
				branchLine += "┬" + strings.Repeat("─", 26)
			}
			if numBranches > 1 {
				branchLine += "┬" + strings.Repeat("─", 18) + "┐"
			}
			
			content.WriteString(leftPadding)
			content.WriteString(mutedStyle.Render(branchLine))
			content.WriteString("\n")
			
			// Vertical lines down to boxes
			arrowLine := "       │                   "
			for i := 1; i < numBranches-1; i++ {
				arrowLine += "│                          "
			}
			if numBranches > 1 {
				arrowLine += "│                   │"
			}
			content.WriteString(leftPadding)
			content.WriteString(mutedStyle.Render(arrowLine))
			content.WriteString("\n")
			
			// Downward arrows
			arrow := m.getAnimatedArrow()
			arrowTips := "       " + arrow + "                   "
			for i := 1; i < numBranches-1; i++ {
				arrowTips += arrow + "                          "
			}
			if numBranches > 1 {
				arrowTips += arrow + "                   " + arrow
			}
			content.WriteString(leftPadding)
			content.WriteString(mutedStyle.Render(arrowTips))
			content.WriteString("\n")
		}
	}
	
	// Render subnet boxes in rows (max 3 per row)
	for i := 0; i < len(subnets); i += 3 {
		var boxes []string
		
		for j := 0; j < 3 && i+j < len(subnets); j++ {
			subnet := subnets[i+j]
			group := subnetGroups[subnet]
			boxes = append(boxes, m.renderSubnetBox(group))
		}
		
		// Join boxes horizontally and add left padding
		joinedBoxes := m.joinBoxesHorizontally(boxes)
		for _, line := range strings.Split(joinedBoxes, "\n") {
			content.WriteString(leftPadding + line + "\n")
		}
	}
	
	// Navigation help
	content.WriteString("\n")
	content.WriteString(leftPadding)
	content.WriteString(mutedStyle.Render("  Navigation: V - Cycle Views | E - Expand Subnets | Q - Quit"))
	
	return content.String()
}

// renderC2Box renders the C2 infrastructure box
func (m model) renderC2Box(c2Servers map[string]int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TitleColor).
		Padding(0, 1).
		Width(50).
		Align(lipgloss.Center)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true)
	
	var lines []string
	lines = append(lines, titleStyle.Render("🔥 C2 INFRASTRUCTURE"))
	
	for server, count := range c2Servers {
		lines = append(lines, fmt.Sprintf("%s (%d agents)", server, count))
	}
	
	return boxStyle.Render(strings.Join(lines, "\n"))
}

// renderSubnetBox renders a single subnet group box
func (m model) renderSubnetBox(group *SubnetGroup) string {
	// Check if this subnet is expanded
	isExpanded := m.expandedSubnets[group.Subnet]
	
	// Adjust height based on expansion state
	boxHeight := 10
	if isExpanded && len(group.Agents) > 4 {
		boxHeight = 8 + len(group.Agents) // Dynamic height for expanded view
		if boxHeight > 20 {
			boxHeight = 20 // Max height
		}
	}
	
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(24).
		Height(boxHeight)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true)
	
	mutedStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	var lines []string
	
	// Title with subnet and expand/collapse indicator
	expandIndicator := "▶" // Collapsed
	if isExpanded {
		expandIndicator = "▼" // Expanded
	}
	lines = append(lines, titleStyle.Render(expandIndicator+" "+group.Subnet))
	lines = append(lines, mutedStyle.Render(strings.Repeat("─", 20)))
	
	// Count agents by type
	sessionCount := 0
	beaconCount := 0
	privilegedCount := 0
	deadCount := 0
	
	for _, agent := range group.Agents {
		if agent.IsDead {
			deadCount++
		} else if agent.IsSession {
			sessionCount++
		} else {
			beaconCount++
		}
		if agent.IsPrivileged {
			privilegedCount++
		}
	}
	
	// Show agents based on expansion state
	maxShow := 4
	if isExpanded {
		maxShow = len(group.Agents) // Show all when expanded
		if maxShow > 15 {
			maxShow = 15 // Reasonable limit
		}
	}
	
	for i, agent := range group.Agents {
		if i >= maxShow {
			remaining := len(group.Agents) - maxShow
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("  ... +%d more", remaining)))
			break
		}
		
		icon := "◇"
		color := m.theme.BeaconColor
		if agent.IsDead {
			icon = "◇"
			color = m.theme.DeadColor
		} else if agent.IsSession {
			icon = "◆"
			color = m.theme.SessionColor
		}
		
		privilege := ""
		if agent.IsPrivileged {
			privilege = " 💎"
		}
		
		agentStyle := lipgloss.NewStyle().Foreground(color)
		
		hostname := agent.Hostname
		if len(hostname) > 12 {
			hostname = hostname[:12]
		}
		
		lines = append(lines, fmt.Sprintf("%s %s%s", 
			agentStyle.Render(icon), 
			hostname,
			privilege))
	}
	
	lines = append(lines, "")
	
	// Summary line
	summaryParts := []string{}
	if sessionCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d◆", sessionCount))
	}
	if beaconCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d◇", beaconCount))
	}
	if privilegedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d💎", privilegedCount))
	}
	if deadCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d⚠️", deadCount))
	}
	
	if len(summaryParts) > 0 {
		lines = append(lines, mutedStyle.Render(strings.Join(summaryParts, " ")))
	}
	
	// Pivot indicator
	if group.HasPivots {
		lines = append(lines, mutedStyle.Render("🔗 PROXIED"))
	}
	
	return boxStyle.Render(strings.Join(lines, "\n"))
}

// joinBoxesHorizontally joins multiple box strings side by side
func (m model) joinBoxesHorizontally(boxes []string) string {
	if len(boxes) == 0 {
		return ""
	}
	
	// Split each box into lines
	boxLines := make([][]string, len(boxes))
	maxLines := 0
	
	for i, box := range boxes {
		boxLines[i] = strings.Split(box, "\n")
		if len(boxLines[i]) > maxLines {
			maxLines = len(boxLines[i])
		}
	}
	
	// Join lines horizontally
	var result []string
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var lineParts []string
		for boxIdx := 0; boxIdx < len(boxLines); boxIdx++ {
			if lineIdx < len(boxLines[boxIdx]) {
				lineParts = append(lineParts, boxLines[boxIdx][lineIdx])
			} else {
				lineParts = append(lineParts, strings.Repeat(" ", 24)) // Pad empty box space
			}
		}
		result = append(result, "  "+strings.Join(lineParts, "  "))
	}
	
	return strings.Join(result, "\n")
}

// renderDashboard renders the dashboard view with analytics panels
func (m model) renderDashboard() string {
	var content strings.Builder
	
	// Page names
	pageNames := []string{
		"OVERVIEW",
		"NETWORK INTEL", 
		"OPERATIONS",
		"SECURITY",
		"ANALYTICS",
	}
	
	// Dashboard header with page indicator
	headerStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true).
		Underline(true)
	
	pageStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalMuted)
	
	currentPageStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true)
	
	// Build page tabs
	var pageTabs []string
	for i, name := range pageNames {
		if i == m.dashboardPage {
			pageTabs = append(pageTabs, currentPageStyle.Render(fmt.Sprintf("[F%d:%s]", i+1, name)))
		} else {
			pageTabs = append(pageTabs, pageStyle.Render(fmt.Sprintf(" F%d:%s ", i+1, name)))
		}
	}
	
	content.WriteString(headerStyle.Render("📊 DASHBOARD"))
	content.WriteString("  ")
	content.WriteString(strings.Join(pageTabs, " "))
	content.WriteString("\n")
	content.WriteString(pageStyle.Render("Navigate: Tab/Shift+Tab or F1-F5"))
	content.WriteString("\n\n")
	
	// Render different pages based on dashboardPage
	switch m.dashboardPage {
	case 0: // Overview
		content.WriteString(m.renderOverviewPage())
	case 1: // Network Intel
		content.WriteString(m.renderNetworkIntelPage())
	case 2: // Operations
		content.WriteString(m.renderOperationsPage())
	case 3: // Security
		content.WriteString(m.renderSecurityPage())
	case 4: // Analytics
		content.WriteString(m.renderAnalyticsPage())
	}
	
	return content.String()
}

// renderOverviewPage shows quick summary stats
func (m model) renderOverviewPage() string {
	// Quick stats panel + recent activity
	archPanel := m.renderArchitecturePanel()
	taskQueuePanel := m.renderTaskQueuePanel()
	sparklinePanel := m.renderSparklinePanel()
	
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, archPanel, "  ", taskQueuePanel, "  ", sparklinePanel)
	
	// Add a summary panel
	summaryPanel := m.renderQuickStatsPanel()
	
	return topRow + "\n\n" + summaryPanel
}

// renderNetworkIntelPage shows network topology and C2 infrastructure
func (m model) renderNetworkIntelPage() string {
	c2Panel := m.renderC2InfrastructurePanel()
	networkPanel := m.renderNetworkTopologyPanel()
	
	// Could add more network-related panels here
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, c2Panel, "  ", networkPanel)
	
	return topRow
}

// renderOperationsPage shows task queues and beacon activity  
func (m model) renderOperationsPage() string {
	taskQueuePanel := m.renderTaskQueuePanel()
	
	// For now, show task queue - can add more operational panels
	return taskQueuePanel
}

// renderSecurityPage shows security status and privilege tracking
func (m model) renderSecurityPage() string {
	securityPanel := m.renderSecurityStatusPanel()
	archPanel := m.renderArchitecturePanel()
	
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, securityPanel, "  ", archPanel)
	
	return topRow
}

// renderAnalyticsPage shows historical data and trends
func (m model) renderAnalyticsPage() string {
	sparklinePanel := m.renderSparklinePanel()
	
	// For now just sparklines - can add more analytics panels
	return sparklinePanel
}

// renderQuickStatsPanel shows a summary of key metrics
func (m model) renderQuickStatsPanel() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.TacticalBorder).
		Padding(1, 2).
		Width(120). // Wide panel for overview
		Height(10)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.TitleColor).
		Bold(true)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalSection)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(m.theme.TacticalValue).
		Bold(true)
	
	var lines []string
	lines = append(lines, titleStyle.Render("📈 QUICK STATS"))
	lines = append(lines, "")
	
	// Count stats
	totalAgents := 0
	sessions := 0
	beacons := 0
	privileged := 0
	dead := 0
	
	for _, agent := range m.agents {
		totalAgents++
		if agent.IsDead {
			dead++
			continue
		}
		if agent.IsSession {
			sessions++
		} else {
			beacons++
		}
		if agent.IsPrivileged {
			privileged++
		}
	}
	
	activeAgents := totalAgents - dead
	
	// Build stats line
	stats := fmt.Sprintf("%s %s  |  %s %s  |  %s %s  |  %s %s  |  %s %s",
		labelStyle.Render("Total:"),
		valueStyle.Render(fmt.Sprintf("%d", activeAgents)),
		labelStyle.Render("Sessions:"),
		valueStyle.Render(fmt.Sprintf("%d", sessions)),
		labelStyle.Render("Beacons:"),
		valueStyle.Render(fmt.Sprintf("%d", beacons)),
		labelStyle.Render("Privileged:"),
		valueStyle.Render(fmt.Sprintf("%d", privileged)),
		labelStyle.Render("Dead:"),
		lipgloss.NewStyle().Foreground(m.theme.DeadColor).Bold(true).Render(fmt.Sprintf("%d", dead)))
	
	lines = append(lines, stats)
	
	return panelStyle.Render(strings.Join(lines, "\n"))
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
	
	// Cyan bar style matching task queue
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")) // Dark turquoise
	
	// Privilege color styles
	privStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff5555")). // Red for privileged
		Bold(true)
	
	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50fa7b")) // Green for user-level
	
	var lines []string
	lines = append(lines, titleStyle.Render("💻 OS & PRIVILEGE MATRIX"))
	lines = append(lines, "")
	
	// Count OS types, architectures, and privilege levels
	type OSArchKey struct {
		OS   string
		Arch string
	}
	
	osArchData := make(map[OSArchKey]struct {
		privileged int
		user       int
		sessions   int
		beacons    int
	})
	
	totalAgents := 0
	totalPrivileged := 0
	totalSessions := 0
	
	for _, agent := range m.agents {
		if agent.IsDead {
			continue // Skip dead agents
		}
		
		totalAgents++
		if agent.IsPrivileged {
			totalPrivileged++
		}
		if agent.IsSession {
			totalSessions++
		}
		
		os := agent.OS
		if os == "" {
			os = "unknown"
		}
		// Simplify OS names
		if strings.Contains(strings.ToLower(os), "windows") {
			os = "Windows"
		} else if strings.Contains(strings.ToLower(os), "linux") {
			os = "Linux"
		} else if strings.Contains(strings.ToLower(os), "darwin") {
			os = "macOS"
		}
		
		arch := agent.Arch
		if arch == "" {
			arch = "unknown"
		}
		
		key := OSArchKey{OS: os, Arch: arch}
		data := osArchData[key]
		
		if agent.IsPrivileged {
			data.privileged++
		} else {
			data.user++
		}
		
		if agent.IsSession {
			data.sessions++
		} else {
			data.beacons++
		}
		
		osArchData[key] = data
	}
	
	if totalAgents == 0 {
		lines = append(lines, mutedStyle.Render("No agents available"))
		lines = append(lines, "")
		lines = append(lines, mutedStyle.Render("Waiting for connections..."))
	} else {
		// Summary stats at top
		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Total Agents:"),
			valueStyle.Render(fmt.Sprintf("%d", totalAgents))))
		lines = append(lines, fmt.Sprintf("%s %s | %s",
			mutedStyle.Render("├─"),
			privStyle.Render(fmt.Sprintf("💎 %d priv", totalPrivileged)),
			userStyle.Render(fmt.Sprintf("👤 %d std", totalAgents-totalPrivileged))))
		lines = append(lines, fmt.Sprintf("%s %s | %s",
			mutedStyle.Render("└─"),
			labelStyle.Render(fmt.Sprintf("◆ %d sess", totalSessions)),
			mutedStyle.Render(fmt.Sprintf("◇ %d beac", totalAgents-totalSessions))))
		lines = append(lines, "")
		
		// Display each OS/Arch combination with detailed breakdown
		for osArch, data := range osArchData {
			total := data.privileged + data.user
			
			// OS icon
			icon := "🖥️"
			if osArch.OS == "Linux" {
				icon = "🐧"
			} else if osArch.OS == "macOS" {
				icon = "🍎"
			} else if osArch.OS == "Windows" {
				icon = "🪟"
			}
			
			// Calculate percentage
			percentage := float64(total) / float64(totalAgents) * 100
			
			// Show OS + Arch with count
			lines = append(lines, fmt.Sprintf("%s %s",
				icon,
				labelStyle.Render(fmt.Sprintf("%s (%s)", osArch.OS, osArch.Arch))))
			
			// Percentage bar
			barLength := int(percentage / 10) // 10% per block
			if barLength > 10 {
				barLength = 10
			}
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)
			lines = append(lines, fmt.Sprintf("  %s %s",
				barStyle.Render(bar),
				mutedStyle.Render(fmt.Sprintf("%.0f%%", percentage))))
			
			// Privilege breakdown on same line
			lines = append(lines, fmt.Sprintf("  %s%s │ %s%s",
				privStyle.Render("💎"),
				valueStyle.Render(fmt.Sprintf("%-2d", data.privileged)),
				userStyle.Render("👤"),
				mutedStyle.Render(fmt.Sprintf("%-2d", data.user))))
			
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
		Width(38). // Consistent width with other panels for 3x2 grid
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
	
	sparklineWidth := 28 // Adjusted width for narrower panel (38 char panel)
	
	if len(samples) == 0 {
		lines = append(lines, mutedStyle.Render("Collecting data... (first sample in 10min)"))
		return panelStyle.Render(strings.Join(lines, "\n"))
	}
	
	// Calculate statistics
	stats := calculateActivityStats(samples)
	
	// Use cached sparklines if available and samples haven't changed
	var sessionsSparkline, beaconsSparkline, newSparkline, privilegedSparkline, timeAxis string
	if m.sparklineCache.lastSampleCount == len(samples) && 
	   time.Since(m.sparklineCache.lastUpdate) < 30*time.Second {
		// Use cached sparklines
		sessionsSparkline = m.sparklineCache.sessionSparkline
		beaconsSparkline = m.sparklineCache.beaconSparkline
		newSparkline = m.sparklineCache.newAgentsSparkline
		privilegedSparkline = m.sparklineCache.privilegedSparkline
		timeAxis = m.sparklineCache.timeAxis
	} else {
		// Generate new sparklines and cache them
		sessionsSparkline = generateHistoricalSparkline(samples, "sessions", sparklineWidth)
		beaconsSparkline = generateHistoricalSparkline(samples, "beacons", sparklineWidth)
		newSparkline = generateHistoricalSparkline(samples, "new", sparklineWidth)
		privilegedSparkline = generateHistoricalSparkline(samples, "privileged", sparklineWidth)
		timeAxis = generateTimeAxis(samples, sparklineWidth, m.activityTracker.StartTime)
		
		// Update cache
		m.sparklineCache.sessionSparkline = sessionsSparkline
		m.sparklineCache.beaconSparkline = beaconsSparkline
		m.sparklineCache.newAgentsSparkline = newSparkline
		m.sparklineCache.privilegedSparkline = privilegedSparkline
		m.sparklineCache.timeAxis = timeAxis
		m.sparklineCache.lastSampleCount = len(samples)
		m.sparklineCache.lastUpdate = time.Now()
	}
	
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
	
	// Time axis (aligned with sparkline) - use cached value
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updateViewportContent updates the viewport with the current agent list
func (m *model) updateViewportContent() {
	// Skip re-render if content hasn't changed (unless forced)
	if !m.contentDirty && m.cachedContent != "" {
		return
	}
	
	var content string
	
	// Check if we're in dashboard view
	if m.view.Type == config.ViewTypeDashboard {
		content = m.renderDashboard()
	} else if m.view.Type == config.ViewTypeNetworkMap {
		content = m.renderNetworkMapView()
	} else {
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
		
		content = strings.Join(contentLines, "\n")
	}
	
	// Cache the rendered content
	m.cachedContent = content
	m.contentDirty = false
	
	// Set viewport content
	m.viewport.SetContent(content)
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
			lines = append(lines, indent+lipgloss.NewStyle().Foreground(connectorColor).Render("   ╰──"+m.getAnimatedHorizontalArrow()+" ")+agentLines[0])
			
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
						connector := lipgloss.NewStyle().Foreground(connectorColor).Render("├──────────────"+m.getAnimatedHorizontalArrow()+" ")
						lines = append(lines, connector+boxLine)
					} else {
						// Last box: use corner
						connector := lipgloss.NewStyle().Foreground(connectorColor).Render("╰──────────────"+m.getAnimatedHorizontalArrow()+" ")
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
	
	line1 := fmt.Sprintf("%s%s%s%s %s %s  %s%s%s",
		connectorStyle.Render("╰────────"),
		protocolBox,
		connectorStyle.Render("────────"),
		connectorStyle.Render(m.getAnimatedHorizontalArrow()),
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

type animationTickMsg struct{}

type domainQueryMsg struct {
	sessionID string
	domain    string
}

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

// animationTickCmd triggers animation frame updates (arrows, etc.)
func animationTickCmd() tea.Msg {
	time.Sleep(150 * time.Millisecond) // Update animation every 150ms
	return animationTickMsg{}
}

// queryDomainCmd queries domain from a session in the background
func queryDomainCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		// Connect to Sliver
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		configPath, err := client.FindConfigFile()
		if err != nil {
			return domainQueryMsg{sessionID: sessionID, domain: ""}
		}
		
		config, err := client.LoadConfig(configPath)
		if err != nil {
			return domainQueryMsg{sessionID: sessionID, domain: ""}
		}
		
		sliverClient := client.NewSliverClient(config)
		if err := sliverClient.Connect(ctx); err != nil {
			return domainQueryMsg{sessionID: sessionID, domain: ""}
		}
		defer sliverClient.Close()
		
		// Query domain with timeout
		queryCtx, queryCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer queryCancel()
		
		domain := sliverClient.QueryDomainFromSession(queryCtx, sessionID)
		
		return domainQueryMsg{
			sessionID: sessionID,
			domain:    domain,
		}
	}
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
		dnsCache:        make(map[string]string), // Initialize DNS cache
		domainCache:     make(map[string]string), // Initialize domain cache (sessionID -> domain)
	}

	// Create and run program with alt screen
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
