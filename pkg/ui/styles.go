package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Enhanced color palette with better contrast and visual appeal
	colorPrimary   = lipgloss.Color("#00AA88") // Muted cyan/teal
	colorSecondary = lipgloss.Color("#666666") // Muted gray (no more purple!)
	colorSuccess   = lipgloss.Color("#00AA00") // Muted green (not too bright)
	colorWarning   = lipgloss.Color("#FFAA00") // Orange
	colorError     = lipgloss.Color("#FF0000") // Red
	colorInfo      = lipgloss.Color("#00AAFF") // Blue
	colorActive    = lipgloss.Color("#00CCAA") // Cyan - active elements
	colorDimmed    = lipgloss.Color("#666666") // Dimmed gray
	colorBorder    = lipgloss.Color("#444444") // Default border color
	colorBlack     = lipgloss.Color("#000000") // Black for text on colored backgrounds

	// Title styles - make it pop!
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			BorderBottom(true)

	// Panel styles - using black text on colored backgrounds for better readability
	PanelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlack).
			Background(colorPrimary).
			Padding(0, 1)

	PanelTitleActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBlack).
				Background(colorActive).
				Padding(0, 1)

	PanelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(1, 2)

	PanelBorderActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorActive).
				Padding(1, 2)

	// List item styles
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Faint(true)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(colorActive).
				Bold(true).
				Padding(0, 1).
				MarginLeft(1)

	// Status styles
	StatusHealthyStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	StatusWarningStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	// Help text style - polished bar
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			BorderTop(true)

	// Key binding styles
	KeyStyle = lipgloss.NewStyle().
			Foreground(colorActive).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)
)

// StatusStyle returns the appropriate style for a status string
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "healthy", "ready":
		return StatusHealthyStyle
	case "warning":
		return StatusWarningStyle
	case "error", "failed":
		return StatusErrorStyle
	default:
		return lipgloss.NewStyle()
	}
}

// RenderPanelWithTitle renders a panel with the title embedded in the top border line
func RenderPanelWithTitle(title string, content string, width, height int, active bool) string {
	borderColor := colorBorder
	if active {
		borderColor = colorActive
	}

	// Border characters
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	// Calculate available space for title and border lines
	// width-4 because we have 2 chars for left/right borders + 2 for padding
	innerWidth := width - 4

	// Title with spacing: " [1] Title "
	titleWithSpaces := " " + title + " "
	titleLen := len(titleWithSpaces)

	// Calculate border chars - left-justify the title with minimal left border
	// Total horizontal chars (excluding corners) should be innerWidth + 2 to match bottom border
	totalHorizontalChars := innerWidth + 2
	remainingWidth := totalHorizontalChars - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	leftBorderLen := 1 // Just one dash on the left for left-justify
	rightBorderLen := remainingWidth - leftBorderLen
	if rightBorderLen < 0 {
		rightBorderLen = 0
	}

	// Construct top border with embedded title (left-justified)
	topBorder := topLeft + strings.Repeat(horizontal, leftBorderLen) + titleWithSpaces + strings.Repeat(horizontal, rightBorderLen) + topRight

	// Apply color to border
	styledTopBorder := lipgloss.NewStyle().Foreground(borderColor).Render(topBorder)

	// Constrain content - height minus top border (1), bottom border (1), and padding (2)
	contentHeight := height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(contentHeight)
	styledContent := contentStyle.Render(content)

	// Split content into lines and add side borders
	contentLines := strings.Split(styledContent, "\n")

	// Ensure we have exactly contentHeight lines
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}

	var borderedLines []string
	borderedLines = append(borderedLines, styledTopBorder)

	for _, line := range contentLines {
		// Pad line to exact width
		lineLen := lipgloss.Width(line)
		if lineLen < innerWidth {
			line = line + strings.Repeat(" ", innerWidth-lineLen)
		} else if lineLen > innerWidth {
			// Truncate if too long
			line = line[:innerWidth]
		}
		sideBorder := lipgloss.NewStyle().Foreground(borderColor).Render(vertical)
		borderedLines = append(borderedLines, sideBorder+" "+line+" "+sideBorder)
	}

	// Bottom border
	bottomBorder := bottomLeft + strings.Repeat(horizontal, innerWidth+2) + bottomRight
	styledBottomBorder := lipgloss.NewStyle().Foreground(borderColor).Render(bottomBorder)
	borderedLines = append(borderedLines, styledBottomBorder)

	return strings.Join(borderedLines, "\n")
}
