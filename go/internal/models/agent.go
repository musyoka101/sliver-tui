package models

import "time"

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
	PID            int32  // Process ID
	Filename       string // Process filename/path (Argv[0])
	Arch           string // Architecture (x64, x86, arm64, etc.)
	Version        string // Implant version
	ActiveC2       string // Active C2 server URL
	Interval       int64  // Beacon check-in interval (nanoseconds)
	Jitter         int64  // Beacon jitter
	NextCheckin    int64  // Next beacon check-in time (unix timestamp)
	TasksCount     int64  // Total tasks queued
	TasksCompleted int64  // Tasks completed
	LastCheckin    int64  // Last check-in time (unix timestamp)
	Evasion        bool   // Evasion mode enabled
	Burned         bool   // Marked as compromised
}

// Stats holds statistics
type Stats struct {
	Sessions    int
	Beacons     int
	Hosts       int
	Compromised int
}
