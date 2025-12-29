package ui

import (
	"strings"
	"testing"
	"time"
)

func TestNewOperationsPanel(t *testing.T) {
	panel := NewOperationsPanel()

	if panel == nil {
		t.Fatal("NewOperationsPanel() returned nil")
	}

	if len(panel.logs) != 0 {
		t.Errorf("Initial logs length = %v, want 0", len(panel.logs))
	}
}

func TestOperationsPanel_AddLog(t *testing.T) {
	panel := NewOperationsPanel()

	panel.AddLog("info", "Test message")

	if len(panel.logs) != 1 {
		t.Fatalf("Logs length = %v, want 1", len(panel.logs))
	}

	log := panel.logs[0]
	if log.Level != "info" {
		t.Errorf("Log level = %v, want info", log.Level)
	}
	if log.Message != "Test message" {
		t.Errorf("Log message = %v, want 'Test message'", log.Message)
	}
	if log.Timestamp.IsZero() {
		t.Error("Log timestamp should not be zero")
	}
}

func TestOperationsPanel_ConvenienceMethods(t *testing.T) {
	panel := NewOperationsPanel()

	panel.Info("Info message")
	panel.Success("Success message")
	panel.Warning("Warning message")
	panel.Error("Error message")

	if len(panel.logs) != 4 {
		t.Fatalf("Logs length = %v, want 4", len(panel.logs))
	}

	tests := []struct {
		index   int
		level   string
		message string
	}{
		{0, "info", "Info message"},
		{1, "success", "Success message"},
		{2, "warning", "Warning message"},
		{3, "error", "Error message"},
	}

	for _, tt := range tests {
		log := panel.logs[tt.index]
		if log.Level != tt.level {
			t.Errorf("Log %d level = %v, want %v", tt.index, log.Level, tt.level)
		}
		if log.Message != tt.message {
			t.Errorf("Log %d message = %v, want %v", tt.index, log.Message, tt.message)
		}
	}
}

func TestOperationsPanel_LogLimit(t *testing.T) {
	panel := NewOperationsPanel()

	// Add 150 logs (limit is 100)
	for i := 0; i < 150; i++ {
		panel.Info("Log entry")
	}

	if len(panel.logs) != 100 {
		t.Errorf("Logs length = %v, want 100 (should be limited)", len(panel.logs))
	}
}

func TestOperationsPanel_LogOrder(t *testing.T) {
	panel := NewOperationsPanel()

	panel.Info("First")
	time.Sleep(1 * time.Millisecond)
	panel.Info("Second")
	time.Sleep(1 * time.Millisecond)
	panel.Info("Third")

	// Logs should be in chronological order
	if panel.logs[0].Message != "First" {
		t.Errorf("First log message = %v, want 'First'", panel.logs[0].Message)
	}
	if panel.logs[2].Message != "Third" {
		t.Errorf("Third log message = %v, want 'Third'", panel.logs[2].Message)
	}

	// Timestamps should be in order
	if !panel.logs[0].Timestamp.Before(panel.logs[1].Timestamp) {
		t.Error("Log timestamps should be in chronological order")
	}
}

func TestOperationsPanel_SetSize(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 20)

	if panel.width != 100 {
		t.Errorf("Width = %v, want 100", panel.width)
	}
	if panel.height != 20 {
		t.Errorf("Height = %v, want 20", panel.height)
	}
}

func TestOperationsPanel_Render_Empty(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(80, 20)

	output := panel.Render(false)

	if output == "" {
		t.Error("Render() should not return empty string")
	}

	if !strings.Contains(output, "No operations") {
		t.Error("Empty panel should show 'No operations' message")
	}
}

func TestOperationsPanel_Render_WithLogs(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 20)

	panel.Info("Information log")
	panel.Success("Success log")
	panel.Warning("Warning log")
	panel.Error("Error log")

	output := panel.Render(false)

	// Should contain all log messages
	if !strings.Contains(output, "Information log") {
		t.Error("Render() should contain info message")
	}
	if !strings.Contains(output, "Success log") {
		t.Error("Render() should contain success message")
	}
	if !strings.Contains(output, "Warning log") {
		t.Error("Render() should contain warning message")
	}
	if !strings.Contains(output, "Error log") {
		t.Error("Render() should contain error message")
	}
}

func TestOperationsPanel_Render_LevelIndicators(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 20)

	panel.Success("Success")
	panel.Warning("Warning")
	panel.Error("Error")

	output := panel.Render(false)

	// Should contain level-specific indicators
	// Success: ✓, Warning: ⚠, Error: ✗
	indicators := []string{"✓", "⚠", "✗"}
	for _, indicator := range indicators {
		if !strings.Contains(output, indicator) {
			t.Errorf("Render() should contain indicator '%v'", indicator)
		}
	}
}

func TestOperationsPanel_Render_ShowsTimestamps(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 20)

	panel.Info("Test message")

	output := panel.Render(false)

	// Should contain timestamp in HH:MM:SS format
	// Check for : separator (timestamps have colons)
	colonCount := strings.Count(output, ":")
	if colonCount < 2 { // At least HH:MM:SS has 2 colons
		t.Error("Render() should show timestamps")
	}
}

func TestOperationsPanel_Render_ActiveState(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 20)

	panel.Info("Test")

	inactiveOutput := panel.Render(false)
	activeOutput := panel.Render(true)

	// Both should render successfully
	if inactiveOutput == "" {
		t.Error("Inactive render should not be empty")
	}
	if activeOutput == "" {
		t.Error("Active render should not be empty")
	}

	// Both should contain the panel title
	if !strings.Contains(inactiveOutput, "Operations") {
		t.Error("Render should contain 'Operations' title")
	}
	if !strings.Contains(activeOutput, "Operations") {
		t.Error("Active render should contain 'Operations' title")
	}
}

func TestOperationsPanel_Render_LimitsDisplayedEntries(t *testing.T) {
	panel := NewOperationsPanel()
	panel.SetSize(100, 10) // Small height

	// Add many log entries
	for i := 0; i < 50; i++ {
		panel.Info("Log entry")
	}

	output := panel.Render(false)

	// Output should not be excessively long
	lines := strings.Split(output, "\n")

	// Should be roughly limited by panel height
	// (exact number depends on rendering, but should be reasonable)
	if len(lines) > 20 {
		t.Errorf("Render() output has %v lines, should be limited by panel height", len(lines))
	}
}

func TestLogEntry_Timestamp(t *testing.T) {
	before := time.Now()
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
	}
	after := time.Now()

	// Timestamp should be between before and after
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Error("LogEntry timestamp should be current time")
	}
}

func BenchmarkOperationsPanel_AddLog(b *testing.B) {
	panel := NewOperationsPanel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.AddLog("info", "Benchmark message")
	}
}

func BenchmarkOperationsPanel_Render(b *testing.B) {
	panel := NewOperationsPanel()
	panel.SetSize(120, 40)

	// Populate with logs
	for i := 0; i < 50; i++ {
		panel.Info("Log message")
		panel.Success("Success message")
		panel.Warning("Warning message")
		panel.Error("Error message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = panel.Render(i%2 == 0)
	}
}
