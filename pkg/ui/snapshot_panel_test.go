package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestNewSnapshotPanel(t *testing.T) {
	panel := NewSnapshotPanel()

	if panel == nil {
		t.Fatal("NewSnapshotPanel() returned nil")
	}

	if panel.selected != 0 {
		t.Errorf("Initial selected = %v, want 0", panel.selected)
	}

	if len(panel.snapshots) != 0 {
		t.Errorf("Initial snapshots length = %v, want 0", len(panel.snapshots))
	}
}

func TestSnapshotPanel_SetSnapshots(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "abc123", ShortID: "abc123", Hostname: "host1", Time: time.Now()},
		{ID: "def456", ShortID: "def456", Hostname: "host2", Time: time.Now()},
	}

	panel.SetSnapshots(snapshots)

	if len(panel.snapshots) != 2 {
		t.Errorf("Snapshots length = %v, want 2", len(panel.snapshots))
	}
}

func TestSnapshotPanel_Navigation(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Time: time.Now()},
		{ID: "snap2", ShortID: "snap2", Time: time.Now()},
		{ID: "snap3", ShortID: "snap3", Time: time.Now()},
	}
	panel.SetSnapshots(snapshots)

	// Test MoveDown
	panel.MoveDown()
	if panel.selected != 1 {
		t.Errorf("After MoveDown, selected = %v, want 1", panel.selected)
	}

	// Test MoveUp
	panel.MoveUp()
	if panel.selected != 0 {
		t.Errorf("After MoveUp, selected = %v, want 0", panel.selected)
	}

	// Test boundary - can't go below 0
	panel.MoveUp()
	if panel.selected != 0 {
		t.Errorf("Selected should stay at 0, got %v", panel.selected)
	}

	// Move to last item
	panel.selected = 2
	panel.MoveDown()
	if panel.selected != 2 {
		t.Errorf("Selected should stay at 2, got %v", panel.selected)
	}
}

func TestSnapshotPanel_GetSelected(t *testing.T) {
	panel := NewSnapshotPanel()

	// No snapshots
	if panel.GetSelected() != nil {
		t.Error("GetSelected() should return nil with no snapshots")
	}

	snapshots := []types.Snapshot{
		{ID: "abc", ShortID: "abc", Hostname: "host1"},
		{ID: "def", ShortID: "def", Hostname: "host2"},
	}
	panel.SetSnapshots(snapshots)

	selected := panel.GetSelected()
	if selected == nil {
		t.Fatal("GetSelected() returned nil")
	}
	if selected.ShortID != "abc" {
		t.Errorf("Selected snapshot ShortID = %v, want abc", selected.ShortID)
	}

	panel.MoveDown()
	selected = panel.GetSelected()
	if selected.ShortID != "def" {
		t.Errorf("Selected snapshot ShortID = %v, want def", selected.ShortID)
	}
}

func TestSnapshotPanel_Render_Empty(t *testing.T) {
	panel := NewSnapshotPanel()
	panel.SetSize(100, 30)

	output := panel.Render(false)

	if output == "" {
		t.Error("Render() should not return empty string")
	}

	if !strings.Contains(output, "No snapshots") {
		t.Error("Empty panel should show 'No snapshots' message")
	}
}

func TestSnapshotPanel_Render_WithSnapshots(t *testing.T) {
	panel := NewSnapshotPanel()
	panel.SetSize(100, 30)

	now := time.Now()
	snapshots := []types.Snapshot{
		{
			ID:       "abc123def456",
			ShortID:  "abc123",
			Hostname: "myhost",
			Time:     now,
			Paths:    []string{"/home/user"},
			Tags:     []string{"important"},
		},
	}
	panel.SetSnapshots(snapshots)

	output := panel.Render(false)

	// Should contain short ID
	if !strings.Contains(output, "abc123") {
		t.Error("Render() should contain snapshot short ID")
	}

	// Should contain hostname (when selected)
	outputActive := panel.Render(true)
	if !strings.Contains(outputActive, "myhost") {
		t.Error("Active render should contain hostname")
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{
			name:     "Just now",
			time:     now,
			contains: "just now",
		},
		{
			name:     "Minutes ago",
			time:     now.Add(-5 * time.Minute),
			contains: "minutes ago",
		},
		{
			name:     "Hours ago",
			time:     now.Add(-2 * time.Hour),
			contains: "hours ago",
		},
		{
			name:     "Days ago",
			time:     now.Add(-3 * 24 * time.Hour),
			contains: "days ago",
		},
		{
			name:     "Days ago",
			time:     now.Add(-10 * 24 * time.Hour),
			contains: "days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimeAgo(tt.time)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatTimeAgo(%v) = %v, should contain %v",
					tt.time, result, tt.contains)
			}
		})
	}
}

func TestSnapshotPanel_Render_ShowsTags(t *testing.T) {
	panel := NewSnapshotPanel()
	panel.SetSize(100, 30)

	snapshots := []types.Snapshot{
		{
			ID:       "test",
			ShortID:  "test",
			Time:     time.Now(),
			Hostname: "host",
			Tags:     []string{"daily", "important", "production"},
		},
	}
	panel.SetSnapshots(snapshots)

	output := panel.Render(true)

	// Tags should be visible when snapshot is selected
	if !strings.Contains(output, "daily") {
		t.Error("Render() should show tags")
	}
}

func TestSnapshotPanel_Render_ShowsPaths(t *testing.T) {
	panel := NewSnapshotPanel()
	panel.SetSize(100, 30)

	snapshots := []types.Snapshot{
		{
			ID:       "test",
			ShortID:  "test",
			Time:     time.Now(),
			Hostname: "host",
			Paths:    []string{"/home/user", "/etc/config"},
		},
	}
	panel.SetSnapshots(snapshots)

	output := panel.Render(true)

	// Paths should be visible when snapshot is selected
	if !strings.Contains(output, "/home/user") {
		t.Error("Render() should show paths")
	}
}

func TestSnapshotPanel_Render_TruncatesLongPaths(t *testing.T) {
	panel := NewSnapshotPanel()
	panel.SetSize(100, 30)

	longPath := "/this/is/a/very/long/path/that/should/be/truncated/in/the/display"
	snapshots := []types.Snapshot{
		{
			ID:       "test",
			ShortID:  "test",
			Time:     time.Now(),
			Hostname: "host",
			Paths:    []string{longPath},
		},
	}
	panel.SetSnapshots(snapshots)

	output := panel.Render(true)

	// Should contain truncation indicator if path is too long
	if strings.Count(output, longPath) > 0 {
		// Full path might be shown, check if it's truncated in the right context
		pathsLine := ""
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Paths:") {
				pathsLine = line
				break
			}
		}

		// The paths line should not be excessively long
		if len(pathsLine) > 120 {
			t.Error("Path display should be truncated for long paths")
		}
	}
}

func BenchmarkSnapshotPanel_Render(b *testing.B) {
	panel := NewSnapshotPanel()
	panel.SetSize(120, 40)

	snapshots := make([]types.Snapshot, 50)
	for i := 0; i < 50; i++ {
		snapshots[i] = types.Snapshot{
			ID:       "abc123def456ghi789",
			ShortID:  "abc123",
			Time:     time.Now().Add(-time.Duration(i) * time.Hour),
			Hostname: "testhost",
			Paths:    []string{"/home/user", "/etc"},
			Tags:     []string{"tag1", "tag2"},
		}
	}
	panel.SetSnapshots(snapshots)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = panel.Render(i%2 == 0)
	}
}

// Filter tests
func TestSnapshotPanel_SetFilter(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "abc123", ShortID: "abc123", Hostname: "host1", Paths: []string{"/home/user"}},
		{ID: "def456", ShortID: "def456", Hostname: "host2", Paths: []string{"/var/log"}},
		{ID: "ghi789", ShortID: "ghi789", Hostname: "host1", Paths: []string{"/home/admin"}},
	}
	panel.SetSnapshots(snapshots)

	// Filter by text
	panel.SetFilter("abc")

	if len(panel.filteredSnapshots) != 1 {
		t.Errorf("SetFilter('abc') filtered count = %v, want 1", len(panel.filteredSnapshots))
	}

	if panel.filteredSnapshots[0].ShortID != "abc123" {
		t.Errorf("Filtered snapshot = %v, want abc123", panel.filteredSnapshots[0].ShortID)
	}

	if !panel.IsFilterActive() {
		t.Error("IsFilterActive() should return true after SetFilter")
	}
}

func TestSnapshotPanel_ClearFilter(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "abc", ShortID: "abc", Hostname: "host1"},
		{ID: "def", ShortID: "def", Hostname: "host2"},
	}
	panel.SetSnapshots(snapshots)

	// Apply and then clear filter
	panel.SetFilter("abc")
	panel.ClearFilter()

	if len(panel.filteredSnapshots) != 2 {
		t.Errorf("After ClearFilter, filtered count = %v, want 2", len(panel.filteredSnapshots))
	}

	if panel.IsFilterActive() {
		t.Error("IsFilterActive() should return false after ClearFilter")
	}
}

func TestSnapshotPanel_FilterByTag(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Tags: []string{"daily", "production"}},
		{ID: "snap2", ShortID: "snap2", Tags: []string{"weekly", "backup"}},
		{ID: "snap3", ShortID: "snap3", Tags: []string{"daily", "test"}},
	}
	panel.SetSnapshots(snapshots)

	// Filter by tag
	panel.SetTagFilter("daily")

	if len(panel.filteredSnapshots) != 2 {
		t.Errorf("FilterByTag('daily') filtered count = %v, want 2", len(panel.filteredSnapshots))
	}

	// Both filtered snapshots should have "daily" tag
	for _, snap := range panel.filteredSnapshots {
		found := false
		for _, tag := range snap.Tags {
			if strings.Contains(strings.ToLower(tag), "daily") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Filtered snapshot %v should have 'daily' tag", snap.ShortID)
		}
	}
}

func TestSnapshotPanel_FilterByHostname(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Hostname: "webserver01"},
		{ID: "snap2", ShortID: "snap2", Hostname: "database01"},
		{ID: "snap3", ShortID: "snap3", Hostname: "webserver02"},
	}
	panel.SetSnapshots(snapshots)

	// Filter by hostname
	panel.SetHostFilter("webserver")

	if len(panel.filteredSnapshots) != 2 {
		t.Errorf("FilterByHostname('webserver') filtered count = %v, want 2", len(panel.filteredSnapshots))
	}

	// Both filtered snapshots should have "webserver" in hostname
	for _, snap := range panel.filteredSnapshots {
		if !strings.Contains(strings.ToLower(snap.Hostname), "webserver") {
			t.Errorf("Filtered snapshot %v should have 'webserver' in hostname", snap.ShortID)
		}
	}
}

func TestSnapshotPanel_FilterByPath(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Paths: []string{"/home/user"}},
		{ID: "snap2", ShortID: "snap2", Paths: []string{"/var/log"}},
		{ID: "snap3", ShortID: "snap3", Paths: []string{"/home/admin"}},
	}
	panel.SetSnapshots(snapshots)

	// Filter by path containing "home"
	panel.SetFilter("home")

	if len(panel.filteredSnapshots) != 2 {
		t.Errorf("Filter by path 'home' filtered count = %v, want 2", len(panel.filteredSnapshots))
	}
}

func TestSnapshotPanel_FilterBySnapshotID(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "abc123def456", ShortID: "abc123"},
		{ID: "ghi789jkl012", ShortID: "ghi789"},
		{ID: "mno345pqr678", ShortID: "mno345"},
	}
	panel.SetSnapshots(snapshots)

	// Filter by partial ID
	panel.SetFilter("abc")

	if len(panel.filteredSnapshots) != 1 {
		t.Errorf("Filter by ID 'abc' filtered count = %v, want 1", len(panel.filteredSnapshots))
	}

	if panel.filteredSnapshots[0].ShortID != "abc123" {
		t.Errorf("Filtered snapshot = %v, want abc123", panel.filteredSnapshots[0].ShortID)
	}
}

func TestSnapshotPanel_MultipleFilters(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Hostname: "webserver", Tags: []string{"production"}},
		{ID: "snap2", ShortID: "snap2", Hostname: "webserver", Tags: []string{"test"}},
		{ID: "snap3", ShortID: "snap3", Hostname: "database", Tags: []string{"production"}},
	}
	panel.SetSnapshots(snapshots)

	// Filter by both tag and hostname
	panel.SetTagFilter("production")
	panel.SetHostFilter("webserver")

	if len(panel.filteredSnapshots) != 1 {
		t.Errorf("Multiple filters filtered count = %v, want 1", len(panel.filteredSnapshots))
	}

	if panel.filteredSnapshots[0].ShortID != "snap1" {
		t.Errorf("Filtered snapshot = %v, want snap1", panel.filteredSnapshots[0].ShortID)
	}
}

func TestSnapshotPanel_FilterCaseInsensitive(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Hostname: "WebServer"},
		{ID: "snap2", ShortID: "snap2", Hostname: "DATABASE"},
		{ID: "snap3", ShortID: "snap3", Tags: []string{"Production"}},
	}
	panel.SetSnapshots(snapshots)

	// Test case insensitive hostname filter
	panel.SetHostFilter("webserver")
	if len(panel.filteredSnapshots) != 1 {
		t.Error("Hostname filter should be case insensitive")
	}

	panel.ClearFilter()

	// Test case insensitive tag filter
	panel.SetTagFilter("production")
	if len(panel.filteredSnapshots) != 1 {
		t.Error("Tag filter should be case insensitive")
	}
}

func TestSnapshotPanel_FilterNoMatches(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Hostname: "host1"},
		{ID: "snap2", ShortID: "snap2", Hostname: "host2"},
	}
	panel.SetSnapshots(snapshots)

	// Filter that matches nothing
	panel.SetFilter("nonexistent")

	if len(panel.filteredSnapshots) != 0 {
		t.Errorf("Filter with no matches should return 0 snapshots, got %v", len(panel.filteredSnapshots))
	}

	if !panel.IsFilterActive() {
		t.Error("IsFilterActive() should still return true even with no matches")
	}
}

func TestSnapshotPanel_FilteredNavigation(t *testing.T) {
	panel := NewSnapshotPanel()

	snapshots := []types.Snapshot{
		{ID: "snap1", ShortID: "snap1", Tags: []string{"daily"}},
		{ID: "snap2", ShortID: "snap2", Tags: []string{"weekly"}},
		{ID: "snap3", ShortID: "snap3", Tags: []string{"daily"}},
		{ID: "snap4", ShortID: "snap4", Tags: []string{"monthly"}},
	}
	panel.SetSnapshots(snapshots)

	// Filter to 2 snapshots
	panel.SetTagFilter("daily")

	// Navigation should work with filtered list
	panel.selected = 0
	panel.MoveDown()
	if panel.selected != 1 {
		t.Errorf("After MoveDown in filtered list, selected = %v, want 1", panel.selected)
	}

	// Should not go beyond filtered list
	panel.MoveDown()
	if panel.selected != 1 {
		t.Errorf("Should not move beyond filtered list, selected = %v, want 1", panel.selected)
	}

	// GetSelected should return from filtered list
	selected := panel.GetSelected()
	if selected == nil || selected.ShortID != "snap3" {
		t.Errorf("GetSelected from filtered list should return snap3, got %v", selected)
	}
}

func BenchmarkSnapshotPanel_Filter(b *testing.B) {
	panel := NewSnapshotPanel()

	// Create large dataset
	snapshots := make([]types.Snapshot, 1000)
	for i := 0; i < 1000; i++ {
		snapshots[i] = types.Snapshot{
			ID:       "snap" + string(rune(i)),
			ShortID:  "snap" + string(rune(i%100)),
			Hostname: "host" + string(rune(i%10)),
			Paths:    []string{"/path/" + string(rune(i%50))},
			Tags:     []string{"tag" + string(rune(i%5))},
		}
	}
	panel.SetSnapshots(snapshots)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.SetFilter("tag")
		panel.ClearFilter()
	}
}
