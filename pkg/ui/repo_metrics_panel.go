package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// RepoMetricsPanel displays metrics for the currently selected repository
type RepoMetricsPanel struct {
	width      int
	height     int
	repository *types.Repository
	active     bool
}

// NewRepoMetricsPanel creates a new repository metrics panel
func NewRepoMetricsPanel() *RepoMetricsPanel {
	return &RepoMetricsPanel{
		width:      80,
		height:     10,
		repository: nil,
		active:     false,
	}
}

// SetSize updates the panel dimensions
func (p *RepoMetricsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetRepository updates the repository being displayed
func (p *RepoMetricsPanel) SetRepository(repo *types.Repository) {
	p.repository = repo
}

// SetActive sets whether this panel is active
func (p *RepoMetricsPanel) SetActive(active bool) {
	p.active = active
}

// GetWidth returns the panel width
func (p *RepoMetricsPanel) GetWidth() int {
	return p.width
}

// GetHeight returns the panel height
func (p *RepoMetricsPanel) GetHeight() int {
	return p.height
}

// Render returns the panel's view
func (p *RepoMetricsPanel) Render() string {
	title := "[2] Metrics"

	// If no repository selected
	if p.repository == nil {
		content := "\nNo repository selected"
		return RenderPanelWithTitle(title, content, p.width, p.height, p.active)
	}

	// Build metrics display
	lines := []string{}

	// Add top margin/padding for breathing room
	lines = append(lines, "")

	// Repository name and path
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render(p.repository.Name))
	lines = append(lines, lipgloss.NewStyle().Foreground(colorDimmed).Render(p.repository.Path))
	lines = append(lines, "") // Blank line

	// Metrics in columns
	col1 := []string{}
	col2 := []string{}

	// Column 1: Snapshots and Status
	col1 = append(col1, lipgloss.NewStyle().Foreground(colorInfo).Render("Snapshots:"))
	col1 = append(col1, fmt.Sprintf("  %d", p.repository.SnapshotCount))
	col1 = append(col1, "")
	col1 = append(col1, lipgloss.NewStyle().Foreground(colorInfo).Render("Status:"))
	col1 = append(col1, "  "+StatusStyle(p.repository.Status).Render(p.repository.Status))

	// Column 2: Size and Files
	col2 = append(col2, lipgloss.NewStyle().Foreground(colorInfo).Render("Total Size:"))
	col2 = append(col2, fmt.Sprintf("  %s", formatBytes(p.repository.Size)))
	col2 = append(col2, "")
	col2 = append(col2, lipgloss.NewStyle().Foreground(colorInfo).Render("Total Files:"))
	col2 = append(col2, fmt.Sprintf("  %d", p.repository.TotalFiles))

	// Join columns
	col1Str := strings.Join(col1, "\n")
	col2Str := strings.Join(col2, "\n")

	colWidth := (p.width - 10) / 2
	col1Styled := lipgloss.NewStyle().Width(colWidth).Render(col1Str)
	col2Styled := lipgloss.NewStyle().Width(colWidth).Render(col2Str)

	metricsRow := lipgloss.JoinHorizontal(lipgloss.Top, col1Styled, col2Styled)
	lines = append(lines, metricsRow)

	// Last backup time
	if !p.repository.LastBackup.IsZero() {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorInfo).Render("Last Backup:"))
		lines = append(lines, "  "+FormatTimeAgo(p.repository.LastBackup))
	}

	// Render panel with embedded title
	return RenderPanelWithTitle(title, strings.Join(lines, "\n"), p.width, p.height, p.active)
}
