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
func (m model) renderAgentTreeWithView(agent Agent, depth int, viewType ViewType) []string {
	return m.renderAgentTreeWithViewAndContext(agent, depth, viewType, false, false)
}

// renderAgentTreeWithViewAndContext renders agent tree with context about siblings
func (m model) renderAgentTreeWithViewAndContext(agent Agent, depth int, viewType ViewType, hasNextSibling bool, isLastChild bool) []string {
	var lines []string

	// Render current agent based on view type
	agentLines := m.renderAgentInView(agent, viewType)

	if viewType == ViewTypeBox {
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
	if depth == 0 && len(agent.Children) == 0 && viewType == ViewTypeTree {
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
