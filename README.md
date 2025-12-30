# LazyRestic

A beautiful Terminal User Interface (TUI) for managing [restic](https://restic.net/) backup repositories, inspired by lazydocker and lazygit.

## Features

- **Visual Repository Management**: Browse and manage multiple restic repositories
- **Snapshot Browser**: View and explore snapshots with an intuitive interface
- **Interactive Backup & Restore**: Create backups and restore snapshots with guided forms
- **Real-time Progress Tracking**: Watch backup/restore operations with live progress updates
- **Smart Filtering**: Quickly find snapshots by ID, path, tag, or hostname with instant search
- **Repository Statistics**: View snapshot counts, sizes, file counts, and last backup time
- **Real-time Operations Log**: Monitor backup operations and see what's happening
- **Keyboard-Driven**: Vim-style navigation for efficient workflow
- **Multi-Panel Layout**: See repositories, snapshots, and operations at a glance

## Installation

### Prerequisites

- Go 1.21 or later
- [restic](https://restic.readthedocs.io/en/latest/020_installation.html) installed and in PATH

### Install via Go

```bash
go install github.com/craigderington/lazyrestic@latest
```

This will install the `lazyrestic` binary to `$GOPATH/bin` (or `$HOME/go/bin`).

### Build from Source

```bash
git clone https://github.com/craigderington/lazyrestic
cd lazyrestic
go build -o lazyrestic
sudo mv lazyrestic /usr/local/bin/  # Optional: install system-wide
```

## Quick Start

### 1. Create a Test Repository

```bash
# Create a test restic repository
restic -r /tmp/restic-test init
# Set password when prompted (e.g., "testpassword")

# Create a test snapshot
restic -r /tmp/restic-test backup ~/.bashrc
```

### 2. Configure LazyRestic

```bash
# Create config directory
mkdir -p ~/.config/lazyrestic

# Copy example config
cp config.example.yaml ~/.config/lazyrestic/config.yaml

# Edit the config with your repositories
$EDITOR ~/.config/lazyrestic/config.yaml
```

Example minimal configuration:

```yaml
repositories:
  - name: test-repo
    path: /tmp/restic-test
    password_file: ~/.config/lazyrestic/passwords/test-repo.txt
```

**Note:** Plain-text passwords are no longer supported for security reasons. See the [Password Security](#password-security) section below.

### 3. Run LazyRestic

```bash
./lazyrestic
```

## Usage

### Keyboard Shortcuts

**Navigation:**
- `↑`/`k` - Move up
- `↓`/`j` - Move down
- `Tab` or `→`/`l` - Next panel
- `Shift+Tab` or `←`/`h` - Previous panel

**Actions:**
- `Enter` - Select item / View details
- `b` - Start a backup (opens backup configuration dialog)
- `R` - Restore selected snapshot (Shift+r)
- `r` - Refresh data
- `?` - Toggle help screen
- `q` or `Ctrl+C` - Quit

**Filtering (in Snapshots panel):**
- `/` - Enter filter mode (search by ID, path, tag, or hostname)
- `Esc` or `c` - Clear active filter
- While in filter mode:
  - Type to search in real-time
  - `Enter` to apply filter
  - `Esc` to cancel

### Panel Overview

```
┌────────────────────────────────────────────────────────┐
│  LazyRestic - TUI Backup Manager                      │
├──────────────────────┬─────────────────────────────────┤
│  📦 Repositories     │  📸 Snapshots                   │
│                      │                                 │
│  ▶ local-backup      │  ▶ a1b2c3d4 - 2 hours ago       │
│      [healthy]       │    Host: myserver               │
│      /path/to/repo   │    Paths: /home/user            │
│                      │                                 │
│      Snapshots: 15   │    d5e6f7g8 - 1 day ago         │
│      Size: 2.3 GiB   │                                 │
│      Files: 45,231   │                                 │
│      Last: 2 hrs ago │                                 │
│                      │                                 │
│  • s3-backup         │                                 │
│      [healthy]       │                                 │
├──────────────────────┴─────────────────────────────────┤
│  📋 Operations                                         │
│                                                        │
│  10:30:45 ✓ Found restic 0.16.2                       │
│  10:30:46 ✓ Loaded 2 repositories                     │
│  10:30:47 ✓ Loaded 15 snapshots                       │
└────────────────────────────────────────────────────────┘
```

### Creating Backups

To create a new backup:

1. Select a repository from the left panel using `↑`/`↓` or `j`/`k`
2. Press `b` to open the backup configuration dialog
3. Enter the paths you want to backup (comma-separated)
4. Optionally add tags and exclude patterns
5. Navigate to "Start Backup" using `Tab` or `↓`
6. Press `Enter` to start the backup

The backup will run in the background and progress will be displayed in the Operations panel at the bottom. Once complete, the snapshots panel will automatically refresh to show the new backup.

### Restoring Snapshots

To restore a snapshot:

1. Navigate to the Snapshots panel (right panel) using `Tab` or `→`/`l`
2. Select the snapshot you want to restore using `↑`/`↓` or `j`/`k`
3. Press `R` (Shift+r) to open the restore configuration dialog
4. Choose restore location:
   - Press `Space` to toggle "Restore to original location" (⚠️ this will overwrite files!)
   - Or enter a custom target directory path
5. Optionally specify specific files/paths to restore (leave empty to restore all)
6. Navigate to "Restore Snapshot" using `Tab` or `↓`
7. Press `Enter` to start the restore

The restore will run and completion status will be displayed in the Operations panel.

### Filtering Snapshots

When you have many snapshots, filtering makes it easy to find what you need:

**Quick Search:**
1. Navigate to the Snapshots panel (right panel)
2. Press `/` to enter filter mode
3. Start typing - the list filters in real-time as you type
4. Press `Enter` to keep the filter active, or `Esc` to cancel

**What You Can Search:**
- **Snapshot ID**: Find specific snapshots by their ID (e.g., "abc123")
- **Paths**: Search by backup paths (e.g., "/home" will show all snapshots containing /home paths)
- **Tags**: Find snapshots with specific tags (e.g., "daily", "production")
- **Hostname**: Filter by the host that created the backup (e.g., "webserver")

**Filter Examples:**
- Type `daily` - Shows all snapshots tagged with "daily"
- Type `home` - Shows snapshots backing up paths containing "home"
- Type `abc` - Shows snapshots with IDs containing "abc"
- Type `webserver` - Shows snapshots from hosts named "webserver"

**Filter Indicators:**
- Active filters are displayed in orange in the panel title (e.g., `📸 Snapshots [text=home]`)
- Filtered count is shown below the title (e.g., `[3 of 50 snapshots shown]`)
- To clear a filter, press `Esc` or `c`

Filters are case-insensitive and search across multiple fields, making it easy to find snapshots quickly even in repositories with hundreds of backups.

### Repository Statistics

When you select a repository in the left panel, LazyRestic automatically displays comprehensive statistics:

- **Snapshot Count**: Total number of backups in the repository
- **Repository Size**: Total deduplicated storage space used
- **Total Files**: Number of unique files across all snapshots
- **Last Backup**: Human-readable time since the most recent backup (e.g., "2 hours ago", "3 days ago")
- **Status**: Repository health indicator (healthy, warning, error)

These statistics refresh automatically when you press `r` or when you create a new backup.

## Configuration

Configuration file: `~/.config/lazyrestic/config.yaml`

### Password Security

LazyRestic enforces secure password management and **does not support plain-text passwords** in the configuration file. You must use one of these two secure methods:

#### Method 1: Password File (Recommended)

Store your password in a separate file with restrictive permissions:

```yaml
repositories:
  - name: my-backup
    path: /path/to/repo
    password_file: ~/.config/lazyrestic/passwords/my-backup.txt
```

**Creating a password file manually:**
```bash
# Create password directory
mkdir -p ~/.config/lazyrestic/passwords

# Create password file (replace YOUR_PASSWORD with your actual password)
echo 'YOUR_PASSWORD' > ~/.config/lazyrestic/passwords/my-backup.txt

# Set secure permissions (read-only for owner)
chmod 400 ~/.config/lazyrestic/passwords/my-backup.txt
```

**Or use LazyRestic's built-in auto-generation:**
When creating a new repository in the TUI, LazyRestic can automatically:
- Generate a cryptographically secure random password
- Create the password file with proper permissions (0400)
- Store it at `~/.config/lazyrestic/passwords/<repo-name>.txt`

#### Method 2: Password Command (For Password Managers)

Use a password manager like `pass`, `1password`, or `lastpass`:

```yaml
repositories:
  - name: my-backup
    path: /path/to/repo
    password_command: pass show restic/my-backup
```

**Example with different password managers:**
```yaml
# Using 'pass' (password-store)
password_command: pass show restic/my-backup

# Using 1Password CLI
password_command: op read "op://vault/restic-backup/password"

# Using macOS Keychain
password_command: security find-generic-password -a restic -s my-backup -w
```

### Repository Configuration Options

```yaml
repositories:
  - name: my-backup           # Display name
    path: /path/to/repo       # Repository path (local or remote)

    # Password options (choose ONE):
    password_file: ~/.config/lazyrestic/passwords/my-backup.txt  # Recommended
    password_command: pass show restic/my-backup                  # For password managers
```

**Important Security Notes:**
- Config file must have `0600` permissions
- Password files must have `0400` or `0600` permissions
- Never commit password files to version control
- Add `~/.config/lazyrestic/passwords/` to your `.gitignore`

### Supported Repository Types

- **Local**: `/path/to/repo`
- **SFTP**: `sftp:user@host:/path/to/repo`
- **S3**: `s3:s3.amazonaws.com/bucket/path`
- **B2**: `b2:bucketname:path`
- **Azure**: `azure:container:path`
- **GCS**: `gs:bucket:/path`
- **REST**: `rest:http://host:8000/`

See [restic documentation](https://restic.readthedocs.io/en/latest/030_preparing_a_new_repo.html) for more details.

## Development

### Project Structure

```
lazyrestic/
├── main.go              # Entry point
├── pkg/
│   ├── model/          # Bubbletea model (application state)
│   ├── ui/             # UI components (panels, styles)
│   ├── restic/         # Restic command execution
│   ├── config/         # Configuration parsing
│   └── types/          # Shared types
└── CLAUDE.md           # Development guide
```

### Running Tests

```bash
go test ./...
```

### Development Mode

```bash
# Run without building
go run main.go

# Build and run
go build && ./lazyrestic
```

## Roadmap

### Phase 1: Foundation ✅
- [x] Basic Bubbletea setup
- [x] Panel-based UI layout
- [x] Repository listing
- [x] Snapshot browser
- [x] Configuration parser

### Phase 2: Core Operations ✅
- [x] Backup initiation with streaming progress tracking
- [x] Interactive backup configuration dialog
- [x] Real-time log streaming
- [x] Snapshot restore workflow with interactive form
- [x] Restore to original or custom location
- [x] Repository statistics display (size, count, last backup)
- [x] Repository health checks

### Phase 3: Advanced Features
- [ ] Snapshot mounting and file browsing
- [ ] Diff between snapshots
- [ ] Prune/forget operations
- [ ] Search and filtering

### Phase 4: Polish
- [ ] Systemd timer integration
- [ ] Configuration management UI
- [ ] Comprehensive error handling
- [ ] Performance optimization

## Similar Projects

- [lazydocker](https://github.com/jesseduffield/lazydocker) - Docker management TUI
- [lazygit](https://github.com/jesseduffield/lazygit) - Git management TUI
- [k9s](https://github.com/derailed/k9s) - Kubernetes management TUI

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

TBD (likely MIT or Apache 2.0)

## Author

Craig Derington

---

*Making backups beautiful, one terminal at a time* ✨
