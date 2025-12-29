package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// RepositoryPanel represents the repository list panel
type RepositoryPanel struct {
	repositories []types.Repository
	selected     int
	width        int
	height       int
}

// NewRepositoryPanel creates a new repository panel
func NewRepositoryPanel() *RepositoryPanel {
	return &RepositoryPanel{
		repositories: []types.Repository{},
		selected:     0,
	}
}

// SetRepositories updates the list of repositories
func (p *RepositoryPanel) SetRepositories(repos []types.Repository) {
	p.repositories = repos
	if p.selected >= len(repos) && len(repos) > 0 {
		p.selected = len(repos) - 1
	}
}

// SetSize updates the panel dimensions
func (p *RepositoryPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// MoveUp moves the selection up
func (p *RepositoryPanel) MoveUp() {
	if p.selected > 0 {
		p.selected--
	}
}

// MoveDown moves the selection down
func (p *RepositoryPanel) MoveDown() {
	if p.selected < len(p.repositories)-1 {
		p.selected++
	}
}

// GetSelected returns the currently selected repository
func (p *RepositoryPanel) GetSelected() *types.Repository {
	if p.selected >= 0 && p.selected < len(p.repositories) {
		return &p.repositories[p.selected]
	}
	return nil
}

// Render renders the repository panel
func (p *RepositoryPanel) Render(active bool) string {
	var b strings.Builder

	title := "[1] Repositories"

	// Add top margin/padding for breathing room
	b.WriteString("\n")

	// Repository list
	if len(p.repositories) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No repositories configured\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Add repositories to ~/.config/lazyrestic/config.yaml"))
	} else {
		for i, repo := range p.repositories {
			var line string
			if i == p.selected && active {
				line = ListItemSelectedStyle.Render(fmt.Sprintf("▶ %s", repo.Name))
			} else if i == p.selected {
				line = ListItemStyle.Render(fmt.Sprintf("• %s", repo.Name))
			} else {
				line = ListItemStyle.Render(fmt.Sprintf("  %s", repo.Name))
			}

			b.WriteString(line + "\n")

			// Show path in dimmed color (for all repos, not just selected)
			// Using lipgloss MarginBottom for proper spacing
			pathStyle := lipgloss.NewStyle().
				Foreground(colorDimmed).
				PaddingLeft(2).
				MarginBottom(1) // Proper spacing between items

			// Truncate long paths
			displayPath := repo.Path
			maxLen := p.width - 8
			if len(displayPath) > maxLen && maxLen > 10 {
				displayPath = "..." + displayPath[len(displayPath)-(maxLen-3):]
			}

			b.WriteString(pathStyle.Render(displayPath) + "\n")
		}
	}

	// Render panel with embedded title
	return RenderPanelWithTitle(title, b.String(), p.width, p.height, active)
}
