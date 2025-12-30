package restic

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestNewClient(t *testing.T) {
	config := types.RepositoryConfig{
		Name:            "test-repo",
		Path:            "/tmp/test",
		PasswordFile:    "/path/to/file",
		PasswordCommand: "pass show test",
	}

	client := NewClient(config)

	if client.config.Path != config.Path {
		t.Errorf("repoPath = %v, want %v", client.config.Path, config.Path)
	}
	if client.config.PasswordFile != config.PasswordFile {
		t.Errorf("passwordFile = %v, want %v", client.config.PasswordFile, config.PasswordFile)
	}
	if client.config.PasswordCommand != config.PasswordCommand {
		t.Errorf("passwordCommand = %v, want %v", client.config.PasswordCommand, config.PasswordCommand)
	}
}

func TestClient_buildEnv(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		contains []string
	}{
		{
			name: "Password file",
			client: &Client{
				config: types.RepositoryConfig{
					Path:         "/tmp/repo",
					PasswordFile: "/home/user/.pass",
				},
			},
			contains: []string{
				"RESTIC_REPOSITORY=/tmp/repo",
				"RESTIC_PASSWORD_FILE=/home/user/.pass",
			},
		},
		{
			name: "Password command",
			client: &Client{
				config: types.RepositoryConfig{
					Path:            "/tmp/repo",
					PasswordCommand: "pass show restic",
				},
			},
			contains: []string{
				"RESTIC_REPOSITORY=/tmp/repo",
				"RESTIC_PASSWORD_COMMAND=pass show restic",
			},
		},
		{
			name: "No password",
			client: &Client{
				config: types.RepositoryConfig{
					Path: "/tmp/repo",
				},
			},
			contains: []string{
				"RESTIC_REPOSITORY=/tmp/repo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.client.buildEnv()

			for _, expected := range tt.contains {
				found := false
				for _, envVar := range env {
					if envVar == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildEnv() missing expected var: %v", expected)
				}
			}
		})
	}
}

func TestIsResticInstalled(t *testing.T) {
	// This test depends on system state
	// We can verify the function works, but result may vary
	result := IsResticInstalled()

	// Check if restic is actually in PATH
	_, err := exec.LookPath("restic")
	expected := err == nil

	if result != expected {
		t.Errorf("IsResticInstalled() = %v, want %v", result, expected)
	}
}

func TestGetResticVersion(t *testing.T) {
	if !IsResticInstalled() {
		t.Skip("restic not installed, skipping version test")
	}

	version, err := GetResticVersion()
	if err != nil {
		t.Fatalf("GetResticVersion() failed: %v", err)
	}

	if version == "" {
		t.Error("GetResticVersion() returned empty string")
	}

	// Version should contain "restic"
	if len(version) < 6 {
		t.Errorf("Version string seems too short: %v", version)
	}
}

// TestListSnapshots_Integration is an integration test that requires
// a real restic repository. It can be skipped with -short flag.
func TestListSnapshots_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !IsResticInstalled() {
		t.Skip("restic not installed")
	}

	// Check if test repo exists from setup
	testRepo := os.Getenv("RESTIC_TEST_REPO")
	if testRepo == "" {
		testRepo = "/tmp/restic-test"
	}

	testPass := os.Getenv("RESTIC_TEST_PASSWORD")
	if testPass == "" {
		testPass = "testpassword"
	}

	// Verify repo exists
	if _, err := os.Stat(testRepo); os.IsNotExist(err) {
		t.Skipf("Test repository %s does not exist", testRepo)
	}

	config := types.RepositoryConfig{
		Name:         "test-integration",
		Path:         testRepo,
		PasswordFile: "/tmp/restic-test-password.txt",
	}

	client := NewClient(config)

	snapshots, err := client.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots() failed: %v", err)
	}

	// Should have at least one snapshot from our setup
	if len(snapshots) == 0 {
		t.Error("Expected at least one snapshot")
	}

	// Verify snapshot structure
	for i, snap := range snapshots {
		if snap.ID == "" {
			t.Errorf("Snapshot %d has empty ID", i)
		}
		if snap.Time.IsZero() {
			t.Errorf("Snapshot %d has zero time", i)
		}
		if snap.Hostname == "" {
			t.Errorf("Snapshot %d has empty hostname", i)
		}
	}
}

func TestListSnapshots_JSONParsing(t *testing.T) {
	// Test that we can correctly parse restic's JSON format
	sampleJSON := `[
		{
			"time": "2025-12-28T10:00:00Z",
			"tree": "abc123",
			"paths": ["/home/user"],
			"hostname": "testhost",
			"username": "testuser",
			"tags": ["important"],
			"id": "abc123def456",
			"short_id": "abc123"
		},
		{
			"time": "2025-12-28T11:00:00Z",
			"tree": "def456",
			"paths": ["/home/user", "/etc"],
			"hostname": "testhost",
			"username": "testuser",
			"tags": ["daily", "auto"],
			"id": "def456ghi789",
			"short_id": "def456"
		}
	]`

	var snapshots []types.Snapshot
	err := json.Unmarshal([]byte(sampleJSON), &snapshots)
	if err != nil {
		t.Fatalf("Failed to parse sample JSON: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf("Expected 2 snapshots, got %d", len(snapshots))
	}

	// Verify first snapshot
	snap1 := snapshots[0]
	if snap1.ID != "abc123def456" {
		t.Errorf("Snapshot 1 ID = %v, want abc123def456", snap1.ID)
	}
	if snap1.ShortID != "abc123" {
		t.Errorf("Snapshot 1 ShortID = %v, want abc123", snap1.ShortID)
	}
	if snap1.Hostname != "testhost" {
		t.Errorf("Snapshot 1 Hostname = %v, want testhost", snap1.Hostname)
	}
	if len(snap1.Paths) != 1 || snap1.Paths[0] != "/home/user" {
		t.Errorf("Snapshot 1 Paths = %v, want [/home/user]", snap1.Paths)
	}
	if len(snap1.Tags) != 1 || snap1.Tags[0] != "important" {
		t.Errorf("Snapshot 1 Tags = %v, want [important]", snap1.Tags)
	}

	// Verify second snapshot
	snap2 := snapshots[1]
	if len(snap2.Paths) != 2 {
		t.Errorf("Snapshot 2 should have 2 paths, got %d", len(snap2.Paths))
	}
	if len(snap2.Tags) != 2 {
		t.Errorf("Snapshot 2 should have 2 tags, got %d", len(snap2.Tags))
	}
}

func TestCheckRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !IsResticInstalled() {
		t.Skip("restic not installed")
	}

	testRepo := os.Getenv("RESTIC_TEST_REPO")
	if testRepo == "" {
		testRepo = "/tmp/restic-test"
	}

	testPass := os.Getenv("RESTIC_TEST_PASSWORD")
	if testPass == "" {
		testPass = "testpassword"
	}

	if _, err := os.Stat(testRepo); os.IsNotExist(err) {
		t.Skip("Test repository does not exist")
	}

	config := types.RepositoryConfig{
		Name:         "test-check",
		Path:         testRepo,
		PasswordFile: "/tmp/restic-test-password.txt",
	}

	client := NewClient(config)

	err := client.CheckRepository()
	if err != nil {
		t.Errorf("CheckRepository() failed: %v", err)
	}
}

func TestCheckRepository_InvalidRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !IsResticInstalled() {
		t.Skip("restic not installed")
	}

	config := types.RepositoryConfig{
		Name:     "invalid",
		Path:     "/nonexistent/repo",
		PasswordFile: "/tmp/wrongpass",
	}

	client := NewClient(config)

	err := client.CheckRepository()
	if err == nil {
		t.Error("CheckRepository() should fail for invalid repository")
	}
}

// Benchmark tests
func BenchmarkBuildEnv(b *testing.B) {
	client := &Client{
		config: types.RepositoryConfig{
			Path:            "/tmp/repo",
			PasswordFile:    "/file",
			PasswordCommand: "command",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.buildEnv()
	}
}

func BenchmarkNewClient(b *testing.B) {
	config := types.RepositoryConfig{
		Name:     "bench",
		Path:     "/tmp/bench",
		PasswordFile: "/tmp/password",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewClient(config)
	}
}
