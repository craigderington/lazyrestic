package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// SnapshotPanel represents the snapshot list panel
type SnapshotPanel struct {
	snapshots         []types.Snapshot // All snapshots
	filteredSnapshots []types.Snapshot // Filtered view
	selected          int
	width             int
	height            int
	scrollOffset      int // Viewport scroll offset

	// Filter state
	filterActive bool
	filterText   string
	filterTag    string
	filterHost   string
}

// NewSnapshotPanel creates a new snapshot panel
func NewSnapshotPanel() *SnapshotPanel {
	return &SnapshotPanel{
		snapshots: []types.Snapshot{},
		selected:  0,
	}
}

// SetSnapshots updates the list of snapshots
func (p *SnapshotPanel) SetSnapshots(snapshots []types.Snapshot) {
	p.snapshots = snapshots
	p.ApplyFilter()

	// Adjust selection to fit within filtered list
	listLen := len(p.filteredSnapshots)
	if p.selected >= listLen && listLen > 0 {
		p.selected = listLen - 1
	}

	// Reset scroll offset
	p.scrollOffset = 0
}

// ApplyFilter applies the current filter settings to the snapshot list
func (p *SnapshotPanel) ApplyFilter() {
	// If no filter is active, show all snapshots
	if !p.filterActive || (p.filterText == "" && p.filterTag == "" && p.filterHost == "") {
		p.filteredSnapshots = p.snapshots
		p.scrollOffset = 0
		return
	}

	// Filter snapshots based on active filters
	p.filteredSnapshots = []types.Snapshot{}
	for _, snap := range p.snapshots {
		if p.matchesFilter(snap) {
			p.filteredSnapshots = append(p.filteredSnapshots, snap)
		}
	}

	// Reset selection and scroll if current selection is out of bounds
	if p.selected >= len(p.filteredSnapshots) && len(p.filteredSnapshots) > 0 {
		p.selected = 0
		p.scrollOffset = 0
	}
}

// matchesFilter checks if a snapshot matches the current filter criteria
func (p *SnapshotPanel) matchesFilter(snap types.Snapshot) bool {
	// Filter by tag
	if p.filterTag != "" {
		found := false
		for _, tag := range snap.Tags {
			if strings.Contains(strings.ToLower(tag), strings.ToLower(p.filterTag)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by hostname
	if p.filterHost != "" {
		if !strings.Contains(strings.ToLower(snap.Hostname), strings.ToLower(p.filterHost)) {
			return false
		}
	}

	// Filter by text (search in snapshot ID and paths)
	if p.filterText != "" {
		filterLower := strings.ToLower(p.filterText)

		// Check snapshot ID
		if strings.Contains(strings.ToLower(snap.ID), filterLower) {
			return true
		}
		if strings.Contains(strings.ToLower(snap.ShortID), filterLower) {
			return true
		}

		// Check paths
		for _, path := range snap.Paths {
			if strings.Contains(strings.ToLower(path), filterLower) {
				return true
			}
		}

		// Check tags
		for _, tag := range snap.Tags {
			if strings.Contains(strings.ToLower(tag), filterLower) {
				return true
			}
		}

		// No match found
		return false
	}

	return true
}

// SetFilter sets a text filter and applies it
func (p *SnapshotPanel) SetFilter(text string) {
	p.filterText = text
	p.filterActive = true
	p.ApplyFilter()
}

// SetTagFilter sets a tag filter and applies it
func (p *SnapshotPanel) SetTagFilter(tag string) {
	p.filterTag = tag
	p.filterActive = true
	p.ApplyFilter()
}

// SetHostFilter sets a hostname filter and applies it
func (p *SnapshotPanel) SetHostFilter(host string) {
	p.filterHost = host
	p.filterActive = true
	p.ApplyFilter()
}

// ClearFilter removes all filters
func (p *SnapshotPanel) ClearFilter() {
	p.filterActive = false
	p.filterText = ""
	p.filterTag = ""
	p.filterHost = ""
	p.ApplyFilter()
}

// IsFilterActive returns true if any filter is currently active
func (p *SnapshotPanel) IsFilterActive() bool {
	return p.filterActive && (p.filterText != "" || p.filterTag != "" || p.filterHost != "")
}

// SetSize updates the panel dimensions
func (p *SnapshotPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// GetWidth returns the panel width
func (p *SnapshotPanel) GetWidth() int {
	return p.width
}

// GetHeight returns the panel height
func (p *SnapshotPanel) GetHeight() int {
	return p.height
}

// MoveUp moves the selection up
func (p *SnapshotPanel) MoveUp() {
	if p.selected > 0 {
		p.selected--
		// Adjust scroll offset to keep selection visible
		if p.selected < p.scrollOffset {
			p.scrollOffset = p.selected
		}
	}
}

// MoveDown moves the selection down
func (p *SnapshotPanel) MoveDown() {
	listLen := len(p.filteredSnapshots)
	if p.selected < listLen-1 {
		p.selected++
		// Adjust scroll offset to keep selection visible
		// Calculate visible area (approximate - account for header, borders, padding)
		visibleLines := p.height - 6 // Rough estimate: height minus borders/title/padding
		if visibleLines < 1 {
			visibleLines = 1
		}
		if p.selected >= p.scrollOffset+visibleLines {
			p.scrollOffset = p.selected - visibleLines + 1
		}
	}
}

// GetSelected returns the currently selected snapshot
func (p *SnapshotPanel) GetSelected() *types.Snapshot {
	listLen := len(p.filteredSnapshots)
	if p.selected >= 0 && p.selected < listLen {
		return &p.filteredSnapshots[p.selected]
	}
	return nil
}

// Render renders the snapshot panel
func (p *SnapshotPanel) Render(active bool) string {
	var b strings.Builder

	title := "[3] Snapshots"

	// Add filter indicator if active
	if p.IsFilterActive() {
		filterParts := []string{}
		if p.filterText != "" {
			filterParts = append(filterParts, fmt.Sprintf("text=%s", p.filterText))
		}
		if p.filterTag != "" {
			filterParts = append(filterParts, fmt.Sprintf("tag=%s", p.filterTag))
		}
		if p.filterHost != "" {
			filterParts = append(filterParts, fmt.Sprintf("host=%s", p.filterHost))
		}
		filterInfo := strings.Join(filterParts, ", ")
		title += fmt.Sprintf(" [%s]", filterInfo)
	}

	// Add top margin/padding for breathing room
	b.WriteString("\n")

	// Snapshot list
	if len(p.snapshots) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No snapshots found\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Select a repository to view snapshots"))
	} else if len(p.filteredSnapshots) == 0 {
		// No snapshots match the filter
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("No snapshots match the current filter\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Press Esc to clear filter"))
	} else {
		// Show filter count if active
		if p.IsFilterActive() {
			countStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
			b.WriteString(countStyle.Render(fmt.Sprintf("[%d of %d snapshots shown]\n\n",
				len(p.filteredSnapshots), len(p.snapshots))))
		}

		// Calculate visible area for viewport scrolling
		visibleLines := p.height - 6 // Account for borders, title, padding
		if visibleLines < 1 {
			visibleLines = 1
		}

		totalSnapshots := len(p.filteredSnapshots)

		// Show scroll indicators
		if p.scrollOffset > 0 {
			scrollTopStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
			b.WriteString(scrollTopStyle.Render("  ▲ more above...\n"))
		}

		// Calculate viewport bounds
		startIdx := p.scrollOffset
		endIdx := p.scrollOffset + visibleLines
		if endIdx > totalSnapshots {
			endIdx = totalSnapshots
		}

		// Render only visible snapshots
		for i := startIdx; i < endIdx; i++ {
			snapshot := p.filteredSnapshots[i]
			var line string

			// Truncate ID for display
			shortID := snapshot.ShortID
			if shortID == "" && len(snapshot.ID) >= 8 {
				shortID = snapshot.ID[:8]
			}

			timeStr := FormatTimeAgo(snapshot.Time)

			if i == p.selected && active {
				line = ListItemSelectedStyle.Render(fmt.Sprintf("▶ %s", shortID))
			} else if i == p.selected {
				line = ListItemStyle.Render(fmt.Sprintf("• %s", shortID))
			} else {
				line = ListItemStyle.Render(fmt.Sprintf("  %s", shortID))
			}

			// Add timestamp
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			line += timeStyle.Render(fmt.Sprintf(" - %s", timeStr))

			b.WriteString(line + "\n")
		}

		// Show scroll indicator for more content below
		if endIdx < totalSnapshots {
			scrollBottomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
			b.WriteString(scrollBottomStyle.Render("  ▼ more below...\n"))
		}
	}

	// Render panel with embedded title
	return RenderPanelWithTitle(title, b.String(), p.width, p.height, active)
}
