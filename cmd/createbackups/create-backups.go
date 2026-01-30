package createbackups

import (
	"log"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
)

type CreateBackupsMessage struct {
	Ok  bool
	Err error
}

type CreateBackupsModel struct {
	data    parameters.InputData
	done    bool
	success bool
	err     error
}

func (m *CreateBackupsModel) createBackups() tea.Msg {
	log.Println("creating Backups")

	return CreateBackupsMessage{
		Ok:  true,
		Err: nil,
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	return m.createBackups
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {

	case CreateBackupsMessage:
		m.done = true
		m.success = msg.Ok
		m.err = msg.Err
	}

	return m, tea.Batch(cmds...)
}

func (m CreateBackupsModel) View() string {
	return "Creating Backups not yet implemented"
}

func InitialCreateBackupsModel(data parameters.InputData, paths []string) CreateBackupsModel {
	return CreateBackupsModel{
		data: data,
		done: false,
	}
}
