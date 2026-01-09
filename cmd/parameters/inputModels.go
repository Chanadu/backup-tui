package parameters

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ParametersDoneMsg struct{}

func ParametersDoneCmd() tea.Msg {
	return ParametersDoneMsg{}
}

type ParametersModel struct {
	inputs     []InputModel
	focusIndex int
}

func (m ParametersModel) Init() tea.Cmd {
	return nil
}

func (m ParametersModel) Update(msg tea.Msg) (ParametersModel, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strMsg := msg.String(); strMsg {
		case "tab", "shift+tab", "up", "down", "ctrl+j", "ctrl+k", "enter":
			if strMsg == "enter" && m.focusIndex == len(m.inputs)-1 {
				isDone := true
				for i := range m.inputs {
					if m.inputs[i].ti.Value() == "" {
						isDone = false
						break
					}
				}

				if isDone {
					return m, ParametersDoneCmd
				}
			}

			m.inputs[m.focusIndex].ti.Blur()
			if strMsg == "up" || strMsg == "ctrl+k" || strMsg == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			m.inputs[m.focusIndex].ti.Focus()
		}
	}

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)

		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ParametersModel) View() string {
	var s strings.Builder

	for i := range m.inputs {
		if i == m.focusIndex {
			s.WriteString("> ")
		} else {
			s.WriteString("  ")
		}
		s.WriteString(m.inputs[i].View())
		s.WriteString("\n")
	}
	s.WriteString("Press tab to switch, enter to submit.\n")

	return s.String()
}

func InitialParametersInputs() ParametersModel {
	// user, server, password
	inputs := []InputModel{}
	inputs = append(inputs, InitalInputModel("user", "User: ", "ex: pi", false))
	inputs = append(inputs, InitalInputModel("server", "Server: ", "ex: 192.168.1.1 or raspberrypi", false))
	inputs = append(inputs, InitalInputModel("password", "Password: ", "ex: 1234", true))

	inputs[0].ti.Focus()
	return ParametersModel{
		inputs: inputs,
	}
}
