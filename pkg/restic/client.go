package restic

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/craigderington/lazyrestic/pkg/types"
)

// Client handles restic command execution
type Client struct {
	config types.RepositoryConfig
}

// NewClient creates a new restic client for a repository
func NewClient(config types.RepositoryConfig) *Client {
	return &Client{
		config: config,
	}
}

// buildEnv creates environment variables for restic commands
func (c *Client) buildEnv() []string {
	env := []string{
		fmt.Sprintf("RESTIC_REPOSITORY=%s", c.config.Path),
	}

	// Only password_file and password_command are supported (no plain-text passwords)
	if c.config.PasswordFile != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD_FILE=%s", c.config.PasswordFile))
	}
	if c.config.PasswordCommand != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD_COMMAND=%s", c.config.PasswordCommand))
	}

	return env
}

// execCommand executes a restic command and returns the output
func (c *Client) execCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("restic", args...)

	// Start with parent environment and add our custom vars
	cmd.Env = append(os.Environ(), c.buildEnv()...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return both the error and output for better debugging
		return output, fmt.Errorf("restic command failed: %w (output: %s)", err, string(output))
	}

	return output, nil
}

// ListSnapshots retrieves all snapshots from the repository
func (c *Client) ListSnapshots() ([]types.Snapshot, error) {
	output, err := c.execCommand("snapshots", "--json")
	if err != nil {
		return nil, err
	}

	var snapshots []types.Snapshot
	if err := json.Unmarshal(output, &snapshots); err != nil {
		return nil, fmt.Errorf("failed to parse snapshots JSON: %w (output: %s)", err, string(output))
	}

	return snapshots, nil
}

// ListFiles lists all files in a snapshot
// If path is empty, lists all files in the snapshot
// If path is specified, lists files in that directory
func (c *Client) ListFiles(snapshotID string, path string) ([]types.FileNode, error) {
	args := []string{"ls", snapshotID, "--json"}
	if path != "" {
		args = append(args, path)
	}

	cmd := exec.Command("restic", args...)
	cmd.Env = append(os.Environ(), c.buildEnv()...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ls command: %w", err)
	}

	var nodes []types.FileNode
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse the JSON line directly to FileNode
		var node types.FileNode
		if err := json.Unmarshal(line, &node); err != nil {
			continue // Skip malformed lines
		}

		// Check message type - we only want "node" entries
		if node.MessageType != "node" {
			continue // Skip non-node entries (like snapshot metadata)
		}

		nodes = append(nodes, node)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ls output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ls command failed: %w", err)
	}

	return nodes, nil
}

// CheckRepository verifies repository integrity
func (c *Client) CheckRepository() error {
	_, err := c.execCommand("check")
	return err
}

// CleanupCache removes old cache entries
func (c *Client) CleanupCache() (string, error) {
	output, err := c.execCommand("cache", "--cleanup")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Unlock removes stale locks from the repository
func (c *Client) Unlock() (string, error) {
	output, err := c.execCommand("unlock")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetStats retrieves repository statistics
func (c *Client) GetStats() (*types.RepositoryStats, error) {
	output, err := c.execCommand("stats", "--json")
	if err != nil {
		return nil, err
	}

	var stats types.RepositoryStats
	if err := json.Unmarshal(output, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats JSON: %w", err)
	}

	return &stats, nil
}

// GetRepositoryInfo retrieves comprehensive repository information
func (c *Client) GetRepositoryInfo() (*types.Repository, error) {
	repo := &types.Repository{
		Status: "unknown",
	}

	// Get repository stats
	stats, err := c.GetStats()
	if err != nil {
		repo.Status = "error"
		return repo, err
	}

	repo.Size = stats.TotalSize
	repo.TotalFiles = stats.TotalFileCount
	repo.SnapshotCount = stats.SnapshotsCount

	// Get snapshots to find the last backup time
	snapshots, err := c.ListSnapshots()
	if err != nil {
		repo.Status = "warning" // Stats work but can't get snapshots
		return repo, nil
	}

	// Find the most recent snapshot
	if len(snapshots) > 0 {
		mostRecent := snapshots[0]
		for _, snap := range snapshots {
			if snap.Time.After(mostRecent.Time) {
				mostRecent = snap
			}
		}
		repo.LastBackup = mostRecent.Time
	}

	// Check repository health
	if err := c.CheckRepository(); err != nil {
		repo.Status = "warning"
	} else {
		repo.Status = "healthy"
	}

	return repo, nil
}

// IsResticInstalled checks if restic binary is available
func IsResticInstalled() bool {
	_, err := exec.LookPath("restic")
	return err == nil
}

// GetResticVersion returns the installed restic version
func GetResticVersion() (string, error) {
	cmd := exec.Command("restic", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// Init initializes a new restic repository
func (c *Client) Init() error {
	_, err := c.execCommand("init")
	return err
}

// BackupProgressCallback is called for each progress update during backup
type BackupProgressCallback func(progress *types.BackupProgress, summary *types.BackupSummary) error

// BackupMessage represents a message from the backup operation
type BackupMessage struct {
	Progress *types.BackupProgress
	Summary  *types.BackupSummary
	Error    error
}

// RestoreMessage represents a message from the restore operation
type RestoreMessage struct {
	Progress *types.RestoreProgress
	Summary  *types.RestoreSummary
	Error    error
}

// BackupWithChannel performs a backup and sends updates through a channel
func (c *Client) BackupWithChannel(ctx context.Context, opts types.BackupOptions, updates chan<- BackupMessage) {
	defer close(updates)

	// Build command arguments
	args := []string{"backup", "--json"}

	// Add tags
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	// Add excludes
	for _, exclude := range opts.Exclude {
		args = append(args, "--exclude", exclude)
	}

	// Add paths
	args = append(args, opts.Paths...)

	// Create command
	cmd := exec.CommandContext(ctx, "restic", args...)
	cmd.Env = append(os.Environ(), c.buildEnv()...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		updates <- BackupMessage{Error: fmt.Errorf("failed to create stdout pipe: %w", err)}
		return
	}

	// Get stderr pipe for errors
	stderr, err := cmd.StderrPipe()
	if err != nil {
		updates <- BackupMessage{Error: fmt.Errorf("failed to create stderr pipe: %w", err)}
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		updates <- BackupMessage{Error: fmt.Errorf("failed to start backup: %w", err)}
		return
	}

	// Read and process JSON output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Try to parse as generic JSON to determine message type
		var msgType struct {
			MessageType string `json:"message_type"`
		}

		if err := json.Unmarshal(line, &msgType); err != nil {
			// Skip non-JSON lines
			continue
		}

		switch msgType.MessageType {
		case "status":
			// Progress update
			var progress types.BackupProgress
			if err := json.Unmarshal(line, &progress); err != nil {
				continue
			}
			updates <- BackupMessage{Progress: &progress}

		case "summary":
			// Final summary
			var summary types.BackupSummary
			if err := json.Unmarshal(line, &summary); err != nil {
				continue
			}
			updates <- BackupMessage{Summary: &summary}
		}
	}

	if err := scanner.Err(); err != nil {
		updates <- BackupMessage{Error: fmt.Errorf("error reading backup output: %w", err)}
		return
	}

	// Check for errors on stderr
	stderrData, _ := io.ReadAll(stderr)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		updates <- BackupMessage{Error: fmt.Errorf("backup failed: %w (stderr: %s)", err, string(stderrData))}
		return
	}
}

// Backup performs a backup operation with progress tracking
func (c *Client) Backup(opts types.BackupOptions, progressCallback BackupProgressCallback) error {
	// Build command arguments
	args := []string{"backup", "--json"}

	// Add tags
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	// Add excludes
	for _, exclude := range opts.Exclude {
		args = append(args, "--exclude", exclude)
	}

	// Add paths
	args = append(args, opts.Paths...)

	// Create command
	cmd := exec.Command("restic", args...)
	cmd.Env = append(os.Environ(), c.buildEnv()...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Get stderr pipe for errors
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start backup: %w", err)
	}

	// Read and process JSON output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Try to parse as generic JSON to determine message type
		var msgType struct {
			MessageType string `json:"message_type"`
		}

		if err := json.Unmarshal(line, &msgType); err != nil {
			// Skip non-JSON lines
			continue
		}

		switch msgType.MessageType {
		case "status":
			// Progress update
			var progress types.BackupProgress
			if err := json.Unmarshal(line, &progress); err != nil {
				continue
			}
			if progressCallback != nil {
				if err := progressCallback(&progress, nil); err != nil {
					return err
				}
			}

		case "summary":
			// Final summary
			var summary types.BackupSummary
			if err := json.Unmarshal(line, &summary); err != nil {
				continue
			}
			if progressCallback != nil {
				if err := progressCallback(nil, &summary); err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading backup output: %w", err)
	}

	// Check for errors on stderr
	stderrData, _ := io.ReadAll(stderr)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("backup failed: %w (stderr: %s)", err, string(stderrData))
	}

	return nil
}

// RestoreWithChannel performs a restore and sends updates through a channel
func (c *Client) RestoreWithChannel(ctx context.Context, opts types.RestoreOptions, updates chan<- RestoreMessage) {
	defer close(updates)

	// Build command arguments
	args := []string{"restore", opts.SnapshotID}

	// Add target directory
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	// Add include paths if specified
	for _, include := range opts.Include {
		args = append(args, "--include", include)
	}

	// Create command
	cmd := exec.CommandContext(ctx, "restic", args...)
	cmd.Env = append(os.Environ(), c.buildEnv()...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		updates <- RestoreMessage{Error: fmt.Errorf("failed to create stdout pipe: %w", err)}
		return
	}

	// Get stderr pipe for errors
	stderr, err := cmd.StderrPipe()
	if err != nil {
		updates <- RestoreMessage{Error: fmt.Errorf("failed to create stderr pipe: %w", err)}
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		updates <- RestoreMessage{Error: fmt.Errorf("failed to start restore: %w", err)}
		return
	}

	// Read and process output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// restic restore doesn't output JSON by default, just text progress
		// We'll send simple progress updates based on the text output
		// Format is typically: "restoring <snapshot-id> to /path/to/target"
		if strings.Contains(line, "restoring") {
			updates <- RestoreMessage{
				Progress: &types.RestoreProgress{
					MessageType: "status",
				},
			}
		}
	}

	if err := scanner.Err(); err != nil {
		updates <- RestoreMessage{Error: fmt.Errorf("error reading restore output: %w", err)}
		return
	}

	// Check for errors on stderr
	stderrData, _ := io.ReadAll(stderr)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		updates <- RestoreMessage{Error: fmt.Errorf("restore failed: %w (stderr: %s)", err, string(stderrData))}
		return
	}

	// Send completion summary
	updates <- RestoreMessage{
		Summary: &types.RestoreSummary{
			MessageType: "summary",
		},
	}
}

// Restore performs a restore operation (synchronous version for compatibility)
func (c *Client) Restore(opts types.RestoreOptions) error {
	args := []string{"restore", opts.SnapshotID}

	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	for _, include := range opts.Include {
		args = append(args, "--include", include)
	}

	_, err := c.execCommand(args...)
	return err
}

// ForgetDryRun performs a dry-run of forget to preview what would be removed
func (c *Client) ForgetDryRun(policy types.ForgetPolicy) ([]types.ForgetResult, error) {
	args := []string{"forget", "--dry-run", "--json"}

	// Add policy flags
	if policy.KeepLast > 0 {
		args = append(args, "--keep-last", fmt.Sprintf("%d", policy.KeepLast))
	}
	if policy.KeepHourly > 0 {
		args = append(args, "--keep-hourly", fmt.Sprintf("%d", policy.KeepHourly))
	}
	if policy.KeepDaily > 0 {
		args = append(args, "--keep-daily", fmt.Sprintf("%d", policy.KeepDaily))
	}
	if policy.KeepWeekly > 0 {
		args = append(args, "--keep-weekly", fmt.Sprintf("%d", policy.KeepWeekly))
	}
	if policy.KeepMonthly > 0 {
		args = append(args, "--keep-monthly", fmt.Sprintf("%d", policy.KeepMonthly))
	}
	if policy.KeepYearly > 0 {
		args = append(args, "--keep-yearly", fmt.Sprintf("%d", policy.KeepYearly))
	}
	if policy.KeepWithin != "" {
		args = append(args, "--keep-within", policy.KeepWithin)
	}

	// Add filters
	if policy.Host != "" {
		args = append(args, "--host", policy.Host)
	}
	for _, tag := range policy.Tags {
		args = append(args, "--tag", tag)
	}
	for _, path := range policy.Paths {
		args = append(args, "--path", path)
	}

	output, err := c.execCommand(args...)
	if err != nil {
		return nil, err
	}

	var results []types.ForgetResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse forget results: %w", err)
	}

	return results, nil
}

// Forget removes snapshots according to policy
func (c *Client) Forget(policy types.ForgetPolicy) error {
	args := []string{"forget"}

	// Add policy flags (same as dry-run)
	if policy.KeepLast > 0 {
		args = append(args, "--keep-last", fmt.Sprintf("%d", policy.KeepLast))
	}
	if policy.KeepHourly > 0 {
		args = append(args, "--keep-hourly", fmt.Sprintf("%d", policy.KeepHourly))
	}
	if policy.KeepDaily > 0 {
		args = append(args, "--keep-daily", fmt.Sprintf("%d", policy.KeepDaily))
	}
	if policy.KeepWeekly > 0 {
		args = append(args, "--keep-weekly", fmt.Sprintf("%d", policy.KeepWeekly))
	}
	if policy.KeepMonthly > 0 {
		args = append(args, "--keep-monthly", fmt.Sprintf("%d", policy.KeepMonthly))
	}
	if policy.KeepYearly > 0 {
		args = append(args, "--keep-yearly", fmt.Sprintf("%d", policy.KeepYearly))
	}
	if policy.KeepWithin != "" {
		args = append(args, "--keep-within", policy.KeepWithin)
	}

	// Add filters
	if policy.Host != "" {
		args = append(args, "--host", policy.Host)
	}
	for _, tag := range policy.Tags {
		args = append(args, "--tag", tag)
	}
	for _, path := range policy.Paths {
		args = append(args, "--path", path)
	}

	_, err := c.execCommand(args...)
	return err
}

// PruneDryRun performs a dry-run of prune to preview what would be removed
func (c *Client) PruneDryRun() (string, error) {
	output, err := c.execCommand("prune", "--dry-run")
	return string(output), err
}

// Prune removes unreferenced data from the repository
func (c *Client) Prune() error {
	_, err := c.execCommand("prune")
	return err
}
