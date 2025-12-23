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


def is_privileged(username, uid, os_name):
    """Detect if session is running with elevated privileges"""
    username_lower = username.lower() if username else ''
    os_lower = os_name.lower() if os_name else ''
    uid_str = str(uid) if uid else ''
    
    # Windows privilege detection
    if 'windows' in os_lower:
        # Check for well-known privileged accounts
        privileged_names = [
            'administrator',
            'system',
            'nt authority\\system',
            'admin',
        ]
        
        # Check username
        for priv_name in privileged_names:
            if priv_name in username_lower:
                return True
        
        # Check Windows SID (Security Identifier)
        # S-1-5-18 = SYSTEM
        # S-1-5-19 = LOCAL SERVICE
        # S-1-5-20 = NETWORK SERVICE
        # S-1-5-*-500 = Administrator (RID 500)
        if 'S-1-5-18' in uid_str:  # SYSTEM
            return True
        if uid_str.endswith('-500'):  # Administrator account (RID 500)
            return True
    
    # Linux/Unix privilege detection
    elif 'linux' in os_lower or 'unix' in os_lower or 'darwin' in os_lower:
        # Check for root username
        if username_lower == 'root':
            return True
        
        # Check UID 0 (root)
        if uid_str == '0':
            return True
    
    return False


def is_dead_or_late(is_dead, next_checkin, agent_type):
    """
    Check if agent is dead
    
    Args:
        is_dead: Boolean indicating if agent is marked as dead
        next_checkin: Next expected check-in time (not used for beacons due to jitter)
        agent_type: 'session' or 'beacon'
    
    Returns:
        'dead' if agent is marked as dead by Sliver
        'alive' if agent is active
        
    Note: We don't check 'late' for beacons because they use sleep intervals
    with jitter, so NextCheckin is not a reliable indicator of health.
    """
    if is_dead:
        return 'dead'
    
    return 'alive'


def draw_graph(agent_tree, total_sessions, total_beacons, last_update=None):
    """Draw the main network graph visualization with hierarchical tree"""
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
    total_hosts = total_sessions + total_beacons
    
    if total_hosts == 0:
        output.append(f"{Colors.RED}                    ⚠️  [No Active Hosts Connected]{Colors.ENDC}")
        output.append("")
        return '\n'.join(output)
    
    # Draw the C2 server logo on the left
    logo_lines = draw_sliver_logo()
    
    # Calculate starting position for logo (center it vertically)
    total_lines = sum(1 + len(agent.get('children', [])) for agent in agent_tree)
    logo_start = total_lines // 2 - len(logo_lines) // 2
    if logo_start < 0:
        logo_start = 0
    
    line_count = 0
    
    # Draw each root agent and its children
    for idx, agent in enumerate(agent_tree):
        # Draw root agent
        is_session = agent['type'] == 'session'
        host_id = agent['ID'][:8]
        hostname = agent.get('Hostname', 'Unknown')
        username = agent.get('Username', 'Unknown')
        os_name = agent.get('OS', 'Unknown')
        transport = agent.get('Transport', 'unknown')
        uid = agent.get('UID', '')
        
        # Check if privileged
        privileged = is_privileged(username, uid, os_name)
        
        # Check if agent is dead or late
        agent_status = is_dead_or_late(
            agent.get('IsDead', False),
            agent.get('NextCheckin', 0),
            agent['type']
        )
        
        # Determine colors based on status
        if agent_status == 'dead':
            host_color = Colors.GRAY
            username_color = Colors.GRAY
            protocol_color = Colors.GRAY
            status_marker = f" {Colors.RED}💀{Colors.ENDC}"
            type_label = "session [DEAD]" if is_session else "beacon [DEAD]"
        else:
            # Alive - normal colors
            if is_session:
                host_color = Colors.GREEN
                type_label = "session"
            else:
                host_color = Colors.YELLOW
                type_label = "beacon"
            
            protocol_color = get_protocol_color(transport)
            username_color = Colors.RED if privileged else Colors.CYAN
            status_marker = ""
        
        # Build the line for root agent
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
        
        # Status indicator - smaller diamond icons
        status_icon = "◆" if is_session else "◇"
        
        # Privilege indicator - gem badge for high-value targets
        priv_badge = f" {Colors.CYAN}💎{Colors.ENDC}" if privileged else ""
        
        # Check if this agent has children (is a pivot point)
        has_children = len(agent.get('children', [])) > 0
        tree_branch = "├─" if has_children else "──"
        
        # Format root agent line
        output.append(
            f"  {logo_line}      {tree_branch}─[ {protocol_color}{transport.upper():^6}{Colors.ENDC} ]───▶ "
            f"{host_color}{status_icon}{Colors.ENDC} {host_color}{pc_icon}{Colors.ENDC} "
            f"{Colors.BOLD}{username_color}{username}@{hostname}{Colors.ENDC}{priv_badge}{status_marker}  "
            f"{Colors.GRAY}{host_id} ({type_label}){Colors.ENDC}"
        )
        line_count += 1
        
        # Draw children (pivoted agents)
        children = agent.get('children', [])
        for child_idx, child in enumerate(children):
            is_last_child = child_idx == len(children) - 1
            
            child_is_session = child['type'] == 'session'
            child_id = child['ID'][:8]
            child_hostname = child.get('Hostname', 'Unknown')
            child_username = child.get('Username', 'Unknown')
            child_os = child.get('OS', 'Unknown')
            child_transport = child.get('Transport', 'unknown')
            child_uid = child.get('UID', '')
            
            # Check if child is privileged
            child_privileged = is_privileged(child_username, child_uid, child_os)
            
            # Check if child is dead or late
            child_status_check = is_dead_or_late(
                child.get('IsDead', False),
                child.get('NextCheckin', 0),
                child['type']
            )
            
            # Determine child colors based on status
            if child_status_check == 'dead':
                child_color = Colors.GRAY
                child_username_color = Colors.GRAY
                child_protocol_color = Colors.GRAY
                child_status_marker = f" {Colors.RED}💀{Colors.ENDC}"
                child_type = "session [DEAD]" if child_is_session else "beacon [DEAD]"
            else:
                # Alive - normal colors
                if child_is_session:
                    child_color = Colors.GREEN
                    child_type = "session"
                else:
                    child_color = Colors.YELLOW
                    child_type = "beacon"
                
                child_protocol_color = get_protocol_color(child_transport)
                child_username_color = Colors.RED if child_privileged else Colors.CYAN
                child_status_marker = ""
            
            # Logo line
            if line_count >= logo_start and line_count < logo_start + len(logo_lines):
                logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
            else:
                logo_line = " " * 12
            
            # Child emoji icons
            if 'windows' in child_os.lower():
                child_icon = "🖥️ " if child_is_session else "💻"
            elif 'linux' in child_os.lower():
                child_icon = "🐧" if child_is_session else "🖳 "
            else:
                child_icon = "💻" if child_is_session else "🖥️ "
            
            # Child status indicator - smaller diamond icons
            child_status = "◆" if child_is_session else "◇"
            
            # Child privilege indicator - gem badge for high-value targets
            child_priv_badge = f" {Colors.CYAN}💎{Colors.ENDC}" if child_privileged else ""
            
            # Tree branch for child
            child_branch = "└─" if is_last_child else "├─"
            
            # Format child line with indentation
            output.append(
                f"  {logo_line}      │  {child_branch}[ {child_protocol_color}{child_transport.upper():^6}{Colors.ENDC} ]──▶ "
                f"{child_color}{child_status}{Colors.ENDC} {child_color}{child_icon}{Colors.ENDC} "
                f"{Colors.BOLD}{child_username_color}{child_username}@{child_hostname}{Colors.ENDC}{child_priv_badge}{child_status_marker}  "
                f"{Colors.GRAY}{child_id} ({child_type}) {Colors.MAGENTA}↪ pivoted{Colors.ENDC}"
            )
            line_count += 1
        
        # Add spacing between root agents
        if idx < len(agent_tree) - 1:
            if line_count >= logo_start and line_count < logo_start + len(logo_lines):
                logo_line = f"{Colors.MAGENTA}{logo_lines[line_count - logo_start]:12}{Colors.ENDC}"
            else:
                logo_line = " " * 12
            output.append(f"  {logo_line}")
            line_count += 1
    
    output.append("")
    output.append(f"{Colors.GRAY}{'─' * 80}{Colors.ENDC}")
    output.append(f"  🟢 Active Sessions: {Colors.BOLD}{Colors.GREEN}{total_sessions}{Colors.ENDC}  🟡 Active Beacons: {Colors.BOLD}{Colors.YELLOW}{total_beacons}{Colors.ENDC}  🔵 Total Compromised: {Colors.BOLD}{Colors.CYAN}{total_hosts}{Colors.ENDC}")
    output.append("")
    
    return '\n'.join(output)


def build_agent_tree(sessions, beacons):
    """Build hierarchical tree structure from agents based on pivot relationships"""
    # Convert to dict format with pivot info
    all_agents = {}
    
    for s in sessions:
        agent_id = s.ID
        all_agents[agent_id] = {
            'ID': s.ID,
            'Hostname': s.Hostname,
            'Username': s.Username,
            'OS': s.OS,
            'Transport': s.Transport,
            'ProxyURL': s.ProxyURL,
            'PeerID': s.PeerID,
            'RemoteAddress': s.RemoteAddress,
            'UID': s.UID,
            'IsDead': s.IsDead,
            'LastCheckin': getattr(s, 'LastCheckin', 0),
            'NextCheckin': 0,  # Sessions don't have NextCheckin
            'type': 'session',
            'children': []
        }
    
    for b in beacons:
        agent_id = b.ID
        all_agents[agent_id] = {
            'ID': b.ID,
            'Hostname': b.Hostname,
            'Username': b.Username,
            'OS': b.OS,
            'Transport': b.Transport,
            'ProxyURL': b.ProxyURL,
            'RemoteAddress': b.RemoteAddress,
            'UID': b.UID,
            'IsDead': b.IsDead,
            'LastCheckin': b.LastCheckin,
            'NextCheckin': b.NextCheckin,
            'type': 'beacon',
            'children': []
        }
    
    # Build tree by detecting parent-child relationships
    # Agents with ProxyURL or pivot transports are children
    root_agents = []
    pivoted_agents = []
    
    for agent_id, agent in all_agents.items():
        transport_lower = agent['Transport'].lower()
        
        # Check if agent is pivoted
        is_pivoted = (
            agent['ProxyURL'] or  # Has proxy configured
            'pivot' in transport_lower or  # TCP pivot, etc.
            'namedpipe' in transport_lower or  # Windows named pipe
            'bind' in transport_lower  # Bind connection
        )
        
        if is_pivoted:
            pivoted_agents.append(agent)
        else:
            root_agents.append(agent)
    
    # Try to match pivoted agents to their parents
    for pivoted in pivoted_agents:
        parent_found = False
        
        # If ProxyURL is set, try to extract parent info
        if pivoted['ProxyURL']:
            # ProxyURL format might be: tcp://parentID or similar
            proxy_url = pivoted['ProxyURL']
            
            # Try to find parent by matching hostname or ID in proxy URL
            for root in root_agents:
                if root['ID'] in proxy_url or root['Hostname'] in proxy_url:
                    root['children'].append(pivoted)
                    parent_found = True
                    break
        
        # If transport indicates pivot type, try to match by network
        if not parent_found:
            transport = pivoted['Transport'].lower()
            
            # For named pipes or TCP pivots, try to match by hostname
            if 'namedpipe' in transport or 'pivot' in transport:
                for root in root_agents:
                    # Match if same hostname (likely pivoting through that host)
                    if root['Hostname'] == pivoted['Hostname']:
                        root['children'].append(pivoted)
                        parent_found = True
                        break
        
        # If still no parent found, add to root (fallback)
        if not parent_found:
            root_agents.append(pivoted)
    
    return root_agents


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
            
            # Build hierarchical tree structure
            agent_tree = build_agent_tree(sessions, beacons)
            
            # Clear screen and draw the graph
            clear_screen()
            graph = draw_graph(agent_tree, len(sessions), len(beacons), time.time())
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
            agent_tree = build_agent_tree(sessions, beacons)
            
            clear_screen()
            graph = draw_graph(agent_tree, len(sessions), len(beacons), time.time())
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
