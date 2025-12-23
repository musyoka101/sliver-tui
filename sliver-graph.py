#!/usr/bin/env python3
"""
Sliver C2 Network Topology Visualizer
A standalone tool to visualize compromised hosts in Sliver C2

Usage:
    python3 sliver-graph.py
    
This tool connects to the Sliver RPC and displays an ASCII network topology
showing all active sessions and beacons.
"""

import asyncio
import os
import sys
from pathlib import Path

# Import Sliver protobuf definitions
try:
    # Try to import from sliver-py if installed
    from sliver import SliverClientConfig, SliverClient
    HAS_SLIVER_PY = True
except ImportError:
    HAS_SLIVER_PY = False
    print("[!] sliver-py not found. Install it with: pip3 install sliver-py")
    print("[!] Alternatively, this script can work with direct gRPC if you have the protobuf files")


class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    GRAY = '\033[90m'
    RED = '\033[91m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    MAGENTA = '\033[95m'
    CYAN = '\033[96m'
    WHITE = '\033[97m'


def draw_computer(os_name):
    """Draw a computer ASCII art based on OS"""
    os_lower = os_name.lower() if os_name else ''
    
    if 'windows' in os_lower:
        return [
            "  ┌─────┐",
            "  │ ▄▄  │",
            "  │ ▀▀  │",
            "  └─────┘",
            "  ┌─────┐",
            "  └─────┘"
        ]
    elif 'linux' in os_lower:
        return [
            "  ┌─────┐",
            "  │ ╱╲  │",
            "  │▕  ▏ │",
            "  └─────┘",
            "  ┌─────┐",
            "  └─────┘"
        ]
    else:
        return [
            "  ┌─────┐",
            "  │ ▄▄▄ │",
            "  │ ███ │",
            "  └─────┘",
            "  ┌─────┐",
            "  └─────┘"
        ]


def draw_sliver_logo():
    """Draw Sliver C2 logo"""
    return [
        "   ▄████▄   ",
        "  ████████  ",
        "  ▀██████▀  ",
        "    ▀██▀    ",
        "     ▀      "
    ]


def get_protocol_color(transport):
    """Get color for transport protocol"""
    transport_lower = transport.lower() if transport else ''
    if 'http' in transport_lower:
        return Colors.GREEN
    elif 'mtls' in transport_lower or 'tls' in transport_lower:
        return Colors.CYAN
    elif 'dns' in transport_lower:
        return Colors.YELLOW
    elif 'tcp' in transport_lower:
        return Colors.BLUE
    else:
        return Colors.WHITE


def draw_graph(sessions, beacons):
    """Draw the main network graph visualization"""
    output = []
    
    # Header
    output.append("")
    output.append(f"{Colors.BOLD}{Colors.CYAN}╔══════════════════════════════════════════════════════════════════════════════╗{Colors.ENDC}")
    output.append(f"{Colors.BOLD}{Colors.CYAN}║            SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION                        ║{Colors.ENDC}")
    output.append(f"{Colors.BOLD}{Colors.CYAN}╚══════════════════════════════════════════════════════════════════════════════╝{Colors.ENDC}")
    output.append("")
    
    # Calculate totals
    total_hosts = len(sessions) + len(beacons)
    
    if total_hosts == 0:
        output.append(f"{Colors.RED}                    [No Active Hosts Connected]{Colors.ENDC}")
        output.append("")
        return '\n'.join(output)
    
    # Combine all hosts
    all_hosts = []
    for session in sessions:
        all_hosts.append({
            'type': 'session',
            'data': session
        })
    for beacon in beacons:
        all_hosts.append({
            'type': 'beacon',
            'data': beacon
        })
    
    # Draw the C2 server logo on the left
    logo_lines = draw_sliver_logo()
    
    # Calculate starting position for logo (center it vertically)
    logo_start = (len(all_hosts) * 7) // 2 - 2
    
    line_count = 0
    
    # Draw each host with connections
    for idx, host in enumerate(all_hosts):
        host_data = host['data']
        is_session = host['type'] == 'session'
        
        # Extract host info
        host_id = host_data['ID'][:8] if 'ID' in host_data else 'N/A'
        hostname = host_data.get('Hostname', 'Unknown')
        username = host_data.get('Username', 'Unknown')
        os_name = host_data.get('OS', 'Unknown')
        transport = host_data.get('Transport', 'unknown')
        
        # Determine colors
        if is_session:
            host_color = Colors.GREEN
            type_label = "session"
        else:
            host_color = Colors.YELLOW
            type_label = "beacon"
        
        protocol_color = get_protocol_color(transport)
        
        # Draw computer icon
        computer = draw_computer(os_name)
        
        # Build the lines for this host
        host_lines = []
        
        # Line 0: connection start
        if line_count >= logo_start and line_count < logo_start + len(logo_lines):
            logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
        else:
            logo_line = " " * 12
        host_lines.append(f"  {logo_line}  {Colors.GRAY}·········╲{Colors.ENDC}")
        line_count += 1
        
        # Line 1: protocol label
        if line_count >= logo_start and line_count < logo_start + len(logo_lines):
            logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
        else:
            logo_line = " " * 12
        proto_label = f"{transport.upper()} (p0)"
        host_lines.append(f"  {logo_line}   ────────── {protocol_color}{proto_label:^12}{Colors.ENDC} ──────────▶")
        line_count += 1
        
        # Line 2: connection end with computer start
        if line_count >= logo_start and line_count < logo_start + len(logo_lines):
            logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
        else:
            logo_line = " " * 12
        host_lines.append(f"  {logo_line}  {Colors.GRAY}·········╱{Colors.ENDC}               {host_color}{computer[0]}{Colors.ENDC}")
        line_count += 1
        
        # Line 3-7: computer and info
        for i in range(1, 6):
            if line_count >= logo_start and line_count < logo_start + len(logo_lines):
                logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
            else:
                logo_line = " " * 12
            
            if i == 1:
                info_text = f"  {username}@{hostname}"
                host_lines.append(f"  {logo_line}                        {host_color}{computer[i]}{Colors.ENDC}{Colors.GRAY}{info_text}{Colors.ENDC}")
            elif i == 2:
                info_text = f"  {host_id} ({type_label})"
                host_lines.append(f"  {logo_line}                        {host_color}{computer[i]}{Colors.ENDC}{Colors.GRAY}{info_text}{Colors.ENDC}")
            else:
                host_lines.append(f"  {logo_line}                        {host_color}{computer[i]}{Colors.ENDC}")
            line_count += 1
        
        # Add lines to output
        for line in host_lines:
            output.append(line)
        
        # Add spacing between hosts
        if idx < len(all_hosts) - 1:
            if line_count >= logo_start and line_count < logo_start + len(logo_lines):
                logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
            else:
                logo_line = " " * 12
            output.append(f"  {logo_line}")
            line_count += 1
    
    output.append("")
    output.append(f"{Colors.GRAY}{'─' * 80}{Colors.ENDC}")
    output.append(f"  {Colors.GREEN}●{Colors.ENDC} Active Sessions: {len(sessions)}  {Colors.YELLOW}●{Colors.ENDC} Active Beacons: {len(beacons)}  {Colors.CYAN}●{Colors.ENDC} Total: {total_hosts}")
    output.append("")
    
    return '\n'.join(output)


async def get_sliver_data(config_file):
    """Connect to Sliver and get sessions/beacons"""
    # Connect to Sliver
    config = SliverClientConfig.parse_config_file(str(config_file))
    client = SliverClient(config)
    await client.connect()
    
    # Get sessions and beacons
    sessions = await client.sessions()
    beacons = await client.beacons()
    
    return sessions, beacons


def main():
    """Main entry point"""
    if not HAS_SLIVER_PY:
        print(f"{Colors.RED}[!] This script requires sliver-py{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Install it with: pip3 install sliver-py{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Or install from source: https://github.com/moloch--/sliver-py{Colors.ENDC}")
        sys.exit(1)
    
    # Look for Sliver config file
    config_path = Path.home() / ".sliver-client" / "configs"
    
    if not config_path.exists():
        print(f"{Colors.RED}[!] Sliver client config not found at {config_path}{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Make sure you have run Sliver client at least once{Colors.ENDC}")
        sys.exit(1)
    
    # Get the first config file
    config_files = list(config_path.glob("*.cfg"))
    if not config_files:
        print(f"{Colors.RED}[!] No Sliver config files found{Colors.ENDC}")
        sys.exit(1)
    
    config_file = config_files[0]
    print(f"{Colors.CYAN}[*] Using config: {config_file.name}{Colors.ENDC}")
    
    try:
        # Get data using async
        sessions, beacons = asyncio.run(get_sliver_data(config_file))
        
        # Convert to dict format for drawing
        session_list = []
        for s in sessions:
            session_list.append({
                'ID': s.ID,
                'Hostname': s.Hostname,
                'Username': s.Username,
                'OS': s.OS,
                'Transport': s.Transport,
            })
        
        beacon_list = []
        for b in beacons:
            beacon_list.append({
                'ID': b.ID,
                'Hostname': b.Hostname,
                'Username': b.Username,
                'OS': b.OS,
                'Transport': b.Transport,
            })
        
        # Draw the graph
        graph = draw_graph(session_list, beacon_list)
        print(graph)
        
    except Exception as e:
        print(f"{Colors.RED}[!] Error connecting to Sliver: {e}{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Make sure Sliver server is running{Colors.ENDC}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
