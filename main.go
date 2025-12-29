package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/craigderington/lazyrestic/pkg/model"
)

const version = "0.1.0"

func main() {
	// Create the initial model
	m := model.NewModel()

	// Create the Bubbletea program with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
