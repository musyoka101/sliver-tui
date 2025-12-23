#!/usr/bin/env python3
"""
Sliver C2 Network Topology Visualizer
A standalone tool to visualize compromised hosts in Sliver C2

Usage:
    python3 sliver-graph.py
    
This tool connects to the Sliver RPC and displays an ASCII network topology
showing all active sessions and beacons. It refreshes automatically every few seconds.
Press Ctrl+C to exit.
"""

import asyncio
import os
import sys
import time
import signal
from pathlib import Path
from datetime import datetime

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
            "┌─────┐",
            "│ ▄▄  │",
            "│ ▀▀  │",
            "└─────┘",
            "┌─────┐",
            "└─────┘"
        ]
    elif 'linux' in os_lower:
        return [
            "┌─────┐",
            "│ ╱╲  │",
            "│▕  ▏ │",
            "└─────┘",
            "┌─────┐",
            "└─────┘"
        ]
    else:
        return [
            "┌─────┐",
            "│ ▄▄▄ │",
            "│ ███ │",
            "└─────┘",
            "┌─────┐",
            "└─────┘"
        ]


def draw_sliver_logo():
    """Draw Sliver C2 logo with emoji"""
    return [
        "   🎯 C2    ",
        "  ▄████▄   ",
        "  ████████  ",
        "  ▀██████▀  ",
        "    ▀██▀    "
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


def clear_screen():
    """Clear the terminal screen"""
    os.system('clear' if os.name != 'nt' else 'cls')


def draw_graph(sessions, beacons, last_update=None):
    """Draw the main network graph visualization"""
    output = []
    
    # Header
    output.append("")
    output.append(f"{Colors.BOLD}{Colors.CYAN}╔════════════════════════════════════════════════════════════════════════════╗{Colors.ENDC}")
    output.append(f"{Colors.BOLD}{Colors.CYAN}║  🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION                            ║{Colors.ENDC}")
    output.append(f"{Colors.BOLD}{Colors.CYAN}╚════════════════════════════════════════════════════════════════════════════╝{Colors.ENDC}")
    
    # Show last update time
    if last_update:
        update_time = datetime.fromtimestamp(last_update).strftime("%Y-%m-%d %H:%M:%S")
        output.append(f"{Colors.GRAY}  ⏰ Last Update: {update_time}  |  Press Ctrl+C to exit{Colors.ENDC}")
    
    output.append("")
    
    # Calculate totals
    total_hosts = len(sessions) + len(beacons)
    
    if total_hosts == 0:
        output.append(f"{Colors.RED}                    ⚠️  [No Active Hosts Connected]{Colors.ENDC}")
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
    
    # Calculate starting position for logo (center it vertically with new single-line design)
    logo_start = (len(all_hosts) * 1) // 2 - len(logo_lines) // 2
    if logo_start < 0:
        logo_start = 0
    
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
        
        # Build the lines for this host - using single arrow design with emojis
        host_lines = []
        proto_label = f"{transport.upper()}"
        
        # Single line with arrow, protocol, icon and info
        if line_count >= logo_start and line_count < logo_start + len(logo_lines):
            logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
        else:
            logo_line = " " * 12
        
        # Emoji icons based on OS and type
        if 'windows' in os_name.lower():
            pc_icon = "🖥️ " if is_session else "💻"
        elif 'linux' in os_name.lower():
            pc_icon = "🐧" if is_session else "🖳 "
        else:
            pc_icon = "💻" if is_session else "🖥️ "
        
        # Status indicator
        status_icon = "🟢" if is_session else "🟡"
        
        # Format: Logo  ──[ PROTOCOL ]──▶ 🟢 💻 username@hostname  id (type)
        host_lines.append(
            f"  {logo_line}      ───[ {protocol_color}{proto_label:^6}{Colors.ENDC} ]───▶ "
            f"{status_icon} {host_color}{pc_icon}{Colors.ENDC} "
            f"{Colors.BOLD}{Colors.CYAN}{username}@{hostname}{Colors.ENDC}  "
            f"{Colors.GRAY}{host_id} ({type_label}){Colors.ENDC}"
        )
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
    output.append(f"  🟢 Active Sessions: {Colors.BOLD}{Colors.GREEN}{len(sessions)}{Colors.ENDC}  🟡 Active Beacons: {Colors.BOLD}{Colors.YELLOW}{len(beacons)}{Colors.ENDC}  🔵 Total Compromised: {Colors.BOLD}{Colors.CYAN}{total_hosts}{Colors.ENDC}")
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


def signal_handler(sig, frame):
    """Handle Ctrl+C gracefully"""
    print(f"\n\n{Colors.YELLOW}[*] Exiting... Goodbye!{Colors.ENDC}\n")
    sys.exit(0)


async def monitor_loop(config_file, refresh_interval=5):
    """Main monitoring loop that refreshes the display"""
    while True:
        try:
            # Get data from Sliver
            sessions, beacons = await get_sliver_data(config_file)
            
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
            
            # Clear screen and draw the graph
            clear_screen()
            graph = draw_graph(session_list, beacon_list, time.time())
            print(graph)
            
            # Wait before next refresh
            await asyncio.sleep(refresh_interval)
            
        except KeyboardInterrupt:
            print(f"\n\n{Colors.YELLOW}[*] Exiting... Goodbye!{Colors.ENDC}\n")
            sys.exit(0)
        except Exception as e:
            clear_screen()
            print(f"{Colors.RED}[!] Error: {e}{Colors.ENDC}")
            print(f"{Colors.YELLOW}[*] Retrying in {refresh_interval} seconds...{Colors.ENDC}")
            await asyncio.sleep(refresh_interval)


def main():
    """Main entry point"""
    import argparse
    
    parser = argparse.ArgumentParser(
        description='Sliver C2 Network Topology Visualizer - Live monitoring dashboard',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
Examples:
  python3 sliver-graph.py              # Default 5 second refresh
  python3 sliver-graph.py -r 10        # Refresh every 10 seconds
  python3 sliver-graph.py --once       # Run once without loop
        '''
    )
    parser.add_argument('-r', '--refresh', type=int, default=5, 
                        help='Refresh interval in seconds (default: 5)')
    parser.add_argument('--once', action='store_true',
                        help='Run once and exit (no live monitoring)')
    
    args = parser.parse_args()
    
    if not HAS_SLIVER_PY:
        print(f"{Colors.RED}[!] This script requires sliver-py{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Install it with: pip3 install sliver-py{Colors.ENDC}")
        print(f"{Colors.YELLOW}[*] Or install from source: https://github.com/moloch--/sliver-py{Colors.ENDC}")
        sys.exit(1)
    
    # Set up signal handler for Ctrl+C
    signal.signal(signal.SIGINT, signal_handler)
    
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
    
    if args.once:
        # Run once and exit
        print(f"{Colors.CYAN}[*] Using config: {config_file.name}{Colors.ENDC}")
        
        async def run_once():
            sessions, beacons = await get_sliver_data(config_file)
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
            
            clear_screen()
            graph = draw_graph(session_list, beacon_list, time.time())
            print(graph)
        
        try:
            asyncio.run(run_once())
        except Exception as e:
            print(f"{Colors.RED}[!] Error: {e}{Colors.ENDC}")
            sys.exit(1)
    else:
        # Run in monitoring mode
        print(f"{Colors.CYAN}[*] Using config: {config_file.name}{Colors.ENDC}")
        print(f"{Colors.CYAN}[*] Refresh interval: {args.refresh} seconds{Colors.ENDC}")
        print(f"{Colors.CYAN}[*] Starting live monitoring... (Press Ctrl+C to exit){Colors.ENDC}")
        time.sleep(2)
        
        try:
            # Run the monitoring loop
            asyncio.run(monitor_loop(config_file, refresh_interval=args.refresh))
            
        except KeyboardInterrupt:
            print(f"\n\n{Colors.YELLOW}[*] Exiting... Goodbye!{Colors.ENDC}\n")
            sys.exit(0)
        except Exception as e:
            print(f"{Colors.RED}[!] Error: {e}{Colors.ENDC}")
            import traceback
            traceback.print_exc()
            sys.exit(1)


if __name__ == '__main__':
    main()
