package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/craigderington/lazyrestic/pkg/types"
)

// LogEntry represents a log message
type LogEntry struct {
	Timestamp time.Time
	Level     string // "info", "success", "warning", "error"
	Message   string
}

// OperationsPanel represents the operations/logs panel
type OperationsPanel struct {
	logs             []LogEntry
	width            int
	height           int
	backupProgress   *types.BackupProgress
	backupInProgress bool
}

// NewOperationsPanel creates a new operations panel
func NewOperationsPanel() *OperationsPanel {
	return &OperationsPanel{
		logs: []LogEntry{},
	}
}

// AddLog adds a log entry
func (p *OperationsPanel) AddLog(level, message string) {
	p.logs = append(p.logs, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	})

	// Keep only last 100 entries
	if len(p.logs) > 100 {
		p.logs = p.logs[len(p.logs)-100:]
	}
}

// Info adds an info log
func (p *OperationsPanel) Info(message string) {
	p.AddLog("info", message)
}

// Dimmed adds a dimmed info log
func (p *OperationsPanel) Dimmed(message string) {
	p.AddLog("dimmed", message)
}

// Success adds a success log
func (p *OperationsPanel) Success(message string) {
	p.AddLog("success", message)
}

// Warning adds a warning log
func (p *OperationsPanel) Warning(message string) {
	p.AddLog("warning", message)
}

// Error adds an error log
func (p *OperationsPanel) Error(message string) {
	p.AddLog("error", message)
}

// SetBackupProgress updates the backup progress
func (p *OperationsPanel) SetBackupProgress(progress *types.BackupProgress) {
	p.backupProgress = progress
	p.backupInProgress = true
}

// ClearBackupProgress clears the backup progress
func (p *OperationsPanel) ClearBackupProgress() {
	p.backupProgress = nil
	p.backupInProgress = false
}

// SetSize updates the panel dimensions
func (p *OperationsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// renderProgressBar renders a progress bar
func renderProgressBar(percent float64, width int) string {
	if width < 10 {
		width = 10
	}

	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	return barStyle.Render(bar) + " " + percentStyle.Render(fmt.Sprintf("%.1f%%", percent))
}

// Render renders the operations panel
func (p *OperationsPanel) Render(active bool) string {
	var b strings.Builder

	title := "[4] Operations"

	// Add top margin/padding for breathing room
	b.WriteString("\n")

	// Show backup progress if active
	if p.backupInProgress && p.backupProgress != nil {
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

		b.WriteString(progressStyle.Render("Backup in Progress") + "\n\n")

		// Progress bar
		barWidth := p.width - 20
		if barWidth < 10 {
			barWidth = 10
		}
		b.WriteString(renderProgressBar(p.backupProgress.PercentDone, barWidth) + "\n\n")

		// Statistics
		b.WriteString(labelStyle.Render(fmt.Sprintf("Files: %d/%d  ",
			p.backupProgress.FilesDone, p.backupProgress.TotalFiles)))
		b.WriteString(labelStyle.Render(fmt.Sprintf("Data: %s/%s\n",
			formatBytes(p.backupProgress.BytesDone), formatBytes(p.backupProgress.TotalBytes))))

		// Current file (if available)
		if len(p.backupProgress.CurrentFiles) > 0 {
			currentFile := p.backupProgress.CurrentFiles[0]
			if len(currentFile) > 60 {
				currentFile = "..." + currentFile[len(currentFile)-57:]
			}
			b.WriteString(labelStyle.Render(fmt.Sprintf("Processing: %s\n", currentFile)))
		}

		b.WriteString("\n")
	}

	// Log entries
	if len(p.logs) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No operations yet"))
	} else {
		// Show last N entries that fit in the panel
		maxEntries := (p.height - 8) / 2 // Each entry takes ~2 lines
		if maxEntries < 1 {
			maxEntries = 1
		}

		startIdx := len(p.logs) - maxEntries
		if startIdx < 0 {
			startIdx = 0
		}

		for i := startIdx; i < len(p.logs); i++ {
			entry := p.logs[i]

			// Style based on level
			var levelStyle lipgloss.Style
			var levelPrefix string

			switch entry.Level {
			case "success":
				levelStyle = StatusHealthyStyle
				levelPrefix = "✓"
			case "warning":
				levelStyle = StatusWarningStyle
				levelPrefix = "⚠"
			case "error":
				levelStyle = StatusErrorStyle
				levelPrefix = "✗"
			case "dimmed":
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Faint(true) // Dimmed and faint
				levelPrefix = "•"
			default:
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
				levelPrefix = "•"
			}

			timestamp := entry.Timestamp.Format("15:04:05")
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

			line := timeStyle.Render(timestamp) + " " +
				levelStyle.Render(levelPrefix) + " " +
				entry.Message

			b.WriteString(line + "\n")
		}
	}

	// Render panel with embedded title
	return RenderPanelWithTitle(title, b.String(), p.width, p.height, active)
}
