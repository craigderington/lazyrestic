package restic

import (
	"testing"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestClient_Integration_ListSnapshots(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := types.RepositoryConfig{
		Path:     "/tmp/test-repo",
		Password: "testpassword",
	}

	client := NewClient(config)

	snapshots, err := client.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots() failed: %v", err)
	}

	if len(snapshots) == 0 {
		t.Error("Expected at least one snapshot")
	}

	// Check that snapshot has expected fields
	snap := snapshots[0]
	if snap.ID == "" {
		t.Error("Snapshot ID should not be empty")
	}
	if snap.Time.IsZero() {
		t.Error("Snapshot time should not be zero")
	}
}

func TestClient_Integration_GetRepositoryInfo(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := types.RepositoryConfig{
		Path:     "/tmp/test-repo",
		Password: "testpassword",
	}

	client := NewClient(config)

	info, err := client.GetRepositoryInfo()
	if err != nil {
		t.Fatalf("GetRepositoryInfo() failed: %v", err)
	}

	if info.Size == 0 {
		t.Error("Repository size should not be zero")
	}
	if info.SnapshotCount == 0 {
		t.Error("Snapshot count should not be zero")
	}
}
