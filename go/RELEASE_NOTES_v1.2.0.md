# Sliver C2 TUI v1.2.0 - Process Display & Click Improvements

🎉 **Release Date:** January 2026

## 🌟 What's New

### Process Display Feature
**NEW: Toggleable Process Path Display**
- Agent details panel now shows process information
- **Smart display**: Shows filename by default (e.g., `powershell.exe`)
- **Press 'p'**: Toggle to view full path (e.g., `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`)
- **Cross-platform**: Works with both Windows backslash (`\`) and Unix forward slash (`/`) paths
- **Visual indicator**: Filename is highlighted and underlined with helpful hint text
- **Per-agent state**: Toggle state is tracked individually for each selected agent

### Mouse Interaction Improvements
**FIXED: Click Detection Accuracy**
- Dramatically improved click detection precision
- Clicks now only register within actual agent box boundaries
- Fixed issue where clicking empty space would select agents
- Added intelligent boundary checks:
  - Minimum X coordinate (skips logo area in Box view)
  - Maximum X coordinate (excludes empty space)
  - Proper viewport bounds checking
- No more accidental agent selection when clicking outside content area

### Documentation Updates
**Comprehensive Theme Documentation**
- Updated documentation to reflect all 14 available themes (previously listed only 6)
- Complete theme list now includes:
  1. Default (Dracula)
  2. Rainbow
  3. Cyberpunk
  4. Matrix
  5. Tactical
  6. Pastel
  7. Heatmap
  8. Lip Gloss
  9. Nord
  10. Gruvbox
  11. Tokyo Night
  12. Monokai
  13. Catppuccin Mocha
  14. Catppuccin Macchiato
  15. Catppuccin Frappé

**Visual Documentation**
- Added 6 high-quality screenshots to repository
- Screenshots showcase all major features:
  - Agent details panel
  - Alert panel
  - Box view
  - Dashboard view
  - Help menu
  - Table view

### Help Menu Enhancement
**New Section: Agent Details Panel Controls**
- Added dedicated help section for agent detail interactions
- Documents 'p' key for process path toggling
- Documents ESC key for closing agent details panel

## 🐛 Bug Fixes

### Click Detection
- **Fixed:** Mouse clicks registering outside intended click areas
- **Fixed:** Clicking in empty space selecting agents
- **Fixed:** Clicks on right panel (agent details/tactical) affecting main content
- **Fixed:** Logo area clicks triggering agent selection

### Process Display
- **Fixed:** Text overflow when displaying long process paths
- **Fixed:** Hint text extending beyond panel boundaries
- **Improved:** Shortened hint text from "(press 'p' for full path)" to "('p')"

### Network Map View
- **Restored:** Network Map view to view cycling (was accidentally removed)
- View order now: Box → Table → Dashboard → Network Map

## 🔑 Keyboard Controls (Updated)

| Key | Action |
|-----|--------|
| `p` | **NEW:** Toggle process path (filename ↔ full path) when agent selected |
| `r` | Manual refresh |
| `v` | Cycle through views |
| `d` | Quick dashboard access |
| `t` | Change theme |
| `i` | Toggle icon style (Nerd Font/Emoji) |
| `e` | Expand/collapse subnets (Dashboard/Network Map) |
| `?` | Toggle help menu |
| `ESC` | Deselect agent / Close help menu |
| `Tab` / `F1-F5` | Direct view access |
| `↑↓` / `j/k` | Scroll up/down |
| `Home` / `g` | Jump to top |
| `End` / `G` | Jump to bottom |
| `q` / `Ctrl+C` | Quit |

## 🖱️ Mouse Controls (Improved)

| Action | Function |
|--------|----------|
| Left Click | Select/deselect agent (shows details panel) |
| Click Agent Box | More precise detection - only within actual box boundaries |
| Click Alert | Jump to agent associated with alert |
| Scroll Wheel | Scroll content up/down (or help menu when open) |

## 📸 New Screenshots

The repository now includes comprehensive screenshots showing:
- **agent-details-panel.png** - Process display and agent information
- **alert-panel.png** - Real-time alert system
- **box-view.png** - Default box layout with agent hierarchy
- **dashboard-view.png** - Advanced analytics dashboard
- **help-menu.png** - Comprehensive keyboard shortcuts
- **table-view.png** - Tabular agent display

## 🔧 Technical Improvements

### Code Quality
- Added cross-platform path parsing helper function
- Improved mouse coordinate calculation logic
- Better boundary detection for click events
- Enhanced state management for process path expansion

### Performance
- No performance regression
- Maintains fast startup time (<100ms)
- Low memory footprint (~10MB)
- Efficient content caching

## 📦 Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/musyoka101/sliver-tui/master/go/install.sh | bash
```

### Manual Install

```bash
# Clone repository
git clone https://github.com/musyoka101/sliver-tui.git
cd sliver-tui/go

# Run installer
./install.sh

# Or build manually
./build.sh
```

### Upgrade from v1.1.0

```bash
# If installed via install.sh
cd sliver-tui/go
git pull
./install.sh

# Binary will be updated at ~/.local/bin/sliver-tui
```

## 🆕 What's Changed Since v1.1.0

### Features
- ✨ Process name display in agent details panel
- ✨ Toggleable process path with 'p' key
- ✨ Cross-platform path parsing (Windows/Unix)
- ✨ Agent Details Panel section in help menu

### Fixes
- 🐛 Mouse click detection now properly bounded
- 🐛 Text overflow in process display
- 🐛 Network Map view restored to cycling
- 🐛 Click detection in empty areas

### Documentation
- 📚 Complete theme list (14 themes documented)
- 📚 6 new screenshots added
- 📚 Updated README with all features
- 📚 Enhanced help menu content

## 🔮 Coming Soon

Planned for future releases:
- Agent filtering capabilities
- Search functionality
- Extended process information (command line arguments)
- Task history viewer
- Export functionality (CSV/JSON)
- Custom theme editor

## 🤝 Contributing

We welcome contributions! Areas for improvement:
- Additional agent detail fields
- More visualization options
- Performance optimizations
- UI/UX enhancements
- Bug reports and feature requests

## 📝 Commit History

```
7028324 fix: restrict agent click detection to actual content area
16d864c fix: prevent agent selection when clicking outside content area
14f971f fix: shorten process toggle hint to prevent overflow
8c000ab feat: add toggleable process path display (press 'p')
8b53f61 feat: simplify process display to show full path
e472c2b docs: update theme documentation with all 14 available themes
0565b8c docs: add comprehensive screenshots to repository
e055c0e feat: restore Network Map view to view cycling
```

## 📄 License

MIT License

## 🔗 Links

- **GitHub (Public):** https://github.com/musyoka101/sliver-tui
- **GitHub (Dev):** https://github.com/musyoka101/sliver-graphs
- **Issues:** https://github.com/musyoka101/sliver-tui/issues
- **Sliver C2:** https://github.com/BishopFox/sliver

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Powered by [Sliver C2](https://github.com/BishopFox/sliver)

---

**Full Changelog:** v1.1.0...v1.2.0

**⚠️ Security Note:** This tool is designed for authorized security testing and red team operations only. Always ensure proper authorization before use.
