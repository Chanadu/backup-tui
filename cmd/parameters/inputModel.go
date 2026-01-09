package parameters

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModel struct {
	Name       string
	promptText string
	ti         textinput.Model
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
	s.WriteString(m.promptText)
	s.WriteString(m.ti.View())

	return s.String()
}

func InitalInputModel(name string, promptText string, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder

	return InputModel{
		Name:       name,
		promptText: promptText,
		ti:         ti,
	}
}
