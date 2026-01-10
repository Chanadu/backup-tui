package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
)

type CheckServerMessage struct {
	ok  bool
	err error
}

func CheckServerCmd(ok bool, err error) tea.Cmd {
	return func() tea.Msg {
		return CheckServerMessage{
			ok:  ok,
			err: err,
		}
	}
}

type CheckServerModel struct {
	data parameters.InputData
}

func (m CheckServerModel) checkServer() tea.Msg {

	if err := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf("sshpass -p %s ssh %s StrictHostKeyChecking=no 'echo 1; exit'", m.data.Password, m.data.Server),
	).Run(); err != nil {
		return CheckServerMessage{
			ok:  false,
			err: fmt.Errorf("error checking server: %v", err),
		}
	}

	return CheckServerMessage{
		ok:  true,
		err: nil,
	}
}

func (m CheckServerModel) Init() tea.Cmd {
	return m.checkServer
}

func (m CheckServerModel) Update(msg tea.Msg) (CheckServerModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	//
	// 	// handle key messages
	// }

	return m, tea.Batch(cmds...)
}

func (m CheckServerModel) View() string {
	var s strings.Builder
	s.WriteString("Checking Server")
	return s.String()
}

func InitialCheckServerModel(data parameters.InputData) CheckServerModel {
	return CheckServerModel{
		data: data,
	}
}
