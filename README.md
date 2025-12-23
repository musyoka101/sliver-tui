# Sliver C2 Network Topology Visualizer

A terminal-based emoji-enhanced network topology visualization tool for Sliver C2 that displays compromised hosts in a hierarchical tree structure, similar to Havoc and Cobalt Strike visualizations.

## Features

- 🎯 **Hierarchical tree visualization** showing pivot relationships
- 🟢 **Color-coded agents**:
  - Green for active sessions (real-time)
  - Yellow for active beacons (callback-based)
- 🖥️  **OS-specific emojis**:
  - 🖥️  Desktop for Windows sessions
  - 💻 Laptop for Windows beacons
  - 🐧 Penguin for Linux systems
- 🔄 **Live monitoring** with auto-refresh (configurable interval)
- 🌳 **Pivot chain detection**:
  - Automatically detects agents connected through other agents
  - Shows parent-child relationships in tree format
  - Highlights pivoted connections with ↪ indicator
- 📊 **Real-time statistics** footer
- 🎨 **Protocol-specific colors**:
  - Cyan for mTLS/TLS
  - Green for HTTP/HTTPS
  - Yellow for DNS
  - Blue for TCP
  - White for SMB

## Requirements

- Python 3.6+
- Sliver C2 server running
- Sliver client configured (have run `sliver-client` at least once)
- Terminal with ANSI color and emoji support

## Installation

### Option 1: Quick Start (Using included launcher)

Simply run the launcher script - it will automatically set up everything:

```bash
./graph
```

The launcher will:
1. Create a Python virtual environment if it doesn't exist
2. Install `sliver-py` dependency
3. Run the visualization tool

### Option 2: Manual Installation

1. Create a virtual environment (optional but recommended):
```bash
python3 -m venv .venv
source .venv/bin/activate
```

2. Install the required Python library:
```bash
pip3 install sliver-py
```

3. Run the visualization tool:
```bash
python3 sliver-graph.py
```

## Usage

### Live Monitoring (Default)

The tool runs in live monitoring mode by default, refreshing every 5 seconds:

```bash
./graph
# or
python3 sliver-graph.py
```

Press `Ctrl+C` to exit.

### Custom Refresh Interval

Change the refresh interval (in seconds):

```bash
./graph -r 10           # Refresh every 10 seconds
python3 sliver-graph.py --refresh 10
```

### Run Once (No Loop)

Run once and exit without live monitoring:

```bash
./graph --once
python3 sliver-graph.py --once
```

## Example Output

### Direct Connections (No Pivots)
```
╔════════════════════════════════════════════════════════════════════════════╗
║  🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION                            ║
╚════════════════════════════════════════════════════════════════════════════╝
  ⏰ Last Update: 2025-12-23 17:50:51  |  Press Ctrl+C to exit

   🎯 C2         ───[ MTLS ]───▶ 🟢 🖥️  M3C\Administrator@m3dc  d0189a61 (session)
  ▄████▄   
  ████████      ───[ MTLS ]───▶ 🟡 💻 M3C\Administrator@m3dc  4370d26a (beacon)
  ▀██████▀  
    ▀██▀        ───[HTTPS]───▶ 🟡 💻 admin@webserver  b773f522 (beacon)

────────────────────────────────────────────────────────────────────────────────
  🟢 Active Sessions: 1  🟡 Active Beacons: 2  🔵 Total Compromised: 3
```

### With Pivot Chains
```
╔════════════════════════════════════════════════════════════════════════════╗
║  🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION                            ║
╚════════════════════════════════════════════════════════════════════════════╝
  ⏰ Last Update: 2025-12-23 17:50:51  |  Press Ctrl+C to exit

   🎯 C2         ├──[ MTLS ]───▶ 🟢 🖥️  Administrator@WebServer  a1b2c3d4 (session)
  ▄████▄        │  └─[ SMB ]──▶ 🟡 💻 SYSTEM@Database  e5f6g7h8 (beacon) ↪ pivoted
  ████████      │  ├─[ TCP ]──▶ 🟡 💻 User@Workstation  i9j0k1l2 (beacon) ↪ pivoted
  ▀██████▀      │
    ▀██▀        ├──[HTTPS]───▶ 🟢 🐧 root@linux-server  m3n4o5p6 (session)
                │  └─[NAMEDPIPE]▶ 🟡 🖥️  Admin@DC  q7r8s9t0 (beacon) ↪ pivoted

────────────────────────────────────────────────────────────────────────────────
  🟢 Active Sessions: 2  🟡 Active Beacons: 3  🔵 Total Compromised: 5
```

## Pivot Detection

The tool automatically detects pivot relationships based on:

1. **ProxyURL** - When an agent has a proxy URL set, it's routing through another agent
2. **Transport types**:
   - `namedpipe` - Windows named pipe pivoting
   - `tcp-pivot` - TCP pivot connections
   - `bind` - Bind connections
3. **Network matching** - Agents on the same host may indicate pivoting

Pivoted agents are displayed indented under their parent with:
- Tree branch indicators (`├─`, `└─`)
- ↪ pivoted label in magenta
- Indented connection lines showing hierarchy

## Troubleshooting

### "sliver-py not found"
Install it with:
```bash
pip3 install sliver-py
```

### "Sliver client config not found"
Make sure you've run the Sliver client at least once:
```bash
sliver-client
```

This will create the config file in `~/.sliver-client/configs/`

### "'coroutine' object is not iterable"
This has been fixed in the latest version. Make sure you're using the updated `sliver-graph.py` script with async support.

### Colors don't display correctly
- Ensure your terminal supports ANSI colors
- Try using a different terminal emulator (e.g., Windows Terminal, iTerm2, gnome-terminal, terminator)

## Files

- `sliver-graph.py` - Main visualization script
- `graph` - Launcher script with auto-setup
- `README.md` - This file
- `.venv/` - Python virtual environment (created automatically)

## How It Works

This tool is **not a Sliver extension** (which would require compiled binaries), but rather a standalone Python tool that:
1. Reads your Sliver client configuration from `~/.sliver-client/configs/`
2. Connects to Sliver's gRPC API using the `sliver-py` library
3. Fetches all active sessions and beacons
4. Renders them in an ASCII network topology visualization

## Why Not a Sliver Extension?

Sliver extensions must be compiled binaries (DLL/EXE/SO files) written in languages like C/C++/Go. While this is powerful for executing code on implants, it's unnecessarily complex for a simple visualization tool that runs on the operator's machine. This standalone approach is:
- Easier to develop and maintain
- Cross-platform without compilation
- Simple to install and use
- Achieves the same visualization goal

## Author

Cybernetics Team

## License

Use at your own risk for authorized testing only.
