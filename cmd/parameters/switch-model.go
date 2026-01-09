package parameters

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SwitchModel struct {
	name    string
	prompt  string
	enabled bool
	focused bool
}

func (m SwitchModel) Init() tea.Cmd {
	return nil
}

func (m SwitchModel) Update(msg tea.Msg) (SwitchModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	log.Printf("Focused: %s\n", m.String())

	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strMsg := msg.String(); strMsg {
		case " ":
			m.enabled = !m.enabled
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SwitchModel) View() string {
	var s strings.Builder
	if m.enabled {
		s.WriteString("[x]")
	} else {
		s.WriteString("[ ]")
	}
	s.WriteString(m.prompt)

	return s.String()
}

func InitialSwitchModel(name string, prompt string, enabled bool) SwitchModel {
	return SwitchModel{
		name:    name,
		prompt:  prompt,
		enabled: enabled,
		focused: false,
	}
}

func (m *SwitchModel) Focus() {
	m.focused = true
}

func (m *SwitchModel) Blur() {
	m.focused = false
}

func (m *SwitchModel) String() string {
	return fmt.Sprintf("{name: %s, enabled: %t, focused: %t}", m.name, m.enabled, m.focused)
}
