# Sliver C2 Network Topology Visualizer v1.0.0

A modern, high-performance TUI (Terminal User Interface) for visualizing and monitoring Sliver C2 infrastructure built with Go and Bubble Tea.

## 🎯 What's New in v1.0.0

This is the first stable release of the Go/Bubble Tea implementation!

### ✨ Core Features

**Beautiful Terminal Interface**
- Professional TUI built with Bubble Tea framework
- Multiple color themes (Dracula, Monokai, Nord, Solarized, Cyberpunk)
- Smooth animations and loading indicators
- Real-time auto-refresh every 5 seconds

**Multiple View Modes**
- 🌲 **Tree View** - Hierarchical topology with pivot relationships
- 📦 **Box View** - Compact display with side connectors
- 📊 **Table View** - Professional tabular layout
- 📈 **Dashboard** - Advanced analytics with 5 panels
- 🗺️ **Network Map** - Subnet-based topology visualization

**Advanced Dashboard**
- C2 Infrastructure mapping with agent counts
- Architecture distribution charts (x64, x86, arm64)
- Beacon task queue monitoring with progress tracking
- Security status panel (STEALTH/BURNED agents)
- 12-hour activity tracking with sparkline graphs
- Real-time alert system with tactical notifications

**Real-Time Alert System**
- Military-style notification panel
- Color-coded severity levels (Critical, Warning, Success, Info)
- Event tracking: agent acquisition, privilege escalation, tasks, disconnections
- Smart deduplication and auto-expiration
- Column-aligned display for easy scanning

**Agent Intelligence**
- NEW agent badges (✨) for recently connected agents
- Dead beacon detection (💀)
- Privilege detection (💎)
- Lost agents tracking (5-minute window)
- Security status (🕵️ STEALTH, 🔥 BURNED)
- OS-specific icons with Nerd Font support
- Protocol color coding (MTLS, HTTP, DNS, TCP)

**Performance**
- ⚡ Single compiled binary (~19MB)
- 🚀 Instant startup (<100ms)
- 💾 Low memory footprint (~10MB)
- 🔄 Auto-refresh with change detection
- 📦 No external dependencies

## 🎮 Keyboard Controls

| Key | Action |
|-----|--------|
| `r` | Manual refresh |
| `v` | Cycle through views |
| `d` | Quick dashboard access |
| `t` | Change theme |
| `i` | Toggle icon style (Nerd Font/Emoji) |
| `Tab` / `F1-F5` | Direct view access |
| `↑↓` / `j/k` | Scroll up/down |
| `Home` / `g` | Jump to top |
| `End` / `G` | Jump to bottom |
| `q` / `Ctrl+C` | Quit |

## 📦 Installation

### Quick Install (One-Line)

```bash
curl -sSL https://raw.githubusercontent.com/musyoka101/sliver-graphs/master/go/install.sh | bash
```

### Manual Install

```bash
git clone https://github.com/musyoka101/sliver-graphs.git
cd sliver-graphs/go
./install.sh
```

### Run

```bash
# If installed to PATH
sliver-tui

# Or from build directory
./sliver-graph
```

## 🔧 Requirements

- Go 1.21+ (for building)
- Sliver C2 server running
- Sliver client configured (`~/.sliver-client/configs/*.cfg`)
- Terminal with 120x30+ size recommended
- Optional: Nerd Font for best icon display

## 📸 Screenshots

### Dashboard View
```
╔════════════════════════════════════════════════════════════════╗
║ 🌐 C2 INFRASTRUCTURE        🔹 ARCHITECTURE      📋 TASKS     ║
║ mtls://10.0.1.5:8443       amd64    ████ 100%   ✅ 5/5 done  ║
║   └─ 5 agents              arm64    ▓▓░░  67%   ⏳ 2 pending ║
╠════════════════════════════════════════════════════════════════╣
║ 🔒 SECURITY STATUS          📊 ACTIVITY METRICS (12 Hours)   ║
║ ✓ All agents operating      Sessions  ▁▂▃▄▅▆▇█  Peak: 2      ║
║   normally                  Beacons   ▁▂▃▄▅▆▇█  Peak: 3      ║
║                             New       ▁▂▃▃▃▃▄▅  Peak: 5      ║
║ 0 STEALTH  0 BURNED         Privileged▁▂▃▃▃▃▃▄  Peak: 3      ║
╚════════════════════════════════════════════════════════════════╝
```

### Tree View with Alerts
```
🔥 C2 Server: mtls://10.10.14.15:8443
  ├─[MTLS]──▶ ◆ 🖥️  root@webserver 💎 ✨
  │           └─ ID: a1b2c3d4 (session)
  │           └─ IP: 192.168.1.100
  └─[HTTP]──▶ ◇ 💻  admin@workstation
              └─ ID: e5f6g7h8 (beacon)
              └─ IP: 192.168.1.101

╔════════════════════════════════════════╗
║ ⚠ ALERTS ◉ ACTIVE                     ║
║ ║█║ 10:41 PRIVILEGED SESSION ACQUIRED ║
║ ║▓║ 10:42 TASK COMPLETE webserver     ║
║ ║░║ 10:43 BEACON ACQUIRED workstation ║
╚════════════════════════════════════════╝
```

## 🏗️ Architecture

**Built with:**
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Sliver SDK](https://github.com/BishopFox/sliver) - Official Go client
- gRPC + mTLS - Secure C2 communication

**Project Structure:**
```
go/
├── main.go              - UI rendering & Bubble Tea app
├── internal/
│   ├── alerts/         - Alert system
│   ├── client/         - Sliver gRPC client
│   ├── config/         - Themes & views
│   ├── models/         - Data structures
│   ├── tracking/       - Activity & change tracking
│   └── tree/           - Tree builder
├── install.sh          - Installation script
└── build.sh            - Build script
```

## 🐛 Known Issues

None reported! This is the first stable release.

## 🔮 Future Enhancements

Potential features for future releases:
- Agent filtering by OS/privilege/protocol
- Search functionality (find by hostname/IP)
- Agent detail inspector view
- Sorting options
- Export to CSV/JSON
- Quick actions menu for C2 operations
- Mouse support
- Persistent activity storage

## 🤝 Contributing

Contributions welcome! This project demonstrates modern TUI development with Go.

Areas for contribution:
- Additional visualization types
- New dashboard panels
- Export functionality
- Performance optimizations
- Additional themes

## 📝 License

MIT License

## 🙏 Credits

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Powered by [Sliver](https://github.com/BishopFox/sliver)

---

**Repository:** https://github.com/musyoka101/sliver-graphs  
**Documentation:** See [README.md](README.md) for detailed usage  
**Issues:** https://github.com/musyoka101/sliver-graphs/issues

## 🔐 Security Note

This tool is designed for authorized security testing and red team operations only. Always ensure you have proper authorization before using this tool against any systems.
