package getfiles

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/getfiles/filepicker"
	"github.com/Chanadu/backup-tui/cmd/utils"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Message sent when files are selected
type FilesSelectedMsg struct {
	Paths []string
}

type FileSelectorModel struct {
	Picker           filepicker.Model
	SelectedPaths    []string
	Prompting        bool
	Done             bool
	Search           textinput.Model
	Dir              string
	UsingFilteredDir bool

	tempDir string
}

func initialFilePicker(dir string) filepicker.Model {
	fp := filepicker.New()
	fp.AllowedTypes = nil
	fp.ShowHidden = true
	fp.AllowedTypes = nil
	fp.DirAllowed = true
	fp.FileAllowed = true
	fp.ShowPermissions = false
	fp.ShowSize = true
	fp.SetHeight(10)

	fp.KeyMap.Up.SetKeys("up", "ctrl+k")
	fp.KeyMap.Down.SetKeys("down", "ctrl+j")
	fp.KeyMap.Back.SetKeys("left", "ctrl+h")
	// fp.KeyMap.Open.SetKeys("right", "ctrl+l")
	// fp.KeyMap.Select.SetKeys("enter")

	var err error

	if dir != "" {
		fp.CurrentDirectory = filepath.Clean(dir)
		return fp
	}

	fp.CurrentDirectory, err = os.UserHomeDir()
	if err != nil {
		fp.CurrentDirectory = "/"
		log.Printf("Failed to get user home directory, error: %e", err)
	}

	return fp
}

func InitialFilesSelectorModel(paths []string, tempDir string) FileSelectorModel {
	search := textinput.New()
	search.Placeholder = "Search files..."
	search.CharLimit = 64
	search.Width = 30
	search.Focus()

	fp := initialFilePicker("")

	return FileSelectorModel{
		Picker:           fp,
		SelectedPaths:    paths,
		Search:           search,
		Dir:              fp.CurrentDirectory,
		UsingFilteredDir: false,
		tempDir:          tempDir,
	}
}

func (m FileSelectorModel) Init() tea.Cmd {
	return m.Picker.Init()
}

func (m FileSelectorModel) createFilteredSymlinkDir(srcDir string, query string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	lowerQuery := strings.ToLower(query)
	for _, file := range files {
		name := strings.ToLower(file.Name())
		ext := strings.ToLower(filepath.Ext(name))
		if lowerQuery == "" || strings.Contains(name, lowerQuery) ||
			strings.Contains(ext, lowerQuery) {

			srcPath := filepath.Join(srcDir, name)
			dstPath := filepath.Join(m.tempDir, name)

			err := os.Symlink(srcPath, dstPath)

			if err != nil {
				log.Printf("error creating symlink: %v", err)
				continue
			}
		}
	}
	return nil
}

func (m FileSelectorModel) handleSearch(msg tea.Msg) (FileSelectorModel, tea.Cmd) {

	oldSearch := m.Search.Value()

	var cmd tea.Cmd
	m.Search, cmd = m.Search.Update(msg)
	newSearch := m.Search.Value()

	if oldSearch == newSearch {
		return m, cmd
	}

	utils.ClearDir(m.tempDir)

	err := m.createFilteredSymlinkDir(m.Dir, newSearch)

	if err != nil {
		log.Printf("Couldn't create filteredDir, error: %v", err)
		return m, cmd
	}

	log.Printf("Switching to filtered dir: %s", m.tempDir)
	m.Picker.SetCurrentDirectory(m.tempDir)
	return m, tea.Batch(m.Picker.Init(), cmd)
}

func (m FileSelectorModel) Update(msg tea.Msg) (FileSelectorModel, tea.Cmd) {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd

	if m.Done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		strMsg := msg.String()

		if !m.Prompting && strMsg != " " {
			m, cmd = m.handleSearch(msg)
			cmds = append(cmds, cmd)
			break
		}

		switch strMsg {
		case "y":
			m = InitialFilesSelectorModel(m.SelectedPaths, m.tempDir)
			return m, m.Picker.Init()
		case "n":
			m.Done = true
			utils.ClearDir(m.tempDir)
			return m, func() tea.Msg {
				return FilesSelectedMsg{Paths: m.SelectedPaths}
			}
		}
	}

	if m.Prompting {
		return m, tea.Batch(cmds...)
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

	if !m.Prompting {
		s.WriteString("Search: ")
		s.WriteString(m.Search.View())
		s.WriteString("\n")
	}

	s.WriteString("Selected:\n")

	renders := []string{}
	for _, path := range m.SelectedPaths {
		renders = append(renders, m.Picker.Styles.Selected.Render(path))
	}
	s.WriteString(strings.Join(renders, "\n"))

	if len(renders) != 0 {
		s.WriteString("\n")
	}
	s.WriteString("\n")
	if m.UsingFilteredDir {
		s.WriteString("(Filtered View)\n")

	}

	if m.Done {
		s.WriteString("File selection complete.")
	} else if m.Prompting {
		s.WriteString("Pick another file? (y/n)\n")
	} else {
		s.WriteString(m.Picker.View())
	}

	return s.String()
}
