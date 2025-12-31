# Sliver C2 Network Topology Visualizer - Python Edition

A lightweight Python script to visualize Sliver C2 network topology in the terminal.

## Features

- ASCII network topology visualization
- Real-time agent tracking
- Change detection with NEW badges
- Dead beacon detection
- Lost agents tracking
- Automatic refresh every 5 seconds
- Color-coded status indicators

## Requirements

- Python 3.7+
- sliver-py library
- Sliver C2 server running
- Sliver client configured (`~/.sliver-client/configs/*.cfg`)

## Installation

```bash
# Install dependencies
pip3 install sliver-py

# Make executable
chmod +x graph

# Run
./graph
```

Or run directly:
```bash
python3 sliver-graph.py
```

## Usage

```bash
# Start the visualizer
./graph

# Press Ctrl+C to exit
```

## Keyboard Controls

- `Ctrl+C` - Quit

## Output Example

```
═══════════════════════════════════════════════════════
    🔥 SLIVER C2 NETWORK TOPOLOGY 🔥
═══════════════════════════════════════════════════════
Last Updated: 2024-12-31 10:30:45
Active Sessions: 2 | Active Beacons: 3

🖥️  C2 Server: mtls://10.10.14.15:8443
    │
    ├─ [SESSION] root@webserver ✨ NEW!
    │  └─ ID: a1b2c3d4 | IP: 192.168.1.100
    │
    └─ [BEACON] admin@workstation
       └─ ID: e5f6g7h8 | IP: 192.168.1.101
```

## Comparison with Go Version

| Feature | Python | Go |
|---------|--------|-----|
| Speed | Moderate | Fast |
| Memory | ~50MB | ~10MB |
| Setup | pip install | Single binary |
| Dashboard | ❌ | ✅ |
| Themes | 1 | 5 |
| Views | 1 | 5 |

For a more feature-rich experience with dashboards, analytics, and multiple views, see the [Go version](../go/README.md).

## License

MIT License - See [LICENSE](../go/LICENSE)

## Credits

- Powered by [Sliver](https://github.com/BishopFox/sliver)
- Uses [sliver-py](https://github.com/moloch--/sliver-py)
