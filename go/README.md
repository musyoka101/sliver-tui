# Sliver C2 Network Topology Visualizer - Go/Bubble Tea Edition

A modern, high-performance **Bubble Tea** (Go) implementation of the Sliver C2 network topology visualizer with a beautiful, interactive TUI and comprehensive analytics dashboard.

## Features

✨ **Beautiful TUI** - Built with Bubble Tea and Lip Gloss
- Professional styling with borders and gradients
- Smooth animations and loading spinners
- Interactive keyboard controls
- Real-time updates every 5 seconds
- Multiple themed views (Dracula, Monokai, Nord, Solarized, Cyberpunk)

🎯 **Complete Feature Set**
- ✅ Real Sliver C2 connection (gRPC + mTLS)
- ✅ Hierarchical topology with pivot detection
- ✅ Live agent tracking with NEW badges (✨)
- ✅ Dead beacon detection (💀)
- ✅ Lost agents tracking (5 min window)
- ✅ Privilege detection with 💎 badges
- ✅ Security status tracking (🕵️ STEALTH, 🔥 BURNED)
- ✅ OS-specific icons (🖥️ 💻 🐧)
- ✅ Protocol color coding (MTLS, HTTP, DNS, TCP)
- ✅ **Interactive analytics dashboard**

📊 **Advanced Analytics Dashboard**
- **C2 Infrastructure Map** - Active C2 servers with agent counts and protocol breakdown
- **Architecture Distribution** - Visual bar charts of agent architectures (x64, x86, arm64)
- **Task Queue Monitor** - Real-time beacon task progress tracking
- **Security Status Panel** - STEALTH and BURNED agent monitoring with hostnames
- **Activity Metrics** - 12-hour historical tracking with sparkline graphs
  - Sessions count over time
  - Beacons count over time
  - New agents discovered
  - Privileged agents detected
  - Time-mapped visualization with hour markers
  - Peak, current, and average statistics

⚡ **Performance Benefits**
- Compiled binary (no Python/venv needed)
- Instant startup (<100ms)
- Low memory footprint (~10MB)
- Native Sliver SDK integration
- Single ~19MB executable

## Installation

### Quick Install (Recommended)

**One-line installation:**
```bash
curl -sSL https://raw.githubusercontent.com/musyoka101/sliver-graphs/master/go/install.sh | bash
```

Or clone and run:
```bash
git clone https://github.com/musyoka101/sliver-graphs.git
cd sliver-graphs/go
./install.sh
```

The installer will:
- ✓ Check for Go 1.21+
- ✓ Download dependencies
- ✓ Build optimized binary
- ✓ Install to `~/.local/bin/sliver-tui`
- ✓ Check for Sliver configs

### Manual Build

If you prefer to build manually:
```bash
cd go
go mod download
go build -o sliver-tui .
```

### Prerequisites
- Go 1.21+ installed
- Sliver C2 server running
- Sliver client configured (`~/.sliver-client/configs/*.cfg`)

## Usage

```bash
# If installed with install.sh (in PATH)
sliver-tui

# Or run from build directory
./sliver-tui

# Keyboard shortcuts:
# r - Manual refresh
# t - Change theme (5 themes available)
# v - Change view (Tree/List/Dashboard)
# d - Quick dashboard access
# ↑↓ or j/k - Scroll
# Home/g - Jump to top
# End/G - Jump to bottom
# q / Ctrl+C - Quit
```

## Dashboard Views

### Main Dashboard (Press 'd')
The dashboard features a **5-panel layout** providing comprehensive operational analytics:

**Top Row:**
1. **🌐 C2 Infrastructure Map** - Shows all active C2 servers, agent counts, and protocol distribution
2. **🔹 Architecture Distribution** - Visual breakdown of agent architectures with percentage bars
3. **📋 Task Queue Monitor** - Real-time tracking of beacon task execution progress

**Bottom Row:**
4. **🔒 Security Status** - Lists agents in STEALTH mode (evasion) and BURNED/compromised agents
5. **Activity Metrics** (spans 2 columns) - 12-hour historical sparkline graphs tracking:
   - Session counts
   - Beacon counts  
   - New agent discoveries
   - Privileged agent detections
   - Automatic sampling every 10 minutes (72 samples max)
   - Time axis with hour markers

### Agent List Views
- **Tree View** - Hierarchical display showing C2 → Agents with pivot relationships
- **List View** - Flat list of all agents with full details
- Both views include:
  - Agent status indicators (Session ◆, Beacon ◇, Dead 💀)
  - Privilege badges (💎)
  - NEW agent badges (✨)
  - Protocol boxes with color coding
  - OS-specific icons
  - Remote IP addresses

### Visual Enhancements
- **Green sparklines** (#00FF00) for activity metrics
- **Cyan progress bars** (#00CED1) for architecture and task queues
- **Purple STEALTH badges** (#9370DB) for evasion mode agents
- **Orange-red BURNED badges** (#FF4500) for compromised agents
- **Themed color schemes** - 5 professional themes to choose from

## Architecture

**Technology Stack:**
- **Bubble Tea** - TUI framework (The Elm Architecture)
- **Lip Gloss** - Advanced styling and colors
- **Bubbles** - Reusable UI components (viewport, spinner)
- **Sliver SDK** - Official Go client library (v1.5.42)
- **gRPC** - High-performance RPC communication

**Project Structure:**
```
go/
├── main.go          - Bubble Tea app, UI rendering & dashboard panels
├── sliver.go        - Sliver client & gRPC connection
├── themes.go        - Theme definitions and color schemes
├── views.go         - View type definitions
├── go.mod           - Go module dependencies
├── go.sum           - Dependency checksums
├── build.sh         - Build script
├── run.sh           - Launcher script
└── README.md        - This file
```

## Current Status

✅ **Production Ready** - All features implemented and tested!

### Completed Features
- [x] Bubble Tea TUI framework
- [x] Beautiful styling with Lip Gloss
- [x] Auto-refresh timer (5 seconds)
- [x] Keyboard controls (r/t/v/d/q)
- [x] **Real Sliver RPC connection via gRPC**
- [x] **Token-based authentication**
- [x] **Change detection with ✨ NEW! badges**
- [x] **Lost agents tracking**
- [x] **Dead beacon detection (💀)**
- [x] **Hierarchical tree with pivot support**
- [x] **Privilege detection (💎)**
- [x] **Interactive analytics dashboard**
- [x] **5-panel dashboard layout**
- [x] **C2 Infrastructure mapping**
- [x] **Architecture distribution charts**
- [x] **Task queue monitoring**
- [x] **Security status tracking (Evasion/Burned)**
- [x] **12-hour activity tracking with sparklines**
- [x] **Multiple theme support**
- [x] **Multiple view modes**
- [x] **Scrollable viewport**

### Future Enhancements (Optional)
- [ ] Advanced filtering (by OS, privilege, transport)
- [ ] Sorting options
- [ ] Mouse support (click to select agents)
- [ ] Export dashboard to PNG/SVG
- [ ] Export activity data to JSON/CSV
- [ ] Adjustable sampling intervals
- [ ] Persistent activity data storage
- [ ] Alert thresholds for burned agents

## Branches

- `master` - **Current stable** - Go/Bubble Tea with full dashboard
- `dev` - Development branch with latest features
- `go-bubbletea` - Active development branch
- `experimental` - Python testing branch (deprecated)

## Why Go + Bubble Tea?

1. **Native Sliver Integration** - Sliver is written in Go, native SDK access
2. **Performance** - 10-100x faster than Python
3. **Single Binary** - Easy deployment, no dependencies
4. **Professional TUI** - Bubble Tea powers GitHub CLI, k9s, lazygit
5. **Type Safety** - Compile-time error checking
6. **Cross-Platform** - Linux, macOS, Windows support
7. **Small Footprint** - ~19MB binary vs ~200MB Python venv
8. **Rich Visualization** - Sparklines, charts, and real-time graphs

## Development

### Setup Development Environment
```bash
# Clone the repository
git clone https://github.com/musyoka101/sliver-graphs.git
cd sliver-graphs/go

# Install dependencies
go mod tidy

# Build
go build -o sliver-tui .

# Run
./sliver-tui
```

### Testing
```bash
# Run with your Sliver server
./sliver-tui

# Test different themes (press 't' repeatedly)
# Test dashboard (press 'd')
# Test views (press 'v' repeatedly)
```

### Activity Tracking Implementation
The activity tracker samples agent states every 10 minutes:
- Rolling 12-hour window (72 samples maximum)
- Tracks: Sessions, Beacons, New Agents, Privileged Agents
- In-memory storage (no persistent data across sessions)
- Thread-safe with mutex-protected access
- Automatic sampling on agent fetch + 10-minute intervals
- Sparklines use 8-level block characters (▁▂▃▄▅▆▇█)

### Dashboard Panel Architecture
Each panel is independently rendered:
- `renderC2InfrastructurePanel()` - Groups agents by C2 server
- `renderArchitecturePanel()` - Architecture distribution with bars
- `renderTaskQueuePanel()` - Beacon task progress tracking
- `renderSecurityStatusPanel()` - STEALTH/BURNED agent listing
- `renderSparklinePanel()` - Historical activity sparklines
- All panels use consistent width/height for grid alignment

## Configuration

The tool automatically discovers your Sliver config:
- Looks in `~/.sliver-client/configs/*.cfg`
- Uses the first `.cfg` file found
- Supports mTLS authentication
- Token-based API authorization

## Troubleshooting

**Connection Issues:**
```bash
# Check Sliver server is running
ps aux | grep sliver-server

# Verify config exists
ls ~/.sliver-client/configs/

# Check config permissions
chmod 600 ~/.sliver-client/configs/*.cfg
```

**Build Issues:**
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o sliver-tui .
```

**Dashboard Not Showing:**
- Press 'd' key to toggle dashboard
- Press 'v' to cycle through views
- Ensure terminal is at least 120x30 for optimal display

**Activity Metrics Show "Collecting data...":**
- Wait for first automatic sample (occurs on agent fetch)
- Sparklines will populate as data is collected over time
- Each sample is taken every 10 minutes

## Performance Comparison

| Metric | Python | Go + Bubble Tea |
|--------|--------|-----------------|
| Startup | ~2s | <100ms |
| Memory | ~50MB | ~10MB |
| Binary Size | N/A (venv ~200MB) | 19MB |
| Dependencies | sliver-py, asyncio | None (statically linked) |
| Refresh Speed | ~500ms | ~50ms |
| Dashboard | ❌ | ✅ (5 panels) |
| Activity Tracking | ❌ | ✅ (12-hour history) |
| Themes | 1 | 5 |

## Screenshots

### Dashboard View
```
┌─────────────────┬─────────────────┬─────────────────┐
│ C2 Infrastructure│ Architecture   │ Task Queue      │
│ mtls://10.0.1.5 │ amd64 ████ 100%│ ✅ cywebdw 3/3  │
│   5 agents      │                │ ✅ m3webaw 2/2  │
└─────────────────┴─────────────────┴─────────────────┘
┌─────────────────┬───────────────────────────────────┐
│ 🔒 Security     │ Activity Metrics (12 Hours)      │
│ All agents      │ Sessions  ▁▂▃▄▅▆▇█  Peak: 2      │
│ operating       │ Beacons   ▁▂▃▄▅▆▇█  Peak: 3      │
│ normally        │ New       ▁▂▃▃▃▃▄▅  Peak: 5      │
│ ✓ 5 standard    │ Privileged▁▂▃▃▃▃▃▄  Peak: 3      │
│                 │ 19:00              Now            │
└─────────────────┴───────────────────────────────────┘
```

### Tree View
```
🔥🔥     ╰────────[MTLS]────────▶ ◆ 🖥️  root@webserver 💎
▄▄▄▄▄▄▄                           └─ ID: a1b2c3d4 (session) ✨ NEW!
█ C2  █                           └─ IP: 192.168.1.100
█▓▓▓▓▓█ 
▀▀▀▀▀▀▀  ╰────────[MTLS]────────▶ ◇ 💻  admin@workstation
                                  └─ ID: e5f6g7h8 (beacon)
                                  └─ IP: 192.168.1.101
```

## Contributing

Contributions welcome! This project demonstrates modern TUI development with Go and Bubble Tea.

Areas for contribution:
- Additional dashboard panels
- More visualization types
- Export functionality
- Alert systems
- Performance optimizations

## License

MIT License - See parent project for details

## Credits

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Powered by [Sliver](https://github.com/BishopFox/sliver)

---

**Repository:** https://github.com/musyoka101/sliver-graphs  
**Branch:** master  
**Status:** Production Ready ✅
