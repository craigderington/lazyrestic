package model

import (
	"github.com/craigderington/lazyrestic/pkg/restic"
	"github.com/craigderington/lazyrestic/pkg/types"
	"github.com/craigderington/lazyrestic/pkg/ui"
)

// Model represents the application state
type Model struct {
	width    int
	height   int
	ready    bool
	tooSmall bool // Terminal too small to display properly

	// Configuration
	config *types.ResticConfig

	// Current state
	activePanel        types.Panel
	repositories       []types.Repository
	currentRepoIndex   int
	loadingSnapshots   bool
	loadingRepositories bool

	// UI Panels
	repoPanel    *ui.RepositoryPanel
	metricsPanel *ui.RepoMetricsPanel
	snapPanel    *ui.SnapshotPanel
	opsPanel     *ui.OperationsPanel
	showHelp     bool

	// Repo creation
	showRepoForm bool
	repoForm     *ui.RepoForm

	// Backup state
	showBackupForm        bool
	backupForm            *ui.BackupForm
	backupInProgress      bool
	currentBackupProgress *types.BackupProgress

	// Restore state
	showRestoreForm        bool
	restoreForm            *ui.RestoreForm
	restoreInProgress      bool
	currentRestoreProgress *types.RestoreProgress

	// Filter state
	filterInputActive bool
	filterInputText   string

	// File browser state
	showFileBrowser bool
	fileBrowser     *ui.FileBrowser

	// Found repos state
	showFoundRepos bool
	foundRepos     []types.RepositoryConfig
	selectedFound  int

	// Forget/Prune state
	showForgetForm       bool
	forgetForm           *ui.ForgetForm
	showForgetPreview    bool
	forgetPreview        *ui.ForgetPreview
	showForgetConfirm    bool
	forgetConfirmDialog  *ui.ConfirmationDialog
	forgetPreviewResults []types.ForgetResult
	forgetPolicy         types.ForgetPolicy
	showPruneConfirm     bool
	pruneConfirmDialog   *ui.ConfirmationDialog
	pruneDryRunOutput    string

	// Remove repository state
	showRemoveConfirm   bool
	removeConfirmDialog *ui.ConfirmationDialog
	repoToRemove        string // Name of repository to remove
}

// RepositoriesLoadedMsg is sent when repositories are loaded
type RepositoriesLoadedMsg struct {
	Repositories []types.Repository
}

// SnapshotsLoadStartMsg is sent when snapshot loading starts
type SnapshotsLoadStartMsg struct {
	RepoName string
	RepoPath string
}

// SnapshotsLoadedMsg is sent when snapshots are loaded
type SnapshotsLoadedMsg struct {
	Snapshots     []types.Snapshot
	Error         error
	FilteredCount int
	CmdLog        SnapshotsLoadStartMsg
}

// FilesLoadedMsg is sent when files are loaded from a snapshot
type FilesLoadedMsg struct {
	Files []types.FileNode
	Error error
}

// BackupProgressMsg is sent during backup operations
type BackupProgressMsg struct {
	Progress *types.BackupProgress
	Updates  <-chan restic.BackupMessage // Channel to continue listening
}

// BackupSummaryMsg is sent when backup completes
type BackupSummaryMsg struct {
	Summary *types.BackupSummary
	Error   error
}

// RestoreProgressMsg is sent during restore operations
type RestoreProgressMsg struct {
	Progress *types.RestoreProgress
	Updates  <-chan restic.RestoreMessage
}

// RestoreSummaryMsg is sent when restore completes
type RestoreSummaryMsg struct {
	Summary *types.RestoreSummary
	Error   error
}

// ForgetDryRunMsg is sent when forget dry-run completes
type ForgetDryRunMsg struct {
	Results []types.ForgetResult
	Policy  types.ForgetPolicy
	Error   error
}

// ForgetCompleteMsg is sent when forget operation completes
type ForgetCompleteMsg struct {
	Error error
}

// PruneDryRunMsg is sent when prune dry-run completes
type PruneDryRunMsg struct {
	Output string
	Error  error
}

// PruneCompleteMsg is sent when prune operation completes
type PruneCompleteMsg struct {
	Error error
}

// ScannedReposMsg is sent when repository scanning completes
type ScannedReposMsg struct {
	FoundRepos []types.RepositoryConfig
}

// CacheCleanupMsg is sent when cache cleanup completes
type CacheCleanupMsg struct {
	Output string
	Error  error
}

// UnlockMsg is sent when repository unlock completes
type UnlockMsg struct {
	Output string
	Error  error
}

// RepoRemovedMsg is sent when a repository is removed from config
type RepoRemovedMsg struct {
	RepoName string
	Error    error
}
