package ui

// Terminal size constraints
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 20
)

// Panel layout ratios
const (
	LeftPanelWidthRatio    = 1.0 / 3.0  // Left panel is 1/3 of width
	TopPanelHeightRatio    = 2.0 / 3.0  // Top panels are 2/3 of height
	BottomPanelHeightRatio = 1.0 / 3.0  // Bottom panel is 1/3 of height
	FormWidthRatio         = 2.0 / 3.0  // Forms are 2/3 of screen width
	FormHeightRatio        = 2.0 / 3.0  // Forms are 2/3 of screen height
	DialogWidthRatio       = 3.0 / 4.0  // Dialogs are 3/4 of screen width
	DialogHeightRatio      = 3.0 / 4.0  // Dialogs are 3/4 of screen height
)

// UI spacing
const (
	TitleAndHelpHeight = 4  // Space reserved for title and help text
	MaxLogEntries      = 100
	MaxPathDisplayLen  = 50
)

// Progress bar
const (
	ProgressBarWidth = 40
)

// Visual indicators
const (
	IconActive       = "◆ "
	IconInactive     = "○ "
	IconSuccess      = "✓"
	IconWarning      = "⚠"
	IconError        = "✗"
	IconInfo         = "•"
	IconSnapshot     = "📸"
	IconBackup       = "💾"
	IconRestore      = "♻"
	IconFolder       = "📁"
	IconFile         = "📄"
)

// Status indicators
const (
	StatusHealthy = "healthy"
	StatusWarning = "warning"
	StatusError   = "error"
	StatusPending = "pending"
)

// Color scheme (muted, easy on the eyes)
const (
	ColorSuccess      = "#00AA00"
	ColorWarning      = "#FFAA00"
	ColorError        = "#FF0000"
	ColorInfo         = "#00AAFF"
	ColorDimmed       = "#666666"
	ColorPrimary      = "#00AA88"
	ColorSecondary    = "#666666"
	ColorBorder       = "#444444"
	ColorBorderActive = "#00CCAA"
)
