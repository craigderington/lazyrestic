package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/config"
	"github.com/craigderington/lazyrestic/pkg/restic"
	"github.com/craigderington/lazyrestic/pkg/types"
	"github.com/craigderington/lazyrestic/pkg/ui"
)

// NewModel creates a new instance of the application model
func NewModel() Model {
	// Load configuration
	cfg := config.LoadOrDefault("")

	// Initialize panels
	repoPanel := ui.NewRepositoryPanel()
	metricsPanel := ui.NewRepoMetricsPanel()
	snapPanel := ui.NewSnapshotPanel()
	opsPanel := ui.NewOperationsPanel()
	backupForm := ui.NewBackupForm()
	repoForm := ui.NewRepoForm()

	// Initial log messages - polished startup
	opsPanel.Success("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	opsPanel.Success("✓ LazyRestic TUI started successfully")
	opsPanel.Dimmed("Version 0.1.0 - Terminal UI for restic backup management")

	if !restic.IsResticInstalled() {
		opsPanel.Error("✗ restic binary not found in PATH")
		opsPanel.Warning("Please install restic: https://restic.net")
	} else {
		if version, err := restic.GetResticVersion(); err == nil {
			opsPanel.Success(fmt.Sprintf("✓ %s detected", version))
			opsPanel.Dimmed("Ready for backup operations")
		}
	}
	opsPanel.Info("Press '?' for help or 'q' to quit")
	opsPanel.Success("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return Model{
		ready:                  false,
		config:                 cfg,
		activePanel:            types.PanelRepositories,
		repositories:           []types.Repository{},
		currentRepoIndex:       0,
		loadingSnapshots:       false,
		loadingRepositories:    true, // Start with loading state
		repoPanel:              repoPanel,
		metricsPanel:           metricsPanel,
		snapPanel:              snapPanel,
		opsPanel:               opsPanel,
		showHelp:               false,
		showRepoForm:           false,
		repoForm:               repoForm,
		showBackupForm:         false,
		backupForm:             backupForm,
		backupInProgress:       false,
		currentBackupProgress:  nil,
		showRestoreForm:        false,
		restoreForm:            nil, // Created when needed
		restoreInProgress:      false,
		currentRestoreProgress: nil,
	}
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return m.loadRepositories
}

// loadRepositories loads repository information
func (m Model) loadRepositories() tea.Msg {
	var repos []types.Repository

	for _, repoConfig := range m.config.Repositories {
		client := restic.NewClient(repoConfig)

		// Get comprehensive repository information
		repoInfo, err := client.GetRepositoryInfo()
		if err != nil {
			// If we can't get info, create a minimal repo entry
			repo := types.Repository{
				Name:   repoConfig.Name,
				Path:   repoConfig.Path,
				Status: "error",
			}
			repos = append(repos, repo)
			continue
		}

		// Set the name and path from config
		repoInfo.Name = repoConfig.Name
		repoInfo.Path = repoConfig.Path

		repos = append(repos, *repoInfo)
	}

	return RepositoriesLoadedMsg{Repositories: repos}
}

// loadSnapshotsWithMessage shows loading message and loads snapshots
func (m *Model) loadSnapshotsWithMessage() tea.Cmd {
	m.loadingSnapshots = true
	m.opsPanel.Info("Loading snapshots...")
	return m.loadSnapshots
}

// loadSnapshots loads snapshots for the current repository
func (m Model) loadSnapshots() tea.Msg {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return SnapshotsLoadedMsg{Error: fmt.Errorf("no repository selected")}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]

	// Log the command being executed
	cmdLog := SnapshotsLoadStartMsg{
		RepoName: repoConfig.Name,
		RepoPath: repoConfig.Path,
	}

	client := restic.NewClient(repoConfig)

	snapshots, err := client.ListSnapshots()

	// Filter out systemd-private snapshots
	var filteredCount int
	if err == nil {
		filtered := make([]types.Snapshot, 0)
		for _, snap := range snapshots {
			// Check if any path starts with /systemd-private or contains systemd-private
			shouldInclude := true
			for _, path := range snap.Paths {
				if strings.Contains(path, "systemd-private") {
					shouldInclude = false
					filteredCount++
					break
				}
			}
			if shouldInclude {
				filtered = append(filtered, snap)
			}
		}
		snapshots = filtered
	}

	return SnapshotsLoadedMsg{
		Snapshots:     snapshots,
		Error:         err,
		FilteredCount: filteredCount,
		CmdLog:        cmdLog,
	}
}

// cleanupCache runs restic cache --cleanup for the current repository
func (m Model) cleanupCache() tea.Cmd {
	return func() tea.Msg {
		if m.currentRepoIndex >= len(m.config.Repositories) {
			return CacheCleanupMsg{Error: fmt.Errorf("no repository selected")}
		}

		repoConfig := m.config.Repositories[m.currentRepoIndex]
		client := restic.NewClient(repoConfig)

		output, err := client.CleanupCache()
		return CacheCleanupMsg{
			Output: output,
			Error:  err,
		}
	}
}

// unlockRepository runs restic unlock for the current repository
func (m Model) unlockRepository() tea.Cmd {
	return func() tea.Msg {
		if m.currentRepoIndex >= len(m.config.Repositories) {
			return UnlockMsg{Error: fmt.Errorf("no repository selected")}
		}

		repoConfig := m.config.Repositories[m.currentRepoIndex]
		client := restic.NewClient(repoConfig)

		output, err := client.Unlock()
		return UnlockMsg{
			Output: output,
			Error:  err,
		}
	}
}

// removeRepository removes a repository from the configuration
func (m Model) removeRepository() tea.Cmd {
	return func() tea.Msg {
		// Remove from config
		removed := config.RemoveRepository(m.config, m.repoToRemove)
		if !removed {
			return RepoRemovedMsg{
				RepoName: m.repoToRemove,
				Error:    fmt.Errorf("repository not found in configuration"),
			}
		}

		// Save updated config
		configPath := config.DefaultConfigPath()
		if err := config.Save(m.config, configPath); err != nil {
			return RepoRemovedMsg{
				RepoName: m.repoToRemove,
				Error:    fmt.Errorf("failed to save config: %w", err),
			}
		}

		return RepoRemovedMsg{
			RepoName: m.repoToRemove,
			Error:    nil,
		}
	}
}

// scanForRepositories scans common locations for restic repositories
func (m Model) scanForRepositories() tea.Cmd {
	return func() tea.Msg {
		foundRepos := []types.RepositoryConfig{}

		// Common locations to scan
		scanPaths := []string{
			"/mnt",
			"/media",
			"/run/media",
			"./",
			"~/Documents",
			"~/Downloads",
			"~/Backup",
			"/tmp",
		}

		for _, basePath := range scanPaths {
			// Expand ~ to home
			if strings.HasPrefix(basePath, "~/") {
				home, _ := os.UserHomeDir()
				basePath = filepath.Join(home, basePath[2:])
			}

			// Scan directory for restic repos
			foundRepos = append(foundRepos, scanDirectoryForRepos(basePath)...)
		}

		return ScannedReposMsg{FoundRepos: foundRepos}
	}
}

// scanDirectoryForRepos recursively scans a directory for restic repositories
func scanDirectoryForRepos(basePath string) []types.RepositoryConfig {
	var repos []types.RepositoryConfig

	// Check if basePath itself is a restic repo
	if isResticRepo(basePath) {
		repoName := filepath.Base(basePath)
		if repoName == "." {
			repoName = "local-repo"
		}
		// Filter out systemd repos
		if !strings.HasPrefix(repoName, "systemd") && !strings.Contains(basePath, "systemd-private") {
			repos = append(repos, types.RepositoryConfig{
				Name: repoName,
				Path: basePath,
			})
		}
	}

	// Walk the directory tree (but not too deep to avoid performance issues)
	filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Don't go too deep
		relPath, _ := filepath.Rel(basePath, path)
		depth := strings.Count(relPath, string(filepath.Separator))
		if depth > 2 { // Max depth 2
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() && isResticRepo(path) {
			repoName := filepath.Base(path)
			// Filter out systemd repos
			if !strings.HasPrefix(repoName, "systemd") && !strings.Contains(path, "systemd-private") {
				repos = append(repos, types.RepositoryConfig{
					Name: repoName,
					Path: path,
				})
			}
			return filepath.SkipDir // Don't scan inside repos
		}

		return nil
	})

	return repos
}

// isResticRepo checks if a directory contains a restic repository
func isResticRepo(path string) bool {
	requiredFiles := []string{"config", "data", "keys", "snapshots"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(path, file)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// loadFiles loads files from the current path in the file browser
func (m Model) loadFiles() tea.Msg {
	if m.currentRepoIndex >= len(m.config.Repositories) {
		return FilesLoadedMsg{Error: fmt.Errorf("no repository selected")}
	}

	if m.fileBrowser == nil || m.fileBrowser.GetSnapshot() == nil {
		return FilesLoadedMsg{Error: fmt.Errorf("no snapshot selected for browsing")}
	}

	repoConfig := m.config.Repositories[m.currentRepoIndex]
	client := restic.NewClient(repoConfig)

	currentPath := m.fileBrowser.GetCurrentPath()
	files, err := client.ListFiles(m.fileBrowser.GetSnapshot().ID, currentPath)

	return FilesLoadedMsg{
		Files: files,
		Error: err,
	}

}

// logSelectedSnapshot logs details about the currently selected snapshot to the operations panel
func (m *Model) logSelectedSnapshot() {
	snapshot := m.snapPanel.GetSelected()
	if snapshot == nil {
		return
	}

	// Log snapshot details with visual separator
	m.opsPanel.Success("─────────────────────────────────────────────────────────")
	m.opsPanel.Success(fmt.Sprintf("📸 Snapshot: %s", snapshot.ShortID))
	m.opsPanel.Dimmed(fmt.Sprintf("Full ID: %s", snapshot.ID))
	m.opsPanel.Info(fmt.Sprintf("Created: %s", snapshot.Time.Format("2006-01-02 15:04:05")))
	m.opsPanel.Info(fmt.Sprintf("Hostname: %s", snapshot.Hostname))

	if len(snapshot.Paths) > 0 {
		m.opsPanel.Info(fmt.Sprintf("Paths: %s", strings.Join(snapshot.Paths, ", ")))
	}

	if len(snapshot.Tags) > 0 {
		m.opsPanel.Info(fmt.Sprintf("Tags: %s", strings.Join(snapshot.Tags, ", ")))
	}

	if snapshot.Username != "" {
		m.opsPanel.Dimmed(fmt.Sprintf("User: %s", snapshot.Username))
	}

	m.opsPanel.Dimmed(fmt.Sprintf("Time ago: %s", ui.FormatTimeAgo(snapshot.Time)))
	m.opsPanel.Success("─────────────────────────────────────────────────────────")
}

// listenForRestoreUpdates returns a command that listens for more restore updates

// Update handles incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Check if terminal is too small
		if m.width < ui.MinTerminalWidth || m.height < ui.MinTerminalHeight {
			m.tooSmall = true
			return m, nil
		}
		m.tooSmall = false

		// New 4-panel layout:
		// - Left column (1/3 width, full height): Repos / Metrics / Snapshots stacked
		// - Right column (2/3 width, full height): Operations

		leftWidth := int(float64(m.width) * ui.LeftPanelWidthRatio)
		rightWidth := m.width - leftWidth

		// Account for title and help (already includes margins in TitleAndHelpHeight)
		panelHeight := m.height - ui.TitleAndHelpHeight

		// Left column: balanced distribution
		// Repos: 40%, Metrics: 36%, Snapshots: 24%
		repoHeight := int(float64(panelHeight) * 0.40)
		metricsHeight := int(float64(panelHeight) * 0.36)
		snapshotsHeight := panelHeight - repoHeight - metricsHeight // Remainder goes to snapshots

		m.repoPanel.SetSize(leftWidth, repoHeight)
		m.metricsPanel.SetSize(leftWidth, metricsHeight)
		m.snapPanel.SetSize(leftWidth, snapshotsHeight)

		// Right column: operations takes full height
		m.opsPanel.SetSize(rightWidth, panelHeight)

		formWidth := int(float64(m.width) * ui.FormWidthRatio)
		formHeight := int(float64(m.height) * ui.FormHeightRatio)
		m.repoForm.SetSize(formWidth, formHeight)
		m.backupForm.SetSize(formWidth, formHeight)

		return m, nil

	case RepositoriesLoadedMsg:
		m.loadingRepositories = false
		m.repositories = msg.Repositories
		m.opsPanel.Success(fmt.Sprintf("✓ Loaded %d repositories from config", len(msg.Repositories)))
		if len(msg.Repositories) == 0 {
			m.opsPanel.Dimmed("No repositories configured")
			m.opsPanel.Info("Press 'a' to add repository or 's' to scan for existing repos")
			m.opsPanel.Dimmed("Config: ~/.config/lazyrestic/config.yaml")
			m.metricsPanel.SetRepository(nil)
			return m, nil
		} else {
			// Update metrics panel with currently selected repo
			if m.currentRepoIndex < len(m.repositories) {
				selectedRepo := &m.repositories[m.currentRepoIndex]
				m.metricsPanel.SetRepository(selectedRepo)
				m.opsPanel.Info(fmt.Sprintf("Selected repository: '%s' at %s", selectedRepo.Name, selectedRepo.Path))
			}
			// Load snapshots for the selected repository
			return m, m.loadSnapshotsWithMessage()
		}

	case SnapshotsLoadedMsg:
		m.loadingSnapshots = false
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Failed to load snapshots from '%s': %v", msg.CmdLog.RepoName, msg.Error))
			m.opsPanel.Dimmed(fmt.Sprintf("Repository: %s", msg.CmdLog.RepoPath))
		} else {
			m.snapPanel.SetSnapshots(msg.Snapshots)
			m.opsPanel.Success(fmt.Sprintf("✓ Loaded %d snapshots from '%s'", len(msg.Snapshots), msg.CmdLog.RepoName))
			m.opsPanel.Dimmed(fmt.Sprintf("Repository path: %s", msg.CmdLog.RepoPath))
			if msg.FilteredCount > 0 {
				m.opsPanel.Dimmed(fmt.Sprintf("Filtered %d systemd-private snapshots", msg.FilteredCount))
			}
			m.opsPanel.Info(fmt.Sprintf("Command: restic -r %s snapshots --json", msg.CmdLog.RepoPath))

			// Log the currently selected snapshot details
			if len(msg.Snapshots) > 0 {
				m.logSelectedSnapshot()
			}
		}
		return m, nil

	case FilesLoadedMsg:
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Failed to load files: %v", msg.Error))
		} else if m.fileBrowser != nil {
			m.fileBrowser.SetFiles(msg.Files)
			m.opsPanel.Info(fmt.Sprintf("Loaded %d files/directories", len(msg.Files)))
		}
		return m, nil

	case BackupProgressMsg:
		m.currentBackupProgress = msg.Progress

		// Update operations panel with progress
		if msg.Progress != nil {
			m.opsPanel.SetBackupProgress(msg.Progress)
		}

		// Continue listening for more updates if channel is still open
		if msg.Updates != nil {
			return m, listenForBackupUpdates(msg.Updates)
		}

		return m, nil

	case BackupSummaryMsg:
		m.backupInProgress = false
		m.currentBackupProgress = nil
		m.opsPanel.ClearBackupProgress()

		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Backup failed: %v", msg.Error))
		} else if msg.Summary != nil {
			m.opsPanel.Success(fmt.Sprintf("Backup completed! New: %d, Changed: %d, Unmodified: %d",
				msg.Summary.FilesNew, msg.Summary.FilesChanged, msg.Summary.FilesUnmodified))
		} else {
			m.opsPanel.Success("Backup completed successfully")
		}

		// Reload snapshots to show the new backup
		return m, m.loadSnapshotsWithMessage()

	case RestoreProgressMsg:
		m.currentRestoreProgress = msg.Progress

		// Update operations panel with progress
		if msg.Progress != nil {
			m.opsPanel.Info("Restoring snapshot...")
		}

		// Continue listening for more updates if channel is still open
		if msg.Updates != nil {
			return m, listenForRestoreUpdates(msg.Updates)
		}

		return m, nil

	case RestoreSummaryMsg:
		m.restoreInProgress = false
		m.currentRestoreProgress = nil

		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Restore failed: %v", msg.Error))
		} else if msg.Summary != nil {
			m.opsPanel.Success("Restore completed successfully")
		} else {
			m.opsPanel.Success("Restore completed")
		}

		return m, nil

	case ForgetDryRunMsg:
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Forget dry-run failed: %v", msg.Error))
			m.showForgetForm = false
			return m, nil
		}

		// Show preview with results
		m.forgetPreviewResults = msg.Results
		m.forgetPolicy = msg.Policy
		m.forgetPreview = ui.NewForgetPreview(msg.Results, msg.Policy)
		m.forgetPreview.SetSize(m.width*3/4, m.height*3/4)
		m.showForgetForm = false
		m.showForgetPreview = true

		totalRemove := m.forgetPreview.GetTotalToRemove()
		m.opsPanel.Info(fmt.Sprintf("Dry-run complete: %d snapshots will be removed", totalRemove))
		return m, nil

	case ForgetCompleteMsg:
		m.showForgetConfirm = false
		m.forgetConfirmDialog = nil

		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Forget failed: %v", msg.Error))
		} else {
			totalRemoved := 0
			for _, result := range m.forgetPreviewResults {
				totalRemoved += len(result.SnapshotsToRemove)
			}
			m.opsPanel.Success(fmt.Sprintf("✓ Forget completed: %d snapshots removed", totalRemoved))
		}

		// Reload snapshots
		return m, m.loadSnapshotsWithMessage()

	case PruneDryRunMsg:
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Prune dry-run failed: %v", msg.Error))
			return m, nil
		}

		// Store dry-run output and show confirmation
		m.pruneDryRunOutput = msg.Output
		m.pruneConfirmDialog = ui.NewConfirmationDialog(
			"PRUNE REPOSITORY",
			"You are about to PRUNE the repository.\n\nThis will permanently remove unreferenced data.\nThis operation CANNOT be undone!\n\n"+msg.Output,
			"PRUNE",
		)
		m.pruneConfirmDialog.SetSize(m.width*3/4, m.height*3/4)
		m.showPruneConfirm = true
		m.opsPanel.Info("Prune dry-run complete - review and confirm")
		return m, nil

	case PruneCompleteMsg:
		m.showPruneConfirm = false
		m.pruneConfirmDialog = nil

		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Prune failed: %v", msg.Error))
		} else {
			m.opsPanel.Success("Prune completed successfully")
		}
		return m, m.loadRepositories

	case ScannedReposMsg:
		if len(msg.FoundRepos) == 0 {
			m.opsPanel.Info("No restic repositories found in scanned locations")
			m.opsPanel.Dimmed("Scanned: /mnt, /media, /backup, /srv, /opt")
		} else {
			m.opsPanel.Success(fmt.Sprintf("✓ Found %d potential repositories", len(msg.FoundRepos)))
			m.opsPanel.Info("Select a repository and press Enter to add it")
			m.showFoundRepos = true
			m.foundRepos = msg.FoundRepos
			m.selectedFound = 0
		}
		return m, nil

	case CacheCleanupMsg:
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Cache cleanup failed: %v", msg.Error))
		} else {
			m.opsPanel.Success("✓ Cache cleanup completed successfully")
			if msg.Output != "" {
				m.opsPanel.Info(msg.Output)
			}
			m.opsPanel.Dimmed("Removed old/unused cache entries")
		}
		return m, nil

	case UnlockMsg:
		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("Unlock failed: %v", msg.Error))
		} else {
			m.opsPanel.Success("✓ Repository unlocked successfully")
			if msg.Output != "" {
				m.opsPanel.Info(msg.Output)
			}
			m.opsPanel.Dimmed("Stale locks removed - repository is now accessible")
			// Refresh repository info after unlock
			return m, m.loadRepositories
		}
		return m, nil

	case RepoRemovedMsg:
		m.showRemoveConfirm = false
		m.removeConfirmDialog = nil
		m.repoToRemove = ""

		if msg.Error != nil {
			m.opsPanel.Error(fmt.Sprintf("✗ Failed to remove repository: %v", msg.Error))
			m.opsPanel.Dimmed("Repository was not removed from configuration")
		} else {
			m.opsPanel.Success("─────────────────────────────────────────────────────────")
			m.opsPanel.Success(fmt.Sprintf("✓ Repository '%s' removed from LazyRestic", msg.RepoName))
			configPath := config.DefaultConfigPath()
			m.opsPanel.Dimmed(fmt.Sprintf("Configuration file updated: %s", configPath))
			m.opsPanel.Info("Repository files are still on disk - only config entry removed")
			m.opsPanel.Success("─────────────────────────────────────────────────────────")
			// Refresh repository list
			return m, m.loadRepositories
		}
		return m, nil

	case tea.KeyMsg:
		if m.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		// Handle backup form interactions
		if m.showBackupForm {
			switch msg.String() {
			case "esc":
				m.showBackupForm = false
				return m, nil

			case "enter":
				// Check which field is focused
				if m.backupForm.IsValid() {
					// Start backup
					opts := types.BackupOptions{
						Paths:   m.backupForm.GetPaths(),
						Tags:    m.backupForm.GetTags(),
						Exclude: m.backupForm.GetExclude(),
					}

					m.showBackupForm = false
					m.backupInProgress = true
					m.opsPanel.Info(fmt.Sprintf("Starting backup of %d paths...", len(opts.Paths)))

					return m, m.executeBackup(opts)
				}
			}

			// Pass other keys to the form
			var cmd tea.Cmd
			cmd = m.backupForm.Update(msg)
			return m, cmd
		}

		// Handle restore form interactions
		if m.showRestoreForm {
			switch msg.String() {
			case "esc":
				m.showRestoreForm = false
				return m, nil

			case "enter":
				// Check if form is valid
				if m.restoreForm.IsValid() {
					// Get selected snapshot
					selectedSnapshot := m.snapPanel.GetSelected()
					if selectedSnapshot == nil {
						m.opsPanel.Error("No snapshot selected")
						m.showRestoreForm = false
						return m, nil
					}

					// Start restore
					opts := types.RestoreOptions{
						SnapshotID: selectedSnapshot.ID,
						Target:     m.restoreForm.GetTarget(),
						Include:    m.restoreForm.GetInclude(),
					}

					m.showRestoreForm = false
					m.restoreInProgress = true
					m.opsPanel.Info(fmt.Sprintf("Starting restore of snapshot %s...", selectedSnapshot.ShortID))

					return m, m.executeRestore(opts)
				}
			}

			// Pass other keys to the form
			var cmd tea.Cmd
			cmd = m.restoreForm.Update(msg)
			return m, cmd
		}

		// Handle filter input mode
		if m.filterInputActive {
			switch msg.String() {
			case "esc":
				// Cancel filter input
				m.filterInputActive = false
				m.filterInputText = ""
				return m, nil

			case "enter":
				// Apply the filter
				m.snapPanel.SetFilter(m.filterInputText)
				m.filterInputActive = false
				m.opsPanel.Info(fmt.Sprintf("Filter applied: %s", m.filterInputText))
				return m, nil

			case "backspace":
				// Remove last character
				if len(m.filterInputText) > 0 {
					m.filterInputText = m.filterInputText[:len(m.filterInputText)-1]
					// Apply filter in real-time as user types
					if m.filterInputText == "" {
						m.snapPanel.ClearFilter()
					} else {
						m.snapPanel.SetFilter(m.filterInputText)
					}
				}
				return m, nil

			default:
				// Add typed character to filter
				if len(msg.String()) == 1 {
					m.filterInputText += msg.String()
					// Apply filter in real-time as user types
					m.snapPanel.SetFilter(m.filterInputText)
				}
				return m, nil
			}
		}

		// Handle file browser interactions
		if m.showFileBrowser && m.fileBrowser != nil {
			switch msg.String() {
			case "esc":
				// Close file browser
				m.showFileBrowser = false
				m.opsPanel.Info("Closed file browser")
				return m, nil

			case "j", "down":
				// Move down in file list
				m.fileBrowser.MoveDown()
				return m, nil

			case "k", "up":
				// Move up in file list
				m.fileBrowser.MoveUp()
				return m, nil

			case "h", "left":
				// Go to parent directory
				if m.fileBrowser.CanGoUp() {
					m.fileBrowser.GoUp()
					return m, m.loadFiles
				}
				return m, nil

			case "n", "pgdown":
				// Next page
				m.fileBrowser.NextPage()
				return m, nil

			case "p", "pgup":
				// Previous page
				m.fileBrowser.PrevPage()
				return m, nil

			case "l", "right", "enter":
				// Enter directory or do nothing for files
				if newPath, entered := m.fileBrowser.EnterDirectory(); entered {
					m.opsPanel.Info(fmt.Sprintf("Navigating to %s...", newPath))
					return m, m.loadFiles
				}
				return m, nil

			case " ", "space":
				// Toggle file selection
				m.fileBrowser.ToggleSelection()
				return m, nil

			case "r":
				// Restore selected files
				selectedFiles := m.fileBrowser.GetSelectedFiles()
				if len(selectedFiles) == 0 {
					m.opsPanel.Warning("No files selected - press Space to select files")
					return m, nil
				}

				// Create paths list from selected files
				var paths []string
				for _, file := range selectedFiles {
					paths = append(paths, file.Path)
				}

				// Open restore form with selected paths pre-filled
				snapshot := m.fileBrowser.GetSnapshot()
				m.restoreForm = ui.NewRestoreForm(snapshot)
				m.restoreForm.SetSize(m.width*2/3, m.height*2/3)
				// Pre-fill with selected file paths
				m.restoreForm.SetIncludePaths(paths)
				m.showRestoreForm = true
				m.showFileBrowser = false
				m.opsPanel.Info(fmt.Sprintf("Restoring %d selected files...", len(paths)))
				return m, nil
			}
		}

		// Handle found repos selection
		if m.showFoundRepos {
			switch msg.String() {
			case "esc":
				// Close found repos list
				m.showFoundRepos = false
				m.foundRepos = nil
				m.selectedFound = 0
				m.opsPanel.Info("Cancelled repository selection")
				return m, nil

			case "j", "down":
				// Move down in found repos list
				if m.selectedFound < len(m.foundRepos)-1 {
					m.selectedFound++
				}
				return m, nil

			case "k", "up":
				// Move up in found repos list
				if m.selectedFound > 0 {
					m.selectedFound--
				}
				return m, nil

			case "enter":
				// Add selected repo
				if m.selectedFound >= 0 && m.selectedFound < len(m.foundRepos) {
					selectedRepo := m.foundRepos[m.selectedFound]
					m.showFoundRepos = false
					m.foundRepos = nil
					m.selectedFound = 0

					// Open repo form pre-filled with the selected repo
					m.repoForm = ui.NewRepoForm()
					m.repoForm.SetPath(selectedRepo.Path)
					m.repoForm.SetName(selectedRepo.Name)
					m.showRepoForm = true
					m.opsPanel.Info(fmt.Sprintf("Adding repository: %s", selectedRepo.Name))
				}
				return m, nil
			}
		}

		// Handle repo form interactions
		// Handle remove confirmation dialog
		if m.showRemoveConfirm && m.removeConfirmDialog != nil {
			switch msg.String() {
			case "esc":
				// Cancel removal
				m.showRemoveConfirm = false
				m.removeConfirmDialog = nil
				m.repoToRemove = ""
				m.opsPanel.Info("Cancelled repository removal")
				return m, nil

			case "enter":
				// Check if user typed the confirmation word
				if m.removeConfirmDialog.IsConfirmed() {
					m.opsPanel.Info(fmt.Sprintf("✓ Confirmed - removing '%s' from configuration...", m.repoToRemove))
					m.opsPanel.Dimmed("Updating configuration file...")
					return m, m.removeRepository()
				}
				return m, nil
			}

			// Pass other keys to the dialog
			var cmd tea.Cmd
			cmd = m.removeConfirmDialog.Update(msg)
			return m, cmd
		}

		if m.showRepoForm && m.repoForm != nil {
			switch msg.String() {
			case "esc":
				// Cancel repo creation
				m.showRepoForm = false
				m.repoForm = ui.NewRepoForm() // Reset form
				m.opsPanel.Info("Cancelled repository creation")
				return m, nil

			case "enter":
				// Submit form
				if m.repoForm.GetFocusedField() == ui.FieldSubmit {
					// Get form data
					name := m.repoForm.GetName()
					path := m.repoForm.GetPath()
					passwordMethod := m.repoForm.GetPasswordMethod()
					password := m.repoForm.GetPassword()

					if name == "" || path == "" {
						m.opsPanel.Error("Name and path are required")
						return m, nil
					}

					// Create repository config
					repoConfig := types.RepositoryConfig{
						Name: name,
						Path: path,
					}

					switch passwordMethod {
					case "file":
						var passwordFilePath string

						// Auto-generate password file if requested
						if m.repoForm.ShouldAutoGeneratePasswordFile() {
							// Generate password file path
							home, err := os.UserHomeDir()
							if err != nil {
								m.opsPanel.Error(fmt.Sprintf("Failed to get home directory: %v", err))
								return m, nil
							}

							passwordDir := filepath.Join(home, ".config", "lazyrestic", "passwords")
							passwordFilePath = filepath.Join(passwordDir, name+".txt")

							// Create password directory if it doesn't exist
							if err := os.MkdirAll(passwordDir, 0700); err != nil {
								m.opsPanel.Error(fmt.Sprintf("Failed to create password directory: %v", err))
								return m, nil
							}

							// Generate secure random password
							generatedPassword, err := generateSecurePassword(32)
							if err != nil {
								m.opsPanel.Error(fmt.Sprintf("Failed to generate password: %v", err))
								return m, nil
							}

							// Write password file with secure permissions (0400)
							if err := os.WriteFile(passwordFilePath, []byte(generatedPassword), 0400); err != nil {
								m.opsPanel.Error(fmt.Sprintf("Failed to write password file: %v", err))
								return m, nil
							}

							m.opsPanel.Success(fmt.Sprintf("Created password file: %s", passwordFilePath))
						} else {
							// Use manually specified password file path
							if password == "" {
								m.opsPanel.Error("Password file path is required")
								return m, nil
							}
							passwordFilePath = password
						}

						repoConfig.PasswordFile = passwordFilePath

					case "command":
						if password == "" {
							m.opsPanel.Error("Password command is required")
							return m, nil
						}
						repoConfig.PasswordCommand = password
					}

					// Add to config
					m.config.Repositories = append(m.config.Repositories, repoConfig)

					// Save config
					if err := config.Save(m.config, ""); err != nil {
						m.opsPanel.Error(fmt.Sprintf("Failed to save config: %v", err))
						return m, nil
					}

					// Initialize repository if requested
					if m.repoForm.ShouldInitialize() {
						client := restic.NewClient(repoConfig)
						if err := client.Init(); err != nil {
							m.opsPanel.Error(fmt.Sprintf("Failed to initialize repository: %v", err))
							// Still close form since config was saved
						} else {
							m.opsPanel.Success(fmt.Sprintf("Repository '%s' created and initialized", name))
						}
					} else {
						m.opsPanel.Success(fmt.Sprintf("Added repository '%s'", name))
					}

					// Close form and refresh
					m.showRepoForm = false
					m.repoForm = ui.NewRepoForm() // Reset for next use
					m.opsPanel.Info("Refreshing repository list...")
					return m, m.loadRepositories
				}
				fallthrough

			default:
				// Let the form handle the key
				cmd := m.repoForm.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "tab", "l", "right":
			// Cycle panels forward (4 panels: Repos, Metrics, Snapshots, Operations)
			m.activePanel = (m.activePanel + 1) % 4
			return m, nil

		case "shift+tab", "h", "left":
			// Cycle panels backward (4 panels: Repos, Metrics, Snapshots, Operations)
			m.activePanel = (m.activePanel + 3) % 4
			return m, nil

		case "j", "down":
			// Move down in active panel
			switch m.activePanel {
			case types.PanelRepositories:
				m.repoPanel.MoveDown()
				m.currentRepoIndex = m.GetSelected()
				// Update metrics panel with newly selected repo
				if m.currentRepoIndex < len(m.repositories) {
					m.metricsPanel.SetRepository(&m.repositories[m.currentRepoIndex])
				}
				// Load snapshots for selected repo
				return m, m.loadSnapshotsWithMessage()
			case types.PanelSnapshots:
				m.snapPanel.MoveDown()
				m.logSelectedSnapshot()
			}
			return m, nil

		case "k", "up":
			// Move up in active panel
			switch m.activePanel {
			case types.PanelRepositories:
				m.repoPanel.MoveUp()
				m.currentRepoIndex = m.GetSelected()
				// Update metrics panel with newly selected repo
				if m.currentRepoIndex < len(m.repositories) {
					m.metricsPanel.SetRepository(&m.repositories[m.currentRepoIndex])
				}
				return m, m.loadSnapshotsWithMessage()
			case types.PanelSnapshots:
				m.snapPanel.MoveUp()
				m.logSelectedSnapshot()
			}
			return m, nil

		case "enter":
			// Action on selected item
			if m.activePanel == types.PanelRepositories {
				return m, m.loadSnapshotsWithMessage()
			}
			// Open file browser for selected snapshot
			if m.activePanel == types.PanelSnapshots {
				selectedSnapshot := m.snapPanel.GetSelected()
				if selectedSnapshot != nil {
					m.fileBrowser = ui.NewFileBrowser(selectedSnapshot)
					m.fileBrowser.SetSize(m.width*2/3, m.height*2/3)
					m.showFileBrowser = true
					m.opsPanel.Info(fmt.Sprintf("Browsing snapshot %s...", selectedSnapshot.ShortID))
					return m, m.loadFiles
				}
			}
			return m, nil

		case "a":
			// Add new repository (only in repositories panel)
			if m.activePanel == types.PanelRepositories {
				m.showRepoForm = true
				m.opsPanel.Info("Add new repository")
				return m, nil
			}
			return m, nil

		case "s":
			// Scan for repositories (only in repositories panel)
			if m.activePanel == types.PanelRepositories {
				m.opsPanel.Info("Scanning for repositories...")
				return m, m.scanForRepositories()
			}
			return m, nil

		case "r":
			// Refresh
			m.opsPanel.Info("Refreshing repositories and snapshots...")
			m.opsPanel.Dimmed("Reloading configuration and rescanning repository stats")
			return m, tea.Batch(m.loadRepositories, m.loadSnapshotsWithMessage())

		case "C":
			// Cache cleanup
			if m.currentRepoIndex >= len(m.repositories) {
				m.opsPanel.Warning("No repository selected for cache cleanup")
				return m, nil
			}
			repo := m.repositories[m.currentRepoIndex]
			m.opsPanel.Info(fmt.Sprintf("Running cache cleanup for '%s'...", repo.Name))
			m.opsPanel.Dimmed(fmt.Sprintf("Command: restic -r %s cache --cleanup", repo.Path))
			return m, m.cleanupCache()

		case "u":
			// Unlock repository
			if m.currentRepoIndex >= len(m.repositories) {
				m.opsPanel.Warning("No repository selected for unlock")
				return m, nil
			}
			repo := m.repositories[m.currentRepoIndex]
			m.opsPanel.Info(fmt.Sprintf("Unlocking repository '%s'...", repo.Name))
			m.opsPanel.Dimmed(fmt.Sprintf("Removing stale locks from: %s", repo.Path))
			m.opsPanel.Dimmed(fmt.Sprintf("Command: restic -r %s unlock", repo.Path))
			return m, m.unlockRepository()

		case "x":
			// Remove repository from LazyRestic config
			if m.currentRepoIndex >= len(m.repositories) {
				m.opsPanel.Warning("No repository selected to remove")
				return m, nil
			}
			repo := m.repositories[m.currentRepoIndex]
			m.repoToRemove = repo.Name
			m.removeConfirmDialog = ui.NewConfirmationDialog(
				"REMOVE REPOSITORY",
				fmt.Sprintf("Remove '%s' from LazyRestic?\n\nPath: %s\n\nThis will only remove it from the LazyRestic configuration.\nThe repository files will NOT be deleted from disk.", repo.Name, repo.Path),
				"yes",
			)
			m.removeConfirmDialog.SetSize(m.width*3/4, m.height*3/4)
			m.showRemoveConfirm = true
			m.opsPanel.Success("─────────────────────────────────────────────────────────")
			m.opsPanel.Info(fmt.Sprintf("Removal requested for repository: %s", repo.Name))
			m.opsPanel.Dimmed(fmt.Sprintf("Path: %s", repo.Path))
			m.opsPanel.Warning("⚠️  Type 'yes' to confirm removal from configuration")
			m.opsPanel.Success("─────────────────────────────────────────────────────────")
			return m, nil

		case "b":
			// Show backup form (only if a repository is selected and not already backing up)
			if !m.backupInProgress && len(m.repositories) > 0 {
				m.showBackupForm = true
				return m, nil
			} else if m.backupInProgress {
				m.opsPanel.Warning("Backup already in progress")
			} else {
				m.opsPanel.Warning("No repository selected")
			}
			return m, nil

		case "R":
			// Show restore form (only if a snapshot is selected and not already restoring)
			selectedSnapshot := m.snapPanel.GetSelected()
			if !m.restoreInProgress && selectedSnapshot != nil {
				m.restoreForm = ui.NewRestoreForm(selectedSnapshot)
				m.restoreForm.SetSize(m.width*2/3, m.height*2/3)
				m.showRestoreForm = true
				return m, nil
			} else if m.restoreInProgress {
				m.opsPanel.Warning("Restore already in progress")
			} else {
				m.opsPanel.Warning("No snapshot selected")
			}
			return m, nil

		case "/":
			// Enter filter mode (only when snapshot panel is active)
			if m.activePanel == types.PanelSnapshots {
				m.filterInputActive = true
				m.filterInputText = ""
				m.opsPanel.Info("Filter mode: type to search, Enter to confirm, Esc to cancel")
				return m, nil
			}
			return m, nil

		case "esc":
			// Clear filter if active and not in input mode
			if m.activePanel == types.PanelSnapshots && m.snapPanel.IsFilterActive() {
				m.snapPanel.ClearFilter()
				m.opsPanel.Info("Filter cleared")
				return m, nil
			}
			return m, nil

		case "c":
			// Alternative shortcut to clear filter
			if m.activePanel == types.PanelSnapshots && m.snapPanel.IsFilterActive() {
				m.snapPanel.ClearFilter()
				m.opsPanel.Info("Filter cleared")
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

// GetSelected returns the index of the currently selected repository
func (m Model) GetSelected() int {
	if repo := m.repoPanel.GetSelected(); repo != nil {
		// Find index in config
		for i, r := range m.repositories {
			if r.Name == repo.Name {
				return i
			}
		}
	}
	return 0
}

// View renders the UI
// renderLoadingPanel renders a loading placeholder panel
func (m Model) renderLoadingPanel(title string, width, height int) string {
	loadingText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorInfo)).
		Bold(true).
		Render("Loading...")

	content := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(loadingText)

	return ui.RenderPanelWithTitle(title, content, width, height, false)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing LazyRestic..."
	}

	if m.tooSmall {
		return "Terminal window too small. Please resize to at least 80x20 characters."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.showBackupForm {
		return m.renderBackupForm()
	}

	if m.showRestoreForm {
		return m.renderRestoreForm()
	}

	if m.showRepoForm {
		return m.renderRepoForm()
	}

	if m.showFileBrowser {
		return m.renderFileBrowser()
	}

	if m.showFoundRepos {
		return m.renderFoundRepos()
	}

	if m.showRemoveConfirm {
		return m.renderRemoveConfirm()
	}

	// Update repository panel data
	m.repoPanel.SetRepositories(m.repositories)

	// Title bar with version - full width
	titleText := "📦 LazyRestic - TUI Backup Manager"
	versionText := "v0.1.0"

	// Calculate padding to push version to the right
	titleLen := len(titleText)
	versionLen := len(versionText)
	paddingNeeded := m.width - titleLen - versionLen - 6 // 6 for margins/padding
	if paddingNeeded < 1 {
		paddingNeeded = 1
	}

	titleLeft := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.TitleStyle.GetForeground()).
		Render(titleText)

	versionRight := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render(versionText)

	titleContent := titleLeft + strings.Repeat(" ", paddingNeeded) + versionRight

	title := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a1a")).
		Width(m.width - 4). // Leave small margin on sides
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00AA88")).
		BorderBottom(true).
		MarginTop(1).
		MarginBottom(1).
		Render(titleContent)

	// Render panels in new 4-panel layout
	// Left column: Repos / Metrics / Snapshots stacked vertically
	repoPanel := m.repoPanel.Render(m.activePanel == types.PanelRepositories)

	var metricsPanel string
	if m.loadingRepositories || (len(m.repositories) == 0 && m.currentRepoIndex == 0) {
		metricsPanel = m.renderLoadingPanel("[2] Metrics", m.metricsPanel.GetWidth(), m.metricsPanel.GetHeight())
	} else {
		m.metricsPanel.SetActive(m.activePanel == types.PanelMetrics)
		metricsPanel = m.metricsPanel.Render()
	}

	var snapshotsPanel string
	if m.loadingSnapshots {
		snapshotsPanel = m.renderLoadingPanel("[3] Snapshots", m.snapPanel.GetWidth(), m.snapPanel.GetHeight())
	} else {
		snapshotsPanel = m.snapPanel.Render(m.activePanel == types.PanelSnapshots)
	}

	// Stack repos, metrics, snapshots vertically in left column
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, repoPanel, metricsPanel, snapshotsPanel)

	// Right column: Operations panel (full height)
	rightColumn := m.opsPanel.Render(m.activePanel == types.PanelOperations)

	// Join left and right columns side by side
	allPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Help hint or filter input prompt
	var helpHint string
	if m.filterInputActive {
		filterPromptStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Bold(true)
		filterInputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // White
			Background(lipgloss.Color("236")). // Dark gray
			Padding(0, 1)

		helpHint = filterPromptStyle.Render("Filter: ") +
			filterInputStyle.Render(m.filterInputText+"_") +
			ui.HelpStyle.Render(" • Enter to apply • Esc to cancel")
	} else {
		helpHint = ui.HelpStyle.Render("?:help  q:quit  a:add  x:rm  s:scan  b:backup  R:restore  u:unlock  C:cache  /:filter  r:refresh")
	}

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		allPanels,
		helpHint,
	)

	// Ensure content doesn't exceed terminal height
	if m.height > 0 {
		content = lipgloss.NewStyle().
			MaxHeight(m.height).
			Render(content)
	}

	return content
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	// Make width responsive to terminal size
	helpWidth := m.width - 10
	if helpWidth > 100 {
		helpWidth = 100
	}
	if helpWidth < 60 {
		helpWidth = 60
	}

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(helpWidth)

	help := `LazyRestic v0.1.0 - Keyboard Shortcuts

Navigation:
  ↑/k        Move up
  ↓/j        Move down
  Tab/→/l    Next panel
  Shift+Tab/←/h  Previous panel

Actions:
   Enter      Select / View details
   a          Add new repository (repositories panel)
   b          Start a backup
   R          Restore selected snapshot (Shift+r)
   r          Refresh data
   ?          Toggle this help
   q/Ctrl+C   Quit

Filtering (in Snapshots panel):
  /          Enter filter mode
  Esc/c      Clear active filter

   When in filter mode:
     Type to search by ID, path, tag, or hostname
     Enter to apply, Esc to cancel

Panels:
  Left:   Repositories list
  Right:  Snapshots for selected repository
  Bottom: Operations and logs

Press ? or Esc to close this help.
`

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		helpStyle.Render(help),
	)
}

// renderBackupForm renders the backup configuration form
func (m Model) renderBackupForm() string {
	form := m.backupForm.Render()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)
}

// renderRestoreForm renders the restore configuration form
func (m Model) renderRestoreForm() string {
	form := m.restoreForm.Render()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)
}

// renderRepoForm renders the repository creation form
func (m Model) renderRepoForm() string {
	form := m.repoForm.Render()

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)
}

// renderFoundRepos renders the found repositories selection list
func (m Model) renderRemoveConfirm() string {
	// Render confirmation dialog centered
	dialog := m.removeConfirmDialog.Render()

	// Center the dialog on screen
	dialogWidth := lipgloss.Width(dialog)
	dialogHeight := lipgloss.Height(dialog)

	horizontalPadding := (m.width - dialogWidth) / 2
	if horizontalPadding < 0 {
		horizontalPadding = 0
	}

	verticalPadding := (m.height - dialogHeight) / 2
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	// Add padding to center
	centeredStyle := lipgloss.NewStyle().
		PaddingLeft(horizontalPadding).
		PaddingTop(verticalPadding)

	return centeredStyle.Render(dialog)
}

func (m Model) renderFoundRepos() string {
	var b strings.Builder

	// Title
	titleStyle := ui.TitleStyle
	title := titleStyle.Render("Select Repository to Add")
	b.WriteString(title + "\n\n")

	// Instructions
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(infoStyle.Render("Found repositories - press Enter to add, Esc to cancel\n\n"))

	// List found repos
	for i, repo := range m.foundRepos {
		var line string

		// Selection indicator
		if i == m.selectedFound {
			line = "▶ "
		} else {
			line = "  "
		}

		// Repo info
		line += fmt.Sprintf("%s (%s)", repo.Name, repo.Path)

		// Style
		if i == m.selectedFound {
			line = ui.ListItemSelectedStyle.Render(line)
		} else {
			line = ui.ListItemStyle.Render(line)
		}

		b.WriteString(line + "\n")
	}

	// Wrap in styled box
	content := b.String()
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.width - 10)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		borderStyle.Render(content),
	)
}

// renderFileBrowser renders the file browser view
func (m Model) renderFileBrowser() string {
	if m.fileBrowser == nil {
		return ""
	}

	browser := m.fileBrowser.Render(true)

	// Add help hint at bottom
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	help := helpStyle.Render("↑/↓ navigate • ←/h back • →/l enter dir • Space select • r restore • Esc close")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		browser,
		"\n"+help,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 for a readable password
	// Base64 encoding makes it URL-safe and easier to handle
	password := base64.URLEncoding.EncodeToString(bytes)

	// Trim to requested length
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}
