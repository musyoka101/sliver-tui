package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/rpcpb"
	"github.com/musyoka101/sliver-graphs/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// SliverConfig represents the Sliver client configuration
type SliverConfig struct {
	Operator      string `json:"operator"`
	LHost         string `json:"lhost"`
	LPort         int    `json:"lport"`
	CACertificate string `json:"ca_certificate"`
	Certificate   string `json:"certificate"`
	PrivateKey    string `json:"private_key"`
	Token         string `json:"token,omitempty"`
}

// SliverClient wraps the gRPC client
type SliverClient struct {
	config *SliverConfig
	conn   *grpc.ClientConn
	rpc    rpcpb.SliverRPCClient
}

// LoadConfig loads the Sliver config from file
func LoadConfig(configPath string) (*SliverConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config SliverConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// FindConfigFile looks for Sliver config in standard location
func FindConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".sliver-client", "configs")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return "", fmt.Errorf("config directory not found: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cfg") {
			return filepath.Join(configDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no .cfg files found in %s", configDir)
}

// Connect establishes a connection to the Sliver server
func (c *SliverClient) Connect(ctx context.Context) error {
	// Create TLS credentials
	tlsConfig, err := c.buildTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to build TLS config: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)

	// Connect to server
	target := fmt.Sprintf("%s:%d", c.config.LHost, c.config.LPort)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	c.conn = conn
	c.rpc = rpcpb.NewSliverRPCClient(conn)

	return nil
}

// buildTLSConfig creates TLS configuration from the Sliver config
func (c *SliverClient) buildTLSConfig() (*tls.Config, error) {
	// Parse CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(c.config.CACertificate)) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Parse client certificate and key
	cert, err := tls.X509KeyPair(
		[]byte(c.config.Certificate),
		[]byte(c.config.PrivateKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	return &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}, nil
}

// GetSessions fetches all active sessions from Sliver
func (c *SliverClient) GetSessions(ctx context.Context) ([]*clientpb.Session, error) {
	// Add token to context if available
	if c.config.Token != "" {
		md := metadata.New(map[string]string{
			"Authorization": "Bearer " + c.config.Token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	resp, err := c.rpc.GetSessions(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return resp.Sessions, nil
}

// GetBeacons fetches all beacons from Sliver
func (c *SliverClient) GetBeacons(ctx context.Context) ([]*clientpb.Beacon, error) {
	// Add token to context if available
	if c.config.Token != "" {
		md := metadata.New(map[string]string{
			"Authorization": "Bearer " + c.config.Token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	resp, err := c.rpc.GetBeacons(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get beacons: %w", err)
	}

	return resp.Beacons, nil
}

// Close closes the connection
func (c *SliverClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NewSliverClient creates a new Sliver client with the given config
func NewSliverClient(config *SliverConfig) *SliverClient {
	return &SliverClient{
		config: config,
	}
}

// ConvertToAgents converts Sliver sessions and beacons to our models.Agent type
func ConvertToAgents(sessions []*clientpb.Session, beacons []*clientpb.Beacon) ([]models.Agent, models.Stats) {
	var agents []models.Agent
	hostMap := make(map[string]bool)

	// Convert sessions
	for _, s := range sessions {
		agent := models.Agent{
			ID:            s.ID,
			Hostname:      s.Hostname,
			Username:      s.Username,
			OS:            s.OS,
			Transport:     s.Transport,
			RemoteAddress: s.RemoteAddress,
			IsSession:     true,
			IsPrivileged:  isPrivileged(s.Username, s.OS),
			IsDead:        false,
			ProxyURL:      s.ProxyURL,
			// Additional fields
			PID:           s.PID,
			Filename:      s.Filename,
			Arch:          s.Arch,
			Version:       s.Version,
			ActiveC2:      s.ActiveC2,
			LastCheckin:   s.LastCheckin,
			Evasion:       s.Evasion,
			Burned:        s.Burned,
		}
		agents = append(agents, agent)
		hostMap[s.Hostname] = true
	}

	// Convert beacons
	for _, b := range beacons {
		// Calculate if beacon is dead based on last check-in time
		// A beacon is considered dead if it hasn't checked in for 3x its interval
		// Note: b.Interval is in nanoseconds (time.Duration)
		isDead := b.IsDead
		
		if !isDead && b.LastCheckin > 0 && b.Interval > 0 {
			lastCheckin := time.Unix(b.LastCheckin, 0)
			// Interval is already a time.Duration in nanoseconds, don't multiply by time.Second
			deadThreshold := time.Duration(b.Interval * 3)
			timeSinceCheckin := time.Since(lastCheckin)
			
			if timeSinceCheckin > deadThreshold {
				isDead = true
			}
		}
		
		agent := models.Agent{
			ID:            b.ID,
			Hostname:      b.Hostname,
			Username:      b.Username,
			OS:            b.OS,
			Transport:     b.Transport,
			RemoteAddress: b.RemoteAddress,
			IsSession:     false,
			IsPrivileged:  isPrivileged(b.Username, b.OS),
			IsDead:        isDead,
			ProxyURL:      b.ProxyURL,
			// Additional fields
			PID:            b.PID,
			Filename:       b.Filename,
			Arch:           b.Arch,
			Version:        b.Version,
			ActiveC2:       b.ActiveC2,
			Interval:       b.Interval,
			Jitter:         b.Jitter,
			NextCheckin:    b.NextCheckin,
			TasksCount:     b.TasksCount,
			TasksCompleted: b.TasksCountCompleted,
			LastCheckin:    b.LastCheckin,
			Evasion:        b.Evasion,
			Burned:         b.Burned,
		}
		agents = append(agents, agent)
		hostMap[b.Hostname] = true
	}

	stats := models.Stats{
		Sessions:    len(sessions),
		Beacons:     len(beacons),
		Hosts:       len(hostMap),
		Compromised: len(agents),
	}

	return agents, stats
}

// isPrivileged checks if a user is privileged
func isPrivileged(username, os string) bool {
	userLower := strings.ToLower(username)
	osLower := strings.ToLower(os)

	if strings.Contains(osLower, "windows") {
		// Windows privileged users
		return strings.Contains(userLower, "administrator") ||
			strings.Contains(userLower, "system") ||
			strings.Contains(userLower, "nt authority\\system")
	} else if strings.Contains(osLower, "linux") || strings.Contains(osLower, "darwin") {
		// Unix-like privileged users
		return strings.Contains(userLower, "root") || userLower == "root"
	}

	return false
}

// FetchAgents connects to Sliver and fetches all agents
func FetchAgents(ctx context.Context) ([]models.Agent, models.Stats, error) {
	// Find config file
	configPath, err := FindConfigFile()
	if err != nil {
		return nil, models.Stats{}, fmt.Errorf("config not found: %w", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, models.Stats{}, fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	client := NewSliverClient(config)

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Connect(connectCtx); err != nil {
		return nil, models.Stats{}, fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	// Fetch sessions and beacons
	sessions, err := client.GetSessions(ctx)
	if err != nil {
		return nil, models.Stats{}, fmt.Errorf("failed to get sessions: %w", err)
	}

	beacons, err := client.GetBeacons(ctx)
	if err != nil {
		return nil, models.Stats{}, fmt.Errorf("failed to get beacons: %w", err)
	}

	// Convert to our models.Agent type
	agents, stats := ConvertToAgents(sessions, beacons)

	return agents, stats, nil
}
