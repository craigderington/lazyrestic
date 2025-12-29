package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create temporary password file
	passwordFile := filepath.Join(tmpDir, ".restic-pass")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write test password file: %v", err)
	}

	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: /tmp/test
    password: testpass
  - name: s3-repo
    path: s3:bucket/path
    password_file: %s
`, passwordFile)

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading
	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Validate loaded config
	if len(config.Repositories) != 2 {
		t.Errorf("Repositories count = %v, want 2", len(config.Repositories))
	}

	// Check first repository
	repo1 := config.Repositories[0]
	if repo1.Name != "test-repo" {
		t.Errorf("First repo name = %v, want test-repo", repo1.Name)
	}
	if repo1.Path != "/tmp/test" {
		t.Errorf("First repo path = %v, want /tmp/test", repo1.Path)
	}
	if repo1.Password != "testpass" {
		t.Errorf("First repo password = %v, want testpass", repo1.Password)
	}

	// Check second repository
	repo2 := config.Repositories[1]
	if repo2.Name != "s3-repo" {
		t.Errorf("Second repo name = %v, want s3-repo", repo2.Name)
	}
	if repo2.PasswordFile != passwordFile {
		t.Errorf("Second repo password_file = %v, want %v", repo2.PasswordFile, passwordFile)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() should fail for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `this is not: [valid: yaml: content`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should fail for invalid YAML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Empty file should result in empty config
	if len(config.Repositories) != 0 {
		t.Errorf("Empty config should have 0 repositories, got %v", len(config.Repositories))
	}
}

func TestSave_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	config := &types.ResticConfig{
		Repositories: []types.RepositoryConfig{
			{
				Name:     "test-repo",
				Path:     "/tmp/test",
				Password: "secret",
			},
		},
	}

	// Test saving
	if err := Save(config, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists and has correct permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Saved file not found: %v", err)
	}

	// Check permissions (should be 0600)
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("File permissions = %o, want 0600", mode.Perm())
	}

	// Load it back and verify content
	loadedConfig, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if len(loadedConfig.Repositories) != 1 {
		t.Errorf("Loaded repositories count = %v, want 1", len(loadedConfig.Repositories))
	}

	if loadedConfig.Repositories[0].Name != "test-repo" {
		t.Errorf("Loaded repo name = %v, want test-repo", loadedConfig.Repositories[0].Name)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	config := &types.ResticConfig{
		Repositories: []types.RepositoryConfig{
			{Name: "test", Path: "/tmp"},
		},
	}

	// Should create nested directories
	if err := Save(config, configPath); err != nil {
		t.Fatalf("Save() should create directories: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not created: %v", err)
	}
}

func TestLoadOrDefault_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `repositories:
  - name: my-repo
    path: /my/path
    password: mypass
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config := LoadOrDefault(configPath)

	if len(config.Repositories) != 1 {
		t.Errorf("Repositories count = %v, want 1", len(config.Repositories))
	}

	if config.Repositories[0].Name != "my-repo" {
		t.Errorf("Repo name = %v, want my-repo", config.Repositories[0].Name)
	}
}

func TestLoadOrDefault_NonExistentFile(t *testing.T) {
	config := LoadOrDefault("/nonexistent/config.yaml")

	// Should return empty config when file doesn't exist
	if len(config.Repositories) != 0 {
		t.Errorf("LoadOrDefault() should return empty config, got %d repositories", len(config.Repositories))
	}
}

func TestCreateExample(t *testing.T) {
	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, "example.yaml")

	if err := CreateExample(examplePath); err != nil {
		t.Fatalf("CreateExample() failed: %v", err)
	}

	// Load and verify example config
	config, err := Load(examplePath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	// Example should have at least 2 repositories
	if len(config.Repositories) < 2 {
		t.Errorf("Example config should have at least 2 repositories, got %v", len(config.Repositories))
	}

	// Verify it has examples of different password methods
	hasPasswordCommand := false
	hasPasswordFile := false

	for _, repo := range config.Repositories {
		if repo.PasswordCommand != "" {
			hasPasswordCommand = true
		}
		if repo.PasswordFile != "" {
			hasPasswordFile = true
		}
	}

	if !hasPasswordCommand {
		t.Error("Example config should demonstrate password_command")
	}
	if !hasPasswordFile {
		t.Error("Example config should demonstrate password_file")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	// Should not be empty
	if path == "" {
		t.Error("DefaultConfigPath() should not be empty")
	}

	// Should contain .config/lazyrestic
	if !filepath.IsAbs(path) {
		t.Error("DefaultConfigPath() should return absolute path")
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("DefaultConfigPath() should end with config.yaml, got %v", filepath.Base(path))
	}
}

func TestRoundTrip_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip.yaml")

	original := &types.ResticConfig{
		Repositories: []types.RepositoryConfig{
			{
				Name:            "repo1",
				Path:            "/path/one",
				PasswordCommand: "pass show repo1",
			},
			{
				Name:         "repo2",
				Path:         "s3:bucket/path",
				PasswordFile: "/home/user/.pass",
			},
			{
				Name:     "repo3",
				Path:     "/path/three",
				Password: "direct-password",
			},
		},
	}

	// Save
	if err := Save(original, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Compare
	if len(loaded.Repositories) != len(original.Repositories) {
		t.Fatalf("Repository count mismatch: got %v, want %v",
			len(loaded.Repositories), len(original.Repositories))
	}

	for i := range original.Repositories {
		orig := original.Repositories[i]
		load := loaded.Repositories[i]

		if orig.Name != load.Name {
			t.Errorf("Repo %d: Name = %v, want %v", i, load.Name, orig.Name)
		}
		if orig.Path != load.Path {
			t.Errorf("Repo %d: Path = %v, want %v", i, load.Path, orig.Path)
		}
		if orig.Password != load.Password {
			t.Errorf("Repo %d: Password = %v, want %v", i, load.Password, orig.Password)
		}
		if orig.PasswordFile != load.PasswordFile {
			t.Errorf("Repo %d: PasswordFile = %v, want %v", i, load.PasswordFile, orig.PasswordFile)
		}
		if orig.PasswordCommand != load.PasswordCommand {
			t.Errorf("Repo %d: PasswordCommand = %v, want %v", i, load.PasswordCommand, orig.PasswordCommand)
		}
	}
}
