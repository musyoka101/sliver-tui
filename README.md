# Sliver C2 Network Topology Visualizer

A modern terminal-based visualization tool for Sliver C2 infrastructure. Available in both **Go** (feature-rich TUI) and **Python** (lightweight script) versions.

<p align="center">
  <img src="https://img.shields.io/badge/version-1.0.0-blue.svg" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8.svg" alt="Go">
  <img src="https://img.shields.io/badge/Python-3.7+-3776AB.svg" alt="Python">
</p>

## 🎯 Quick Start

### Go Version (Recommended)

**One-line installation:**
```bash
curl -sSL https://raw.githubusercontent.com/musyoka101/sliver-tui/master/go/install.sh | bash
```

**Features:**
- 🎨 Beautiful TUI with multiple themes
- 📊 Advanced analytics dashboard
- 📈 12-hour activity tracking
- 🚨 Real-time alert system
- ⚡ Single binary, instant startup

[📖 Go Documentation →](go/README.md)

### Python Version

**Installation:**
```bash
pip3 install sliver-py
cd python
./graph
```

**Features:**
- 🐍 Simple Python script
- 🔄 Auto-refresh topology
- ✨ Change detection
- 💀 Dead beacon tracking

[📖 Python Documentation →](python/README.md)

## 📸 Screenshots

### Go Version - Dashboard View
```
╔════════════════════════════════════════════════════════════════╗
║ 🌐 C2 INFRASTRUCTURE        🔹 ARCHITECTURE      📋 TASKS     ║
║ mtls://10.0.1.5:8443       amd64    ████ 100%   ✅ 5/5 done  ║
║   └─ 5 agents              arm64    ▓▓░░  67%   ⏳ 2 pending ║
╠════════════════════════════════════════════════════════════════╣
║ 🔒 SECURITY STATUS          📊 ACTIVITY METRICS (12 Hours)   ║
║ ✓ All agents operating      Sessions  ▁▂▃▄▅▆▇█  Peak: 2      ║
║   normally                  Beacons   ▁▂▃▄▅▆▇█  Peak: 3      ║
╚════════════════════════════════════════════════════════════════╝
```

### Go Version - Tree View with Alerts
```
🔥 C2 Server: mtls://10.10.14.15:8443
  ├─[MTLS]──▶ ◆ 🖥️  root@webserver 💎 ✨
  │           └─ ID: a1b2c3d4 (session)
  └─[HTTP]──▶ ◇ 💻  admin@workstation
              └─ ID: e5f6g7h8 (beacon)

╔════════════════════════════════════════╗
║ ⚠ ALERTS ◉ ACTIVE                     ║
║ ║█║ 10:41 PRIVILEGED SESSION ACQUIRED ║
║ ║▓║ 10:42 TASK COMPLETE webserver     ║
╚════════════════════════════════════════╝
```

## 🚀 Features Comparison

| Feature | Go Version | Python Version |
|---------|-----------|---------------|
| **Views** | 5 (Tree, Box, Table, Dashboard, Network Map) | 1 (Simple list) |
| **Themes** | 5 professional themes | Basic colors |
| **Dashboard** | ✅ 5-panel analytics | ❌ |
| **Activity Tracking** | ✅ 12-hour sparklines | ❌ |
| **Alert System** | ✅ Real-time notifications | ❌ |
| **Performance** | ⚡ ~10MB RAM, <100ms startup | ~50MB RAM |
| **Installation** | Single binary | pip install |
| **Dependencies** | None (statically linked) | sliver-py |
| **Size** | ~19MB | ~33KB |

## 📦 Installation

### Go Version

#### Option 1: Quick Install
```bash
curl -sSL https://raw.githubusercontent.com/musyoka101/sliver-tui/master/go/install.sh | bash
```

#### Option 2: Manual Install
```bash
git clone https://github.com/musyoka101/sliver-tui.git
cd sliver-tui/go
./install.sh
```

#### Option 3: Download Binary
Download the latest release from the [Releases page](https://github.com/musyoka101/sliver-tui/releases).

### Python Version

```bash
# Install sliver-py
pip3 install sliver-py

# Clone and run
git clone https://github.com/musyoka101/sliver-tui.git
cd sliver-tui/python
chmod +x graph
./graph
```

## 🎮 Usage

### Go Version
```bash
# If installed to PATH
sliver-tui

# Or from build directory
./sliver-graph

# Keyboard shortcuts:
# r - Refresh           v - Change view
# t - Change theme      d - Dashboard
# i - Toggle icons      q - Quit
# ↑↓/j/k - Scroll      Tab/F1-F5 - Direct view access
```

### Python Version
```bash
./graph                 # Run with auto-refresh
python3 sliver-graph.py # Or run directly
# Ctrl+C to exit
```

## 🔧 Requirements

### Go Version
- Go 1.21+ (for building only)
- Sliver C2 server running
- Sliver client config (`~/.sliver-client/configs/*.cfg`)
- Terminal 120x30+ recommended

### Python Version
- Python 3.7+
- sliver-py library
- Sliver C2 server running
- Sliver client config

## 📚 Documentation

- **[Go Version Documentation](go/README.md)** - Complete feature guide
- **[Python Version Documentation](python/README.md)** - Python script usage
- **[Themes Guide](go/THEMES.md)** - Available color themes
- **[Nerd Font Icons](go/NERD_FONT_ICONS.md)** - Icon configuration
- **[Release Notes](go/RELEASE_NOTES.md)** - Version history

## 🏗️ Architecture

### Go Version
Built with modern TUI frameworks:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (The Elm Architecture)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Sliver SDK](https://github.com/BishopFox/sliver) - Official Go client
- gRPC + mTLS - Secure C2 communication

### Python Version
Simple and lightweight:
- [sliver-py](https://github.com/moloch--/sliver-py) - Python Sliver client
- Native asyncio for async operations

## 🤝 Contributing

Contributions welcome! Areas for contribution:
- Additional dashboard panels
- New visualization types
- Export functionality
- Performance optimizations
- Additional themes
- Bug fixes and improvements

## 🐛 Known Issues

None reported for v1.0.0!

Report issues at: https://github.com/musyoka101/sliver-tui/issues

## 🔮 Roadmap

Future enhancements under consideration:
- Agent filtering by OS/privilege/protocol
- Search functionality
- Agent detail inspector view
- Sorting options
- Export to CSV/JSON
- Quick actions menu for C2 operations
- Mouse support
- Persistent activity storage

## 📝 License

MIT License - See [LICENSE](go/LICENSE)

Copyright (c) 2024 musyoka101

## 🙏 Credits

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Powered by [Sliver](https://github.com/BishopFox/sliver)
- Python client by [sliver-py](https://github.com/moloch--/sliver-py)

## 🔐 Security & Legal

**⚠️ IMPORTANT:** This tool is designed for **authorized security testing and red team operations only**. 

- Always ensure you have proper authorization before using this tool
- Only use against systems you own or have explicit permission to test
- Follow all applicable laws and regulations
- The authors are not responsible for misuse or illegal activities

## ⭐ Show Your Support

If you find this tool useful, please consider:
- ⭐ Starring the repository
- 🐛 Reporting bugs or issues
- 💡 Suggesting new features
- 🤝 Contributing code improvements

---

**Repository:** https://github.com/musyoka101/sliver-tui  
**Version:** 1.0.0  
**Status:** Production Ready ✅
