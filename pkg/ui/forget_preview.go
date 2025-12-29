package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// ForgetPreview displays the dry-run results before actual deletion
type ForgetPreview struct {
	results      []types.ForgetResult
	policy       types.ForgetPolicy
	width        int
	height       int
	scrollOffset int
}

// NewForgetPreview creates a new preview panel
func NewForgetPreview(results []types.ForgetResult, policy types.ForgetPolicy) *ForgetPreview {
	return &ForgetPreview{
		results: results,
		policy:  policy,
	}
}

// SetSize sets the panel dimensions
func (fp *ForgetPreview) SetSize(width, height int) {
	fp.width = width
	fp.height = height
}

// GetTotalToRemove returns the total number of snapshots that will be removed
func (fp *ForgetPreview) GetTotalToRemove() int {
	total := 0
	for _, result := range fp.results {
		total += len(result.SnapshotsToRemove)
	}
	return total
}

// GetTotalToKeep returns the total number of snapshots that will be kept
func (fp *ForgetPreview) GetTotalToKeep() int {
	total := 0
	for _, result := range fp.results {
		total += len(result.SnapshotsToKeep)
	}
	return total
}

// Render renders the preview panel
func (fp *ForgetPreview) Render() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")). // Red
		Background(lipgloss.Color("52")).   // Dark red background
		Padding(0, 2)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginTop(1)

	keepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("40")). // Green
		Padding(0, 2)

	removeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Padding(0, 2)

	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("52")).
		Padding(1, 2).
		MarginTop(1)

	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		MarginTop(1)

	// Title
	b.WriteString(titleStyle.Render("⚠️  DRY-RUN PREVIEW - NO CHANGES MADE YET") + "\n\n")

	// Warning
	warning := "DANGER: The following snapshots will be PERMANENTLY DELETED!\nThis action CANNOT be undone!"
	b.WriteString(warningStyle.Render(warning) + "\n")

	// Policy summary
	b.WriteString(headerStyle.Render("Applied Policy:") + "\n")
	policyLines := fp.formatPolicy()
	for _, line := range policyLines {
		b.WriteString("  " + line + "\n")
	}

	totalRemove := fp.GetTotalToRemove()
	totalKeep := fp.GetTotalToKeep()

	// Summary
	summary := fmt.Sprintf("Will KEEP: %d snapshots  |  Will DELETE: %d snapshots", totalKeep, totalRemove)
	b.WriteString("\n" + summaryStyle.Render(summary) + "\n\n")

	if totalRemove == 0 {
		noDeleteStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Bold(true)
		b.WriteString(noDeleteStyle.Render("✓ No snapshots will be deleted with this policy.") + "\n")
	} else {
		// Show snapshots to be removed
		b.WriteString(headerStyle.Render(fmt.Sprintf("Snapshots to DELETE (%d):", totalRemove)) + "\n")

		count := 0
		maxShow := 10 // Limit display to avoid overwhelming
		for _, result := range fp.results {
			for _, snap := range result.SnapshotsToRemove {
				if count >= maxShow {
					remaining := totalRemove - maxShow
					b.WriteString(removeStyle.Render(fmt.Sprintf("  ... and %d more snapshots", remaining)) + "\n")
					goto endRemoveList
				}

				timeStr := FormatTimeAgo(snap.Time)
				tags := ""
				if len(snap.Tags) > 0 {
					tags = fmt.Sprintf(" [tags: %s]", strings.Join(snap.Tags, ", "))
				}
				line := fmt.Sprintf("  ✗ %s - %s%s", snap.ShortID, timeStr, tags)
				b.WriteString(removeStyle.Render(line) + "\n")
				count++
			}
		}
	endRemoveList:

		b.WriteString("\n")

		// Show snapshots to be kept (brief)
		b.WriteString(headerStyle.Render(fmt.Sprintf("Snapshots to KEEP (%d):", totalKeep)) + "\n")
		if totalKeep <= 5 {
			for _, result := range fp.results {
				for _, snap := range result.SnapshotsToKeep {
					timeStr := FormatTimeAgo(snap.Time)
					line := fmt.Sprintf("  ✓ %s - %s", snap.ShortID, timeStr)
					b.WriteString(keepStyle.Render(line) + "\n")
				}
			}
		} else {
			b.WriteString(keepStyle.Render(fmt.Sprintf("  (showing %d most recent)", 3)) + "\n")
			count := 0
			for _, result := range fp.results {
				for _, snap := range result.SnapshotsToKeep {
					if count >= 3 {
						goto endKeepList
					}
					timeStr := FormatTimeAgo(snap.Time)
					line := fmt.Sprintf("  ✓ %s - %s", snap.ShortID, timeStr)
					b.WriteString(keepStyle.Render(line) + "\n")
					count++
				}
			}
		endKeepList:
		}

		// Confirmation instructions
		b.WriteString("\n")
		confirmText := "To proceed with deletion, you must type exactly: DELETE"
		b.WriteString(confirmStyle.Render(confirmText) + "\n")
	}

	// Border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(fp.width - 4)

	return boxStyle.Render(b.String())
}

// formatPolicy formats the policy for display
func (fp *ForgetPreview) formatPolicy() []string {
	var lines []string
	p := fp.policy

	if p.KeepLast > 0 {
		lines = append(lines, fmt.Sprintf("Keep last %d snapshots", p.KeepLast))
	}
	if p.KeepDaily > 0 {
		lines = append(lines, fmt.Sprintf("Keep %d daily snapshots", p.KeepDaily))
	}
	if p.KeepWeekly > 0 {
		lines = append(lines, fmt.Sprintf("Keep %d weekly snapshots", p.KeepWeekly))
	}
	if p.KeepMonthly > 0 {
		lines = append(lines, fmt.Sprintf("Keep %d monthly snapshots", p.KeepMonthly))
	}
	if p.KeepYearly > 0 {
		lines = append(lines, fmt.Sprintf("Keep %d yearly snapshots", p.KeepYearly))
	}
	if p.KeepWithin != "" {
		lines = append(lines, fmt.Sprintf("Keep snapshots within %s", p.KeepWithin))
	}

	if len(lines) == 0 {
		lines = append(lines, "(No retention rules specified)")
	}

	return lines
}
