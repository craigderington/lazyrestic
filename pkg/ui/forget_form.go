package ui

import (
	"strconv"
	"strings"

	"github.com/craigderington/lazyrestic/pkg/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ForgetFormField represents which field is focused
type ForgetFormField int

const (
	ForgetFieldKeepLast ForgetFormField = iota
	ForgetFieldKeepDaily
	ForgetFieldKeepWeekly
	ForgetFieldKeepMonthly
	ForgetFieldKeepYearly
	ForgetFieldKeepWithin
	ForgetFieldPreview
)

// ForgetForm represents the forget policy configuration form
type ForgetForm struct {
	keepLastInput    textinput.Model
	keepDailyInput   textinput.Model
	keepWeeklyInput  textinput.Model
	keepMonthlyInput textinput.Model
	keepYearlyInput  textinput.Model
	keepWithinInput  textinput.Model

	focusedField ForgetFormField
	width        int
	height       int
	errorMsg     string
}

// NewForgetForm creates a new forget policy form
func NewForgetForm() *ForgetForm {
	keepLast := textinput.New()
	keepLast.Placeholder = "e.g., 10"
	keepLast.CharLimit = 5
	keepLast.Width = 20

	keepDaily := textinput.New()
	keepDaily.Placeholder = "e.g., 7"
	keepDaily.CharLimit = 5
	keepDaily.Width = 20

	keepWeekly := textinput.New()
	keepWeekly.Placeholder = "e.g., 4"
	keepWeekly.CharLimit = 5
	keepWeekly.Width = 20

	keepMonthly := textinput.New()
	keepMonthly.Placeholder = "e.g., 12"
	keepMonthly.CharLimit = 5
	keepMonthly.Width = 20

	keepYearly := textinput.New()
	keepYearly.Placeholder = "e.g., 5"
	keepYearly.CharLimit = 5
	keepYearly.Width = 20

	keepWithin := textinput.New()
	keepWithin.Placeholder = "e.g., 1y6m (1 year 6 months)"
	keepWithin.CharLimit = 20
	keepWithin.Width = 30

	form := &ForgetForm{
		keepLastInput:    keepLast,
		keepDailyInput:   keepDaily,
		keepWeeklyInput:  keepWeekly,
		keepMonthlyInput: keepMonthly,
		keepYearlyInput:  keepYearly,
		keepWithinInput:  keepWithin,
		focusedField:     ForgetFieldKeepLast,
	}

	form.FocusCurrent()
	return form
}

// Update handles input events
func (f *ForgetForm) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.NextField()
			return nil
		case "shift+tab", "up":
			f.PrevField()
			return nil
		}
	}

	// Update the focused input
	switch f.focusedField {
	case ForgetFieldKeepLast:
		f.keepLastInput, cmd = f.keepLastInput.Update(msg)
	case ForgetFieldKeepDaily:
		f.keepDailyInput, cmd = f.keepDailyInput.Update(msg)
	case ForgetFieldKeepWeekly:
		f.keepWeeklyInput, cmd = f.keepWeeklyInput.Update(msg)
	case ForgetFieldKeepMonthly:
		f.keepMonthlyInput, cmd = f.keepMonthlyInput.Update(msg)
	case ForgetFieldKeepYearly:
		f.keepYearlyInput, cmd = f.keepYearlyInput.Update(msg)
	case ForgetFieldKeepWithin:
		f.keepWithinInput, cmd = f.keepWithinInput.Update(msg)
	}

	return cmd
}

// NextField moves to the next field
func (f *ForgetForm) NextField() {
	f.BlurAll()
	f.focusedField = (f.focusedField + 1) % 7
	f.FocusCurrent()
}

// PrevField moves to the previous field
func (f *ForgetForm) PrevField() {
	f.BlurAll()
	f.focusedField = (f.focusedField + 6) % 7
	f.FocusCurrent()
}

// BlurAll blurs all input fields
func (f *ForgetForm) BlurAll() {
	f.keepLastInput.Blur()
	f.keepDailyInput.Blur()
	f.keepWeeklyInput.Blur()
	f.keepMonthlyInput.Blur()
	f.keepYearlyInput.Blur()
	f.keepWithinInput.Blur()
}

// FocusCurrent focuses the current field
func (f *ForgetForm) FocusCurrent() {
	switch f.focusedField {
	case ForgetFieldKeepLast:
		f.keepLastInput.Focus()
	case ForgetFieldKeepDaily:
		f.keepDailyInput.Focus()
	case ForgetFieldKeepWeekly:
		f.keepWeeklyInput.Focus()
	case ForgetFieldKeepMonthly:
		f.keepMonthlyInput.Focus()
	case ForgetFieldKeepYearly:
		f.keepYearlyInput.Focus()
	case ForgetFieldKeepWithin:
		f.keepWithinInput.Focus()
	}
}

// GetPolicy returns the configured policy
func (f *ForgetForm) GetPolicy() types.ForgetPolicy {
	policy := types.ForgetPolicy{}

	if val := f.keepLastInput.Value(); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			policy.KeepLast = n
		}
	}
	if val := f.keepDailyInput.Value(); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			policy.KeepDaily = n
		}
	}
	if val := f.keepWeeklyInput.Value(); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			policy.KeepWeekly = n
		}
	}
	if val := f.keepMonthlyInput.Value(); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			policy.KeepMonthly = n
		}
	}
	if val := f.keepYearlyInput.Value(); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			policy.KeepYearly = n
		}
	}
	policy.KeepWithin = f.keepWithinInput.Value()

	return policy
}

// IsValid checks if at least one retention rule is specified
func (f *ForgetForm) IsValid() bool {
	policy := f.GetPolicy()
	hasRule := policy.KeepLast > 0 || policy.KeepDaily > 0 || policy.KeepWeekly > 0 ||
		policy.KeepMonthly > 0 || policy.KeepYearly > 0 || policy.KeepWithin != ""

	if !hasRule {
		f.errorMsg = "At least one retention rule must be specified"
		return false
	}

	f.errorMsg = ""
	return true
}

// IsPreviewButton returns true if the preview button is focused
func (f *ForgetForm) IsPreviewButton() bool {
	return f.focusedField == ForgetFieldPreview
}

// SetSize sets the form dimensions
func (f *ForgetForm) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// Render renders the form
func (f *ForgetForm) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")). // Orange for warning
		Padding(0, 1)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		Padding(1, 2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(25)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Padding(0, 2).
		MarginTop(1)

	buttonFocusedStyle := buttonStyle.Copy().
		Background(lipgloss.Color("205")).
		Bold(true)

	// Title
	b.WriteString(titleStyle.Render("⚠️  Configure Forget Policy") + "\n\n")

	// Warning
	warning := "This will DELETE snapshots permanently!\nAlways preview before confirming."
	b.WriteString(warningStyle.Render(warning) + "\n\n")

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Width(f.width - 10)
	desc := "Specify retention rules. Snapshots not matching any rule will be REMOVED."
	b.WriteString(descStyle.Render(desc) + "\n\n")

	// Input fields
	b.WriteString(labelStyle.Render("Keep Last N Snapshots:") + "  " + f.keepLastInput.View() + "\n")
	b.WriteString(labelStyle.Render("Keep Daily (last N days):") + "  " + f.keepDailyInput.View() + "\n")
	b.WriteString(labelStyle.Render("Keep Weekly (last N weeks):") + "  " + f.keepWeeklyInput.View() + "\n")
	b.WriteString(labelStyle.Render("Keep Monthly (last N months):") + "  " + f.keepMonthlyInput.View() + "\n")
	b.WriteString(labelStyle.Render("Keep Yearly (last N years):") + "  " + f.keepYearlyInput.View() + "\n")
	b.WriteString(labelStyle.Render("Keep Within Duration:") + "  " + f.keepWithinInput.View() + "\n")

	// Examples
	exampleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)
	b.WriteString(exampleStyle.Render("  Examples: keep-last 10, keep-daily 7, keep-within 1y6m") + "\n")

	// Error message
	if f.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(1)
		b.WriteString(errorStyle.Render("⚠ "+f.errorMsg) + "\n")
	}

	// Preview button
	btnText := "[ Preview Dry-Run ]"
	if f.IsPreviewButton() {
		b.WriteString(buttonFocusedStyle.Render(btnText) + "\n")
	} else {
		b.WriteString(buttonStyle.Render(btnText) + "\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)
	b.WriteString(helpStyle.Render("Tab: next field • Enter: preview • Esc: cancel") + "\n")

	// Border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(f.width - 4)

	return boxStyle.Render(b.String())
}
