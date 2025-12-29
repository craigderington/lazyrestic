package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackupFormField represents which field is being edited
type BackupFormField int

const (
	BackupFieldPaths BackupFormField = iota
	BackupFieldTags
	BackupFieldExclude
	BackupFieldSubmit
)

// BackupForm represents a form for configuring a backup operation
type BackupForm struct {
	pathsInput   textinput.Model
	tagsInput    textinput.Model
	excludeInput textinput.Model
	focusedField BackupFormField
	width        int
	height       int
}

// NewBackupForm creates a new backup configuration form
func NewBackupForm() *BackupForm {
	pathsInput := textinput.New()
	pathsInput.Placeholder = "/home/user, /etc (comma-separated)"
	pathsInput.Focus()
	pathsInput.CharLimit = 500

	tagsInput := textinput.New()
	tagsInput.Placeholder = "config, manual (optional)"
	tagsInput.CharLimit = 200

	excludeInput := textinput.New()
	excludeInput.Placeholder = "*.tmp, *.cache (optional)"
	excludeInput.CharLimit = 200

	return &BackupForm{
		pathsInput:   pathsInput,
		tagsInput:    tagsInput,
		excludeInput: excludeInput,
		focusedField: BackupFieldPaths,
	}
}

// Update handles form input
func (f *BackupForm) Update(msg tea.Msg) tea.Cmd {
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
	case BackupFieldPaths:
		f.pathsInput, cmd = f.pathsInput.Update(msg)
	case BackupFieldTags:
		f.tagsInput, cmd = f.tagsInput.Update(msg)
	case BackupFieldExclude:
		f.excludeInput, cmd = f.excludeInput.Update(msg)
	}

	return cmd
}

// NextField moves to the next form field
func (f *BackupForm) NextField() {
	f.BlurAll()

	f.focusedField++
	if f.focusedField > BackupFieldSubmit {
		f.focusedField = BackupFieldPaths
	}

	f.FocusCurrent()
}

// PrevField moves to the previous form field
func (f *BackupForm) PrevField() {
	f.BlurAll()

	f.focusedField--
	if f.focusedField < BackupFieldPaths {
		f.focusedField = BackupFieldSubmit
	}

	f.FocusCurrent()
}

// BlurAll removes focus from all inputs
func (f *BackupForm) BlurAll() {
	f.pathsInput.Blur()
	f.tagsInput.Blur()
	f.excludeInput.Blur()
}

// FocusCurrent focuses the current field
func (f *BackupForm) FocusCurrent() {
	switch f.focusedField {
	case BackupFieldPaths:
		f.pathsInput.Focus()
	case BackupFieldTags:
		f.tagsInput.Focus()
	case BackupFieldExclude:
		f.excludeInput.Focus()
	}
}

// GetPaths returns the paths to backup as a slice
func (f *BackupForm) GetPaths() []string {
	if f.pathsInput.Value() == "" {
		return []string{}
	}

	paths := strings.Split(f.pathsInput.Value(), ",")
	var trimmedPaths []string
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			trimmedPaths = append(trimmedPaths, trimmed)
		}
	}
	return trimmedPaths
}

// GetTags returns the tags as a slice
func (f *BackupForm) GetTags() []string {
	if f.tagsInput.Value() == "" {
		return []string{}
	}

	tags := strings.Split(f.tagsInput.Value(), ",")
	var trimmedTags []string
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			trimmedTags = append(trimmedTags, trimmed)
		}
	}
	return trimmedTags
}

// GetExclude returns the exclude patterns as a slice
func (f *BackupForm) GetExclude() []string {
	if f.excludeInput.Value() == "" {
		return []string{}
	}

	excludes := strings.Split(f.excludeInput.Value(), ",")
	var trimmedExcludes []string
	for _, e := range excludes {
		trimmed := strings.TrimSpace(e)
		if trimmed != "" {
			trimmedExcludes = append(trimmedExcludes, trimmed)
		}
	}
	return trimmedExcludes
}

// IsValid checks if the form is valid
func (f *BackupForm) IsValid() bool {
	return len(f.GetPaths()) > 0
}

// SetSize sets the form dimensions
func (f *BackupForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.pathsInput.Width = width - 20
	f.tagsInput.Width = width - 20
	f.excludeInput.Width = width - 20
}

// Render renders the form
func (f *BackupForm) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(20)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	title := titleStyle.Render("Configure Backup")
	b.WriteString(title + "\n\n")

	// Paths field
	pathsLabel := labelStyle.Render("Paths to Backup:")
	if f.focusedField == BackupFieldPaths {
		pathsLabel = focusedStyle.Render("▶ Paths to Backup:")
	}
	b.WriteString(pathsLabel + "\n")
	b.WriteString(f.pathsInput.View() + "\n\n")

	// Tags field
	tagsLabel := labelStyle.Render("Tags:")
	if f.focusedField == BackupFieldTags {
		tagsLabel = focusedStyle.Render("▶ Tags:")
	}
	b.WriteString(tagsLabel + "\n")
	b.WriteString(f.tagsInput.View() + "\n\n")

	// Exclude field
	excludeLabel := labelStyle.Render("Exclude Patterns:")
	if f.focusedField == BackupFieldExclude {
		excludeLabel = focusedStyle.Render("▶ Exclude Patterns:")
	}
	b.WriteString(excludeLabel + "\n")
	b.WriteString(f.excludeInput.View() + "\n\n")

	// Submit button
	submitLabel := "  [ Start Backup ]"
	if f.focusedField == BackupFieldSubmit {
		submitLabel = focusedStyle.Render("▶ [ Start Backup ]")
	}
	b.WriteString(submitLabel + "\n\n")

	// Help text
	help := "Tab/↑↓: Navigate • Enter: Start Backup • Esc: Cancel"
	b.WriteString(helpStyle.Render(help))

	// Validation message
	if !f.IsValid() && f.focusedField == BackupFieldSubmit {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString("\n" + errorStyle.Render("⚠ At least one path is required"))
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2).
		Width(f.width - 4)

	return borderStyle.Render(b.String())
}
