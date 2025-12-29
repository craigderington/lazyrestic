package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/craigderington/lazyrestic/pkg/types"
)

func TestNewRepositoryPanel(t *testing.T) {
	panel := NewRepositoryPanel()

	if panel == nil {
		t.Fatal("NewRepositoryPanel() returned nil")
	}

	if panel.selected != 0 {
		t.Errorf("Initial selected = %v, want 0", panel.selected)
	}

	if len(panel.repositories) != 0 {
		t.Errorf("Initial repositories length = %v, want 0", len(panel.repositories))
	}
}

func TestRepositoryPanel_SetRepositories(t *testing.T) {
	panel := NewRepositoryPanel()

	repos := []types.Repository{
		{Name: "repo1", Path: "/tmp/repo1", Status: "healthy"},
		{Name: "repo2", Path: "/tmp/repo2", Status: "error"},
		{Name: "repo3", Path: "/tmp/repo3", Status: "unknown"},
	}

	panel.SetRepositories(repos)

	if len(panel.repositories) != 3 {
		t.Errorf("Repositories length = %v, want 3", len(panel.repositories))
	}

	if panel.repositories[0].Name != "repo1" {
		t.Errorf("First repo name = %v, want repo1", panel.repositories[0].Name)
	}
}

func TestRepositoryPanel_SetRepositories_ResetsSelection(t *testing.T) {
	panel := NewRepositoryPanel()

	// Set initial repos and select last one
	repos1 := []types.Repository{
		{Name: "repo1", Path: "/tmp/1", Status: "healthy"},
		{Name: "repo2", Path: "/tmp/2", Status: "healthy"},
		{Name: "repo3", Path: "/tmp/3", Status: "healthy"},
	}
	panel.SetRepositories(repos1)
	panel.selected = 2

	// Set fewer repos - selection should be adjusted
	repos2 := []types.Repository{
		{Name: "repo1", Path: "/tmp/1", Status: "healthy"},
	}
	panel.SetRepositories(repos2)

	if panel.selected >= len(repos2) {
		t.Errorf("Selection not adjusted: selected = %v, len = %v", panel.selected, len(repos2))
	}
}

func TestRepositoryPanel_MoveDown(t *testing.T) {
	panel := NewRepositoryPanel()
	repos := []types.Repository{
		{Name: "repo1", Path: "/tmp/1", Status: "healthy"},
		{Name: "repo2", Path: "/tmp/2", Status: "healthy"},
		{Name: "repo3", Path: "/tmp/3", Status: "healthy"},
	}
	panel.SetRepositories(repos)

	// Initial selection should be 0
	if panel.selected != 0 {
		t.Fatalf("Initial selected = %v, want 0", panel.selected)
	}

	// Move down
	panel.MoveDown()
	if panel.selected != 1 {
		t.Errorf("After MoveDown, selected = %v, want 1", panel.selected)
	}

	// Move down again
	panel.MoveDown()
	if panel.selected != 2 {
		t.Errorf("After second MoveDown, selected = %v, want 2", panel.selected)
	}

	// Try to move past the end
	panel.MoveDown()
	if panel.selected != 2 {
		t.Errorf("After moving past end, selected = %v, want 2", panel.selected)
	}
}

func TestRepositoryPanel_MoveUp(t *testing.T) {
	panel := NewRepositoryPanel()
	repos := []types.Repository{
		{Name: "repo1", Path: "/tmp/1", Status: "healthy"},
		{Name: "repo2", Path: "/tmp/2", Status: "healthy"},
		{Name: "repo3", Path: "/tmp/3", Status: "healthy"},
	}
	panel.SetRepositories(repos)
	panel.selected = 2

	// Move up
	panel.MoveUp()
	if panel.selected != 1 {
		t.Errorf("After MoveUp, selected = %v, want 1", panel.selected)
	}

	// Move up again
	panel.MoveUp()
	if panel.selected != 0 {
		t.Errorf("After second MoveUp, selected = %v, want 0", panel.selected)
	}

	// Try to move before the beginning
	panel.MoveUp()
	if panel.selected != 0 {
		t.Errorf("After moving before beginning, selected = %v, want 0", panel.selected)
	}
}

func TestRepositoryPanel_GetSelected(t *testing.T) {
	panel := NewRepositoryPanel()

	// No repos - should return nil
	if panel.GetSelected() != nil {
		t.Error("GetSelected() should return nil when no repositories")
	}

	repos := []types.Repository{
		{Name: "repo1", Path: "/tmp/1", Status: "healthy"},
		{Name: "repo2", Path: "/tmp/2", Status: "error"},
	}
	panel.SetRepositories(repos)

	// Get first repo
	selected := panel.GetSelected()
	if selected == nil {
		t.Fatal("GetSelected() returned nil")
	}
	if selected.Name != "repo1" {
		t.Errorf("Selected repo name = %v, want repo1", selected.Name)
	}

	// Move to second repo
	panel.MoveDown()
	selected = panel.GetSelected()
	if selected == nil {
		t.Fatal("GetSelected() returned nil")
	}
	if selected.Name != "repo2" {
		t.Errorf("Selected repo name = %v, want repo2", selected.Name)
	}
}

func TestRepositoryPanel_SetSize(t *testing.T) {
	panel := NewRepositoryPanel()
	panel.SetSize(80, 24)

	if panel.width != 80 {
		t.Errorf("Width = %v, want 80", panel.width)
	}
	if panel.height != 24 {
		t.Errorf("Height = %v, want 24", panel.height)
	}
}

func TestRepositoryPanel_Render_Empty(t *testing.T) {
	panel := NewRepositoryPanel()
	panel.SetSize(80, 24)

	output := panel.Render(false)

	if output == "" {
		t.Error("Render() should not return empty string")
	}

	// Should contain "No repositories" message
	if !strings.Contains(output, "No repositories") {
		t.Error("Empty panel should show 'No repositories' message")
	}
}

func TestRepositoryPanel_Render_WithRepos(t *testing.T) {
	panel := NewRepositoryPanel()
	panel.SetSize(80, 24)

	repos := []types.Repository{
		{Name: "test-repo", Path: "/tmp/test", Status: "healthy"},
		{Name: "backup-repo", Path: "/backup", Status: "error"},
	}
	panel.SetRepositories(repos)

	output := panel.Render(false)

	// Should contain repository names
	if !strings.Contains(output, "test-repo") {
		t.Error("Render() should contain 'test-repo'")
	}
	if !strings.Contains(output, "backup-repo") {
		t.Error("Render() should contain 'backup-repo'")
	}

	// Should contain paths (status moved to metrics panel)
	if !strings.Contains(output, "/tmp/test") {
		t.Error("Render() should contain repository path /tmp/test")
	}
	if !strings.Contains(output, "/backup") {
		t.Error("Render() should contain repository path /backup")
	}
}

func TestRepositoryPanel_Render_ActiveState(t *testing.T) {
	panel := NewRepositoryPanel()
	panel.SetSize(80, 24)

	repos := []types.Repository{
		{Name: "repo", Path: "/tmp", Status: "healthy"},
	}
	panel.SetRepositories(repos)

	// Render inactive
	inactiveOutput := panel.Render(false)

	// Render active
	activeOutput := panel.Render(true)

	// Outputs should be different (different styling)
	if inactiveOutput == activeOutput {
		t.Error("Active and inactive renders should differ")
	}

	// Active should contain selection indicator
	if !strings.Contains(activeOutput, "▶") && !strings.Contains(activeOutput, ">") {
		t.Error("Active render should show selection indicator")
	}
}

func TestRepositoryPanel_Render_ShowsPath(t *testing.T) {
	panel := NewRepositoryPanel()
	panel.SetSize(80, 24)

	repos := []types.Repository{
		{Name: "my-repo", Path: "/very/specific/path", Status: "healthy"},
	}
	panel.SetRepositories(repos)

	output := panel.Render(true) // Active to show path

	// Should show path for selected item
	if !strings.Contains(output, "/very/specific/path") {
		t.Error("Render() should show path for selected repository")
	}
}

func BenchmarkRepositoryPanel_Render(b *testing.B) {
	panel := NewRepositoryPanel()
	panel.SetSize(120, 40)

	repos := make([]types.Repository, 10)
	for i := 0; i < 10; i++ {
		repos[i] = types.Repository{
			Name:       "repo",
			Path:       "/tmp/repo",
			Status:     "healthy",
			Size:       1024 * 1024,
			LastBackup: time.Now(),
		}
	}
	panel.SetRepositories(repos)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = panel.Render(i%2 == 0)
	}
}
