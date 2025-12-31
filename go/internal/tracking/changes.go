package tracking

import (
	"sync"
	"time"

	"github.com/musyoka101/sliver-graphs/internal/models"
)

// Global tracking for agent changes
var (
	agentTracker     = make(map[string]time.Time) // ID -> first seen time
	trackerMutex     sync.RWMutex
	newAgentTimeout  = 5 * time.Minute // Mark as NEW if seen < 5 minutes ago
	lostAgents       = make(map[string]models.Agent)
	lostAgentTimeout = 5 * time.Minute
)

// TrackAgentChanges updates the tracking maps for new/lost agents
func TrackAgentChanges(agents []models.Agent) []models.Agent {
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
				lostAgents[agentID] = models.Agent{
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

// GetLostAgentsCount returns the number of recently lost agents
func GetLostAgentsCount() int {
	trackerMutex.RLock()
	defer trackerMutex.RUnlock()
	return len(lostAgents)
}

// GetLostAgentTimeout returns the timeout duration for lost agents
func GetLostAgentTimeout() time.Duration {
	return lostAgentTimeout
}
