package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// RestoreFormField represents which field is being edited
type RestoreFormField int

const (
	RestoreFieldDestination RestoreFormField = iota
	RestoreFieldInclude
	RestoreFieldSubmit
)

// RestoreForm represents a form for configuring a restore operation
type RestoreForm struct {
	snapshot         *types.Snapshot
	targetInput      textinput.Model
	includeInput     textinput.Model
	focusedField     RestoreFormField
	restoreToOriginal bool
	width            int
	height           int
}

// NewRestoreForm creates a new restore configuration form
func NewRestoreForm(snapshot *types.Snapshot) *RestoreForm {
	targetInput := textinput.New()
	targetInput.Placeholder = "/path/to/restore/location"
	targetInput.Focus()
	targetInput.CharLimit = 500

	includeInput := textinput.New()
	includeInput.Placeholder = "path/to/file, path/to/dir (optional - leave empty for all)"
	includeInput.CharLimit = 500

	return &RestoreForm{
		snapshot:          snapshot,
		targetInput:       targetInput,
		includeInput:      includeInput,
		focusedField:      RestoreFieldDestination,
		restoreToOriginal: false,
	}
}

// Update handles form input
func (f *RestoreForm) Update(msg tea.Msg) tea.Cmd {
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
		case " ":
			// Space to toggle original location when on destination field
			if f.focusedField == RestoreFieldDestination {
				f.restoreToOriginal = !f.restoreToOriginal
				if f.restoreToOriginal {
					f.targetInput.SetValue("")
					f.targetInput.Blur()
				} else {
					f.targetInput.Focus()
				}
				return nil
			}
		}
	}

	// Update the focused input (only if not in "original location" mode)
	if !f.restoreToOriginal || f.focusedField != RestoreFieldDestination {
		switch f.focusedField {
		case RestoreFieldDestination:
			f.targetInput, cmd = f.targetInput.Update(msg)
		case RestoreFieldInclude:
			f.includeInput, cmd = f.includeInput.Update(msg)
		}
	}

	return cmd
}

// NextField moves to the next form field
func (f *RestoreForm) NextField() {
	f.BlurAll()

	f.focusedField++
	if f.focusedField > RestoreFieldSubmit {
		f.focusedField = RestoreFieldDestination
	}

	f.FocusCurrent()
}

// PrevField moves to the previous form field
func (f *RestoreForm) PrevField() {
	f.BlurAll()

	f.focusedField--
	if f.focusedField < RestoreFieldDestination {
		f.focusedField = RestoreFieldSubmit
	}

	f.FocusCurrent()
}

// BlurAll removes focus from all inputs
func (f *RestoreForm) BlurAll() {
	f.targetInput.Blur()
	f.includeInput.Blur()
}

// FocusCurrent focuses the current field
func (f *RestoreForm) FocusCurrent() {
	switch f.focusedField {
	case RestoreFieldDestination:
		if !f.restoreToOriginal {
			f.targetInput.Focus()
		}
	case RestoreFieldInclude:
		f.includeInput.Focus()
	}
}

// GetTarget returns the restore target path
func (f *RestoreForm) GetTarget() string {
	if f.restoreToOriginal {
		return "" // Empty means restore to original location
	}
	return strings.TrimSpace(f.targetInput.Value())
}

// GetInclude returns the specific paths to restore
func (f *RestoreForm) GetInclude() []string {
	if f.includeInput.Value() == "" {
		return []string{} // Empty means restore all
	}

	paths := strings.Split(f.includeInput.Value(), ",")
	var trimmedPaths []string
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			trimmedPaths = append(trimmedPaths, trimmed)
		}
	}
	return trimmedPaths
}

// IsRestoreToOriginal returns true if restoring to original location
func (f *RestoreForm) IsRestoreToOriginal() bool {
	return f.restoreToOriginal
}

// IsValid checks if the form is valid
func (f *RestoreForm) IsValid() bool {
	// Either restore to original or have a target path
	return f.restoreToOriginal || f.GetTarget() != ""
}

// SetSize sets the form dimensions
func (f *RestoreForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.targetInput.Width = width - 20
	f.includeInput.Width = width - 20
}

// SetIncludePaths pre-fills the include paths field with the given paths
func (f *RestoreForm) SetIncludePaths(paths []string) {
	f.includeInput.SetValue(strings.Join(paths, ", "))
}

// Render renders the form
func (f *RestoreForm) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(20)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	title := titleStyle.Render("Restore Snapshot")
	b.WriteString(title + "\n\n")

	// Snapshot info
	if f.snapshot != nil {
		snapshotInfo := infoStyle.Render("Snapshot: " + f.snapshot.ShortID)
		if f.snapshot.Hostname != "" {
			snapshotInfo += " • Host: " + f.snapshot.Hostname
		}
		if len(f.snapshot.Paths) > 0 {
			snapshotInfo += " • Paths: " + strings.Join(f.snapshot.Paths, ", ")
		}
		b.WriteString(snapshotInfo + "\n\n")
	}

	// Destination field
	destLabel := labelStyle.Render("Restore Location:")
	if f.focusedField == RestoreFieldDestination {
		destLabel = focusedStyle.Render("▶ Restore Location:")
	}
	b.WriteString(destLabel + "\n")

	// Original location toggle
	checkBox := "[ ]"
	if f.restoreToOriginal {
		checkBox = "[✓]"
	}
	toggleStyle := labelStyle
	if f.focusedField == RestoreFieldDestination {
		toggleStyle = focusedStyle
	}
	b.WriteString(toggleStyle.Render("  "+checkBox+" Restore to original location") + "\n")

	// Target path input (only if not restoring to original)
	if !f.restoreToOriginal {
		b.WriteString(f.targetInput.View() + "\n")
	}
	b.WriteString("\n")

	// Include paths field
	includeLabel := labelStyle.Render("Specific Paths:")
	if f.focusedField == RestoreFieldInclude {
		includeLabel = focusedStyle.Render("▶ Specific Paths:")
	}
	b.WriteString(includeLabel + "\n")
	b.WriteString(f.includeInput.View() + "\n\n")

	// Submit button
	submitLabel := "  [ Restore Snapshot ]"
	if f.focusedField == RestoreFieldSubmit {
		submitLabel = focusedStyle.Render("▶ [ Restore Snapshot ]")
	}
	b.WriteString(submitLabel + "\n\n")

	// Help text
	help := "Tab/↑↓: Navigate • Space: Toggle original location • Enter: Restore • Esc: Cancel"
	b.WriteString(helpStyle.Render(help))

	// Validation message
	if !f.IsValid() && f.focusedField == RestoreFieldSubmit {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString("\n" + errorStyle.Render("⚠ Destination path is required (or enable original location)"))
	}

	// Warning about original location
	if f.restoreToOriginal {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		b.WriteString("\n" + warningStyle.Render("⚠ Warning: Files will be overwritten in their original locations!"))
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(f.width - 4)

	return borderStyle.Render(b.String())
}
