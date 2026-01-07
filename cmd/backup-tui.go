package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	// choices  []string
	// cursor   int
	// selected map[int]struct{}
	user   string
	server string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.user) > 0 {
				m.user = m.user[:len(m.user)-1]
			}

		case "enter", " ":
			fmt.Printf("Entered User: %s", m.user)
			return m, tea.Quit
		default:
			m.user += msg.String()
		}
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// Send the UI for rendering
	fmt.Fprintf(&s, "User: %s", m.user)
	return s.String()
}

func Start() {
	fmt.Println("BackupTui")

	m := model{
		user:   "",
		server: "",
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		_ = fmt.Errorf("error: %v", err)
		os.Exit(1)
	}

	fmt.Printf("User: %s", m.user)
}
