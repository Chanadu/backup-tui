package parameters

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type InputDoneMsg struct{}

func ParametersDoneCmd() tea.Msg {
	return InputDoneMsg{}
}

type InputModel struct {
	textInputs   []TextModel
	switchInputs []SwitchModel
	focusIndex   int
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) totalItemCount() int {
	return len(m.textInputs) + len(m.switchInputs)
}

func (m InputModel) textInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index < len(m.textInputs)
}

func (m InputModel) switchInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index >= len(m.textInputs)
}

func (m InputModel) textIndex(indexes ...int) int {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index
}

func (m InputModel) switchIndex(indexes ...int) int {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index - len(m.textInputs)
}

func wrap(x, n int) int {
	return ((x % n) + n) % n
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strMsg := msg.String(); strMsg {
		case "tab", "shift+tab", "up", "down", "ctrl+j", "ctrl+k", "enter":

			if strMsg == "enter" && m.focusIndex == m.totalItemCount()-1 {
				isDone := true
				for i := range m.textInputs {
					if m.textInputs[i].ti.Value() == "" {
						isDone = false
						break
					}
				}

				if isDone {
					return m, ParametersDoneCmd
				}
			}
			if m.textInputSelected() {
				m.textInputs[m.textIndex()].ti.Blur()
			} else if m.switchInputSelected() {
				m.switchInputs[m.switchIndex()].Blur()
			}

			if strMsg == "up" || strMsg == "ctrl+k" || strMsg == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			m.focusIndex = wrap(m.focusIndex, m.totalItemCount())

			if m.textInputSelected() {
				m.textInputs[m.textIndex()].ti.Focus()
			} else if m.switchInputSelected() {
				m.switchInputs[m.switchIndex()].Focus()
			}
		}
	}

	var cmd tea.Cmd
	for i := range m.textInputs {
		m.textInputs[i], cmd = m.textInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	for i := range m.switchInputs {
		m.switchInputs[i], cmd = m.switchInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m InputModel) View() string {
	var s strings.Builder

	for i := range m.totalItemCount() {
		if i == m.focusIndex {
			s.WriteString("> ")
		} else {
			s.WriteString("  ")
		}

		if m.textInputSelected(i) {
			s.WriteString(m.textInputs[m.textIndex(i)].View())
		} else if m.switchInputSelected(i) {
			s.WriteString(m.switchInputs[m.switchIndex(i)].View())
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

	textInputs[0].ti.Focus()

	return InputModel{
		textInputs:   textInputs,
		switchInputs: switchInputs,
	}
}
