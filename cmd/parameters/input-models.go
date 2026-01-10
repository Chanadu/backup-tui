package parameters

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		var err error
		switch name {
		case "user":
			data.User = val
		case "server":
			data.Server = val
		case "password":
			data.Password = val
		case "debug":
			data.Debug, err = strconv.ParseBool(val)
			if err != nil {
				_ = fmt.Errorf("error parsing debug value: %v", err)
				log.Printf("error parsing debug value: %v", err)
				os.Exit(1)
			}
		case "commands":
			data.Commands, err = strconv.ParseBool(val)
			if err != nil {
				_ = fmt.Errorf("error parsing commands value: %v", err)
				log.Printf("error parsing commands value: %v", err)
				os.Exit(1)
			}
		case "progress":
			data.Progress, err = strconv.ParseBool(val)
			if err != nil {
				_ = fmt.Errorf("error parsing progress value: %v", err)
				log.Printf("error parsing progress value: %v", err)
				os.Exit(1)
			}
		}
	}
	return InputDataMessage{Data: data}
}

type InputModel struct {
	TextInputs   []TextModel
	SwitchInputs []SwitchModel
	focusIndex   int
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) totalItemCount() int {
	return len(m.TextInputs) + len(m.SwitchInputs)
}

func (m InputModel) textInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index < len(m.TextInputs)
}

func (m InputModel) switchInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.focusIndex)
	}
	index := indexes[0]

	return index >= len(m.TextInputs)
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

	return index - len(m.TextInputs)
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
			if m.textInputSelected() {
				m.TextInputs[m.textIndex()].Ti.Blur()
			} else if m.switchInputSelected() {
				m.SwitchInputs[m.switchIndex()].Blur()
			}

			if strMsg == "up" || strMsg == "ctrl+k" || strMsg == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			m.focusIndex = wrap(m.focusIndex, m.totalItemCount())

			if m.textInputSelected() {
				m.TextInputs[m.textIndex()].Ti.Focus()
			} else if m.switchInputSelected() {
				m.SwitchInputs[m.switchIndex()].Focus()
			}
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
		if i == m.focusIndex {
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
