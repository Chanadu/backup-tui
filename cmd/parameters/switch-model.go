package parameters

import (
	"fmt"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/styles"
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
	// log.Printf("Focused: %s\n", m.String())

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
	check := "[ ]"
	if m.enabled {
		check = "[x]"
	}

	if m.focused {
		s.WriteString(styles.PrimaryStyle.Render(check))
	} else if m.enabled {
		s.WriteString(styles.SuccessStyle.Render(check))
	} else {
		s.WriteString(styles.MutedStyle.Render(check))
	}
	s.WriteString(" ")
	if m.focused {
		s.WriteString(styles.SecondaryStyle.Render(m.prompt))
	} else {
		s.WriteString(styles.InfoStyle.Render(m.prompt))
	}

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
