package restic

import (
	"context"
	"os"
	"testing"

	"github.com/craigderington/lazyrestic/pkg/types"
)

// TestBackupWithChannel tests the channel-based backup approach
func TestBackupWithChannel(t *testing.T) {
	// Skip if RESTIC_TEST_REPO not set
	testRepo := os.Getenv("RESTIC_TEST_REPO")
	if testRepo == "" {
		t.Skip("Skipping integration test: RESTIC_TEST_REPO not set")
	}

	// Create a temp file to backup
	tmpFile, err := os.CreateTemp("", "restic-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("test content for backup"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	config := types.RepositoryConfig{
		Path:     testRepo,
		PasswordFile: "/tmp/restic-test-password.txt",
	}

	client := NewClient(config)

	// Test channel-based backup
	opts := types.BackupOptions{
		Paths: []string{tmpFile.Name()},
		Tags:  []string{"test", "integration"},
	}

	updates := make(chan BackupMessage, 10)

	// Start backup in goroutine
	ctx := context.Background()
	go client.BackupWithChannel(ctx, opts, updates)

	// Collect all messages
	var progressCount int
	var summaryReceived bool
	var errorReceived error

	for msg := range updates {
		if msg.Error != nil {
			errorReceived = msg.Error
			break
		}

		if msg.Progress != nil {
			progressCount++
			t.Logf("Progress: %.1f%% (%d/%d files)",
				msg.Progress.PercentDone,
				msg.Progress.FilesDone,
				msg.Progress.TotalFiles)
		}

		if msg.Summary != nil {
			summaryReceived = true
			t.Logf("Summary: %d new files, %d changed, %d unmodified",
				msg.Summary.FilesNew,
				msg.Summary.FilesChanged,
				msg.Summary.FilesUnmodified)
		}
	}

	if errorReceived != nil {
		t.Fatalf("Backup failed with error: %v", errorReceived)
	}

	if !summaryReceived {
		t.Error("Expected to receive backup summary, but didn't")
	}

	t.Logf("Received %d progress updates", progressCount)
}
