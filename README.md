# Sliver C2 Network Topology Visualizer

A terminal-based ASCII network topology visualization tool for Sliver C2 that displays compromised hosts in a network graph style, similar to Havoc and Cobalt Strike visualizations.

## Features

- ASCII art network topology visualization of all compromised hosts
- C2 server logo on the left with connection lines to hosts
- Color-coded display:
  - **Green** for active sessions
  - **Yellow** for active beacons
  - Protocol-specific colors (HTTP/HTTPS green, mTLS cyan, DNS yellow, TCP blue, SMB white)
- Shows key host information:
  - Session/Beacon ID
  - Hostname and Username
  - Operating System with ASCII computer icons
  - Transport protocol with connection lines
- Real-time statistics footer
- Clean network-style topology layout

## Requirements

- Python 3.6+
- Sliver C2 server running
- Sliver client configured (have run `sliver-client` at least once)
- Terminal with ANSI color support

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

Make sure your Sliver C2 server is running, then:

```bash
./graph
```

Or manually:

```bash
python3 sliver-graph.py
```

The tool will:
- Automatically find your Sliver client configuration
- Connect to the Sliver RPC server
- Fetch all active sessions and beacons
- Display the network topology visualization

## Example Output

```
╔══════════════════════════════════════════════════════════════════════════════╗
║            SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION                        ║
╚══════════════════════════════════════════════════════════════════════════════╝

                ·········╲
                 ────────── MTLS (p0) ──────────▶
                ·········╱               ┌─────┐
                                        │ ▄▄  │  Administrator@m3dc
                                        │ ▀▀  │  d0189a61 (session)
   ▄████▄                               └─────┘
  ████████                              ┌─────┐
  ▀██████▀                              └─────┘
    ▀██▀    
     ▀        ·········╲
                 ────────── HTTPS (p0) ──────────▶
                ·········╱               ┌─────┐
                                        │ ▄▄  │  admin@webserver
                                        │ ▀▀  │  4370d26a (beacon)
                                        └─────┘
                                        ┌─────┐
                                        └─────┘

────────────────────────────────────────────────────────────────────────────────
  ● Active Sessions: 1  ● Active Beacons: 1  ● Total: 2
```

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
