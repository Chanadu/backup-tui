package parameters

import (
	"strings"

	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextModel struct {
	Name string
	Ti   textinput.Model
}

func (m TextModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextModel) Update(msg tea.Msg) (TextModel, tea.Cmd) {
	var cmd tea.Cmd

	m.Ti, cmd = m.Ti.Update(msg)
	return m, cmd
}

func (m TextModel) View() string {
	var s strings.Builder
	s.WriteString(m.Ti.View())

	return s.String()
}

func InitalTextModel(name string, prompt string, placeholder string, isPassword bool) TextModel {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.Placeholder = placeholder
	ti.PromptStyle = styles.SecondaryStyle
	ti.PlaceholderStyle = styles.MutedStyle
	ti.TextStyle = styles.InfoStyle
	ti.Cursor.Style = styles.PrimaryStyle
	if isPassword {
		ti.EchoMode = textinput.EchoPassword
	}
	ti.CharLimit = 64
	ti.Width = 30
	ti.EchoCharacter = '•'

	return TextModel{
		Name: name,
		Ti:   ti,
	}
}
