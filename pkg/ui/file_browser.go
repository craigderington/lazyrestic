package ui

import (
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// FileBrowser represents the file browser panel for browsing snapshot contents
type FileBrowser struct {
	snapshot    *types.Snapshot  // The snapshot being browsed
	currentPath string           // Current directory path
	files       []types.FileNode // Files in current directory
	selected    int              // Selected file index
	width       int
	height      int
	multiSelect bool // Enable multi-selection mode

	// Pagination
	pageSize    int // Number of files per page
	currentPage int // Current page (0-based)
}

// NewFileBrowser creates a new file browser for a snapshot
func NewFileBrowser(snapshot *types.Snapshot) *FileBrowser {
	return &FileBrowser{
		snapshot:    snapshot,
		currentPath: "/",
		files:       []types.FileNode{},
		selected:    0,
		multiSelect: true,
		pageSize:    50, // Show 50 files per page
		currentPage: 0,
	}
}

// SetFiles updates the list of files for the current directory
func (fb *FileBrowser) SetFiles(files []types.FileNode) {
	fb.files = files
	fb.currentPage = 0 // Reset to first page

	// Adjust selection if out of bounds
	totalPages := fb.getTotalPages()
	if fb.currentPage >= totalPages && totalPages > 0 {
		fb.currentPage = totalPages - 1
	}
	if fb.currentPage < 0 {
		fb.currentPage = 0
	}

	// Adjust selected within current page
	filesOnPage := fb.getFilesOnCurrentPage()
	if fb.selected >= len(filesOnPage) && len(filesOnPage) > 0 {
		fb.selected = len(filesOnPage) - 1
	}
	if fb.selected < 0 {
		fb.selected = 0
	}
}

// SetSize updates the panel dimensions
func (fb *FileBrowser) SetSize(width, height int) {
	fb.width = width
	fb.height = height
}

// getTotalPages returns the total number of pages
func (fb *FileBrowser) getTotalPages() int {
	if len(fb.files) == 0 {
		return 1
	}
	return (len(fb.files) + fb.pageSize - 1) / fb.pageSize
}

// getFilesOnCurrentPage returns the files for the current page
func (fb *FileBrowser) getFilesOnCurrentPage() []types.FileNode {
	start := fb.currentPage * fb.pageSize
	end := start + fb.pageSize
	if end > len(fb.files) {
		end = len(fb.files)
	}
	if start >= len(fb.files) {
		return []types.FileNode{}
	}
	return fb.files[start:end]
}

// NextPage moves to the next page
func (fb *FileBrowser) NextPage() {
	totalPages := fb.getTotalPages()
	if fb.currentPage < totalPages-1 {
		fb.currentPage++
		fb.selected = 0 // Reset selection to first item
	}
}

// PrevPage moves to the previous page
func (fb *FileBrowser) PrevPage() {
	if fb.currentPage > 0 {
		fb.currentPage--
		fb.selected = 0 // Reset selection to first item
	}
}

// MoveUp moves the selection up
func (fb *FileBrowser) MoveUp() {
	if fb.selected > 0 {
		fb.selected--
	}
}

// MoveDown moves the selection down
func (fb *FileBrowser) MoveDown() {
	if fb.selected < len(fb.files)-1 {
		fb.selected++
	}
}

// GetSelected returns the currently selected file node
func (fb *FileBrowser) GetSelected() *types.FileNode {
	if fb.selected >= 0 && fb.selected < len(fb.files) {
		return &fb.files[fb.selected]
	}
	return nil
}

// ToggleSelection toggles the selection state of the current file
func (fb *FileBrowser) ToggleSelection() {
	if fb.selected >= 0 && fb.selected < len(fb.files) {
		fb.files[fb.selected].Selected = !fb.files[fb.selected].Selected
	}
}

// GetSelectedFiles returns all files marked as selected
func (fb *FileBrowser) GetSelectedFiles() []types.FileNode {
	var selected []types.FileNode
	for _, file := range fb.files {
		if file.Selected {
			selected = append(selected, file)
		}
	}
	return selected
}

// ClearSelection clears all selections
func (fb *FileBrowser) ClearSelection() {
	for i := range fb.files {
		fb.files[i].Selected = false
	}
}

// GetCurrentPath returns the current directory path
func (fb *FileBrowser) GetCurrentPath() string {
	return fb.currentPath
}

// GetSnapshot returns the snapshot being browsed
func (fb *FileBrowser) GetSnapshot() *types.Snapshot {
	return fb.snapshot
}

// SetCurrentPath sets the current directory path
func (fb *FileBrowser) SetCurrentPath(path string) {
	fb.currentPath = path
}

// CanGoUp returns true if we can navigate to parent directory
func (fb *FileBrowser) CanGoUp() bool {
	return fb.currentPath != "/" && fb.currentPath != ""
}

// GoUp navigates to the parent directory
func (fb *FileBrowser) GoUp() string {
	if !fb.CanGoUp() {
		return fb.currentPath
	}
	fb.currentPath = path.Dir(fb.currentPath)
	if fb.currentPath == "." {
		fb.currentPath = "/"
	}
	return fb.currentPath
}

// EnterDirectory enters the selected directory
func (fb *FileBrowser) EnterDirectory() (string, bool) {
	selected := fb.GetSelected()
	if selected != nil && selected.IsDir() {
		fb.currentPath = selected.Path
		return fb.currentPath, true
	}
	return fb.currentPath, false
}

// Render renders the file browser panel
func (fb *FileBrowser) Render(active bool) string {
	var b strings.Builder

	// Panel title with breadcrumb
	titleStyle := PanelTitleStyle
	borderStyle := PanelBorderStyle
	if active {
		titleStyle = PanelTitleActiveStyle
		borderStyle = PanelBorderActiveStyle
	}

	// Breadcrumb path
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	displayPath := fb.currentPath
	if len(displayPath) > 40 {
		displayPath = "..." + displayPath[len(displayPath)-37:]
	}

	title := titleStyle.Render("📁 Files") + " " + pathStyle.Render(displayPath)
	b.WriteString(title + "\n\n")

	// Show snapshot info
	if fb.snapshot != nil {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		b.WriteString(infoStyle.Render(fmt.Sprintf("Snapshot: %s", fb.snapshot.ShortID)) + "\n\n")
	}

	// File list
	if len(fb.files) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(emptyStyle.Render("No files in this directory\n"))
		if fb.CanGoUp() {
			b.WriteString(emptyStyle.Render("Press ← or h to go back"))
		}
	} else {
		// Add ".." entry if we can go up
		if fb.CanGoUp() {
			backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			b.WriteString(backStyle.Render("  .. (parent directory)") + "\n")
		}

		// Render files (paginated)
		filesOnPage := fb.getFilesOnCurrentPage()
		for i, file := range filesOnPage {
			var line string

			// Selection indicator
			if i == fb.selected && active {
				line = "▶ "
			} else if i == fb.selected {
				line = "• "
			} else {
				line = "  "
			}

			// Multi-select checkbox
			if file.Selected {
				line += "[✓] "
			} else if fb.multiSelect {
				line += "[ ] "
			}

			// Icon and name
			icon := "📄"
			if file.IsDir() {
				icon = "📁"
			}
			line += icon + " " + file.Name

			// Add size and time for files
			if file.IsFile() {
				sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
				line += sizeStyle.Render(fmt.Sprintf(" (%s)", formatBytes(file.Size)))
			}

			// Style the line
			if i == fb.selected {
				line = ListItemSelectedStyle.Render(line)
			} else {
				line = ListItemStyle.Render(line)
			}

			b.WriteString(line + "\n")
		}

		// Selection info at bottom
		selectedCount := len(fb.GetSelectedFiles())
		if selectedCount > 0 {
			selectionStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)
			b.WriteString("\n" + selectionStyle.Render(fmt.Sprintf("%d files selected", selectedCount)))
		}

		// Pagination info
		totalPages := fb.getTotalPages()
		if totalPages > 1 {
			pageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			pageInfo := fmt.Sprintf("Page %d/%d (%d files)", fb.currentPage+1, totalPages, len(fb.files))
			b.WriteString("\n" + pageStyle.Render(pageInfo))
		}
	}

	content := b.String()

	// Apply border
	return borderStyle.
		Width(fb.width - 4).
		Height(fb.height - 4).
		Render(content)
}
