package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBackupFormCreation(t *testing.T) {
	form := NewBackupForm()

	if form == nil {
		t.Fatal("NewBackupForm() returned nil")
	}

	if form.focusedField != BackupFieldPaths {
		t.Errorf("Expected initial focus on BackupFieldPaths, got %v", form.focusedField)
	}
}

func TestBackupFormValidation(t *testing.T) {
	form := NewBackupForm()

	// Form should be invalid without paths
	if form.IsValid() {
		t.Error("Form should be invalid without paths")
	}

	// Simulate entering a path
	form.pathsInput.SetValue("/home/user")

	if !form.IsValid() {
		t.Error("Form should be valid with at least one path")
	}
}

func TestBackupFormGetPaths(t *testing.T) {
	form := NewBackupForm()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single path",
			input:    "/home/user",
			expected: []string{"/home/user"},
		},
		{
			name:     "multiple paths",
			input:    "/home/user, /etc, /var/log",
			expected: []string{"/home/user", "/etc", "/var/log"},
		},
		{
			name:     "paths with extra spaces",
			input:    " /home/user , /etc ",
			expected: []string{"/home/user", "/etc"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form.pathsInput.SetValue(tt.input)
			paths := form.GetPaths()

			if len(paths) != len(tt.expected) {
				t.Errorf("Expected %d paths, got %d", len(tt.expected), len(paths))
				return
			}

			for i, path := range paths {
				if path != tt.expected[i] {
					t.Errorf("Expected path %q at index %d, got %q", tt.expected[i], i, path)
				}
			}
		})
	}
}

func TestBackupFormGetTags(t *testing.T) {
	form := NewBackupForm()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single tag",
			input:    "config",
			expected: []string{"config"},
		},
		{
			name:     "multiple tags",
			input:    "config, manual, important",
			expected: []string{"config", "manual", "important"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form.tagsInput.SetValue(tt.input)
			tags := form.GetTags()

			if len(tags) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(tags))
				return
			}

			for i, tag := range tags {
				if tag != tt.expected[i] {
					t.Errorf("Expected tag %q at index %d, got %q", tt.expected[i], i, tag)
				}
			}
		})
	}
}

func TestBackupFormNavigation(t *testing.T) {
	form := NewBackupForm()

	// Start at paths field
	if form.focusedField != BackupFieldPaths {
		t.Errorf("Expected BackupFieldPaths, got %v", form.focusedField)
	}

	// Navigate forward
	form.NextField()
	if form.focusedField != BackupFieldTags {
		t.Errorf("Expected BackupFieldTags after NextField(), got %v", form.focusedField)
	}

	form.NextField()
	if form.focusedField != BackupFieldExclude {
		t.Errorf("Expected BackupFieldExclude after NextField(), got %v", form.focusedField)
	}

	form.NextField()
	if form.focusedField != BackupFieldSubmit {
		t.Errorf("Expected BackupFieldSubmit after NextField(), got %v", form.focusedField)
	}

	// Should wrap around
	form.NextField()
	if form.focusedField != BackupFieldPaths {
		t.Errorf("Expected BackupFieldPaths after wrap-around, got %v", form.focusedField)
	}

	// Navigate backward
	form.PrevField()
	if form.focusedField != BackupFieldSubmit {
		t.Errorf("Expected BackupFieldSubmit after PrevField(), got %v", form.focusedField)
	}
}

func TestBackupFormUpdate(t *testing.T) {
	form := NewBackupForm()

	// Test tab navigation
	cmd := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	if form.focusedField != BackupFieldTags {
		t.Errorf("Expected BackupFieldTags after tab, got %v", form.focusedField)
	}

	if cmd != nil {
		t.Error("Expected nil command from navigation")
	}

	// Test down arrow navigation
	form.focusedField = BackupFieldPaths
	cmd = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	if form.focusedField != BackupFieldTags {
		t.Errorf("Expected BackupFieldTags after down arrow, got %v", form.focusedField)
	}
}

func TestBackupFormRender(t *testing.T) {
	form := NewBackupForm()
	form.SetSize(80, 24)

	rendered := form.Render()

	if rendered == "" {
		t.Error("Render() returned empty string")
	}

	// Check for key elements in the rendered output
	expectedStrings := []string{
		"Configure Backup",
		"Paths to Backup",
		"Tags",
		"Exclude Patterns",
		"Start Backup",
	}

	for _, expected := range expectedStrings {
		if !contains(rendered, expected) {
			t.Errorf("Rendered output missing expected string: %q", expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
