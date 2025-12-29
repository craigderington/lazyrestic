package restic

import (
	"os"
	"testing"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestGetStats(t *testing.T) {
	testRepo := os.Getenv("RESTIC_TEST_REPO")
	if testRepo == "" {
		t.Skip("Skipping integration test: RESTIC_TEST_REPO not set")
	}

	config := types.RepositoryConfig{
		Path:     testRepo,
		Password: os.Getenv("RESTIC_PASSWORD"),
	}

	client := NewClient(config)

	stats, err := client.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats.TotalSize == 0 {
		t.Error("Expected total_size > 0")
	}

	if stats.SnapshotsCount == 0 {
		t.Error("Expected snapshots_count > 0")
	}

	t.Logf("Stats: Size=%d, Files=%d, Snapshots=%d",
		stats.TotalSize, stats.TotalFileCount, stats.SnapshotsCount)
}

func TestGetRepositoryInfo(t *testing.T) {
	testRepo := os.Getenv("RESTIC_TEST_REPO")
	if testRepo == "" {
		t.Skip("Skipping integration test: RESTIC_TEST_REPO not set")
	}

	config := types.RepositoryConfig{
		Path:     testRepo,
		Password: os.Getenv("RESTIC_PASSWORD"),
	}

	client := NewClient(config)

	repo, err := client.GetRepositoryInfo()
	if err != nil {
		t.Fatalf("GetRepositoryInfo() failed: %v", err)
	}

	if repo.Status == "unknown" || repo.Status == "error" {
		t.Errorf("Expected status to be healthy or warning, got: %s", repo.Status)
	}

	if repo.Size == 0 {
		t.Error("Expected repository size > 0")
	}

	if repo.SnapshotCount == 0 {
		t.Error("Expected snapshot count > 0")
	}

	if repo.LastBackup.IsZero() {
		t.Error("Expected last backup time to be set")
	}

	t.Logf("Repository Info:")
	t.Logf("  Status: %s", repo.Status)
	t.Logf("  Size: %d bytes", repo.Size)
	t.Logf("  Files: %d", repo.TotalFiles)
	t.Logf("  Snapshots: %d", repo.SnapshotCount)
	t.Logf("  Last Backup: %s", repo.LastBackup)
}
