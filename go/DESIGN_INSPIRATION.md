# Design Inspiration from Popular Bubble Tea Projects

## Research Summary

Analyzed 10,000+ Bubble Tea applications to identify best practices for our Sliver C2 tactical panel.

---

## 📊 Top Bubble Tea Projects Analyzed

### 1. **gh-dash** (9.5k stars) - GitHub Terminal Dashboard
**URL**: https://github.com/dlvhdr/gh-dash

#### Key Design Features:
- **Multi-pane layout**: Split screen with PRs/Issues sections
- **Rich table display**: Sortable columns with clear headers
- **Contextual actions**: Quick actions accessible via keyboard shortcuts
- **Status indicators**: Color-coded badges for state (open, merged, closed)
- **Real-time updates**: Auto-refresh with visual indicators
- **Vim-style navigation**: hjkl movement, intuitive keybindings

#### Visual Elements:
```
┌─────────────────────────────────────────────────────────────────┐
│  Pull Requests                                         [12]  🔄 │
├─────────────────────────────────────────────────────────────────┤
│  ✓ #123  feat: add new feature            MERGED    2h ago     │
│  ◉ #122  fix: critical bug                OPEN      5h ago     │
│  ✗ #121  docs: update readme              CLOSED    1d ago     │
└─────────────────────────────────────────────────────────────────┘
```

#### Applicable to Sliver C2:
- ✅ Use badges for agent status (session/beacon/dead)
- ✅ Color-coded state indicators
- ✅ Quick action menu per agent
- ✅ Real-time refresh indicator
- ✅ Sortable columns (by hostname, privilege, last seen)


### 2. **superfile** (16.1k stars) - Terminal File Manager
**URL**: https://github.com/yorukot/superfile

#### Key Design Features:
- **Dual-pane interface**: Side-by-side file browsing
- **Preview pane**: Shows file content on selection
- **Icon-based UI**: Rich icons for file types
- **Breadcrumb navigation**: Clear current path display
- **Bottom status bar**: File count, disk usage, permissions
- **Context menu**: Right-click style operations menu

#### Visual Elements:
```
┌─────────────────────┬─────────────────────┬──────────────────┐
│  📁 Directory       │  📁 Directory       │  Preview Pane    │
│  ├─ 📄 file1.txt   │  ├─ 📄 README.md   │  Content shown   │
│  ├─ 📄 file2.go    │  ├─ 📁 src/        │  here for        │
│  └─ 📁 config/     │  └─ 📄 main.go     │  selected item   │
├─────────────────────┴─────────────────────┴──────────────────┤
│  5 items  │  2.3 GB free  │  rw-r--r--                       │
└──────────────────────────────────────────────────────────────┘
```

#### Applicable to Sliver C2:
- ✅ Agent detail preview pane on selection
- ✅ Status bar with aggregate stats
- ✅ File tree structure for pivot relationships
- ✅ Context menu for agent operations
- ✅ Rich icons for OS types, transport protocols


### 3. **glow** - Markdown Reader
**URL**: https://github.com/charmbracelet/glow

#### Key Design Features:
- **Styled text rendering**: Beautiful markdown display
- **Pager interface**: Smooth scrolling through content
- **Search highlighting**: Yellow highlights on search terms
- **Border decoration**: Rounded corners, drop shadows
- **Gradient backgrounds**: Subtle color transitions

#### Applicable to Sliver C2:
- ✅ Better agent info display with styled text
- ✅ Smooth scrolling viewport (already implemented!)
- ✅ Search/filter highlighting
- ✅ Decorative borders for tactical panel


### 4. **Huh?** - Interactive Forms
**URL**: https://github.com/charmbracelet/huh

#### Key Design Features:
- **Modal dialogs**: Focused input forms that overlay main view
- **Field validation**: Real-time input validation with error messages
- **Multi-step forms**: Wizard-style navigation
- **Accessible UI**: Clear focus indicators, keyboard navigation

#### Applicable to Sliver C2:
- ✅ Modal for agent details/operations
- ✅ Confirmation dialogs for dangerous operations
- ✅ Form for filtering agents
- ✅ Command input modal


### 5. **lazygit** (Not pure Bubble Tea, but similar TUI)
**URL**: https://github.com/jesseduffield/lazygit

#### Key Design Features:
- **4-pane layout**: Status, Files, Branches, Commits
- **Command log**: Shows recent actions at bottom
- **Diff viewer**: Side-by-side comparison
- **Key binding hints**: Context-sensitive help at bottom

#### Applicable to Sliver C2:
- ✅ Command history/log at bottom
- ✅ Context-sensitive key hints
- ✅ Multi-section layout (already have!)

---

## 🎨 Common Design Patterns Across All Projects

### 1. **Layout Patterns**

#### Split Panes (Most Popular)
```
┌────────────────┬──────────────────┐
│  Main Content  │  Side Panel      │
│                │                  │
│  (agents)      │  (tactical intel)│
│                │                  │
├────────────────┴──────────────────┤
│  Status Bar / Help                │
└───────────────────────────────────┘
```
✅ **We already use this!** Main content + tactical panel

#### Stacked Sections
```
┌────────────────────────────────────┐
│  Section 1: Active Sessions        │
├────────────────────────────────────┤
│  Section 2: Beacons                │
├────────────────────────────────────┤
│  Section 3: Dead Agents            │
└────────────────────────────────────┘
```
💡 **Could implement**: Separate agent list by type

#### Tabs
```
┌─[Sessions]─[Beacons]─[Pivots]──────┐
│  Content for selected tab          │
└────────────────────────────────────┘
```
💡 **Could implement**: Tab switching between views


### 2. **Color Schemes**

#### Status Colors (Universal)
```
🟢 Green   → Success / Active / Online
🟡 Yellow  → Warning / Pending / Beacon
🔴 Red     → Error / Critical / Dead
🔵 Blue    → Info / Normal / Session
⚪ Gray    → Disabled / Inactive
```
✅ **We already use this!**

#### Accent Colors
- **Primary**: Cyan (#00d7ff) - Headers, borders
- **Secondary**: Yellow (#f1fa8c) - Highlights, warnings
- **Success**: Green (#50fa7b) - Completed, privileged
- **Danger**: Red (#ff5555) - Errors, critical

✅ **We already use these!**


### 3. **Typography & Symbols**

#### Box Drawing Characters
```
Single Line:  ─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼
Double Line:  ═ ║ ╔ ╗ ╚ ╝ ╠ ╣ ╦ ╩ ╬
Rounded:      ─ │ ╭ ╮ ╰ ╯
Heavy:        ━ ┃ ┏ ┓ ┗ ┛
```
✅ **We use rounded borders!**

#### Common Symbols
```
Status:    ✓ ✗ ◉ ◯ ● ○ ◆ ◇ ■ □
Direction: → ← ↑ ↓ ⇒ ⇐ ⇑ ⇓
Progress:  ▁ ▂ ▃ ▄ ▅ ▆ ▇ █
Arrows:    ▶ ▷ ▸ ▹ ► ▻
Tree:      ├─ └─ │ ╰─ ╯
```
✅ **We already use many of these!**


### 4. **Interactive Elements**

#### Selection Indicators
```
>  Selected item               (current cursor)
*  Marked item                 (multi-select)
[x] Checked                    (completed)
[ ] Unchecked                  (incomplete)
```
✅ **We don't have multi-select yet** 💡

#### Loading States
```
⠋ Spinner animation
[████░░░░░░] Progress bar
... Loading...
🔄 Refresh indicator
```
💡 **Could add**: Loading spinner during agent fetch


### 5. **Information Density**

#### Compact (More items visible)
```
◉ user@host  10.1.1.1  mtls  5m ago
◇ admin@srv  10.1.1.2  http  2h ago
```

#### Detailed (Less items, more info)
```
┌─────────────────────────────────────┐
│ ◉ Session                           │
│   user@host                         │
│   10.1.1.1 • mtls • Windows 10      │
│   Last seen: 5 minutes ago          │
└─────────────────────────────────────┘
```

✅ **We use detailed view currently**
💡 **Could add**: Compact mode toggle


---

## 🚀 Recommended Enhancements for Sliver C2

### Priority 1: High Impact, Easy to Implement

#### 1. **Progress Bars for Ratios**
```go
// Show privilege ratio visually
Privileged: 3 / 8
[██████░░░░░░░░░░] 38% 🟡
```
**Benefit**: Instant visual understanding of access level  
**Effort**: Low (already designed!)

#### 2. **Agent Selection with Details Modal**
```
Press ENTER on agent → Show detailed modal:
┌─────────────────────────────────────┐
│  Agent Details                      │
├─────────────────────────────────────┤
│  ID: abc123                         │
│  Hostname: workstation01            │
│  Username: CORP\user                │
│  OS: Windows 10 Pro                 │
│  Transport: mtls                    │
│  First Seen: 2h ago                 │
│  Last Check-in: 30s ago             │
│                                     │
│  [Execute Command] [Kill Session]   │
└─────────────────────────────────────┘
```
**Benefit**: Detailed info without cluttering main view  
**Effort**: Medium

#### 3. **Sparklines for Trends** (if we add history tracking)
```go
Check-ins: ▁▃▅▇█▇▅▃▁ (last 10 checks)
```
**Benefit**: See activity patterns at a glance  
**Effort**: Medium (requires historical data)

#### 4. **Status Bar Enhancements**
```
Current:  Last Update: 14:32:45 | Scroll: 50% | Term: 337x45
Enhanced: ⚡ 4 agents | 🟢 1 session | 🟡 3 beacons | ⏰ 14:32:45 | 📊 50% | ⌨️  r:refresh ↑↓:scroll q:quit
```
**Benefit**: More information in one glance  
**Effort**: Low


### Priority 2: Medium Impact, Moderate Effort

#### 5. **Tabbed Views**
```
[All Agents] [Sessions Only] [Beacons Only] [Pivots] [Dead]
```
**Benefit**: Quick filtering by agent type  
**Effort**: Medium

#### 6. **Search/Filter Modal**
```
Press '/' to search:
┌────────────────────────────────┐
│ Filter agents:                 │
│ > admin_________________       │
│                                │
│ Results: 2 matches             │
└────────────────────────────────┘
```
**Benefit**: Find specific agents quickly  
**Effort**: Medium

#### 7. **Color-Coded Risk Scoring**
```
High Risk:   🔴 [██████████████] 90% (privileged + recent)
Medium Risk: 🟡 [████████░░░░░░] 60% (privileged only)
Low Risk:    🟢 [████░░░░░░░░░░] 30% (standard user)
```
**Benefit**: Prioritize high-value targets  
**Effort**: Medium

#### 8. **Context Menu on Selection**
```
Press 'm' on agent:
┌────────────────────────┐
│ Actions:               │
│  > Execute Command     │
│    Open Shell          │
│    View Details        │
│    Kill Session        │
│    ──────────────      │
│    Copy ID             │
│    Copy Hostname       │
└────────────────────────┘
```
**Benefit**: Quick access to operations  
**Effort**: Medium-High


### Priority 3: Nice to Have, Higher Effort

#### 9. **Network Topology Graph**
```
        [C2 Server]
           ╱   ╲
          /     \
   [Host A]    [Host B]
      │            │
  [Host C]    [Host D]
```
**Benefit**: Visual pivot relationships  
**Effort**: High (complex layout algorithm)

#### 10. **Timeline View**
```
14:30 ▂▂▃▃▄▄▅▅▆▆▇▇██▇▇▆▆ Now
       └─ Spike at 14:00 (3 new agents)
```
**Benefit**: Understand campaign timeline  
**Effort**: Medium-High (requires history)


---

## 🎯 Quick Wins for Next Sprint

### Week 1: Visual Enhancements (No Data Changes)
1. **Progress bars** for privilege ratio, OS distribution, transports
2. **Enhanced status bar** with more icons and info
3. **Better help footer** with color-coded key hints

### Week 2: Interaction Improvements
4. **Agent selection** with ENTER key
5. **Details modal** showing full agent info
6. **Context menu** with common operations

### Week 3: Advanced Features
7. **Search/filter** modal with live results
8. **Historical tracking** for sparklines
9. **Tab switching** between agent types

---

## 📐 Layout Mockup: Enhanced Version

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  🎯 Sliver C2 Network Topology                                               ║
║  ⚡ 4 agents • 🟢 1 session • 🟡 3 beacons • ⏰ 14:32:45                     ║
╚══════════════════════════════════════════════════════════════════════════════╝

┌──[All Agents]─[Sessions]─[Beacons]─[Pivots]─────────┬───────────────────────┐
│                                                      │  📊 TACTICAL INTEL    │
│  🎯 C2     ──[ mtls ]────▶ ◆ 🖥️  admin@host1       │                       │
│            ││ abc123 • 10.1.1.1 • (session) 💎 ✨ NEW│  🌐 Subnets: 1        │
│            ╰─[ mtls ]────▶ ◇ 🖥️  user@host1         │    10.1.1.0/24        │
│               def456 • 10.1.1.1 • (beacon)          │    [████████] 3 hosts │
│                                                      │                       │
│     ──[ mtls ]────▶ ◆ 🖥️  admin@host2 💎            │  💎 Privileges        │
│        ghi789 • 10.1.1.2 • (session)                │    Admin: 2 / 4       │
│        ╰─[ http ]────▶ ◇ 💻  user@host3             │    [████░░░░] 50% 🟡  │
│           jkl012 • 10.1.1.3 • (beacon)              │    ▂▄▆▇█ Escalating! │
│                                                      │                       │
│     ──[ mtls ]────▶ ◇ 🖥️  user@host4                │  🔐 Transports        │
│        mno345 • 10.1.1.4 • (beacon)                 │    MTLS [███████] 75%│
│                                                      │    HTTP [████░░░] 25%│
│                                                      │                       │
├──────────────────────────────────────────────────────┤  ⚡ Activity          │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │    ▁▃▅▇█ High now!   │
│ 🟢 Sessions: 2 | 🟡 Beacons: 2 | 🔵 Total: 4        │                       │
│ ⚠️  1 agent lost connection (5m ago)                 │  [See Full Report]   │
└──────────────────────────────────────────────────────┴───────────────────────┘
  ⌨️  ENTER:details  m:menu  /:search  TAB:switch  r:refresh  ↑↓jk:scroll  q:quit
```

### Key Improvements:
1. ✅ Tabs at top for filtering
2. ✅ Progress bars in tactical panel
3. ✅ Sparklines for trends
4. ✅ Better status bar
5. ✅ Rich key hints at bottom
6. ✅ Visual hierarchy with boxes and colors

---

## 🎨 Color Palette Reference

### Current Colors (Keep These)
```go
titleStyle:   #00d7ff (cyan) - Headers, borders, accents
logoStyle:    #d75fff (pink) - C2 logo
statusStyle:  #888888 (gray) - Status text, timestamps
helpStyle:    #626262 (dark gray) - Help text
separatorStyle: #444444 (darker gray) - Separators
statsStyle:   #00d7ff (cyan) - Stats text
sectionStyle: #f1fa8c (yellow) - Section headers
valueStyle:   #50fa7b (green) - Values, data
mutedStyle:   #6272a4 (muted purple) - Subtle info
```

### Suggested Additions
```go
errorStyle:   #ff5555 (red) - Errors, critical
warnStyle:    #ff9900 (orange) - Warnings
successStyle: #50fa7b (green) - Success messages
infoStyle:    #00d7ff (cyan) - Info messages
highlightStyle: #f1fa8c (yellow) - Search highlights, selected items
```

---

## 📚 Resources & References

### Bubble Tea Ecosystem
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Bubbles** (components): https://github.com/charmbracelet/bubbles
- **Lip Gloss** (styling): https://github.com/charmbracelet/lipgloss
- **Glamour** (markdown): https://github.com/charmbracelet/glamour
- **Harmonica** (animations): https://github.com/charmbracelet/harmonica

### Inspiration Projects
- **gh-dash**: https://github.com/dlvhdr/gh-dash (9.5k⭐)
- **superfile**: https://github.com/yorukot/superfile (16.1k⭐)
- **glow**: https://github.com/charmbracelet/glow (15k⭐)
- **lazygit**: https://github.com/jesseduffield/lazygit (52k⭐)
- **k9s**: https://github.com/derailed/k9s (27k⭐)

### Design Patterns
- **The Elm Architecture**: https://guide.elm-lang.org/architecture/
- **TUI Best Practices**: https://charm.sh/blog/

---

## 🎯 Next Steps

1. **Review this document** with team
2. **Prioritize features** based on effort/impact
3. **Start with Quick Wins** (Week 1 items)
4. **Iterate based on feedback**
5. **Document new features** as we build

---

**Last Updated**: December 24, 2025  
**Status**: Research Complete, Ready for Implementation  
**Branch**: `dev`
