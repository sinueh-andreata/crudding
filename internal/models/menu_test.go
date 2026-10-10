package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMenuNavigationSelectionAndConfirmation(t *testing.T) {
	current := menu{title: "test", options: []string{"python", "go", "node"}, selected: -1}

	updated, _ := current.Update(tea.KeyMsg{Type: tea.KeyDown})
	current = updated.(menu)
	if current.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", current.cursor)
	}
	updated, _ = current.Update(tea.KeyMsg{Type: tea.KeySpace})
	current = updated.(menu)
	updated, command := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	current = updated.(menu)
	if !current.confirmed || command == nil || current.selected != 1 {
		t.Fatal("menu should select and confirm the focused option")
	}
}