package tracking

import (
	"sync"
	"time"

	"github.com/musyoka101/sliver-graphs/internal/models"
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
	StartTime      time.Time
	Samples        []ActivitySample
	SampleInterval time.Duration // 10 minutes
	MaxSamples     int           // 72 samples (12 hours)
	mutex          sync.RWMutex
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

// SampleCurrentActivity samples the current agent state
func (at *ActivityTracker) SampleCurrentActivity(agents []models.Agent, stats models.Stats) {
	// Count metrics from current agents
	newCount := 0
	privilegedCount := 0

	for _, agent := range agents {
		if agent.IsNew {
			newCount++
		}
		if agent.IsPrivileged {
			privilegedCount++
		}
	}

	// Add sample to tracker
	at.AddSample(stats.Sessions, stats.Beacons, newCount, privilegedCount)
}
