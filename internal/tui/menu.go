package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Menu struct {
	title     string
	options   []string
	selected  int
	cursor    int
	confirmed bool
}

func NewMenu(title string, options []string) Menu {
	return Menu{title: title, options: options, selected: -1}
}

func (m Menu) Cursor() int {
	return m.cursor
}

func (m Menu) Selected() int {
	return m.selected
}

func (m Menu) Confirmed() bool {
	return m.confirmed
}

func (m Menu) Init() tea.Cmd {
	return nil
}

func (m Menu) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case " ":
		m.selected = m.cursor
	case "enter":
		if m.selected >= 0 {
			m.confirmed = true
			return m, tea.Quit
		}
	case "q", "Q", "esc", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Menu) View() string {
	var view strings.Builder
	fmt.Fprintf(&view, "crudding: %s\n\n", m.title)
	for index, option := range m.options {
		cursor := " "
		if index == m.cursor {
			cursor = ">"
		}
		mark := " "
		if index == m.selected {
			mark = "x"
		}
		fmt.Fprintf(&view, "%s (%s) %s\n", cursor, mark, option)
	}
	view.WriteString("\nSetas: navegar | Espaço: selecionar | Enter: continuar | q: sair")
	return view.String()
}

func ChooseOption(title string, options []string) (string, bool, error) {
	initial := NewMenu(title, options)
	final, err := tea.NewProgram(initial).Run()
	if err != nil {
		return "", false, err
	}
	result, ok := final.(Menu)
	if !ok || !result.confirmed || result.selected < 0 || result.selected >= len(options) {
		return "", false, nil
	}
	return options[result.selected], true, nil
}
