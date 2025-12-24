package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View defines how agents are rendered
type View struct {
	Name string
	Type ViewType
}

type ViewType int

const (
	ViewTypeTree ViewType = iota // Current tree layout with arrows
	ViewTypeBox                  // Compact box with side connectors
)

// GetView returns a view by index
func GetView(index int) View {
	views := []View{
		{Name: "Tree", Type: ViewTypeTree},
		{Name: "Box", Type: ViewTypeBox},
	}
	return views[index]
}

// GetViewCount returns the total number of available views
func GetViewCount() int {
	return 2
}

// renderAgentInView renders an agent line based on the current view type
func (m model) renderAgentInView(agent Agent, viewType ViewType) []string {
	switch viewType {
	case ViewTypeBox:
		return m.renderAgentBox(agent)
	case ViewTypeTree:
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
	detailsInfo := fmt.Sprintf("%s │ %s │ %s",
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.ID[:8]),
		lipgloss.NewStyle().Foreground(m.theme.TacticalMuted).Render(agent.RemoteAddress),
		lipgloss.NewStyle().Foreground(m.theme.TacticalValue).Render(agent.Transport),
	)

	// Calculate box width based on content
	userInfoWidth := lipgloss.Width(userInfo)
	detailsInfoWidth := lipgloss.Width(detailsInfo)
	boxContentWidth := userInfoWidth
	if detailsInfoWidth > boxContentWidth {
		boxContentWidth = detailsInfoWidth
	}
	boxContentWidth += 2 // Add padding

	// Border color
	borderColor := m.theme.TacticalBorder
	if agent.IsDead {
		borderColor = m.theme.DeadColor
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Build box lines
	topBorder := borderStyle.Render("╭" + strings.Repeat("─", boxContentWidth) + "╮")
	bottomBorder := borderStyle.Render("╰" + strings.Repeat("─", boxContentWidth) + "╯")

	// Pad content lines to fit box width
	userInfoPadded := userInfo + strings.Repeat(" ", boxContentWidth-userInfoWidth)
	detailsInfoPadded := detailsInfo + strings.Repeat(" ", boxContentWidth-detailsInfoWidth)

	line1 := borderStyle.Render("│") + " " + userInfoPadded + borderStyle.Render("│")
	line2 := borderStyle.Render("│") + " " + detailsInfoPadded + borderStyle.Render("│")

	lines = append(lines, topBorder)
	lines = append(lines, line1)
	lines = append(lines, line2)
	lines = append(lines, bottomBorder)

	return lines
}

// renderAgentTreeWithView renders agent tree with view-specific formatting
func (m model) renderAgentTreeWithView(agent Agent, depth int, viewType ViewType) []string {
	var lines []string

	// Render current agent based on view type
	agentLines := m.renderAgentInView(agent, viewType)

	if viewType == ViewTypeBox {
		// Box view: use vertical connectors for parent-child relationships
		if depth > 0 {
			// Add connector from parent
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
			// Root level - no connector
			lines = append(lines, agentLines...)
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

	// Add spacing between agents at root level
	if depth == 0 && len(agent.Children) == 0 {
		if viewType == ViewTypeBox {
			lines = append(lines, "") // Extra spacing for box view
		} else {
			lines = append(lines, "")
		}
	}

	// Recursively render children
	for _, child := range agent.Children {
		childLines := m.renderAgentTreeWithView(child, depth+1, viewType)
		lines = append(lines, childLines...)
	}

	return lines
}
