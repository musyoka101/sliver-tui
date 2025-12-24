# Theme System - Quick Reference

## Available Themes (Press 't' to cycle)

1. **Default (Dracula)** - Current theme
   - Cyan title, purple logo
   - Green sessions, yellow beacons, gray dead
   - Red privileged, green normal users
   - Cyan protocols

2. **Rainbow** - Vibrant multi-color
   - Magenta title
   - Bright green sessions, orange beacons, red dead
   - Gold privileged, cyan normal users
   - Purple MTLS, blue HTTP, green DNS, pink TCP

3. **Cyberpunk** - Neon aesthetic
   - Neon pink title, purple logo
   - Electric blue sessions, hot pink beacons, dark purple dead
   - Bright cyan privileged, neon yellow normal users
   - Neon green protocols

4. **Matrix** - Classic green terminal
   - Bright green title and logo
   - Matrix green sessions, yellow-green beacons
   - Gold privileged, light green normal users
   - All green protocols

5. **Tactical** - Military colors
   - Orange title, olive green logo
   - Green sessions, yellow beacons, red dead
   - Gold privileged, cyan normal users
   - Steel blue MTLS, teal HTTP, orange DNS

6. **Pastel** - Soft colors (easy on eyes)
   - Soft pink title, lavender logo
   - Mint green sessions, peach beacons, dusty purple dead
   - Soft gold privileged, aqua normal users
   - Soft blue protocols

7. **Heatmap** - Priority-based (red=high, blue=low)
   - White title, red logo
   - Red sessions (highest priority), yellow beacons
   - Red privileged, orange normal users
   - Red MTLS, orange HTTP, yellow DNS, blue TCP

## Controls

- **'t'** - Cycle through themes
- **'r'** - Refresh data
- **'↑↓'** or **'j/k'** - Scroll
- **'q'** - Quit

## Current Status

✅ All 7 themes implemented
✅ Theme switching with 't' hotkey
✅ Theme name displayed in status bar
✅ Default theme is Dracula (current design)
✅ All UI elements themed (title, agents, protocols, tactical panel, stats)

## Testing

Run: `./sliver-graph`

Then press 't' repeatedly to cycle through all themes and see which you like best!

The theme name is shown in the status bar at the top.
