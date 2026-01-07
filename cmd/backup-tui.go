package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input textinput.Model
	done  bool
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		}

	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	// Send the UI for rendering
	// fmt.cprintf(&s, "User: %s", m.user)

	if m.done {
		return fmt.Sprintf(
			"You typed: %s\n",
			m.input.Value(),
		)
	}

	fmt.Fprintf(&s, "Enter text:\n%s\n(Press Enter to submit)", m.input.View())

	return s.String()
}

func initialModel() model {
	ti := textinput.New()

	ti.Placeholder = "Enter Server User"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	m := model{
		input: ti,
	}

	return m
}

func Start() {
	fmt.Println("BackupTui")

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		_ = fmt.Errorf("error: %v", err)
		os.Exit(1)
	}

	// fmt.Printf("User: %s", m.user)
}
