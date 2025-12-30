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
	scrollOffset int // Viewport scroll offset
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
	// Reset scroll when repos change
	p.scrollOffset = 0
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
		// Adjust scroll offset to keep selection visible
		if p.selected < p.scrollOffset {
			p.scrollOffset = p.selected
		}
	}
}

// MoveDown moves the selection down
func (p *RepositoryPanel) MoveDown() {
	if p.selected < len(p.repositories)-1 {
		p.selected++
		// Adjust scroll offset to keep selection visible
		// Each repo takes ~3 lines (name + path + spacing)
		visibleRepos := (p.height - 6) / 3 // Rough estimate
		if visibleRepos < 1 {
			visibleRepos = 1
		}
		if p.selected >= p.scrollOffset+visibleRepos {
			p.scrollOffset = p.selected - visibleRepos + 1
		}
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
		// Calculate visible area for viewport scrolling
		// Each repo takes ~3 lines (name + path + spacing)
		visibleRepos := (p.height - 6) / 3
		if visibleRepos < 1 {
			visibleRepos = 1
		}

		totalRepos := len(p.repositories)

		// Show scroll indicator at top
		if p.scrollOffset > 0 {
			scrollTopStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
			b.WriteString(scrollTopStyle.Render("  ▲ more above...\n"))
		}

		// Calculate viewport bounds
		startIdx := p.scrollOffset
		endIdx := p.scrollOffset + visibleRepos
		if endIdx > totalRepos {
			endIdx = totalRepos
		}

		// Render only visible repositories
		for i := startIdx; i < endIdx; i++ {
			repo := p.repositories[i]
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

		// Show scroll indicator at bottom
		if endIdx < totalRepos {
			scrollBottomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
			b.WriteString(scrollBottomStyle.Render("  ▼ more below...\n"))
		}
	}

	// Render panel with embedded title
	return RenderPanelWithTitle(title, b.String(), p.width, p.height, active)
}
