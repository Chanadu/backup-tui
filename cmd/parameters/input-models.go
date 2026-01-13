package parameters

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type InputData struct {
	User     string
	Server   string
	Password string
	Debug    bool
	Commands bool
	Progress bool
}
type InputDataMessage struct {
	Data InputData
}

func (m InputModel) ParametersDoneCmd() tea.Msg {
	data := InputData{}
	for _, textModel := range m.TextInputs {
		name := textModel.Name
		val := textModel.Ti.Value()
		switch name {
		case "user":
			data.User = val
		case "server":
			data.Server = val
		case "password":
			data.Password = val
		}
	}
	for _, switchModel := range m.SwitchInputs {
		val := switchModel.enabled
		switch switchModel.name {
		case "debug":
			data.Debug = val
		case "commands":
			data.Commands = val
		case "progress":
			data.Progress = val
		}
	}

	return InputDataMessage{Data: data}
}

type InputModel struct {
	TextInputs   []TextModel
	SwitchInputs []SwitchModel
	currentIndex int
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strMsg := msg.String(); strMsg {
		case "tab", "shift+tab", "up", "down", "ctrl+j", "ctrl+k", "enter":

			if strMsg == "enter" && m.currentIndex == m.totalItemCount()-1 {
				isDone := true
				for i := range m.TextInputs {
					if m.TextInputs[i].Ti.Value() == "" {
						isDone = false
						break
					}
				}

				if isDone {
					return m, m.ParametersDoneCmd
				}
			}
			m.blurCurrentIndex()

			if strMsg == "up" || strMsg == "ctrl+k" || strMsg == "shift+tab" {
				m.currentIndex--
			} else {
				m.currentIndex++
			}

			m.currentIndex = wrap(m.currentIndex, m.totalItemCount())

			m.focusCurrentIndex()
		}
	}

	var cmd tea.Cmd
	for i := range m.TextInputs {
		m.TextInputs[i], cmd = m.TextInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	for i := range m.SwitchInputs {
		m.SwitchInputs[i], cmd = m.SwitchInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m InputModel) View() string {
	var s strings.Builder

	for i := range m.totalItemCount() {
		if i == m.currentIndex {
			s.WriteString("> ")
		} else {
			s.WriteString("  ")
		}

		if m.textInputSelected(i) {
			s.WriteString(m.TextInputs[m.textIndex(i)].View())
		} else if m.switchInputSelected(i) {
			s.WriteString(m.SwitchInputs[m.switchIndex(i)].View())
		}
		s.WriteString("\n")
	}

	s.WriteString("Press tab to switch, enter to submit.\n")

	return s.String()
}

func InitialParametersInputs() InputModel {
	textInputs := []TextModel{}
	textInputs = append(textInputs, InitalTextModel("user", "User: ", "ex: pi", false))
	textInputs = append(textInputs, InitalTextModel("server", "Server: ", "ex: 192.168.1.1 or raspberrypi", false))
	textInputs = append(textInputs, InitalTextModel("password", "Password: ", "ex: 1234", true))

	switchInputs := []SwitchModel{}
	switchInputs = append(switchInputs, InitialSwitchModel("debug", "Debug", false))
	switchInputs = append(switchInputs, InitialSwitchModel("commands", "Print Commands", true))
	switchInputs = append(switchInputs, InitialSwitchModel("progress", "Show Progress", true))

	textInputs[0].Ti.Focus()

	return InputModel{
		TextInputs:   textInputs,
		SwitchInputs: switchInputs,
	}
}
