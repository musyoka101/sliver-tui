# Nerd Font Icons Reference for Sliver Dashboard

## OS Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
|  | U+F17C | Linux/Tux | `"\uf17c"` or `""` |
|  | U+F179 | Apple | `"\uf179"` or `""` |
|  | U+F17A | Windows | `"\uf17a"` or `""` |
|  | U+E70E | Windows Alt | `"\ue70e"` or `""` |
|  | U+F17B | Android | `"\uf17b"` or `""` |

## Transport/Connection Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
|  | U+F023 | Lock (TLS/mTLS) | `"\uf023"` or `""` |
| 󰖟 | U+F0599 | Web/HTTP | `"\uf0599"` or `"󰖟"` |
| 󰖩 | U+F05A9 | Network | `"\uf05a9"` or `"󰖩"` |
| 󰚥 | U+F06A5 | DNS/Signal | `"\uf06a5"` or `"󰚥"` |
|  | U+F0C1 | Link/Connection | `"\uf0c1"` or `""` |
| 󰤨 | U+F0928 | Wifi/Wireless | `"\uf0928"` or `"󰤨"` |

## Agent Status Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
| 󰟀 | U+F07C0 | Computer/Laptop | `"\uf07c0"` or `"󰟀"` |
| 󰒋 | U+F048B | Server | `"\uf048b"` or `"󰒋"` |
|  | U+F0E8 | Bug/Implant | `"\uf0e8"` or `""` |
|  | U+2713 | Checkmark/Active | `"\u2713"` or `"✓"` |
|  | U+F00D | Dead/Offline | `"\uf00d"` or `""` |
| 󰚥 | U+F06A5 | Lightning/Active | `"\uf06a5"` or `"󰚥"` |
|  | U+F192 | Dot Circle/Beacon | `"\uf192"` or `""` |

## Privilege Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
| 󰀙 | U+F0019 | Alert/Warning | `"\uf0019"` or `"󰀙"` |
|  | U+F132 | Shield/Protected | `"\uf132"` or `""` |
| 󰯄 | U+F0BC4 | Crown/Admin | `"\uf0bc4"` or `"󰯄"` |
|  | U+F007 | User | `"\uf007"` or `""` |
|  | U+F0C0 | Users/Group | `"\uf0c0"` or `""` |

## View Type Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
|  | U+F1BB | Tree | `"\uf1bb"` or `""` |
|  | U+F466 | Box/Grid | `"\uf466"` or `""` |
|  | U+F0CE | Table/List | `"\uf0ce"` or `""` |
| 󰕮 | U+F056E | Dashboard | `"\uf056e"` or `"󰕮"` |
| 󰺮 | U+F0EAE | Network Map | `"\uf0eae"` or `"󰺮"` |
|  | U+F279 | Map/Topology | `"\uf279"` or `""` |

## Action Icons

| Icon | Unicode | Description | Go Code |
|------|---------|-------------|---------|
|  | U+F021 | Refresh | `"\uf021"` or `""` |
|  | U+F53F | Palette/Theme | `"\uf53f"` or `""` |
|  | U+F011 | Power/Quit | `"\uf011"` or `""` |
|  | U+F013 | Settings | `"\uf013"` or `""` |
| 󰋼 | U+F02FC | Help | `"\uf02fc"` or `"󰋼"` |
|  | U+F002 | Search | `"\uf002"` or `""` |

## Usage in Go

### Method 1: Direct Unicode String
```go
windowsIcon := "\uf17a"  // 
linuxIcon := "\uf17c"    // 
appleIcon := "\uf179"    // 
```

### Method 2: Raw String Literal (Easier)
```go
windowsIcon := ""
linuxIcon := ""
appleIcon := ""
```

### Method 3: Create Icon Constants
```go
const (
    IconWindows   = ""
    IconLinux     = ""
    IconApple     = ""
    IconAndroid   = ""
    IconLock      = ""
    IconNetwork   = "󰖩"
    IconServer    = "󰒋"
    IconComputer  = "󰟀"
    IconDead      = ""
    IconActive    = "󰚥"
)
```

## Terminal Requirements

Your terminal MUST:
1. Use a Nerd Font (you have JetBrainsMono Nerd Font ✓)
2. Support UTF-8 encoding
3. Support Unicode rendering

## Testing Icons

Run this command to test if icons render:
```bash
echo " Linux |  Windows |  Apple"
```

If you see the actual icons, you're good to go!
