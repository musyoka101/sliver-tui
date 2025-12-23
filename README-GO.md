# Sliver C2 Network Topology Visualizer - Go/Bubble Tea Edition

This is a **Bubble Tea** (Go) rewrite of the Sliver topology visualizer with a beautiful, interactive TUI.

## Features

✨ **Beautiful TUI** - Built with Bubble Tea and Lip Gloss
- Professional styling with borders and colors
- Smooth animations and spinners
- Interactive keyboard controls

🎯 **All Original Features**
- Hierarchical topology visualization
- Multi-line agent display (username@host, ID, IP)
- Privilege detection with 💎 badges
- Real-time stats dashboard
- Auto-refresh every 5 seconds

⚡ **Performance Benefits**
- Compiled binary (no Python/venv needed)
- Instant startup
- Better performance
- Single 4MB executable

## Installation

### Prerequisites
- Go 1.21+ installed
- Sliver C2 client configured

### Build
```bash
go build -o sliver-graph main.go
```

## Usage

```bash
# Run the TUI
./sliver-graph

# Keyboard shortcuts:
# r - Manual refresh
# q - Quit
```

## Architecture

- **Bubble Tea** - TUI framework (The Elm Architecture)
- **Lip Gloss** - Styling and colors
- **Bubbles** - UI components (spinner, etc.)

## Current Status

🚧 **Work in Progress**
- [x] Basic Bubble Tea UI structure
- [x] Mock data rendering
- [x] Beautiful styling with Lip Gloss
- [x] Auto-refresh timer
- [x] Keyboard controls
- [ ] Real Sliver client connection
- [ ] Change detection (NEW badges)
- [ ] Lost agents tracking
- [ ] Hierarchical tree with children/pivots
- [ ] Advanced features (filtering, sorting, mouse support)

## Branches

- `master` - Stable Python version (basic)
- `dev` - Python version with all features
- `experimental` - Python testing branch
- `go-bubbletea` - **This branch** - Go/Bubble Tea rewrite

## Why Go + Bubble Tea?

1. **Native Sliver Integration** - Sliver is written in Go
2. **Better Performance** - Compiled vs interpreted
3. **Single Binary** - Easy distribution
4. **Professional TUI** - Bubble Tea is production-ready
5. **No Dependencies** - No Python, venv, pip issues

## Development

```bash
# Switch to this branch
git checkout go-bubbletea

# Install dependencies
go mod download

# Build
go build -o sliver-graph main.go

# Run
./sliver-graph
```

## Next Steps

1. Implement real Sliver client connection
2. Port all features from Python version
3. Add advanced TUI features (tabs, menus, mouse)
4. Performance testing and optimization
5. Release binary builds

## License

Same as parent project
