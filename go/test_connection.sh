#!/bin/bash

# Simple test script to verify Sliver connection works
echo "🎯 Testing Sliver Connection..."
echo "================================"
echo ""

# Build test program
cat > /tmp/test_sliver_conn.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"time"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
	"crypto/tls"
	"crypto/x509"

	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/rpcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type SliverConfig struct {
	Operator      string `json:"operator"`
	LHost         string `json:"lhost"`
	LPort         int    `json:"lport"`
	CACertificate string `json:"ca_certificate"`
	Certificate   string `json:"certificate"`
	PrivateKey    string `json:"private_key"`
	Token         string `json:"token,omitempty"`
}

func main() {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".sliver-client", "configs", "musyoka_127.0.0.1.cfg")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ Failed to read config: %v\n", err)
		return
	}

	var config SliverConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("❌ Failed to parse config: %v\n", err)
		return
	}

	// Create TLS config
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(config.CACertificate))
	
	cert, err := tls.X509KeyPair([]byte(config.Certificate), []byte(config.PrivateKey))
	if err != nil {
		fmt.Printf("❌ Failed to parse certificates: %v\n", err)
		return
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	creds := credentials.NewTLS(tlsConfig)
	target := fmt.Sprintf("%s:%d", config.LHost, config.LPort)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", target, err)
		return
	}
	defer conn.Close()

	rpc := rpcpb.NewSliverRPCClient(conn)

	// Add token to context for authentication
	if config.Token != "" {
		md := metadata.New(map[string]string{
			"Authorization": "Bearer " + config.Token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Get sessions
	sessResp, err := rpc.GetSessions(ctx, &commonpb.Empty{})
	if err != nil {
		fmt.Printf("❌ Failed to get sessions: %v\n", err)
		return
	}

	// Get beacons
	beaconResp, err := rpc.GetBeacons(ctx, &commonpb.Empty{})
	if err != nil {
		fmt.Printf("❌ Failed to get beacons: %v\n", err)
		return
	}

	fmt.Printf("✅ Connected successfully to %s!\n\n", target)
	fmt.Printf("📊 Statistics:\n")
	fmt.Printf("   Sessions: %d\n", len(sessResp.Sessions))
	fmt.Printf("   Beacons:  %d\n\n", len(beaconResp.Beacons))

	if len(sessResp.Sessions) > 0 {
		fmt.Println("🖥️  Active Sessions:")
		for i, s := range sessResp.Sessions {
			priv := ""
			if strings.Contains(strings.ToLower(s.Username), "administrator") ||
				strings.Contains(strings.ToLower(s.Username), "system") ||
				strings.Contains(strings.ToLower(s.Username), "root") {
				priv = " 💎"
			}
			fmt.Printf("   %d. %s@%s [%s]%s\n", i+1, s.Username, s.Hostname, s.Transport, priv)
			fmt.Printf("      ID: %s\n", s.ID)
		}
		fmt.Println()
	}

	if len(beaconResp.Beacons) > 0 {
		fmt.Println("📡 Active Beacons:")
		for i, b := range beaconResp.Beacons {
			priv := ""
			if strings.Contains(strings.ToLower(b.Username), "administrator") ||
				strings.Contains(strings.ToLower(b.Username), "system") ||
				strings.Contains(strings.ToLower(b.Username), "root") {
				priv = " 💎"
			}
			dead := ""
			if b.IsDead {
				dead = " 💀"
			}
			fmt.Printf("   %d. %s@%s [%s]%s%s\n", i+1, b.Username, b.Hostname, b.Transport, priv, dead)
			fmt.Printf("      ID: %s\n", b.ID)
		}
	}
}
EOF

# Build and run
cd "$(dirname "$0")"
go run /tmp/test_sliver_conn.go

# Cleanup
rm -f /tmp/test_sliver_conn.go
