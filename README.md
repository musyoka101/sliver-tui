# Sliver C2 Network Topology Visualizer

A beautiful terminal-based network topology visualization tool for Sliver C2 that displays compromised hosts with hierarchical pivot chains, privilege detection, and real-time statistics.

## 📁 Project Structure

This project has **two implementations**:

```
sliver-graphs/
├── python/          # Python implementation (mature, feature-complete)
│   ├── sliver-graph.py
│   ├── graph        # Launcher script
│   └── .venv/       # Virtual environment
│
└── go/              # Go + Bubble Tea implementation (modern, TUI)
    ├── main.go
    ├── go.mod
    └── README.md
```

## 🐍 Python Version

**Location:** `python/`  
**Status:** ✅ Production-ready, fully featured

### Features
- Real-time agent monitoring with auto-refresh
- Hierarchical topology visualization with C2 logo
- Multi-line agent display (username@host, ID, IP)
- Privilege detection (💎 badges for Administrator/root)
- Change detection (NEW badges, lost agents tracking)
- Dead beacon detection
- Protocol-specific colors (MTLS, HTTP, DNS, TCP)
- Comprehensive statistics dashboard
- OS-specific icons (🖥️ 💻 🐧)

### Installation
```bash
cd python
python3 -m venv .venv
source .venv/bin/activate
pip install sliver-py
```

### Usage
```bash
cd python
source .venv/bin/activate

# Run with auto-refresh (5 seconds)
python3 sliver-graph.py

# Custom refresh interval
python3 sliver-graph.py -r 10

# Run once (no loop)
python3 sliver-graph.py --once

# Or use the launcher script
./graph
```

### Git Branches (Python)
- `master` - Stable, basic features
- `dev` - Production-ready with all features
- `experimental` - Testing new features

## 🚀 Go Version

**Location:** `go/`  
**Status:** 🚧 Work in progress, beautiful TUI

### Features
- Built with Bubble Tea (professional TUI framework)
- Beautiful styling with Lip Gloss
- C2 logo on left side
- Single-line compact agent display
- Interactive keyboard controls (r=refresh, q=quit)
- Compiled binary (no dependencies)

### Installation
```bash
cd go
go mod download
go build -o sliver-graph main.go
```

### Usage
```bash
cd go
./sliver-graph

# Keyboard shortcuts:
# r - Manual refresh
# q - Quit
```

### Git Branches (Go)
- `go-bubbletea` - Go implementation branch

## 🎯 Features Comparison

| Feature | Python | Go |
|---------|--------|-----|
| C2 Logo | ✅ | ✅ |
| Agent Display | Multi-line (detailed) | Single-line (compact) |
| Auto-refresh | ✅ | ✅ |
| Privilege Detection | ✅ | ✅ |
| Change Detection | ✅ | ⏳ TODO |
| Lost Agents | ✅ | ⏳ TODO |
| Pivot Hierarchy | ✅ | ⏳ TODO |
| Real Sliver Connection | ✅ | ⏳ TODO (mock data) |
| Binary Distribution | ❌ (needs Python) | ✅ (single binary) |
| Startup Time | Slow (~1-2s) | Instant (<100ms) |
| Interactive TUI | ❌ | ✅ (Bubble Tea) |

## 🎨 Screenshots

### Python Version
```
╔════════════════════════════════════════════════════════════╗
║  🎯 SLIVER C2 - NETWORK TOPOLOGY VISUALIZATION            ║
╚════════════════════════════════════════════════════════════╝

                    ╰────────[ MTLS ]────────▶ ◆ 🖥️  NT AUTHORITY\NETWORK SERVICE@cywebdw
                                                     └─ ID: 22bf4a82 (session) ✨ NEW!
  🎯 C2                                              └─ IP: 10.10.110.10:50199
 ▄████▄         ╰────────[ MTLS ]────────▶ ◇ 💻 M3C\Administrator@m3dc 💎
 ████████                                             └─ ID: 4370d26a (beacon)
 ▀██████▀                                             └─ IP: 10.10.110.250:63805
   ▀██▀    

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🟢 Sessions: 1  🟡 Beacons: 3  🔵 Hosts: 3  🔴 Privileged: 2
```

### Go Version
```
╭────────────────────────────────────╮
│ 🎯 SLIVER C2 - NETWORK TOPOLOGY   │
╰────────────────────────────────────╯

 🎯 C2       ——[ MTLS ]——▶ ◆ 🖥️  NT AUTHORITY\NETWORK SERVICE@cywebdw  22bf4a82 (session)
▄████▄       ——[ MTLS ]——▶ ◇ 💻  M3C\Administrator@m3dc 💎  4370d26a (beacon)
████████  
▀██████▀  
  ▀██▀    

────────────────────────────────────────────────────────────
🟢 Active Sessions: 1  🟡 Active Beacons: 3  🔵 Total Compromised: 4
```

## 🛠️ Requirements

### Python Version
- Python 3.6+
- sliver-py
- Sliver C2 client configured (~/.sliver-client/configs/*.cfg)

### Go Version
- Go 1.21+
- Bubble Tea, Lip Gloss, Bubbles (auto-installed)
- Sliver C2 client configured

## 🤝 Contributing

Both implementations are actively developed:
- **Python** - Add features to `dev` or `experimental` branches
- **Go** - Work on `go-bubbletea` branch

## 📝 License

MIT

## 🎓 Author

musyoka101 (ianmusyoka101@gmail.com)

## 🔗 Links

- [Sliver C2](https://github.com/BishopFox/sliver)
- [sliver-py](https://github.com/moloch--/sliver-py)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
