package getfiles

import (
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

// Message sent when files are selected
type FilesSelectedMsg struct {
	Paths []string
}

// Message sent when file selection is cancelled
type FilesSelectionCancelledMsg struct{}

func FilesSelectionCancelledCmd() tea.Msg {
	return FilesSelectionCancelledMsg{}
}

type FileSelectorModel struct {
	Picker        filepicker.Model
	SelectedPaths []string
	Prompting     bool // true if prompting user to pick another file
	Done          bool
}

func InitialFileSelectorModel() FileSelectorModel {
	fp := filepicker.New()
	fp.AllowedTypes = nil // allow all files
	fp.ShowHidden = true
	return FileSelectorModel{
		Picker: fp,
	}
}

func (m FileSelectorModel) Init() tea.Cmd {
	return m.Picker.Init()
}

func (m FileSelectorModel) Update(msg tea.Msg) (FileSelectorModel, tea.Cmd) {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		strMsg := msg.String()
		if strMsg == "q" {
			m.Done = true
			return m, FilesSelectionCancelledCmd
		}
		if !m.Prompting {
			break
		}

		switch strMsg {
		case "y":
			m.Picker = filepicker.New()
			m.Picker.AllowedTypes = nil
			m.Picker.ShowHidden = true
			m.Prompting = false

			return m, m.Picker.Init()
		case "n":
			m.Done = true
			return m, func() tea.Msg {
				return FilesSelectedMsg{Paths: m.SelectedPaths}
			}
		}
	}

	m.Picker, cmd = m.Picker.Update(msg)
	cmds = append(cmds, cmd)

	if didSelect, path := m.Picker.DidSelectFile(msg); didSelect {
		m.SelectedPaths = append(m.SelectedPaths, path)
		m.Prompting = true
	}

	return m, tea.Batch(cmds...)
}

func (m FileSelectorModel) View() string {
	var s strings.Builder

	if m.Done {
		s.WriteString("File selection complete.")
		return s.String()
	}
	if m.Prompting {
		s.WriteString("Pick another file? (y/n)\nSelected so far:\n" + strings.Join(m.SelectedPaths, "\n"))
		return s.String()
	}

	s.WriteString(m.Picker.View())
	return s.String()
}
