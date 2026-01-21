package backup

import (
	"fmt"
	"log"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"
)

type CreateBackupsMessage struct {
	Ok  bool
	Err error
}

type TryAgainMessage struct{}

func TryAgainCmd() tea.Msg {
	return TryAgainMessage{}
}

type CreateBackupsModel struct {
	data    parameters.InputData
	done    bool
	success bool
	err     error
}

func (m *CreateBackupsModel) checkServer() tea.Msg {
	log.Println("checking server")
	config := &ssh.ClientConfig{
		User: m.data.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(m.data.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * 1e9,
	}

	client, err := ssh.Dial("tcp", m.data.Server+":22", config)

	if err != nil {
		return CreateBackupsMessage{
			Ok:  false,
			Err: fmt.Errorf("connecting to server: %v", err),
		}
	}
	err = client.Close()
	if err != nil {
		log.Fatalf("error closing connection: %v", err)
	}

	return CreateBackupsMessage{
		Ok:  true,
		Err: nil,
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	return m.checkServer
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {

	case CreateBackupsMessage:
		m.done = true
		m.success = msg.Ok
		m.err = msg.Err
	case tea.KeyMsg:
		if !m.done || m.success {
			break
		}
		strMsg := msg.String()
		log.Printf("Got keypress, %s", msg.String())

		switch strMsg {
		case "enter":
			return m, TryAgainCmd
		case "R":
			m.done = false
			return m, m.checkServer
		}
	}

	return m, tea.Batch(cmds...)
}

func (m CreateBackupsModel) View() string {
	return "Creating Backups not yet implemented"
}

func InitialCheckServerModel(data parameters.InputData) CreateBackupsModel {
	return CreateBackupsModel{
		data: data,
		done: false,
	}
}
