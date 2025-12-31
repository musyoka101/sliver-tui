package config

import (
	"context"
	"net"
	"strings"
	"time"
)

// ResolveDomainFromIP performs reverse DNS lookup to get FQDN from IP address
func ResolveDomainFromIP(ipAddress string) string {
	// Remove port if present (format: "10.10.110.250:5445")
	ip := ipAddress
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	
	// Validate IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}
	
	// Perform reverse DNS lookup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	resolver := &net.Resolver{}
	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	
	// Get the first FQDN returned
	fqdn := names[0]
	
	// Remove trailing dot if present (DNS convention)
	fqdn = strings.TrimSuffix(fqdn, ".")
	
	// Extract domain from FQDN (everything after first dot)
	// Example: "m3sqlw.m3c.local" -> "m3c.local"
	if strings.Contains(fqdn, ".") {
		parts := strings.Split(fqdn, ".")
		if len(parts) >= 2 {
			// Return everything after first part (hostname)
			domain := strings.Join(parts[1:], ".")
			return strings.ToLower(domain)
		}
	}
	
	return ""
}

// ExtractDomainFromHostname extracts domain from hostname if it contains FQDN
func ExtractDomainFromHostname(hostname string) string {
	if strings.Contains(hostname, ".") {
		parts := strings.Split(hostname, ".")
		if len(parts) >= 2 {
			return strings.ToLower(strings.Join(parts[1:], "."))
		}
	}
	return ""
}
