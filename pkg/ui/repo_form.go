package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RepoFormField represents which field is being edited
type RepoFormField int

const (
	FieldName RepoFormField = iota
	FieldPath
	FieldPasswordMethod
	FieldPassword
	FieldGeneratePasswordFile
	FieldInitialize
	FieldSubmit
)

// RepoForm represents a form for creating a new repository
type RepoForm struct {
	nameInput               textinput.Model
	pathInput               textinput.Model
	passwordInput           textinput.Model
	focusedField            RepoFormField
	passwordMethod          string // "file" or "command"
	autoGeneratePasswordFile bool   // Whether to auto-generate password file path
	initializeRepo          bool   // Whether to initialize the repository
	width                   int
	height                  int
}

// NewRepoForm creates a new repository creation form
func NewRepoForm() *RepoForm {
	nameInput := textinput.New()
	nameInput.Placeholder = "my-backup"
	nameInput.Focus()
	nameInput.CharLimit = 50

	pathInput := textinput.New()
	pathInput.Placeholder = "/path/to/repo or s3:bucket/path"
	pathInput.CharLimit = 200

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Will be auto-generated if using file method"
	passwordInput.EchoMode = textinput.EchoNormal
	passwordInput.CharLimit = 200

	return &RepoForm{
		nameInput:               nameInput,
		pathInput:               pathInput,
		passwordInput:           passwordInput,
		focusedField:            FieldName,
		passwordMethod:          "file", // Default to secure file method
		autoGeneratePasswordFile: true,   // Auto-generate by default
	}
}

// Update handles form input
func (f *RepoForm) Update(msg tea.Msg) tea.Cmd {
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
	case FieldName:
		f.nameInput, cmd = f.nameInput.Update(msg)
	case FieldPath:
		f.pathInput, cmd = f.pathInput.Update(msg)
	case FieldPassword:
		f.passwordInput, cmd = f.passwordInput.Update(msg)
	case FieldPasswordMethod:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == " " {
			oldMethod := f.passwordMethod
			f.passwordMethod = f.nextPasswordMethod()
			// Clear password input when changing methods
			if oldMethod != f.passwordMethod {
				f.passwordInput.SetValue("")
				// Update placeholder based on method
				f.updatePasswordPlaceholder()
			}
		}
	case FieldGeneratePasswordFile:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == " " {
			f.autoGeneratePasswordFile = !f.autoGeneratePasswordFile
			// Clear manual input if switching to auto-generate
			if f.autoGeneratePasswordFile {
				f.passwordInput.SetValue("")
			}
		}
	case FieldInitialize:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == " " {
			f.initializeRepo = !f.initializeRepo
		}
	}

	return cmd
}

// NextField moves to the next form field
func (f *RepoForm) NextField() {
	f.BlurAll()

	f.focusedField++
	if f.focusedField > FieldSubmit {
		f.focusedField = FieldName
	}

	// Skip FieldGeneratePasswordFile if password method is "command"
	if f.focusedField == FieldGeneratePasswordFile && f.passwordMethod == "command" {
		f.focusedField++
		if f.focusedField > FieldSubmit {
			f.focusedField = FieldName
		}
	}

	f.FocusCurrent()
}

// PrevField moves to the previous form field
func (f *RepoForm) PrevField() {
	f.BlurAll()

	f.focusedField--
	if f.focusedField < FieldName {
		f.focusedField = FieldSubmit
	}

	// Skip FieldGeneratePasswordFile if password method is "command"
	if f.focusedField == FieldGeneratePasswordFile && f.passwordMethod == "command" {
		f.focusedField--
		if f.focusedField < FieldName {
			f.focusedField = FieldSubmit
		}
	}

	f.FocusCurrent()
}

// BlurAll removes focus from all inputs
func (f *RepoForm) BlurAll() {
	f.nameInput.Blur()
	f.pathInput.Blur()
	f.passwordInput.Blur()
}

// FocusCurrent focuses the current field
func (f *RepoForm) FocusCurrent() {
	switch f.focusedField {
	case FieldName:
		f.nameInput.Focus()
	case FieldPath:
		f.pathInput.Focus()
	case FieldPassword:
		f.passwordInput.Focus()
	}
}

// updatePasswordPlaceholder updates the placeholder text based on method
func (f *RepoForm) updatePasswordPlaceholder() {
	switch f.passwordMethod {
	case "file":
		if f.autoGeneratePasswordFile {
			f.passwordInput.Placeholder = "Will be auto-generated: ~/.config/lazyrestic/passwords/<name>.txt"
		} else {
			f.passwordInput.Placeholder = "/path/to/password-file"
		}
		f.passwordInput.EchoMode = textinput.EchoNormal
	case "command":
		f.passwordInput.Placeholder = "pass show restic/my-repo"
		f.passwordInput.EchoMode = textinput.EchoNormal
	}
}

// GetName returns the repository name
func (f *RepoForm) GetName() string {
	return f.nameInput.Value()
}

// GetPath returns the repository path
func (f *RepoForm) GetPath() string {
	return f.pathInput.Value()
}

// GetPassword returns the password value
func (f *RepoForm) GetPassword() string {
	return f.passwordInput.Value()
}

// GetPasswordMethod returns the password method
func (f *RepoForm) GetPasswordMethod() string {
	return f.passwordMethod
}

// ShouldInitialize returns whether to initialize the repository
func (f *RepoForm) ShouldInitialize() bool {
	return f.initializeRepo
}

// SetPath sets the repository path
func (f *RepoForm) SetPath(path string) {
	f.pathInput.SetValue(path)
}

// SetName sets the repository name
func (f *RepoForm) SetName(name string) {
	f.nameInput.SetValue(name)
}

// IsValid checks if the form is valid
func (f *RepoForm) IsValid() bool {
	// Name and path are always required
	if f.GetName() == "" || f.GetPath() == "" {
		return false
	}

	// For file method with auto-generation, password can be empty
	if f.passwordMethod == "file" && f.autoGeneratePasswordFile {
		return true
	}

	// Otherwise password/command is required
	return f.GetPassword() != ""
}

// nextPasswordMethod cycles to the next password method
func (f *RepoForm) nextPasswordMethod() string {
	switch f.passwordMethod {
	case "file":
		return "command"
	case "command":
		return "file"
	default:
		return "file"
	}
}

// ShouldAutoGeneratePasswordFile returns whether to auto-generate password file
func (f *RepoForm) ShouldAutoGeneratePasswordFile() bool {
	return f.passwordMethod == "file" && f.autoGeneratePasswordFile
}

// GetFocusedField returns the currently focused field
func (f *RepoForm) GetFocusedField() RepoFormField {
	return f.focusedField
}

// SetSize sets the form dimensions
func (f *RepoForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.nameInput.Width = width - 20
	f.pathInput.Width = width - 20
	f.passwordInput.Width = width - 20
}

// Render renders the form
func (f *RepoForm) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
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

	title := titleStyle.Render("Create New Repository")
	b.WriteString(title + "\n\n")

	// Name field
	nameLabel := labelStyle.Render("Repository Name:")
	if f.focusedField == FieldName {
		nameLabel = focusedStyle.Render("▶ Repository Name:")
	}
	b.WriteString(nameLabel + "\n")
	b.WriteString(f.nameInput.View() + "\n\n")

	// Path field
	pathLabel := labelStyle.Render("Repository Path:")
	if f.focusedField == FieldPath {
		pathLabel = focusedStyle.Render("▶ Repository Path:")
	}
	b.WriteString(pathLabel + "\n")
	b.WriteString(f.pathInput.View() + "\n\n")

	// Password method selector
	methodLabel := labelStyle.Render("Password Method:")
	if f.focusedField == FieldPasswordMethod {
		methodLabel = focusedStyle.Render("▶ Password Method:")
	}
	b.WriteString(methodLabel + "\n")

	methods := []string{"file", "command"}
	var methodsDisplay []string
	for _, m := range methods {
		if m == f.passwordMethod {
			methodsDisplay = append(methodsDisplay, fmt.Sprintf("[%s]", m))
		} else {
			methodsDisplay = append(methodsDisplay, m)
		}
	}
	b.WriteString("  " + strings.Join(methodsDisplay, " | ") + "\n")
	if f.focusedField == FieldPasswordMethod {
		b.WriteString(helpStyle.Render("  Press space to cycle") + "\n")
	}
	b.WriteString("\n")

	// Auto-generate password file option (only for file method)
	if f.passwordMethod == "file" {
		genLabel := "  Auto-generate secure password file"
		if f.autoGeneratePasswordFile {
			genLabel = "  [✓] Auto-generate secure password file"
		} else {
			genLabel = "  [ ] Auto-generate secure password file"
		}
		if f.focusedField == FieldGeneratePasswordFile {
			genLabel = focusedStyle.Render("▶ " + genLabel)
		}
		b.WriteString(genLabel + "\n")
		if f.focusedField == FieldGeneratePasswordFile {
			b.WriteString(helpStyle.Render("  Press space to toggle") + "\n")
		}
		b.WriteString("\n")
	}

	// Password field (or command field)
	passwordLabel := labelStyle.Render(f.getPasswordLabel())
	if f.focusedField == FieldPassword {
		passwordLabel = focusedStyle.Render("▶ " + f.getPasswordLabel())
	}
	b.WriteString(passwordLabel + "\n")

	// Show different help text for auto-generated files
	if f.passwordMethod == "file" && f.autoGeneratePasswordFile {
		b.WriteString(helpStyle.Render("  (will be created at: ~/.config/lazyrestic/passwords/"+f.GetName()+".txt)") + "\n")
	} else {
		b.WriteString(f.passwordInput.View() + "\n")
	}
	b.WriteString("\n")

	// Initialize option
	initLabel := "  Initialize repository after creation"
	if f.initializeRepo {
		initLabel = "  [✓] Initialize repository after creation"
	} else {
		initLabel = "  [ ] Initialize repository after creation"
	}
	if f.focusedField == FieldInitialize {
		initLabel = focusedStyle.Render("▶ " + initLabel)
	}
	b.WriteString(initLabel + "\n")
	if f.focusedField == FieldInitialize {
		b.WriteString(helpStyle.Render("  Press space to toggle") + "\n")
	}
	b.WriteString("\n")

	// Submit button
	submitLabel := "  [ Create Repository ]"
	if f.focusedField == FieldSubmit {
		submitLabel = focusedStyle.Render("▶ [ Create Repository ]")
	}
	b.WriteString(submitLabel + "\n\n")

	// Help text
	help := "Tab/↑↓: Navigate • Space: Toggle options • Enter: Submit • Esc: Cancel"
	b.WriteString(helpStyle.Render(help))

	// Validation message
	if !f.IsValid() && f.focusedField == FieldSubmit {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString("\n" + errorStyle.Render("⚠ All fields are required"))
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(f.width - 4)

	return borderStyle.Render(b.String())
}

// getPasswordLabel returns the appropriate label for the password field
func (f *RepoForm) getPasswordLabel() string {
	switch f.passwordMethod {
	case "file":
		if f.autoGeneratePasswordFile {
			return "Password File (auto):"
		}
		return "Password File Path:"
	case "command":
		return "Password Command:"
	default:
		return "Password File Path:"
	}
}
