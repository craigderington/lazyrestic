package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationDialog requires typing a specific word to confirm dangerous operations
type ConfirmationDialog struct {
	title            string
	message          string
	confirmationWord string // Word user must type (e.g., "DELETE", "PRUNE")
	input            textinput.Model
	width            int
	height           int
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(title, message, confirmationWord string) *ConfirmationDialog {
	input := textinput.New()
	input.Placeholder = "Type here..."
	input.CharLimit = 20
	input.Width = 30
	input.Focus()

	return &ConfirmationDialog{
		title:            title,
		message:          message,
		confirmationWord: confirmationWord,
		input:            input,
	}
}

// Update handles input events
func (cd *ConfirmationDialog) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	cd.input, cmd = cd.input.Update(msg)
	return cmd
}

// IsConfirmed returns true if the user typed the correct confirmation word
func (cd *ConfirmationDialog) IsConfirmed() bool {
	return cd.input.Value() == cd.confirmationWord
}

// GetInput returns the current input value
func (cd *ConfirmationDialog) GetInput() string {
	return cd.input.Value()
}

// SetSize sets the dialog dimensions
func (cd *ConfirmationDialog) SetSize(width, height int) {
	cd.width = width
	cd.height = height
	cd.input.Width = width - 20
}

// Render renders the confirmation dialog
func (cd *ConfirmationDialog) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("196")). // Bright red
		Padding(0, 2)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		Padding(1, 2)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Padding(1, 0).
		Width(cd.width - 10)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		MarginTop(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(0, 1).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)

	// Title with emoji
	b.WriteString(titleStyle.Render("⚠️  " + cd.title) + "\n\n")

	// Warning icon and message
	b.WriteString(warningStyle.Render("⛔ DANGER ZONE ⛔") + "\n\n")

	// Message
	b.WriteString(messageStyle.Render(cd.message) + "\n\n")

	// Confirmation prompt
	promptText := "To confirm this action, type exactly:  " + cd.confirmationWord
	b.WriteString(labelStyle.Render(promptText) + "\n")

	// Input field
	inputView := cd.input.View()

	// Show if input matches (visual feedback)
	if cd.IsConfirmed() {
		correctStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Bold(true)
		inputView = correctStyle.Render("✓ " + inputView)
	} else if len(cd.input.Value()) > 0 {
		wrongStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
		inputView = wrongStyle.Render("✗ " + inputView)
	}

	b.WriteString(inputStyle.Render(inputView) + "\n")

	// Help text
	if cd.IsConfirmed() {
		confirmHelpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Bold(true).
			MarginTop(1)
		b.WriteString(confirmHelpStyle.Render("✓ Press Enter to execute • Esc to cancel") + "\n")
	} else {
		b.WriteString(helpStyle.Render("Type the word above, then press Enter • Esc to cancel") + "\n")
	}

	// Border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(2, 3).
		Width(cd.width - 4)

	return boxStyle.Render(b.String())
}
