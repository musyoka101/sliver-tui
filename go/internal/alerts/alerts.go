package alerts

import (
	"sync"
	"time"
)

// AlertType represents the type of alert
type AlertType int

const (
	AlertCritical AlertType = iota // Agent lost, C2 down, security breach
	AlertWarning                    // Beacon late, connection unstable, timeout
	AlertSuccess                    // New agent, privileged access, session established
	AlertInfo                       // Session activity, config changes
	AlertNotice                     // System messages, background updates
)

// AlertCategory represents what the alert is about
type AlertCategory int

const (
	CategoryAgentConnected AlertCategory = iota
	CategoryAgentDisconnected
	CategoryBeaconLate
	CategoryBeaconMissed
	CategoryPrivilegedAccess
	CategorySessionOpened
	CategorySessionClosed
	CategoryC2Connected
	CategoryC2Disconnected
	CategorySecurityBreach
	CategorySystemNotice
)

// Alert represents a single alert/event
type Alert struct {
	ID        string
	Type      AlertType
	Category  AlertCategory
	Message   string
	AgentName string
	Timestamp time.Time
	TTL       time.Duration // How long to display
	IsNew     bool          // For animation purposes
}

// AlertManager manages the alert queue
type AlertManager struct {
	alerts       []Alert
	maxAlerts    int
	mu           sync.RWMutex
	pulseState   int       // For animation: 0, 1, 2 (dim, normal, bright)
	lastPulseAt  time.Time
	pulseDuration time.Duration
}

// NewAlertManager creates a new alert manager
func NewAlertManager(maxAlerts int) *AlertManager {
	return &AlertManager{
		alerts:        make([]Alert, 0, maxAlerts),
		maxAlerts:     maxAlerts,
		pulseState:    0,
		pulseDuration: 500 * time.Millisecond, // Pulse every 500ms
	}
}

// AddAlert adds a new alert to the queue
func (am *AlertManager) AddAlert(alertType AlertType, category AlertCategory, message, agentName string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Set TTL based on alert type
	var ttl time.Duration
	switch alertType {
	case AlertCritical:
		ttl = 30 * time.Second
	case AlertWarning:
		ttl = 20 * time.Second
	case AlertSuccess:
		ttl = 15 * time.Second
	case AlertInfo:
		ttl = 10 * time.Second
	case AlertNotice:
		ttl = 8 * time.Second
	}

	alert := Alert{
		ID:        generateID(),
		Type:      alertType,
		Category:  category,
		Message:   message,
		AgentName: agentName,
		Timestamp: time.Now(),
		TTL:       ttl,
		IsNew:     true,
	}

	// Check for duplicates (deduplication)
	for i := range am.alerts {
		if am.alerts[i].Category == category && 
		   am.alerts[i].AgentName == agentName && 
		   time.Since(am.alerts[i].Timestamp) < 5*time.Second {
			// Duplicate within 5 seconds, skip
			return
		}
	}

	// Add to front of queue
	am.alerts = append([]Alert{alert}, am.alerts...)

	// Trim to max size
	if len(am.alerts) > am.maxAlerts {
		am.alerts = am.alerts[:am.maxAlerts]
	}
}

// GetAlerts returns current alerts (removes expired ones)
func (am *AlertManager) GetAlerts() []Alert {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	validAlerts := make([]Alert, 0, len(am.alerts))

	for i := range am.alerts {
		if now.Sub(am.alerts[i].Timestamp) < am.alerts[i].TTL {
			// Mark as not new after first display
			am.alerts[i].IsNew = false
			validAlerts = append(validAlerts, am.alerts[i])
		}
	}

	am.alerts = validAlerts
	return validAlerts
}

// GetPulseState returns current pulse animation state
func (am *AlertManager) GetPulseState() int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	now := time.Now()
	if now.Sub(am.lastPulseAt) > am.pulseDuration {
		am.mu.RUnlock()
		am.mu.Lock()
		am.lastPulseAt = now
		am.pulseState = (am.pulseState + 1) % 3 // Cycle: 0 → 1 → 2 → 0
		am.mu.Unlock()
		am.mu.RLock()
	}

	return am.pulseState
}

// UpdatePulse cycles the pulse animation state
func (am *AlertManager) UpdatePulse() {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.pulseState = (am.pulseState + 1) % 3 // Cycle: 0 → 1 → 2 → 0
	am.lastPulseAt = time.Now()
}

// HasCriticalAlerts checks if there are any critical alerts
func (am *AlertManager) HasCriticalAlerts() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for i := range am.alerts {
		if am.alerts[i].Type == AlertCritical {
			return true
		}
	}
	return false
}

// ClearAll removes all alerts
func (am *AlertManager) ClearAll() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.alerts = make([]Alert, 0, am.maxAlerts)
}

// generateID creates a unique ID for alerts
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}

// GetIcon returns the military-style icon for the alert type
func (a Alert) GetIcon() string {
	switch a.Type {
	case AlertCritical:
		return "║█║"
	case AlertWarning:
		return "║▒║"
	case AlertSuccess:
		return "║▓║"
	case AlertInfo:
		return "║░║"
	case AlertNotice:
		return "║ ║"
	default:
		return "║ ║"
	}
}

// GetLabel returns a short label for the category
func (a Alert) GetLabel() string {
	switch a.Category {
	case CategoryAgentConnected:
		return "AGENT ACQUIRED"
	case CategoryAgentDisconnected:
		return "AGENT LOST"
	case CategoryBeaconLate:
		return "CHECK-IN LATE"
	case CategoryBeaconMissed:
		return "BEACON MISSED"
	case CategoryPrivilegedAccess:
		return "PRIV ESCALATED"
	case CategorySessionOpened:
		return "SESSION INIT"
	case CategorySessionClosed:
		return "SESSION TERM"
	case CategoryC2Connected:
		return "C2 ONLINE"
	case CategoryC2Disconnected:
		return "C2 OFFLINE"
	case CategorySecurityBreach:
		return "SECURITY ALERT"
	case CategorySystemNotice:
		return "SYS NOTICE"
	default:
		return "EVENT"
	}
}
