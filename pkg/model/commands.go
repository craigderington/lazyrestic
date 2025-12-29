package model

import (
	"context"
	"fmt"

	"github.com/craigderington/lazyrestic/pkg/restic"
	"github.com/craigderington/lazyrestic/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
)

// executeBackup performs a backup operation with progress tracking
func (m Model) executeBackup(opts types.BackupOptions) tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return BackupSummaryMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		// Create a channel for backup updates
		updates := make(chan restic.BackupMessage, 10)

		// Start the backup in a goroutine
		ctx := context.Background()
		go client.BackupWithChannel(ctx, opts, updates)

		// Wait for the first message
		return waitForBackupUpdate(updates)
	}
}

// waitForBackupUpdate waits for a backup update from the channel
func waitForBackupUpdate(updates <-chan restic.BackupMessage) tea.Msg {
	msg, ok := <-updates
	if !ok {
		// Channel closed, backup is done (no summary was sent)
		return BackupSummaryMsg{Error: nil}
	}

	if msg.Error != nil {
		return BackupSummaryMsg{Error: msg.Error}
	}

	if msg.Progress != nil {
		// Return progress and pass the channel along to continue listening
		return BackupProgressMsg{
			Progress: msg.Progress,
			Updates:  updates,
		}
	}

	if msg.Summary != nil {
		return BackupSummaryMsg{Summary: msg.Summary, Error: nil}
	}

	// Empty message, continue listening
	return BackupProgressMsg{Progress: nil, Updates: updates}
}

// listenForBackupUpdates continues listening for backup progress updates
func listenForBackupUpdates(updates <-chan restic.BackupMessage) tea.Cmd {
	return func() tea.Msg {
		return waitForBackupUpdate(updates)
	}
}

// executeRestore performs a restore operation with progress tracking
func (m Model) executeRestore(opts types.RestoreOptions) tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return RestoreSummaryMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		// Create a channel for restore updates
		updates := make(chan restic.RestoreMessage, 10)

		// Start the restore in a goroutine
		ctx := context.Background()
		go client.RestoreWithChannel(ctx, opts, updates)

		// Wait for the first message
		return waitForRestoreUpdate(updates)
	}
}

// waitForRestoreUpdate waits for a restore update from the channel
func waitForRestoreUpdate(updates <-chan restic.RestoreMessage) tea.Msg {
	msg, ok := <-updates
	if !ok {
		// Channel closed, restore is done (no summary was sent)
		return RestoreSummaryMsg{Error: nil}
	}

	if msg.Error != nil {
		return RestoreSummaryMsg{Error: msg.Error}
	}

	if msg.Progress != nil {
		// Return progress and pass the channel along to continue listening
		return RestoreProgressMsg{
			Progress: msg.Progress,
			Updates:  updates,
		}
	}

	if msg.Summary != nil {
		return RestoreSummaryMsg{Summary: msg.Summary, Error: nil}
	}

	// Empty message, continue listening
	return RestoreProgressMsg{Progress: nil, Updates: updates}
}

// listenForRestoreUpdates continues listening for restore progress updates
func listenForRestoreUpdates(updates <-chan restic.RestoreMessage) tea.Cmd {
	return func() tea.Msg {
		return waitForRestoreUpdate(updates)
	}
}

// executeForgetDryRun performs a dry-run of the forget operation
func (m Model) executeForgetDryRun(policy types.ForgetPolicy) tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return ForgetDryRunMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		results, err := client.ForgetDryRun(policy)
		return ForgetDryRunMsg{
			Results: results,
			Policy:  policy,
			Error:   err,
		}
	}
}

// executeForget performs the actual forget operation
func (m Model) executeForget(policy types.ForgetPolicy) tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return ForgetCompleteMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		err := client.Forget(policy)
		return ForgetCompleteMsg{Error: err}
	}
}

// executePruneDryRun performs a dry-run of the prune operation
func (m Model) executePruneDryRun() tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return PruneDryRunMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		output, err := client.PruneDryRun()
		return PruneDryRunMsg{
			Output: output,
			Error:  err,
		}
	}
}

// executePrune performs the actual prune operation
func (m Model) executePrune() tea.Cmd {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return func() tea.Msg {
			return PruneCompleteMsg{Error: fmt.Errorf("no repository selected")}
		}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	return func() tea.Msg {
		err := client.Prune()
		return PruneCompleteMsg{Error: err}
	}
}
