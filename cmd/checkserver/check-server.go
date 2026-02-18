package checkserver

import (
	"fmt"
	"log"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/styles"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"
)

type CheckServerMessage struct {
	Ok  bool
	Err error
}

type TryAgainMessage struct{}

func TryAgainCmd() tea.Msg {
	return TryAgainMessage{}
}

type CheckServerModel struct {
	data     parameters.InputData
	done     bool
	success  bool
	err      error
	attempts int
}

func (m *CheckServerModel) checkServer() tea.Msg {
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
		log.Printf("Connection failed")
		log.Printf("error: %v", err)

		return CheckServerMessage{
			Ok:  false,
			Err: fmt.Errorf("connecting to server: %v", err),
		}
	}
	err = client.Close()
	if err != nil {
		log.Printf("error closing connection: %v", err)
	}

	log.Printf("Connection success")
	return CheckServerMessage{
		Ok:  true,
		Err: nil,
	}
}

func (m CheckServerModel) Init() tea.Cmd {
	return m.checkServer
}

func (m CheckServerModel) Update(msg tea.Msg) (CheckServerModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {

	case CheckServerMessage:
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
			m.attempts += 1
			return m, m.checkServer
		}
	}

	return m, tea.Batch(cmds...)
}

func (m CheckServerModel) View() string {
	// log.Printf("done: %t, success: %t, attempts: %d", m.done, m.success, m.attempts)
	log.Printf("Checking Server Model: done: %t, success: %t, attempts: %d", m.done, m.success, m.attempts)

	var s strings.Builder
	if !m.done {
		s.WriteString(styles.InfoStyle.Render("Checking Server..."))
	} else if !m.success {
		s.WriteString(styles.ErrorStyle.Render("✗ Server Connection Failed."))
		if m.attempts > 1 {
			s.WriteString(styles.MutedStyle.Render(fmt.Sprintf(" (%d)", m.attempts)))
		}
		s.WriteString("\n\n")
		s.WriteString(styles.MutedStyle.Render("Error: "))
		s.WriteString(styles.ErrorStyle.Render(m.err.Error()))
		s.WriteString("\n\n")

		s.WriteString(styles.HelpStyle.Render("Press Enter to change server details."))
		s.WriteString("\n")
		s.WriteString(styles.HelpStyle.Render("Press R to retry."))
	} else {
		s.WriteString(styles.SuccessStyle.Render("✓ Server Connected"))
	}

	s.WriteString("\n")

	return s.String()
}

func InitialCheckServerModel(data parameters.InputData) CheckServerModel {
	return CheckServerModel{
		data:     data,
		done:     false,
		attempts: 1,
	}
}
