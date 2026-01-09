package parameters

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModel struct {
	Name string
	ti   textinput.Model
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd

	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	var s strings.Builder
	s.WriteString(m.ti.View())

	return s.String()
}

func InitalInputModel(name string, prompt string, placeholder string, isPassword bool) InputModel {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.Placeholder = placeholder
	if isPassword {
		ti.EchoMode = textinput.EchoPassword
	}
	ti.CharLimit = 64
	ti.Width = 30
	ti.EchoCharacter = '•'

	return InputModel{
		Name: name,
		ti:   ti,
	}
}
