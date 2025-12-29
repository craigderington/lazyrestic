package types

import "time"

// Repository represents a restic backup repository
type Repository struct {
	Name          string    // User-friendly name
	Path          string    // Repository path (local or remote)
	LastBackup    time.Time // Timestamp of last backup
	Size          int64     // Total repository size in bytes
	TotalFiles    int64     // Total number of files
	SnapshotCount int       // Number of snapshots
	Status        string    // "healthy", "warning", "error", "unknown"
}

// Snapshot represents a restic snapshot
type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	Paths    []string  `json:"paths"`
	Tags     []string  `json:"tags"`
	ShortID  string    `json:"short_id"`
}

// SnapshotStats represents statistics about a snapshot
type SnapshotStats struct {
	TotalSize      int64 `json:"total_size"`
	TotalFileCount int64 `json:"total_file_count"`
}

// RepositoryStats represents statistics about the entire repository
type RepositoryStats struct {
	TotalSize      int64 `json:"total_size"`
	TotalFileCount int64 `json:"total_file_count"`
	SnapshotsCount int   `json:"snapshots_count"`
}

// BackupProgress represents the progress of a backup operation
type BackupProgress struct {
	MessageType      string   `json:"message_type"`
	PercentDone      float64  `json:"percent_done"`
	TotalFiles       int64    `json:"total_files"`
	FilesDone        int64    `json:"files_done"`
	TotalBytes       int64    `json:"total_bytes"`
	BytesDone        int64    `json:"bytes_done"`
	CurrentFiles     []string `json:"current_files"`
	SecondsElapsed   int      `json:"seconds_elapsed"`
	SecondsRemaining int      `json:"seconds_remaining"`
}

// BackupSummary represents the final summary of a backup
type BackupSummary struct {
	MessageType         string `json:"message_type"`
	FilesNew            int64  `json:"files_new"`
	FilesChanged        int64  `json:"files_changed"`
	FilesUnmodified     int64  `json:"files_unmodified"`
	DirsNew             int64  `json:"dirs_new"`
	DirsChanged         int64  `json:"dirs_changed"`
	DirsUnmodified      int64  `json:"dirs_unmodified"`
	DataBlobs           int64  `json:"data_blobs"`
	TreeBlobs           int64  `json:"tree_blobs"`
	DataAdded           int64  `json:"data_added"`
	TotalFilesProcessed int64  `json:"total_files_processed"`
	TotalBytesProcessed int64  `json:"total_bytes_processed"`
	SnapshotID          string `json:"snapshot_id"`
}

// BackupOptions represents options for a backup operation
type BackupOptions struct {
	Paths   []string
	Tags    []string
	Exclude []string
}

// RestoreOptions represents options for a restore operation
type RestoreOptions struct {
	SnapshotID string
	Target     string   // Target directory (empty for original location)
	Include    []string // Specific paths to restore (empty for all)
}

// RestoreProgress represents the progress of a restore operation
type RestoreProgress struct {
	MessageType      string  `json:"message_type"`
	PercentDone      float64 `json:"percent_done"`
	TotalFiles       int64   `json:"total_files"`
	FilesRestored    int64   `json:"files_restored"`
	TotalBytes       int64   `json:"total_bytes"`
	BytesRestored    int64   `json:"bytes_restored"`
	SecondsElapsed   int     `json:"seconds_elapsed"`
	SecondsRemaining int     `json:"seconds_remaining"`
}

// RestoreSummary represents the final summary of a restore
type RestoreSummary struct {
	MessageType    string `json:"message_type"`
	TotalFiles     int64  `json:"total_files"`
	TotalBytes     int64  `json:"total_bytes"`
	SecondsElapsed int    `json:"seconds_elapsed"`
}

// ResticConfig represents the application configuration
type ResticConfig struct {
	Repositories []RepositoryConfig `yaml:"repositories"`
}

// RepositoryConfig represents a configured repository
type RepositoryConfig struct {
	Name            string `yaml:"name"`
	Path            string `yaml:"path"`
	PasswordCommand string `yaml:"password_command,omitempty"`
	PasswordFile    string `yaml:"password_file,omitempty"`
	Password        string `yaml:"password,omitempty"` // Not recommended, but supported
}

// Panel represents which panel is currently focused
type Panel int

const (
	PanelRepositories Panel = iota
	PanelMetrics
	PanelSnapshots
	PanelOperations
)

func (p Panel) String() string {
	switch p {
	case PanelRepositories:
		return "Repositories"
	case PanelMetrics:
		return "Metrics"
	case PanelSnapshots:
		return "Snapshots"
	case PanelOperations:
		return "Operations"
	default:
		return "Unknown"
	}
}

// FileNode represents a file or directory node from restic ls
type FileNode struct {
	MessageType string `json:"message_type,omitempty"` // For filtering node messages

	Name        string    `json:"name"`
	Type        string    `json:"type"` // "file" or "dir"
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        int       `json:"mode"`
	Permissions string    `json:"permissions"`
	ModTime     time.Time `json:"mtime"`
	AccessTime  time.Time `json:"atime"`
	ChangeTime  time.Time `json:"ctime"`
	UID         int       `json:"uid"`
	GID         int       `json:"gid"`
	Inode       uint64    `json:"inode"`

	// Internal fields for UI
	Selected bool `json:"-"` // For multi-select in file browser
}

// IsDir returns true if this node is a directory
func (n FileNode) IsDir() bool {
	return n.Type == "dir"
}

// IsFile returns true if this node is a file
func (n FileNode) IsFile() bool {
	return n.Type == "file"
}

// ForgetPolicy represents retention policy for snapshots
type ForgetPolicy struct {
	KeepLast    int      // Keep the last n snapshots
	KeepHourly  int      // Keep the last n hourly snapshots
	KeepDaily   int      // Keep the last n daily snapshots
	KeepWeekly  int      // Keep the last n weekly snapshots
	KeepMonthly int      // Keep the last n monthly snapshots
	KeepYearly  int      // Keep the last n yearly snapshots
	KeepWithin  string   // Keep snapshots within duration (e.g., "1y5m7d2h")
	KeepTags    []string // Keep all snapshots with these tags
	Host        string   // Only apply to snapshots from this host
	Paths       []string // Only apply to snapshots with these paths
	Tags        []string // Only apply to snapshots with these tags
}

// ForgetResult represents the result of a forget operation
type ForgetResult struct {
	SnapshotsToKeep   []Snapshot `json:"keep"`
	SnapshotsToRemove []Snapshot `json:"remove"`
	Host              string     `json:"host"`
	Paths             []string   `json:"paths"`
	Tags              []string   `json:"tags"`
}

// PruneStats represents statistics from a prune operation
type PruneStats struct {
	TotalBlobs    int64
	UnusedBlobs   int64
	TotalSize     int64
	UnusedSize    int64
	RemovedSize   int64
	RepackedBlobs int64
}
