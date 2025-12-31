package config

import "github.com/charmbracelet/lipgloss"

// Theme defines color scheme for the UI
type Theme struct {
	Name string
	
	// Header colors
	TitleColor      lipgloss.Color
	LogoColor       lipgloss.Color
	StatusColor     lipgloss.Color
	
	// Agent status colors
	SessionColor    lipgloss.Color // Active session icon
	BeaconColor     lipgloss.Color // Beacon icon
	DeadColor       lipgloss.Color // Dead agent icon
	
	// Agent info colors
	PrivilegedUser  lipgloss.Color // Privileged username
	NormalUser      lipgloss.Color // Normal username
	HostnameColor   lipgloss.Color // Hostname
	
	// Protocol colors
	ProtocolMTLS    lipgloss.Color
	ProtocolHTTP    lipgloss.Color
	ProtocolDNS     lipgloss.Color
	ProtocolTCP     lipgloss.Color
	ProtocolDefault lipgloss.Color
	
	// Badge colors
	NewBadgeColor   lipgloss.Color
	PrivBadgeColor  lipgloss.Color
	
	// Panel colors
	TacticalBorder  lipgloss.Color
	TacticalSection lipgloss.Color
	TacticalValue   lipgloss.Color
	TacticalMuted   lipgloss.Color
	
	// Stats colors
	StatsColor      lipgloss.Color
	SeparatorColor  lipgloss.Color
	HelpColor       lipgloss.Color
	
	// Background colors (NEW)
	SessionBg       lipgloss.Color // Background for session agents
	BeaconBg        lipgloss.Color // Background for beacon agents
	DeadBg          lipgloss.Color // Background for dead agents
	PrivilegedBg    lipgloss.Color // Background highlight for privileged users
	NewBg           lipgloss.Color // Background highlight for new agents
	ProtocolBg      lipgloss.Color // Background for protocol boxes
	TacticalPanelBg lipgloss.Color // Background for tactical panel
	HeaderBg        lipgloss.Color // Background for header section
}

// Available themes
var themes = []Theme{
	defaultTheme(),
	rainbowTheme(),
	cyberpunkTheme(),
	matrixTheme(),
	tacticalTheme(),
	pastelTheme(),
	heatmapTheme(),
	lipGlossTheme(),
	nordTheme(),
	gruvboxTheme(),
	tokyoNightTheme(),
	monokaiTheme(),
	catppuccinMochaTheme(),
	catppuccinMacchiatoTheme(),
	catppuccinFrappeTheme(),
	// catppuccinLatteTheme(), // Light theme - disabled (doesn't work on dark terminals)
}

// defaultTheme - Current Dracula-inspired theme (DEFAULT)
func defaultTheme() Theme {
	return Theme{
		Name:            "Default (Dracula)",
		TitleColor:      lipgloss.Color("#00d7ff"),
		LogoColor:       lipgloss.Color("#d75fff"),
		StatusColor:     lipgloss.Color("#888888"),
		SessionColor:    lipgloss.Color("#00ff00"),
		BeaconColor:     lipgloss.Color("#ffff00"),
		DeadColor:       lipgloss.Color("#626262"),
		PrivilegedUser:  lipgloss.Color("#ff5555"),
		NormalUser:      lipgloss.Color("#50fa7b"),
		HostnameColor:   lipgloss.Color("#50fa7b"),
		ProtocolMTLS:    lipgloss.Color("#8be9fd"),
		ProtocolHTTP:    lipgloss.Color("#8be9fd"),
		ProtocolDNS:     lipgloss.Color("#8be9fd"),
		ProtocolTCP:     lipgloss.Color("#8be9fd"),
		ProtocolDefault: lipgloss.Color("#8be9fd"),
		NewBadgeColor:   lipgloss.Color("#f1fa8c"),
		PrivBadgeColor:  lipgloss.Color("#ff79c6"),
		TacticalBorder:  lipgloss.Color("#00d7ff"),
		TacticalSection: lipgloss.Color("#f1fa8c"),
		TacticalValue:   lipgloss.Color("#50fa7b"),
		TacticalMuted:   lipgloss.Color("#6272a4"),
		StatsColor:      lipgloss.Color("#00d7ff"),
		SeparatorColor:  lipgloss.Color("#444444"),
		HelpColor:       lipgloss.Color("#626262"),
		// Background colors
		SessionBg:       lipgloss.Color("#1a3a1a"),    // Dark green tint
		BeaconBg:        lipgloss.Color("#3a3a1a"),    // Dark yellow tint
		DeadBg:          lipgloss.Color("#2a2a2a"),    // Darker gray
		PrivilegedBg:    lipgloss.Color("#3a1a1a"),    // Dark red tint
		NewBg:           lipgloss.Color("#3a3a00"),    // Dark yellow glow
		ProtocolBg:      lipgloss.Color("#1a2a3a"),    // Dark blue for protocol boxes
		TacticalPanelBg: lipgloss.Color("#1a1a2a"),    // Dark purple tint
		HeaderBg:        lipgloss.Color("#1a1a1a"),    // Subtle header background
	}
}

// rainbowTheme - Vibrant rainbow colors
func rainbowTheme() Theme {
	return Theme{
		Name:            "Rainbow",
		TitleColor:      lipgloss.Color("#ff00ff"), // Magenta
		LogoColor:       lipgloss.Color("#ff0080"),
		StatusColor:     lipgloss.Color("#aaaaaa"),
		SessionColor:    lipgloss.Color("#00ff00"), // Bright green
		BeaconColor:     lipgloss.Color("#ff8800"), // Orange
		DeadColor:       lipgloss.Color("#ff0000"), // Red
		PrivilegedUser:  lipgloss.Color("#ffd700"), // Gold
		NormalUser:      lipgloss.Color("#00ffff"), // Cyan
		HostnameColor:   lipgloss.Color("#00ff7f"),
		ProtocolMTLS:    lipgloss.Color("#9d4edd"), // Purple
		ProtocolHTTP:    lipgloss.Color("#4cc9f0"), // Blue
		ProtocolDNS:     lipgloss.Color("#06ffa5"), // Green
		ProtocolTCP:     lipgloss.Color("#ff006e"), // Pink
		ProtocolDefault: lipgloss.Color("#ffbe0b"), // Yellow
		NewBadgeColor:   lipgloss.Color("#ffff00"), // Bright yellow
		PrivBadgeColor:  lipgloss.Color("#ffd700"),
		TacticalBorder:  lipgloss.Color("#ff00ff"),
		TacticalSection: lipgloss.Color("#ff8800"),
		TacticalValue:   lipgloss.Color("#00ff00"),
		TacticalMuted:   lipgloss.Color("#888888"),
		StatsColor:      lipgloss.Color("#ff00ff"),
		SeparatorColor:  lipgloss.Color("#ff8800"),
		HelpColor:       lipgloss.Color("#888888"),
		// Background colors - vibrant but subtle
		SessionBg:       lipgloss.Color("#0a2a0a"),
		BeaconBg:        lipgloss.Color("#2a1a0a"),
		DeadBg:          lipgloss.Color("#2a0a0a"),
		PrivilegedBg:    lipgloss.Color("#2a2a00"),
		NewBg:           lipgloss.Color("#2a2a00"),
		ProtocolBg:      lipgloss.Color("#1a0a2a"),
		TacticalPanelBg: lipgloss.Color("#2a002a"),
		HeaderBg:        lipgloss.Color("#1a001a"),
	}
}

// cyberpunkTheme - Neon cyberpunk aesthetic
func cyberpunkTheme() Theme {
	return Theme{
		Name:            "Cyberpunk",
		TitleColor:      lipgloss.Color("#ff006e"), // Neon pink
		LogoColor:       lipgloss.Color("#8338ec"), // Neon purple
		StatusColor:     lipgloss.Color("#a0a0a0"),
		SessionColor:    lipgloss.Color("#00b4d8"), // Electric blue
		BeaconColor:     lipgloss.Color("#ff006e"), // Hot pink
		DeadColor:       lipgloss.Color("#3c096c"), // Dark purple
		PrivilegedUser:  lipgloss.Color("#00f5ff"), // Bright cyan
		NormalUser:      lipgloss.Color("#ffff00"), // Neon yellow
		HostnameColor:   lipgloss.Color("#00f5ff"),
		ProtocolMTLS:    lipgloss.Color("#39ff14"), // Neon green
		ProtocolHTTP:    lipgloss.Color("#ff006e"),
		ProtocolDNS:     lipgloss.Color("#00b4d8"),
		ProtocolTCP:     lipgloss.Color("#8338ec"),
		ProtocolDefault: lipgloss.Color("#39ff14"),
		NewBadgeColor:   lipgloss.Color("#ffff00"),
		PrivBadgeColor:  lipgloss.Color("#ff006e"),
		TacticalBorder:  lipgloss.Color("#39ff14"),
		TacticalSection: lipgloss.Color("#ff006e"),
		TacticalValue:   lipgloss.Color("#00f5ff"),
		TacticalMuted:   lipgloss.Color("#6a0dad"),
		StatsColor:      lipgloss.Color("#ff006e"),
		SeparatorColor:  lipgloss.Color("#ff00ff"),
		HelpColor:       lipgloss.Color("#8338ec"),
		// Background colors - cyberpunk neon glow
		SessionBg:       lipgloss.Color("#0a1a2a"),
		BeaconBg:        lipgloss.Color("#2a0a1a"),
		DeadBg:          lipgloss.Color("#1a0a1a"),
		PrivilegedBg:    lipgloss.Color("#0a2a2a"),
		NewBg:           lipgloss.Color("#2a2a0a"),
		ProtocolBg:      lipgloss.Color("#0a1a0a"),
		TacticalPanelBg: lipgloss.Color("#1a0a2a"),
		HeaderBg:        lipgloss.Color("#0a0a1a"),
	}
}

// matrixTheme - Matrix green theme
func matrixTheme() Theme {
	return Theme{
		Name:            "Matrix",
		TitleColor:      lipgloss.Color("#00ff41"), // Matrix green
		LogoColor:       lipgloss.Color("#00ff41"),
		StatusColor:     lipgloss.Color("#90ee90"),
		SessionColor:    lipgloss.Color("#00ff41"), // Bright green
		BeaconColor:     lipgloss.Color("#adff2f"), // Yellow-green
		DeadColor:       lipgloss.Color("#0a3d0a"), // Dark green
		PrivilegedUser:  lipgloss.Color("#ffd700"), // Gold accent
		NormalUser:      lipgloss.Color("#90ee90"), // Light green
		HostnameColor:   lipgloss.Color("#90ee90"),
		ProtocolMTLS:    lipgloss.Color("#00ff41"),
		ProtocolHTTP:    lipgloss.Color("#76ff03"),
		ProtocolDNS:     lipgloss.Color("#adff2f"),
		ProtocolTCP:     lipgloss.Color("#90ee90"),
		ProtocolDefault: lipgloss.Color("#00ff41"),
		NewBadgeColor:   lipgloss.Color("#76ff03"), // Lime green
		PrivBadgeColor:  lipgloss.Color("#ffd700"),
		TacticalBorder:  lipgloss.Color("#00ff41"),
		TacticalSection: lipgloss.Color("#ffd700"),
		TacticalValue:   lipgloss.Color("#76ff03"),
		TacticalMuted:   lipgloss.Color("#76ff03"), // Changed from #2d5016 to #76ff03 (lime green) for better visibility
		StatsColor:      lipgloss.Color("#00ff41"),
		SeparatorColor:  lipgloss.Color("#00ff41"),
		HelpColor:       lipgloss.Color("#90ee90"),
		// Background colors - matrix green tints
		SessionBg:       lipgloss.Color("#0a2a0a"),
		BeaconBg:        lipgloss.Color("#1a2a0a"),
		DeadBg:          lipgloss.Color("#0a1a0a"),
		PrivilegedBg:    lipgloss.Color("#2a2a0a"),
		NewBg:           lipgloss.Color("#1a2a00"),
		ProtocolBg:      lipgloss.Color("#0a1a0a"),
		TacticalPanelBg: lipgloss.Color("#0a2a0a"),
		HeaderBg:        lipgloss.Color("#0a1a0a"),
	}
}

// tacticalTheme - Military tactical colors
func tacticalTheme() Theme {
	return Theme{
		Name:            "Tactical",
		TitleColor:      lipgloss.Color("#ff6b35"), // Orange
		LogoColor:       lipgloss.Color("#4a7c59"), // Olive green
		StatusColor:     lipgloss.Color("#a0a0a0"),
		SessionColor:    lipgloss.Color("#06d6a0"), // Green
		BeaconColor:     lipgloss.Color("#ffd60a"), // Yellow
		DeadColor:       lipgloss.Color("#e71d36"), // Red
		PrivilegedUser:  lipgloss.Color("#ffb700"), // Gold
		NormalUser:      lipgloss.Color("#90e0ef"), // Cyan
		HostnameColor:   lipgloss.Color("#90e0ef"),
		ProtocolMTLS:    lipgloss.Color("#457b9d"), // Steel blue
		ProtocolHTTP:    lipgloss.Color("#06d6a0"), // Teal
		ProtocolDNS:     lipgloss.Color("#ff9f1c"), // Orange
		ProtocolTCP:     lipgloss.Color("#4a7c59"),
		ProtocolDefault: lipgloss.Color("#457b9d"),
		NewBadgeColor:   lipgloss.Color("#ffd60a"),
		PrivBadgeColor:  lipgloss.Color("#ffb700"),
		TacticalBorder:  lipgloss.Color("#ff6b35"),
		TacticalSection: lipgloss.Color("#06d6a0"),
		TacticalValue:   lipgloss.Color("#ffd60a"),
		TacticalMuted:   lipgloss.Color("#5a6a62"),
		StatsColor:      lipgloss.Color("#ff6b35"),
		SeparatorColor:  lipgloss.Color("#ff6b35"),
		HelpColor:       lipgloss.Color("#a0a0a0"),
		// Background colors - military tactical tints
		SessionBg:       lipgloss.Color("#0a2a1a"),
		BeaconBg:        lipgloss.Color("#2a2a0a"),
		DeadBg:          lipgloss.Color("#2a0a0a"),
		PrivilegedBg:    lipgloss.Color("#2a1a0a"),
		NewBg:           lipgloss.Color("#2a2a00"),
		ProtocolBg:      lipgloss.Color("#0a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1a2a1a"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// pastelTheme - Soft pastel colors
func pastelTheme() Theme {
	return Theme{
		Name:            "Pastel",
		TitleColor:      lipgloss.Color("#ff99c8"), // Soft pink
		LogoColor:       lipgloss.Color("#a9def9"), // Lavender
		StatusColor:     lipgloss.Color("#b0b0b0"),
		SessionColor:    lipgloss.Color("#b5e48c"), // Mint green
		BeaconColor:     lipgloss.Color("#ffb5a7"), // Peach
		DeadColor:       lipgloss.Color("#9d8189"), // Dusty purple
		PrivilegedUser:  lipgloss.Color("#f4d58d"), // Soft gold
		NormalUser:      lipgloss.Color("#8bd3dd"), // Aqua
		HostnameColor:   lipgloss.Color("#8bd3dd"),
		ProtocolMTLS:    lipgloss.Color("#90dbf4"), // Soft blue
		ProtocolHTTP:    lipgloss.Color("#a9def9"),
		ProtocolDNS:     lipgloss.Color("#b5e48c"),
		ProtocolTCP:     lipgloss.Color("#ffb5a7"),
		ProtocolDefault: lipgloss.Color("#90dbf4"),
		NewBadgeColor:   lipgloss.Color("#f4d58d"),
		PrivBadgeColor:  lipgloss.Color("#ff99c8"),
		TacticalBorder:  lipgloss.Color("#ff99c8"),
		TacticalSection: lipgloss.Color("#a9def9"),
		TacticalValue:   lipgloss.Color("#b5e48c"),
		TacticalMuted:   lipgloss.Color("#7a7a7a"),
		StatsColor:      lipgloss.Color("#ff99c8"),
		SeparatorColor:  lipgloss.Color("#a9def9"),
		HelpColor:       lipgloss.Color("#b0b0b0"),
		// Background colors - soft pastel tints
		SessionBg:       lipgloss.Color("#1a2a1a"),
		BeaconBg:        lipgloss.Color("#2a1a1a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a2a1a"),
		NewBg:           lipgloss.Color("#2a2a00"),
		ProtocolBg:      lipgloss.Color("#1a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1a1a2a"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// heatmapTheme - Heat map colors based on priority
func heatmapTheme() Theme {
	return Theme{
		Name:            "Heatmap",
		TitleColor:      lipgloss.Color("#ffffff"), // White
		LogoColor:       lipgloss.Color("#ff0000"), // Red
		StatusColor:     lipgloss.Color("#aaaaaa"),
		SessionColor:    lipgloss.Color("#ff0000"), // Red - highest priority
		BeaconColor:     lipgloss.Color("#ffff00"), // Yellow
		DeadColor:       lipgloss.Color("#333333"), // Dark gray
		PrivilegedUser:  lipgloss.Color("#ff0000"), // Red - high priority
		NormalUser:      lipgloss.Color("#ff8800"), // Orange
		HostnameColor:   lipgloss.Color("#ff8800"),
		ProtocolMTLS:    lipgloss.Color("#ff4444"),
		ProtocolHTTP:    lipgloss.Color("#ff8800"),
		ProtocolDNS:     lipgloss.Color("#ffff00"),
		ProtocolTCP:     lipgloss.Color("#0096ff"),
		ProtocolDefault: lipgloss.Color("#888888"),
		NewBadgeColor:   lipgloss.Color("#ff0000"),
		PrivBadgeColor:  lipgloss.Color("#ff0000"),
		TacticalBorder:  lipgloss.Color("#ff0000"),
		TacticalSection: lipgloss.Color("#ff8800"),
		TacticalValue:   lipgloss.Color("#ffff00"),
		TacticalMuted:   lipgloss.Color("#666666"),
		StatsColor:      lipgloss.Color("#ff4444"),
		SeparatorColor:  lipgloss.Color("#ff8800"),
		HelpColor:       lipgloss.Color("#aaaaaa"),
		// Background colors - heat gradient
		SessionBg:       lipgloss.Color("#2a0a0a"),
		BeaconBg:        lipgloss.Color("#2a2a0a"),
		DeadBg:          lipgloss.Color("#1a1a1a"),
		PrivilegedBg:    lipgloss.Color("#2a0000"),
		NewBg:           lipgloss.Color("#2a0a00"),
		ProtocolBg:      lipgloss.Color("#1a0a0a"),
		TacticalPanelBg: lipgloss.Color("#2a1a1a"),
		HeaderBg:        lipgloss.Color("#1a0a0a"),
	}
}

// lipGlossTheme - Inspired by Charm's Lip Gloss aesthetic with purple/pink/blue gradients
func lipGlossTheme() Theme {
	return Theme{
		Name:            "Lip Gloss",
		TitleColor:      lipgloss.Color("#d946ef"), // Bright magenta/fuchsia
		LogoColor:       lipgloss.Color("#a855f7"), // Purple
		StatusColor:     lipgloss.Color("#c084fc"), // Light purple
		SessionColor:    lipgloss.Color("#22d3ee"), // Cyan blue
		BeaconColor:     lipgloss.Color("#f0abfc"), // Pink
		DeadColor:       lipgloss.Color("#6b7280"), // Gray
		PrivilegedUser:  lipgloss.Color("#e879f9"), // Bright pink
		NormalUser:      lipgloss.Color("#60a5fa"), // Blue
		HostnameColor:   lipgloss.Color("#60a5fa"),
		ProtocolMTLS:    lipgloss.Color("#a78bfa"), // Violet
		ProtocolHTTP:    lipgloss.Color("#38bdf8"), // Sky blue
		ProtocolDNS:     lipgloss.Color("#2dd4bf"), // Teal
		ProtocolTCP:     lipgloss.Color("#c084fc"), // Purple
		ProtocolDefault: lipgloss.Color("#818cf8"), // Indigo
		NewBadgeColor:   lipgloss.Color("#fbbf24"), // Amber yellow
		PrivBadgeColor:  lipgloss.Color("#f9a8d4"), // Pink
		TacticalBorder:  lipgloss.Color("#d946ef"), // Fuchsia
		TacticalSection: lipgloss.Color("#a78bfa"), // Violet
		TacticalValue:   lipgloss.Color("#22d3ee"), // Cyan
		TacticalMuted:   lipgloss.Color("#9ca3af"), // Gray
		StatsColor:      lipgloss.Color("#d946ef"), // Fuchsia
		SeparatorColor:  lipgloss.Color("#8b5cf6"), // Purple
		HelpColor:       lipgloss.Color("#a1a1aa"), // Light gray
		// Background colors - gradient purple/pink/blue tints
		SessionBg:       lipgloss.Color("#1a1a2a"),
		BeaconBg:        lipgloss.Color("#2a1a2a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a1a2a"),
		NewBg:           lipgloss.Color("#2a2a1a"),
		ProtocolBg:      lipgloss.Color("#1a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1a1a2a"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// GetTheme returns theme by index
func GetTheme(index int) Theme {
	if index < 0 || index >= len(themes) {
		return defaultTheme()
	}
	return themes[index]
}

// GetThemeCount returns total number of themes
func GetThemeCount() int {
	return len(themes)
}

// nordTheme - Popular Nordic-inspired theme with cool blues and grays
func nordTheme() Theme {
	return Theme{
		Name:            "Nord",
		TitleColor:      lipgloss.Color("#88c0d0"), // Frost cyan
		LogoColor:       lipgloss.Color("#81a1c1"), // Frost blue
		StatusColor:     lipgloss.Color("#d8dee9"), // Snow storm light
		SessionColor:    lipgloss.Color("#a3be8c"), // Aurora green
		BeaconColor:     lipgloss.Color("#ebcb8b"), // Aurora yellow
		DeadColor:       lipgloss.Color("#4c566a"), // Polar night dark
		PrivilegedUser:  lipgloss.Color("#bf616a"), // Aurora red
		NormalUser:      lipgloss.Color("#8fbcbb"), // Frost teal
		HostnameColor:   lipgloss.Color("#8fbcbb"), // Frost teal
		ProtocolMTLS:    lipgloss.Color("#5e81ac"), // Frost blue
		ProtocolHTTP:    lipgloss.Color("#88c0d0"), // Frost cyan
		ProtocolDNS:     lipgloss.Color("#81a1c1"), // Frost blue
		ProtocolTCP:     lipgloss.Color("#b48ead"), // Aurora purple
		ProtocolDefault: lipgloss.Color("#8fbcbb"), // Frost teal
		NewBadgeColor:   lipgloss.Color("#ebcb8b"), // Aurora yellow
		PrivBadgeColor:  lipgloss.Color("#bf616a"), // Aurora red
		TacticalBorder:  lipgloss.Color("#88c0d0"), // Frost cyan
		TacticalSection: lipgloss.Color("#81a1c1"), // Frost blue
		TacticalValue:   lipgloss.Color("#a3be8c"), // Aurora green
		TacticalMuted:   lipgloss.Color("#4c566a"), // Polar night
		StatsColor:      lipgloss.Color("#88c0d0"), // Frost cyan
		SeparatorColor:  lipgloss.Color("#434c5e"), // Polar night
		HelpColor:       lipgloss.Color("#616e88"), // Polar night light
		// Background colors - cool Nordic tints
		SessionBg:       lipgloss.Color("#1a2a2a"),
		BeaconBg:        lipgloss.Color("#2a2a1a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a1a1a"),
		NewBg:           lipgloss.Color("#2a2a00"),
		ProtocolBg:      lipgloss.Color("#1a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1a2028"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// gruvboxTheme - Retro warm color scheme inspired by Gruvbox
func gruvboxTheme() Theme {
	return Theme{
		Name:            "Gruvbox",
		TitleColor:      lipgloss.Color("#fe8019"), // Bright orange
		LogoColor:       lipgloss.Color("#d65d0e"), // Orange
		StatusColor:     lipgloss.Color("#a89984"), // Gray
		SessionColor:    lipgloss.Color("#b8bb26"), // Bright green
		BeaconColor:     lipgloss.Color("#fabd2f"), // Bright yellow
		DeadColor:       lipgloss.Color("#665c54"), // Dark gray
		PrivilegedUser:  lipgloss.Color("#fb4934"), // Bright red
		NormalUser:      lipgloss.Color("#83a598"), // Bright blue
		HostnameColor:   lipgloss.Color("#83a598"), // Bright blue
		ProtocolMTLS:    lipgloss.Color("#d3869b"), // Bright purple
		ProtocolHTTP:    lipgloss.Color("#8ec07c"), // Bright aqua
		ProtocolDNS:     lipgloss.Color("#fabd2f"), // Bright yellow
		ProtocolTCP:     lipgloss.Color("#fe8019"), // Bright orange
		ProtocolDefault: lipgloss.Color("#83a598"), // Bright blue
		NewBadgeColor:   lipgloss.Color("#fabd2f"), // Bright yellow
		PrivBadgeColor:  lipgloss.Color("#fb4934"), // Bright red
		TacticalBorder:  lipgloss.Color("#fe8019"), // Bright orange
		TacticalSection: lipgloss.Color("#b8bb26"), // Bright green
		TacticalValue:   lipgloss.Color("#fabd2f"), // Bright yellow
		TacticalMuted:   lipgloss.Color("#7c6f64"), // Gray
		StatsColor:      lipgloss.Color("#fe8019"), // Bright orange
		SeparatorColor:  lipgloss.Color("#665c54"), // Dark gray
		HelpColor:       lipgloss.Color("#a89984"), // Gray
		// Background colors - warm retro tints
		SessionBg:       lipgloss.Color("#1a2a1a"),
		BeaconBg:        lipgloss.Color("#2a2a0a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a1a0a"),
		NewBg:           lipgloss.Color("#2a2a00"),
		ProtocolBg:      lipgloss.Color("#1a2a1a"),
		TacticalPanelBg: lipgloss.Color("#1d2021"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// tokyoNightTheme - Modern dark theme with purple, blue, and pink accents
func tokyoNightTheme() Theme {
	return Theme{
		Name:            "Tokyo Night",
		TitleColor:      lipgloss.Color("#7aa2f7"), // Blue
		LogoColor:       lipgloss.Color("#bb9af7"), // Purple
		StatusColor:     lipgloss.Color("#a9b1d6"), // Light gray-blue
		SessionColor:    lipgloss.Color("#9ece6a"), // Green
		BeaconColor:     lipgloss.Color("#e0af68"), // Yellow
		DeadColor:       lipgloss.Color("#565f89"), // Dark blue-gray
		PrivilegedUser:  lipgloss.Color("#f7768e"), // Red
		NormalUser:      lipgloss.Color("#7dcfff"), // Cyan
		HostnameColor:   lipgloss.Color("#7dcfff"), // Cyan
		ProtocolMTLS:    lipgloss.Color("#bb9af7"), // Purple
		ProtocolHTTP:    lipgloss.Color("#7aa2f7"), // Blue
		ProtocolDNS:     lipgloss.Color("#2ac3de"), // Teal
		ProtocolTCP:     lipgloss.Color("#ff9e64"), // Orange
		ProtocolDefault: lipgloss.Color("#7dcfff"), // Cyan
		NewBadgeColor:   lipgloss.Color("#e0af68"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#f7768e"), // Red
		TacticalBorder:  lipgloss.Color("#7aa2f7"), // Blue
		TacticalSection: lipgloss.Color("#bb9af7"), // Purple
		TacticalValue:   lipgloss.Color("#9ece6a"), // Green
		TacticalMuted:   lipgloss.Color("#565f89"), // Dark blue-gray
		StatsColor:      lipgloss.Color("#7aa2f7"), // Blue
		SeparatorColor:  lipgloss.Color("#414868"), // Dark gray
		HelpColor:       lipgloss.Color("#565f89"), // Dark blue-gray
		// Background colors - Tokyo night tints
		SessionBg:       lipgloss.Color("#1a2a1a"),
		BeaconBg:        lipgloss.Color("#2a2a1a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a1a1a"),
		NewBg:           lipgloss.Color("#2a2a1a"),
		ProtocolBg:      lipgloss.Color("#1a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1a1b26"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// monokaiTheme - Classic Monokai theme with high contrast
func monokaiTheme() Theme {
	return Theme{
		Name:            "Monokai",
		TitleColor:      lipgloss.Color("#66d9ef"), // Cyan
		LogoColor:       lipgloss.Color("#ae81ff"), // Purple
		StatusColor:     lipgloss.Color("#75715e"), // Gray
		SessionColor:    lipgloss.Color("#a6e22e"), // Green
		BeaconColor:     lipgloss.Color("#e6db74"), // Yellow
		DeadColor:       lipgloss.Color("#75715e"), // Gray
		PrivilegedUser:  lipgloss.Color("#f92672"), // Pink/Red
		NormalUser:      lipgloss.Color("#66d9ef"), // Cyan
		HostnameColor:   lipgloss.Color("#66d9ef"), // Cyan
		ProtocolMTLS:    lipgloss.Color("#ae81ff"), // Purple
		ProtocolHTTP:    lipgloss.Color("#66d9ef"), // Cyan
		ProtocolDNS:     lipgloss.Color("#a6e22e"), // Green
		ProtocolTCP:     lipgloss.Color("#fd971f"), // Orange
		ProtocolDefault: lipgloss.Color("#66d9ef"), // Cyan
		NewBadgeColor:   lipgloss.Color("#e6db74"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#f92672"), // Pink/Red
		TacticalBorder:  lipgloss.Color("#66d9ef"), // Cyan
		TacticalSection: lipgloss.Color("#ae81ff"), // Purple
		TacticalValue:   lipgloss.Color("#a6e22e"), // Green
		TacticalMuted:   lipgloss.Color("#75715e"), // Gray
		StatsColor:      lipgloss.Color("#66d9ef"), // Cyan
		SeparatorColor:  lipgloss.Color("#49483e"), // Dark gray
		HelpColor:       lipgloss.Color("#75715e"), // Gray
		// Background colors - Monokai dark tints
		SessionBg:       lipgloss.Color("#1a2a1a"),
		BeaconBg:        lipgloss.Color("#2a2a1a"),
		DeadBg:          lipgloss.Color("#2a2a2a"),
		PrivilegedBg:    lipgloss.Color("#2a1a1a"),
		NewBg:           lipgloss.Color("#2a2a1a"),
		ProtocolBg:      lipgloss.Color("#1a1a2a"),
		TacticalPanelBg: lipgloss.Color("#1e1f1c"),
		HeaderBg:        lipgloss.Color("#1a1a1a"),
	}
}

// catppuccinMochaTheme - Catppuccin Mocha (dark, warm)
func catppuccinMochaTheme() Theme {
	return Theme{
		Name:            "Catppuccin Mocha",
		TitleColor:      lipgloss.Color("#89b4fa"), // Blue
		LogoColor:       lipgloss.Color("#cba6f7"), // Mauve
		StatusColor:     lipgloss.Color("#bac2de"), // Subtext1
		SessionColor:    lipgloss.Color("#a6e3a1"), // Green
		BeaconColor:     lipgloss.Color("#f9e2af"), // Yellow
		DeadColor:       lipgloss.Color("#6c7086"), // Surface2 (muted)
		PrivilegedUser:  lipgloss.Color("#f38ba8"), // Red
		NormalUser:      lipgloss.Color("#89dceb"), // Sky
		HostnameColor:   lipgloss.Color("#89dceb"), // Sky
		ProtocolMTLS:    lipgloss.Color("#cba6f7"), // Mauve
		ProtocolHTTP:    lipgloss.Color("#89b4fa"), // Blue
		ProtocolDNS:     lipgloss.Color("#94e2d5"), // Teal
		ProtocolTCP:     lipgloss.Color("#fab387"), // Peach
		ProtocolDefault: lipgloss.Color("#89dceb"), // Sky
		NewBadgeColor:   lipgloss.Color("#f9e2af"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#f38ba8"), // Red
		TacticalBorder:  lipgloss.Color("#89b4fa"), // Blue
		TacticalSection: lipgloss.Color("#cba6f7"), // Mauve
		TacticalValue:   lipgloss.Color("#a6e3a1"), // Green
		TacticalMuted:   lipgloss.Color("#585b70"), // Surface1
		StatsColor:      lipgloss.Color("#89b4fa"), // Blue
		SeparatorColor:  lipgloss.Color("#45475a"), // Surface0
		HelpColor:       lipgloss.Color("#6c7086"), // Surface2
		// Background colors - Catppuccin Mocha authentic backgrounds
		SessionBg:       lipgloss.Color("#1e1e2e"), // Base with green tint
		BeaconBg:        lipgloss.Color("#1e1e2e"), // Base with yellow tint
		DeadBg:          lipgloss.Color("#181825"), // Mantle
		PrivilegedBg:    lipgloss.Color("#1e1e2e"), // Base with red tint
		NewBg:           lipgloss.Color("#1e1e2e"), // Base with yellow tint
		ProtocolBg:      lipgloss.Color("#181825"), // Mantle
		TacticalPanelBg: lipgloss.Color("#1e1e2e"), // Base
		HeaderBg:        lipgloss.Color("#181825"), // Mantle
	}
}

// catppuccinMacchiatoTheme - Catppuccin Macchiato (dark, cool)
func catppuccinMacchiatoTheme() Theme {
	return Theme{
		Name:            "Catppuccin Macchiato",
		TitleColor:      lipgloss.Color("#8aadf4"), // Blue
		LogoColor:       lipgloss.Color("#c6a0f6"), // Mauve
		StatusColor:     lipgloss.Color("#b8c0e0"), // Subtext1
		SessionColor:    lipgloss.Color("#a6da95"), // Green
		BeaconColor:     lipgloss.Color("#eed49f"), // Yellow
		DeadColor:       lipgloss.Color("#5b6078"), // Surface2 (muted)
		PrivilegedUser:  lipgloss.Color("#ed8796"), // Red
		NormalUser:      lipgloss.Color("#91d7e3"), // Sky
		HostnameColor:   lipgloss.Color("#91d7e3"), // Sky
		ProtocolMTLS:    lipgloss.Color("#c6a0f6"), // Mauve
		ProtocolHTTP:    lipgloss.Color("#8aadf4"), // Blue
		ProtocolDNS:     lipgloss.Color("#8bd5ca"), // Teal
		ProtocolTCP:     lipgloss.Color("#f5a97f"), // Peach
		ProtocolDefault: lipgloss.Color("#91d7e3"), // Sky
		NewBadgeColor:   lipgloss.Color("#eed49f"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#ed8796"), // Red
		TacticalBorder:  lipgloss.Color("#8aadf4"), // Blue
		TacticalSection: lipgloss.Color("#c6a0f6"), // Mauve
		TacticalValue:   lipgloss.Color("#a6da95"), // Green
		TacticalMuted:   lipgloss.Color("#494d64"), // Surface1
		StatsColor:      lipgloss.Color("#8aadf4"), // Blue
		SeparatorColor:  lipgloss.Color("#363a4f"), // Surface0
		HelpColor:       lipgloss.Color("#5b6078"), // Surface2
		// Background colors - Catppuccin Macchiato authentic backgrounds
		SessionBg:       lipgloss.Color("#24273a"), // Base
		BeaconBg:        lipgloss.Color("#24273a"), // Base
		DeadBg:          lipgloss.Color("#1e2030"), // Mantle
		PrivilegedBg:    lipgloss.Color("#24273a"), // Base
		NewBg:           lipgloss.Color("#24273a"), // Base
		ProtocolBg:      lipgloss.Color("#1e2030"), // Mantle
		TacticalPanelBg: lipgloss.Color("#24273a"), // Base
		HeaderBg:        lipgloss.Color("#1e2030"), // Mantle
	}
}

// catppuccinFrappeTheme - Catppuccin Frappé (dark, balanced)
func catppuccinFrappeTheme() Theme {
	return Theme{
		Name:            "Catppuccin Frappé",
		TitleColor:      lipgloss.Color("#8caaee"), // Blue
		LogoColor:       lipgloss.Color("#ca9ee6"), // Mauve
		StatusColor:     lipgloss.Color("#b5bfe2"), // Subtext1
		SessionColor:    lipgloss.Color("#a6d189"), // Green
		BeaconColor:     lipgloss.Color("#e5c890"), // Yellow
		DeadColor:       lipgloss.Color("#626880"), // Surface2 (muted)
		PrivilegedUser:  lipgloss.Color("#e78284"), // Red
		NormalUser:      lipgloss.Color("#99d1db"), // Sky
		HostnameColor:   lipgloss.Color("#99d1db"), // Sky
		ProtocolMTLS:    lipgloss.Color("#ca9ee6"), // Mauve
		ProtocolHTTP:    lipgloss.Color("#8caaee"), // Blue
		ProtocolDNS:     lipgloss.Color("#81c8be"), // Teal
		ProtocolTCP:     lipgloss.Color("#ef9f76"), // Peach
		ProtocolDefault: lipgloss.Color("#99d1db"), // Sky
		NewBadgeColor:   lipgloss.Color("#e5c890"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#e78284"), // Red
		TacticalBorder:  lipgloss.Color("#8caaee"), // Blue
		TacticalSection: lipgloss.Color("#ca9ee6"), // Mauve
		TacticalValue:   lipgloss.Color("#a6d189"), // Green
		TacticalMuted:   lipgloss.Color("#51576d"), // Surface1
		StatsColor:      lipgloss.Color("#8caaee"), // Blue
		SeparatorColor:  lipgloss.Color("#414559"), // Surface0
		HelpColor:       lipgloss.Color("#626880"), // Surface2
		// Background colors - Catppuccin Frappé authentic backgrounds
		SessionBg:       lipgloss.Color("#303446"), // Base
		BeaconBg:        lipgloss.Color("#303446"), // Base
		DeadBg:          lipgloss.Color("#292c3c"), // Mantle
		PrivilegedBg:    lipgloss.Color("#303446"), // Base
		NewBg:           lipgloss.Color("#303446"), // Base
		ProtocolBg:      lipgloss.Color("#292c3c"), // Mantle
		TacticalPanelBg: lipgloss.Color("#303446"), // Base
		HeaderBg:        lipgloss.Color("#292c3c"), // Mantle
	}
}

// catppuccinLatteTheme - Catppuccin Latte (light theme)
func catppuccinLatteTheme() Theme {
	return Theme{
		Name:            "Catppuccin Latte",
		TitleColor:      lipgloss.Color("#1e66f5"), // Blue
		LogoColor:       lipgloss.Color("#8839ef"), // Mauve
		StatusColor:     lipgloss.Color("#5c5f77"), // Subtext1
		SessionColor:    lipgloss.Color("#40a02b"), // Green
		BeaconColor:     lipgloss.Color("#df8e1d"), // Yellow
		DeadColor:       lipgloss.Color("#9ca0b0"), // Surface2 (muted)
		PrivilegedUser:  lipgloss.Color("#d20f39"), // Red
		NormalUser:      lipgloss.Color("#04a5e5"), // Sky
		HostnameColor:   lipgloss.Color("#04a5e5"), // Sky
		ProtocolMTLS:    lipgloss.Color("#8839ef"), // Mauve
		ProtocolHTTP:    lipgloss.Color("#1e66f5"), // Blue
		ProtocolDNS:     lipgloss.Color("#179299"), // Teal
		ProtocolTCP:     lipgloss.Color("#fe640b"), // Peach
		ProtocolDefault: lipgloss.Color("#04a5e5"), // Sky
		NewBadgeColor:   lipgloss.Color("#df8e1d"), // Yellow
		PrivBadgeColor:  lipgloss.Color("#d20f39"), // Red
		TacticalBorder:  lipgloss.Color("#1e66f5"), // Blue
		TacticalSection: lipgloss.Color("#8839ef"), // Mauve
		TacticalValue:   lipgloss.Color("#40a02b"), // Green
		TacticalMuted:   lipgloss.Color("#acb0be"), // Surface1
		StatsColor:      lipgloss.Color("#1e66f5"), // Blue
		SeparatorColor:  lipgloss.Color("#ccd0da"), // Surface0
		HelpColor:       lipgloss.Color("#9ca0b0"), // Surface2
		// Background colors - Catppuccin Latte light backgrounds
		SessionBg:       lipgloss.Color("#eff1f5"), // Base
		BeaconBg:        lipgloss.Color("#eff1f5"), // Base
		DeadBg:          lipgloss.Color("#e6e9ef"), // Mantle
		PrivilegedBg:    lipgloss.Color("#eff1f5"), // Base
		NewBg:           lipgloss.Color("#eff1f5"), // Base
		ProtocolBg:      lipgloss.Color("#e6e9ef"), // Mantle
		TacticalPanelBg: lipgloss.Color("#eff1f5"), // Base
		HeaderBg:        lipgloss.Color("#e6e9ef"), // Mantle
	}
}
