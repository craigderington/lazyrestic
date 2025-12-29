package types

import (
	"testing"
	"time"
)

func TestPanel_String(t *testing.T) {
	tests := []struct {
		name     string
		panel    Panel
		expected string
	}{
		{
			name:     "PanelRepositories",
			panel:    PanelRepositories,
			expected: "Repositories",
		},
		{
			name:     "PanelSnapshots",
			panel:    PanelSnapshots,
			expected: "Snapshots",
		},
		{
			name:     "PanelOperations",
			panel:    PanelOperations,
			expected: "Operations",
		},
		{
			name:     "Unknown panel",
			panel:    Panel(99),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.panel.String()
			if result != tt.expected {
				t.Errorf("Panel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRepository_Validation(t *testing.T) {
	tests := []struct {
		name       string
		repo       Repository
		shouldFail bool
	}{
		{
			name: "Valid repository",
			repo: Repository{
				Name:   "test-repo",
				Path:   "/tmp/test",
				Status: "healthy",
				Size:   1024,
			},
			shouldFail: false,
		},
		{
			name: "Repository with last backup",
			repo: Repository{
				Name:       "backup",
				Path:       "/backup",
				Status:     "healthy",
				LastBackup: time.Now(),
				Size:       2048,
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - repository should have name and path
			if tt.repo.Name == "" && !tt.shouldFail {
				t.Error("Repository name should not be empty")
			}
			if tt.repo.Path == "" && !tt.shouldFail {
				t.Error("Repository path should not be empty")
			}
		})
	}
}

func TestSnapshot_Fields(t *testing.T) {
	now := time.Now()
	snapshot := Snapshot{
		ID:       "abc123def456",
		ShortID:  "abc123",
		Time:     now,
		Hostname: "testhost",
		Username: "testuser",
		Paths:    []string{"/home/user"},
		Tags:     []string{"important", "daily"},
	}

	// Verify all fields are set correctly
	if snapshot.ID != "abc123def456" {
		t.Errorf("ID = %v, want abc123def456", snapshot.ID)
	}
	if snapshot.ShortID != "abc123" {
		t.Errorf("ShortID = %v, want abc123", snapshot.ShortID)
	}
	if snapshot.Time != now {
		t.Errorf("Time mismatch")
	}
	if snapshot.Hostname != "testhost" {
		t.Errorf("Hostname = %v, want testhost", snapshot.Hostname)
	}
	if len(snapshot.Paths) != 1 || snapshot.Paths[0] != "/home/user" {
		t.Errorf("Paths = %v, want [/home/user]", snapshot.Paths)
	}
	if len(snapshot.Tags) != 2 {
		t.Errorf("Tags length = %v, want 2", len(snapshot.Tags))
	}
}

func TestRepositoryConfig_PasswordOptions(t *testing.T) {
	tests := []struct {
		name   string
		config RepositoryConfig
		hasAuth bool
	}{
		{
			name: "Direct password",
			config: RepositoryConfig{
				Name:     "repo1",
				Path:     "/tmp/repo1",
				Password: "secret123",
			},
			hasAuth: true,
		},
		{
			name: "Password file",
			config: RepositoryConfig{
				Name:         "repo2",
				Path:         "/tmp/repo2",
				PasswordFile: "/home/user/.restic-pass",
			},
			hasAuth: true,
		},
		{
			name: "Password command",
			config: RepositoryConfig{
				Name:            "repo3",
				Path:            "/tmp/repo3",
				PasswordCommand: "pass show restic/repo3",
			},
			hasAuth: true,
		},
		{
			name: "No password (invalid but possible)",
			config: RepositoryConfig{
				Name: "repo4",
				Path: "/tmp/repo4",
			},
			hasAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasAuth := tt.config.Password != "" ||
				tt.config.PasswordFile != "" ||
				tt.config.PasswordCommand != ""

			if hasAuth != tt.hasAuth {
				t.Errorf("hasAuth = %v, want %v", hasAuth, tt.hasAuth)
			}
		})
	}
}

func TestResticConfig_MultipleRepositories(t *testing.T) {
	config := ResticConfig{
		Repositories: []RepositoryConfig{
			{
				Name:     "repo1",
				Path:     "/tmp/repo1",
				Password: "pass1",
			},
			{
				Name:         "repo2",
				Path:         "/tmp/repo2",
				PasswordFile: "/pass2",
			},
		},
	}

	if len(config.Repositories) != 2 {
		t.Errorf("Repositories count = %v, want 2", len(config.Repositories))
	}

	if config.Repositories[0].Name != "repo1" {
		t.Errorf("First repo name = %v, want repo1", config.Repositories[0].Name)
	}

	if config.Repositories[1].Path != "/tmp/repo2" {
		t.Errorf("Second repo path = %v, want /tmp/repo2", config.Repositories[1].Path)
	}
}
