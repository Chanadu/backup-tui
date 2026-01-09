package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	params parameters.InputModel
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		strMsg := msg.String()
		switch strMsg {
		case "ctrl+c":
			return m, tea.Quit
		}
	case parameters.InputDoneMsg:
		// Handle the done condition here
		// For now, just quit, or you could transition to another model/state
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.params, cmd = m.params.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(m.params.View())
	s.WriteString("Press Ctrl+C to quit.\n")

	return s.String()
}

func initialModel() model {
	return model{
		params: parameters.InitialParametersInputs(),
	}
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
