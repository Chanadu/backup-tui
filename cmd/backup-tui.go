package cmd

import (
	"fmt"
	"log"
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
		log.Printf("Input Data Collected: %v, %s", m.paramsData, m.stage)
		return m, m.checkServerModel.Init()
	case backup.CheckServerMessage:
		if msg.Ok {
			log.Printf("Connection Succeeded")
		} else {
			log.Printf("Connection Failed")
			log.Printf("error: %v", msg.Err)
		}
	case backup.TryAgainMessage:
		m.paramsInputs.SetCurrentIndex(0)
		m.stage = stage.Input
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
	switch m.stage {
	case stage.Input:
		s.WriteString(m.paramsInputs.View())
	case stage.Check:
		s.WriteString(m.checkServerModel.View())
	case stage.Create:
	case stage.Delete:
	}

	s.WriteString("\nPress Ctrl+C to quit.")

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
	log.Println("=========================BACKUP-TUI=======================================")

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("error: %v", err)
	}

	// fmt.Printf("User: %s", m.user)
}
