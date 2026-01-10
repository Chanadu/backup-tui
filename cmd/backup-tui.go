package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/backup"
	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/stage"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	stage        stage.Stage
	paramsInputs parameters.InputModel
	paramsData   parameters.InputData

	checkServerModel backup.CheckServerModel
}

// Paramters -> check server, create backups, upload to remote server, delete local backups

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
	case parameters.InputDataMessage:
		m.stage++
		m.paramsData = msg.Data
		m.checkServerModel = backup.InitialCheckServerModel(m.paramsData)
	}

	var cmd tea.Cmd
	switch m.stage {
	case stage.Input:
		m.paramsInputs, cmd = m.paramsInputs.Update(msg)
	case stage.Check:
		m.checkServerModel, cmd = m.checkServerModel.Update(msg)
	case stage.Create:
	case stage.Delete:
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(m.paramsInputs.View())
	s.WriteString("Press Ctrl+C to quit.\n")

	return s.String()
}

func initialModel() model {
	return model{
		stage:        stage.Input,
		paramsInputs: parameters.InitialParametersInputs(),
	}
}

func Start() {
	fmt.Println("BackupTui")

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		_ = fmt.Errorf("error: %v", err)
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	// fmt.Printf("User: %s", m.user)
}
