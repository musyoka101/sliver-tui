# Screenshots

This directory contains screenshots of Sliver C2 TUI for the README.

## Adding Screenshots

To add screenshots to the main README:

1. **Take screenshots** of different views:
   - `box-view.png` - Box view (default)
   - `dashboard-view.png` - Dashboard view
   - `table-view.png` - Table view
   - `help-menu.png` - Help menu
   - `alerts.png` - Alert panel
   - `agent-details.png` - Agent details panel

2. **Save them here** in this `screenshots/` directory

3. **Update README.md** by replacing the comments:
   ```markdown
   ### Box View (Default)
   ![Box View](screenshots/box-view.png)
   
   ### Dashboard View
   ![Dashboard View](screenshots/dashboard-view.png)
   ```

## Tips for Good Screenshots

- Use a **clean terminal** with good contrast
- **Resize terminal** to reasonable size (not too small/large)
- **Show actual data** if possible (agents, alerts)
- Use **Matrix or Cyberpunk theme** for visual appeal
- **Crop appropriately** - remove unnecessary borders
- **Use PNG format** for better quality

## Recommended Terminal Settings

- Font: Fira Code, JetBrains Mono, or similar
- Size: 80x24 to 120x30 characters
- Theme: Matrix or Cyberpunk (press 't' to cycle)
- Enable: Unicode, 256 colors

## Screenshot Tools

### Linux
```bash
# Using gnome-screenshot
gnome-screenshot -a

# Using scrot
scrot -s screenshot.png

# Using flameshot
flameshot gui
```

### macOS
```bash
# Command + Shift + 4 (select area)
# Or use built-in Screenshot app
```

## Example File Names

- `box-view.png` - Default Box view
- `box-view-selected.png` - Box view with agent selected
- `dashboard-overview.png` - Dashboard page 1
- `dashboard-network.png` - Dashboard page 2
- `table-view.png` - Table view
- `help-menu.png` - Help menu open
- `help-menu-scrolled.png` - Help menu scrolled
- `alerts-active.png` - Active alerts showing
- `themes-comparison.png` - Multiple themes side-by-side

## Current Screenshots

Screenshots available in this directory:

- [x] box-view.png - Default Box view with agent list
- [x] table-view.png - Table view displaying agents in spreadsheet format
- [x] dashboard-view.png - Dashboard view with analytics panels
- [x] help-menu.png - Help menu with complete keyboard shortcuts
- [x] alert-panel.png - Alert panel showing notifications
- [x] agent-details-panel.png - Agent details panel when agent is selected

**Note:** Network Map view screenshot to be added in future update.
