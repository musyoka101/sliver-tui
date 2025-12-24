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


# Global state for change detection
PREVIOUS_AGENTS = {}  # Store agent IDs from last refresh
AGENT_FIRST_SEEN = {}  # Track when each agent was first detected
LOST_AGENTS = {}  # Track recently lost agents with timestamp
NEW_AGENT_TIMEOUT = 300  # Mark as "new" for 5 minutes (300 seconds)
LOST_AGENT_DISPLAY_TIME = 300  # Show lost agents for 5 minutes


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


def track_agent_changes(agent_tree):
    """
    Track changes in agents between refreshes
    
    Args:
        agent_tree: List of root agents with their children
    
    Returns:
        dict with 'new_count', 'lost_count' keys
    """
    global PREVIOUS_AGENTS, AGENT_FIRST_SEEN, LOST_AGENTS
    
    current_time = time.time()
    
    # Flatten agent tree to get all agent IDs
    current_agents = {}
    for agent in agent_tree:
        current_agents[agent['ID']] = agent
        for child in agent.get('children', []):
            current_agents[child['ID']] = child
    
    current_ids = set(current_agents.keys())
    previous_ids = set(PREVIOUS_AGENTS.keys())
    
    # Detect new agents
    new_ids = current_ids - previous_ids
    for agent_id in new_ids:
        if agent_id not in AGENT_FIRST_SEEN:
            AGENT_FIRST_SEEN[agent_id] = current_time
    
    # Detect lost agents
    lost_ids = previous_ids - current_ids
    for agent_id in lost_ids:
        if agent_id in PREVIOUS_AGENTS:
            LOST_AGENTS[agent_id] = {
                'agent': PREVIOUS_AGENTS[agent_id],
                'lost_time': current_time
            }
    
    # Clean up old lost agents (older than display time)
    expired_lost = []
    for agent_id, data in LOST_AGENTS.items():
        if current_time - data['lost_time'] > LOST_AGENT_DISPLAY_TIME:
            expired_lost.append(agent_id)
    
    for agent_id in expired_lost:
        del LOST_AGENTS[agent_id]
    
    # Update previous state
    PREVIOUS_AGENTS = current_agents.copy()
    
    return {
        'new_count': len(new_ids),
        'lost_count': len(lost_ids)
    }


def is_agent_new(agent_id):
    """
    Check if agent was seen recently (within NEW_AGENT_TIMEOUT)
    
    Args:
        agent_id: Agent ID string
    
    Returns:
        Boolean indicating if agent is new
    """
    if agent_id not in AGENT_FIRST_SEEN:
        return False
    
    time_since_first_seen = time.time() - AGENT_FIRST_SEEN[agent_id]
    return time_since_first_seen < NEW_AGENT_TIMEOUT


def count_unique_hosts(agent_tree):
    """
    Count unique compromised hosts based on hostname
    
    Args:
        agent_tree: List of root agents with their children
    
    Returns:
        Integer count of unique hostnames
    """
    unique_hostnames = set()
    
    for agent in agent_tree:
        # Add root agent hostname
        hostname = agent.get('Hostname', '').lower()
        if hostname:
            unique_hostnames.add(hostname)
        
        # Add children hostnames
        for child in agent.get('children', []):
            child_hostname = child.get('Hostname', '').lower()
            if child_hostname:
                unique_hostnames.add(child_hostname)
    
    return len(unique_hostnames)


def calculate_stats(agent_tree):
    """
    Calculate detailed statistics from agent tree
    
    Returns:
        dict with various statistics
    """
    stats = {
        'total_agents': 0,
        'privileged': 0,
        'unprivileged': 0,
        'windows': 0,
        'linux': 0,
        'other_os': 0,
        'protocols': {},
        'new_agents': 0,
        'dead_agents': 0,
    }
    
    def count_agent(agent):
        stats['total_agents'] += 1
        
        # Count privileged
        username = agent.get('Username', '')
        uid = agent.get('UID', '')
        os_name = agent.get('OS', '')
        
        if is_privileged(username, uid, os_name):
            stats['privileged'] += 1
        else:
            stats['unprivileged'] += 1
        
        # Count OS types
        os_lower = os_name.lower()
        if 'windows' in os_lower:
            stats['windows'] += 1
        elif 'linux' in os_lower:
            stats['linux'] += 1
        else:
            stats['other_os'] += 1
        
        # Count protocols
        transport = agent.get('Transport', 'unknown').upper()
        stats['protocols'][transport] = stats['protocols'].get(transport, 0) + 1
        
        # Count new agents
        if is_agent_new(agent['ID']):
            stats['new_agents'] += 1
        
        # Count dead agents
        if is_dead_or_late(agent.get('IsDead', False), 0, agent['type']) == 'dead':
            stats['dead_agents'] += 1
    
    # Process all agents in tree
    for agent in agent_tree:
        count_agent(agent)
        for child in agent.get('children', []):
            count_agent(child)
    
    return stats


def draw_graph(agent_tree, total_sessions, total_beacons, last_update=None, changes=None):
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
    
    # Show change summary at the top
    if changes:
        if changes['new_count'] > 0:
            output.append(f"{Colors.GREEN}  ✨ {changes['new_count']} new agent(s) detected{Colors.ENDC}")
        if changes['lost_count'] > 0:
            output.append(f"{Colors.RED}  🔴 {changes['lost_count']} agent(s) lost connection{Colors.ENDC}")
    
    output.append("")
    
    # Calculate totals - use unique host count for compromised hosts
    total_agents = total_sessions + total_beacons
    unique_hosts = count_unique_hosts(agent_tree)
    
    if total_agents == 0:
        output.append(f"{Colors.RED}                    ⚠️  [No Active Hosts Connected]{Colors.ENDC}")
        output.append("")
        return '\n'.join(output)
    
    # Draw the C2 server logo
    logo_lines = draw_sliver_logo()
    
    # Calculate total lines needed for content (with spacing)
    content_lines = []
    for idx, agent in enumerate(agent_tree):
        # Root agent: 3 lines (username@host, ID, IP)
        content_lines.append(3)
        
        # Children with vertical connectors
        children = agent.get('children', [])
        if children:
            for child_idx, child in enumerate(children):
                content_lines.append(1)  # Vertical line
                content_lines.append(3)  # Child agent: 3 lines (username@host, ID, IP)
        
        # Spacing between root agents (except last)
        if idx < len(agent_tree) - 1:
            content_lines.append(2)  # Double line spacing
    
    total_content_lines = sum(content_lines)
    logo_start = total_content_lines // 2 - len(logo_lines) // 2
    if logo_start < 0:
        logo_start = 0
    
    line_count = 0
    
    # Draw each root agent and its children
    for idx, agent in enumerate(agent_tree):
        # Extract agent info
        is_session = agent['type'] == 'session'
        host_id = agent['ID'][:8]
        hostname = agent.get('Hostname', 'Unknown')
        username = agent.get('Username', 'Unknown')
        os_name = agent.get('OS', 'Unknown')
        transport = agent.get('Transport', 'unknown')
        uid = agent.get('UID', '')
        
        # Check if privileged
        privileged = is_privileged(username, uid, os_name)
        
        # Check if agent is dead
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
        
        # Get logo line or empty space
        def get_logo_line(line_num):
            if line_num >= logo_start and line_num < logo_start + len(logo_lines):
                return f"{Colors.MAGENTA}{logo_lines[line_num - logo_start]:12}{Colors.ENDC}"
            return " " * 12
        
        # Emoji icons based on OS and type
        if 'windows' in os_name.lower():
            pc_icon = "🖥️ " if is_session else "💻"
        elif 'linux' in os_name.lower():
            pc_icon = "🐧" if is_session else "🖳 "
        else:
            pc_icon = "💻" if is_session else "🖥️ "
        
        # Status indicator
        status_icon = "◆" if is_session else "◇"
        
        # Privilege indicator
        priv_badge = f" {Colors.CYAN}💎{Colors.ENDC}" if privileged else ""
        
        # NEW badge for recently seen agents
        new_badge = ""
        if is_agent_new(agent['ID']):
            new_badge = f" {Colors.GREEN}✨ NEW!{Colors.ENDC}"
        
        # Format root agent line with longer connectors (multi-line display)
        remote_ip = agent.get('RemoteAddress', 'Unknown')
        
        # Line 1: Protocol connector + icon + username@hostname
        logo_line = get_logo_line(line_count)
        output.append(
            f"  {logo_line}      ╰────────[ {protocol_color}{transport.upper():^6}{Colors.ENDC} ]────────▶ "
            f"{host_color}{status_icon}{Colors.ENDC} {host_color}{pc_icon}{Colors.ENDC} "
            f"{Colors.BOLD}{username_color}{username}@{hostname}{Colors.ENDC}{priv_badge}{status_marker}"
        )
        line_count += 1
        
        # Line 2: ID and type with NEW badge
        logo_line = get_logo_line(line_count)
        output.append(
            f"  {logo_line}                                               "
            f"└─ ID: {Colors.GRAY}{host_id} ({type_label}){Colors.ENDC}{new_badge}"
        )
        line_count += 1
        
        # Line 3: IP address
        logo_line = get_logo_line(line_count)
        output.append(
            f"  {logo_line}                                               "
            f"└─ IP: {Colors.CYAN}{remote_ip}{Colors.ENDC}"
        )
        line_count += 1
        
        # Draw children (pivoted agents)
        children = agent.get('children', [])
        for child_idx, child in enumerate(children):
            is_last_child = child_idx == len(children) - 1
            
            # Add vertical connector line
            logo_line = get_logo_line(line_count)
            output.append(f"  {logo_line}           │")
            line_count += 1
            
            # Extract child info
            child_is_session = child['type'] == 'session'
            child_id = child['ID'][:8]
            child_hostname = child.get('Hostname', 'Unknown')
            child_username = child.get('Username', 'Unknown')
            child_os = child.get('OS', 'Unknown')
            child_transport = child.get('Transport', 'unknown')
            child_uid = child.get('UID', '')
            
            # Check if child is privileged
            child_privileged = is_privileged(child_username, child_uid, child_os)
            
            # Check if child is dead
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
            
            # Child emoji icons
            if 'windows' in child_os.lower():
                child_icon = "🖥️ " if child_is_session else "💻"
            elif 'linux' in child_os.lower():
                child_icon = "🐧" if child_is_session else "🖳 "
            else:
                child_icon = "💻" if child_is_session else "🖥️ "
            
            # Child status indicator
            child_status = "◆" if child_is_session else "◇"
            
            # Child privilege indicator
            child_priv_badge = f" {Colors.CYAN}💎{Colors.ENDC}" if child_privileged else ""
            
            # NEW badge for recently seen child agents
            child_new_badge = ""
            if is_agent_new(child['ID']):
                child_new_badge = f" {Colors.GREEN}✨ NEW!{Colors.ENDC}"
            
            # Tree branch for child
            child_branch = "╰─" if is_last_child else "├─"
            
            # Format child line with indentation (multi-line display)
            child_remote_ip = child.get('RemoteAddress', 'Unknown')
            
            # Line 1: Protocol connector + icon + username@hostname
            logo_line = get_logo_line(line_count)
            output.append(
                f"  {logo_line}           {child_branch}───────[ {child_protocol_color}{child_transport.upper():^6}{Colors.ENDC} ]────────▶ "
                f"{child_color}{child_status}{Colors.ENDC} {child_color}{child_icon}{Colors.ENDC} "
                f"{Colors.BOLD}{child_username_color}{child_username}@{child_hostname}{Colors.ENDC}{child_priv_badge}{child_status_marker}"
            )
            line_count += 1
            
            # Line 2: ID and type with NEW badge
            logo_line = get_logo_line(line_count)
            continuation = "│" if not is_last_child else " "
            output.append(
                f"  {logo_line}           {continuation}                                      "
                f"└─ ID: {Colors.GRAY}{child_id} ({child_type}){Colors.ENDC}{child_new_badge}"
            )
            line_count += 1
            
            # Line 3: IP address
            logo_line = get_logo_line(line_count)
            output.append(
                f"  {logo_line}           {continuation}                                      "
                f"└─ IP: {Colors.CYAN}{child_remote_ip}{Colors.ENDC}"
            )
            line_count += 1
        
        # Add spacing between root agents (except last)
        if idx < len(agent_tree) - 1:
            logo_line = get_logo_line(line_count)
            output.append(f"  {logo_line}")
            line_count += 1
            logo_line = get_logo_line(line_count)
            output.append(f"  {logo_line}")
            line_count += 1
    
    # Calculate detailed statistics
    stats = calculate_stats(agent_tree)
    
    output.append("")
    output.append(f"{Colors.GRAY}{'━' * 80}{Colors.ENDC}")
    
    # Compact single-line stats footer
    stats_line = f"🟢 Sessions: {Colors.BOLD}{Colors.GREEN}{total_sessions}{Colors.ENDC}  "
    stats_line += f"🟡 Beacons: {Colors.BOLD}{Colors.YELLOW}{total_beacons}{Colors.ENDC}  "
    stats_line += f"🔵 Hosts: {Colors.BOLD}{Colors.CYAN}{unique_hosts}{Colors.ENDC}  "
    
    if stats['new_agents'] > 0:
        stats_line += f"✨ New: {Colors.BOLD}{Colors.GREEN}{stats['new_agents']}{Colors.ENDC}  "
    
    stats_line += f"🔴 Privileged: {Colors.BOLD}{Colors.RED}{stats['privileged']}{Colors.ENDC}  "
    stats_line += f"🟢 Standard: {Colors.BOLD}{Colors.GREEN}{stats['unprivileged']}{Colors.ENDC}  "
    
    # OS breakdown
    # OS breakdown
    os_parts = []
    if stats['windows'] > 0:
        os_parts.append(f"Windows({Colors.BOLD}{stats['windows']}{Colors.ENDC})")
    if stats['linux'] > 0:
        os_parts.append(f"Linux({Colors.BOLD}{stats['linux']}{Colors.ENDC})")
    if stats['other_os'] > 0:
        os_parts.append(f"Other({Colors.BOLD}{stats['other_os']}{Colors.ENDC})")
    
    if os_parts:
        stats_line += f"💻 OS: " + " ".join(os_parts) + "  "
    
    stats_line += f"  🔗 Protocols: "
    
    # Protocol breakdown
    proto_parts = []
    for proto, count in sorted(stats['protocols'].items()):
        proto_color = get_protocol_color(proto.lower())
        proto_parts.append(f"{proto_color}{proto}{Colors.ENDC}({Colors.BOLD}{count}{Colors.ENDC})")
    stats_line += " ".join(proto_parts)
    
    output.append(stats_line)
    
    # Display recently lost agents
    if LOST_AGENTS:
        output.append("")
        output.append(f"{Colors.RED}🔴 Recently Lost Connections:{Colors.ENDC}")
        for agent_id, data in sorted(LOST_AGENTS.items(), key=lambda x: x[1]['lost_time'], reverse=True):
            agent = data['agent']
            lost_time = data['lost_time']
            time_ago = int(time.time() - lost_time)
            
            # Format time ago
            if time_ago < 60:
                time_str = f"{time_ago}s ago"
            else:
                time_str = f"{time_ago // 60}m ago"
            
            username = agent.get('Username', 'Unknown')
            hostname = agent.get('Hostname', 'Unknown')
            agent_type = agent.get('type', 'unknown')
            
            output.append(f"  {Colors.GRAY}◇ {username}@{hostname}  {agent_id[:8]} ({agent_type}) - {time_str}{Colors.ENDC}")
    
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
            
            # Track changes in agents
            changes = track_agent_changes(agent_tree)
            
            # Clear screen and draw the graph
            clear_screen()
            graph = draw_graph(agent_tree, len(sessions), len(beacons), time.time(), changes)
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
            
            # Track changes (even in once mode, to initialize state)
            changes = track_agent_changes(agent_tree)
            
            clear_screen()
            graph = draw_graph(agent_tree, len(sessions), len(beacons), time.time(), changes)
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
