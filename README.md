# 🎯 Sliver C2 TUI

<div align="center">

![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey.svg)

**A modern, beautiful terminal UI for Sliver C2 framework**

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Screenshots](#-screenshots) • [Documentation](#-documentation)

</div>

---

## 📖 Overview

Sliver C2 TUI is a powerful terminal-based user interface for managing and visualizing your Sliver C2 infrastructure. Built with Go and the Bubble Tea framework, it provides real-time monitoring, multiple view modes, and an intuitive interface for red team operations.

### Why Sliver C2 TUI?

- **🎨 Beautiful UI** - Modern terminal interface with multiple themes
- **⚡ Fast** - Written in Go, starts instantly
- **🔄 Real-time** - Live updates of agent status and statistics
- **🎮 Interactive** - Full keyboard and mouse support
- **📊 Multiple Views** - Box, Table, and Dashboard modes
- **🎯 Smart Alerts** - Click-to-jump alert system
- **🔍 Detailed Stats** - Comprehensive intelligence dashboard
- **🌈 Themes** - 6 color themes including Matrix, Cyberpunk, and Dracula

---

## ✨ Features

### Core Features

- **Real-time Agent Monitoring** - Auto-refresh every 5 seconds
- **Multiple View Modes**:
  - **Box View** (Default) - Compact boxed layout with side connectors
  - **Table View** - Professional spreadsheet-style display
  - **Dashboard View** - 5-page tactical intelligence dashboard
  - **Tree View** (Hidden) - Classic tree layout (Ctrl+T to access)
- **Interactive Alerts** - Click any alert to jump to that agent
- **Agent Details Panel** - Comprehensive information on selected agents
- **Scrollable Help Menu** - Full documentation with keyboard/mouse scrolling
- **Theme System** - 6 beautiful color schemes

### Agent Information

- **Status Indicators**:
  - 🟢 Green - Active session (interactive)
  - 🔵 Blue - Active beacon (check-in based)
  - 🔴 Red - Dead agent (missed check-ins)
- **Privilege Detection**:
  - 💎 Diamond - Privileged access (SYSTEM/root)
- **Activity Tracking**:
  - ✨ Sparkle - Recently connected (new agent)
- **Detailed Metrics**:
  - Hostname, Username, IP, Port
  - Operating System & Architecture
  - Process ID & Transport Protocol
  - Last check-in time & intervals

### Dashboard Analytics

**5-Page Intelligence Dashboard:**

1. **📊 OVERVIEW** - High-level statistics and agent summary
2. **🌐 NETWORK INTEL** - Subnet distribution and compromised networks
3. **⚡ OPERATIONS** - Task queues and operational metrics
4. **🔒 SECURITY** - Privilege analysis and access levels
5. **📈 ANALYTICS** - Activity trends and recent changes

### Alert System

- **🔴 Critical** - Session lost, beacon disconnected
- **🟡 Warning** - Beacon missed check-in
- **🟢 Success** - New connection, privilege escalation
- **🔵 Info** - State changes, task updates
- Auto-expiration after 30 seconds
- Click to jump to agent

---

## 🚀 Installation

### Quick Start (Binary Download)

```bash
# Download the latest release
wget https://github.com/musyoka101/sliver-tui/releases/latest/download/sliver-graph

# Make it executable
chmod +x sliver-graph

# Run it
./sliver-graph
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/musyoka101/sliver-tui.git
cd sliver-tui/go

# Download dependencies
go mod download

# Build
go build -o sliver-graph main.go

# Run
./sliver-graph
```

### Requirements

- **Sliver C2 Server** - Running Sliver C2 instance
- **Sliver Client Config** - Configured at `~/.sliver-client/configs/*.cfg`
- **Go 1.21+** - Only if building from source
- **Terminal** - Modern terminal with Unicode and color support

---

## 🎮 Usage

### Launch

```bash
./sliver-graph
```

The TUI will automatically connect to your Sliver C2 server using your configured client credentials.

### Keyboard Controls

#### General

- `?` - Toggle help menu (scrollable)
- `q` / `Ctrl+C` - Quit application
- `r` - Refresh agents from server
- `ESC` - Deselect agent / Clear number buffer

#### Views

- `v` - Cycle through views (Box → Table → Dashboard)
- `d` - Jump directly to Dashboard
- `Ctrl+T` - Access hidden Tree view 🤫
- `t` - Cycle through color themes
- `i` - Toggle icon style (Nerd Font ↔ Emoji)

#### Dashboard Navigation

- `Tab` - Next dashboard page
- `Shift+Tab` - Previous dashboard page
- `F1` - Jump to OVERVIEW page
- `F2` - Jump to NETWORK INTEL page
- `F3` - Jump to OPERATIONS page
- `F4` - Jump to SECURITY page
- `F5` - Jump to ANALYTICS page

#### Scrolling

- `↑`/`k` - Scroll up
- `↓`/`j` - Scroll down
- `PgUp` / `u` - Page up
- `PgDn` / `d` - Page down
- `Home` / `g` - Go to top
- `End` / `G` - Go to bottom

### Mouse Controls

- **Left Click** - Select/deselect agent (shows details panel)
- **Click Alert** - Jump to agent associated with alert
- **Scroll Wheel** - Scroll content up/down

---

## 📸 Screenshots

### Box View (Default)
<!-- Add your screenshot here -->

### Dashboard View
<!-- Add your screenshot here -->

### Table View
<!-- Add your screenshot here -->

### Help Menu
<!-- Add your screenshot here -->

---

## 🎨 Themes

Six beautiful color themes included:

1. **Default** - Classic green terminal aesthetic
2. **Nord** - Cool, northern bluish palette
3. **Dracula** - Popular purple-tinted dark theme
4. **Matrix** - Iconic green-on-black hacker aesthetic
5. **Cyberpunk** - Vibrant neon pinks and blues
6. **Gruvbox** - Retro warm color scheme

Press `t` to cycle through themes in real-time!

---

## 📚 Documentation

### View Modes Explained

#### Box View
- **Best for**: Quick overview, default view
- **Layout**: Compact boxes with C2 server icon
- **Info**: Hostname, ID, IP, privilege level
- **Navigation**: Clean, organized, professional

#### Table View
- **Best for**: Detailed comparison, many agents
- **Layout**: Spreadsheet-style columns
- **Info**: All agent details in sortable format
- **Navigation**: Easy scanning, data-focused

#### Dashboard View
- **Best for**: Tactical intelligence, analytics
- **Layout**: 5-page intelligence dashboard
- **Info**: Statistics, trends, network maps
- **Navigation**: Tab through pages, F-keys for quick jump

#### Tree View (Hidden)
- **Best for**: Visualizing pivot chains
- **Layout**: Hierarchical tree with C2 logo
- **Info**: Classic topology visualization
- **Navigation**: Access via `Ctrl+T`

---

## 🛠️ Development

### Project Structure

```
sliver-tui/
├── go/
│   ├── main.go              # Main application
│   ├── internal/
│   │   ├── alerts/          # Alert management
│   │   ├── client/          # Sliver client integration
│   │   ├── config/          # Themes and views config
│   │   ├── models/          # Data models
│   │   ├── tracking/        # Activity tracking
│   │   └── tree/            # Tree view builder
│   ├── go.mod
│   └── go.sum
├── README.md
├── LICENSE
└── CONTRIBUTING.md
```

### Tech Stack

- **[Go](https://golang.org/)** - Programming language
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Styling
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components
- **[Sliver](https://github.com/BishopFox/sliver)** - C2 framework

### Building

```bash
cd go
go build -o sliver-graph main.go
```

### Running Tests

```bash
cd go
go test ./...
```

---

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Start

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Commit (`git commit -m 'feat: add amazing feature'`)
5. Push (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

## 📋 Changelog

### v1.1.0 (2026-01-01)

**New Features:**
- Scrollable help menu with keyboard and mouse support
- Theme-aware help styling
- Box view set as default
- Hidden Tree view easter egg (Ctrl+T)
- Enhanced alert clicking

**Improvements:**
- Simplified view cycling (3 views)
- Better viewport responsiveness
- Improved help organization

**Bug Fixes:**
- Fixed help menu double border
- Fixed viewport scroll persistence
- Fixed alert click coordinates

[View all releases](https://github.com/musyoka101/sliver-tui/releases)

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **[BishopFox](https://github.com/BishopFox)** - For the amazing Sliver C2 framework
- **[Charm](https://charm.sh/)** - For the beautiful Bubble Tea TUI framework
- **Red Team Community** - For feedback and feature requests

---

## 📞 Contact

- **Author**: musyoka101
- **Email**: ianmusyoka101@gmail.com
- **Issues**: [GitHub Issues](https://github.com/musyoka101/sliver-tui/issues)
- **Discussions**: [GitHub Discussions](https://github.com/musyoka101/sliver-tui/discussions)

---

## ⚠️ Disclaimer

This tool is for authorized security testing and red team operations only. Always ensure you have proper authorization before testing any systems.

---

<div align="center">

**Made with ❤️ for the Red Team community**

⭐ Star us on GitHub if you find this useful!

</div>
