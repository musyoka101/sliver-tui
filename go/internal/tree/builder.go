package tree

import (
	"strings"

	"github.com/musyoka101/sliver-graphs/internal/models"
)

// BuildAgentTree organizes agents into a hierarchical tree based on pivot relationships
// Optimized to O(n) complexity using parent-child index
func BuildAgentTree(agents []models.Agent) []models.Agent {
	// Create maps for quick lookup
	agentMap := make(map[string]*models.Agent)
	for i := range agents {
		agentMap[agents[i].ID] = &agents[i]
	}

	// Identify parent-child relationships and build index
	// ProxyURL format can be like "socks5://agent-id" or contain agent ID
	var rootAgents []models.Agent
	parentChildIndex := make(map[string][]models.Agent) // parentID -> children

	for i := range agents {
		if agents[i].ProxyURL == "" {
			// This is a root agent (directly connected to C2)
			rootAgents = append(rootAgents, agents[i])
		} else {
			// This agent is pivoted through another
			// Try to extract parent ID from ProxyURL
			agents[i].ParentID = extractParentID(agents[i].ProxyURL, agentMap)
			
			// Add to parent-child index
			if agents[i].ParentID != "" {
				parentChildIndex[agents[i].ParentID] = append(parentChildIndex[agents[i].ParentID], agents[i])
			}
		}
	}

	// Build tree structure using index (O(n) instead of O(n²))
	for i := range rootAgents {
		rootAgents[i].Children = buildChildrenFromIndex(rootAgents[i].ID, parentChildIndex)
	}

	// If no root agents found, return all agents as roots
	if len(rootAgents) == 0 {
		return agents
	}

	return rootAgents
}

// extractParentID tries to extract parent agent ID from ProxyURL
func extractParentID(proxyURL string, agentMap map[string]*models.Agent) string {
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

// buildChildrenFromIndex recursively builds children using pre-built index
// This is O(n) total across all calls instead of O(n²)
func buildChildrenFromIndex(parentID string, parentChildIndex map[string][]models.Agent) []models.Agent {
	children, exists := parentChildIndex[parentID]
	if !exists {
		return nil
	}
	
	// Process each child and recursively get their children
	result := make([]models.Agent, len(children))
	for i, child := range children {
		result[i] = child
		result[i].Children = buildChildrenFromIndex(child.ID, parentChildIndex)
	}
	
	return result
}

// findChildren is deprecated - kept for backwards compatibility
// Use buildChildrenFromIndex with parentChildIndex instead
func findChildren(parentID string, allAgents []models.Agent) []models.Agent {
	var children []models.Agent

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
