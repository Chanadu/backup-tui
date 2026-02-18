package getfiles

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Chanadu/backup-tui/cmd/getfiles/filepicker"
	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/Chanadu/backup-tui/cmd/utils"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Message sent when files are selected
type FilesSelectedMsg struct {
	Paths []string
}

type selectorKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Back   key.Binding
	Select key.Binding
	Remove key.Binding
	Yes    key.Binding
	No     key.Binding
}

type helpBindings struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h helpBindings) ShortHelp() []key.Binding {
	return h.short
}

func (h helpBindings) FullHelp() [][]key.Binding {
	return h.full
}

func (k selectorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Remove, k.No}
}

func (k selectorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Back, k.Select}, {k.Remove, k.Yes, k.No}}
}

func defaultSelectorKeyMap() selectorKeyMap {
	return selectorKeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "ctrl+k"), key.WithHelp("↑/ctrl+k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "ctrl+j"), key.WithHelp("↓/ctrl+j", "down")),
		Back:   key.NewBinding(key.WithKeys("left", "ctrl+h"), key.WithHelp("←/ctrl+h", "back")),
		Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Remove: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "remove last")),
		Yes:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "pick another")),
		No:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "done")),
	}
}

type FileSelectorModel struct {
	Picker           filepicker.Model
	SelectedPaths    []string
	Prompting        bool
	Done             bool
	Notice           string
	Search           textinput.Model
	Help             help.Model
	KeyMap           selectorKeyMap
	Dir              string
	UsingFilteredDir bool

	tempDir string
}

func isSearchEditKey(msg tea.KeyMsg) bool {
	str := msg.String()
	if str == "backspace" || str == "delete" || str == " " {
		return true
	}

	r := []rune(str)
	return len(r) == 1
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
	fp.SetHeight(5)
	fp.SetMaxHeight(5)

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
		Help:             help.New(),
		KeyMap:           defaultSelectorKeyMap(),
		Dir:              fp.CurrentDirectory,
		UsingFilteredDir: false,
		tempDir:          tempDir,
	}
}

func (m FileSelectorModel) helpBindings() helpBindings {
	if m.Prompting {
		short := []key.Binding{m.KeyMap.Yes, m.KeyMap.No, m.KeyMap.Remove}
		return helpBindings{short: short, full: [][]key.Binding{short}}
	}
	short := []key.Binding{m.KeyMap.Up, m.KeyMap.Down, m.KeyMap.Select, m.KeyMap.Remove, m.KeyMap.No}
	full := [][]key.Binding{{m.KeyMap.Up, m.KeyMap.Down, m.KeyMap.Back, m.KeyMap.Select}, {m.KeyMap.Remove, m.KeyMap.No}}
	return helpBindings{short: short, full: full}
}

func (m FileSelectorModel) Init() tea.Cmd {
	return m.Picker.Init()
}

func pathsEquivalent(path1, path2 string) bool {
	fi1, err1 := os.Stat(path1)
	if err1 != nil {
		return filepath.Clean(path1) == filepath.Clean(path2)
	}
	fi2, err2 := os.Stat(path2)
	if err2 != nil {
		return filepath.Clean(path1) == filepath.Clean(path2)
	}
	return os.SameFile(fi1, fi2)
}

func (m FileSelectorModel) hasSelectedPath(candidate string) bool {
	for _, existing := range m.SelectedPaths {
		if pathsEquivalent(existing, candidate) {
			return true
		}
	}
	return false
}

func (m *FileSelectorModel) removeLastSelected() {
	if len(m.SelectedPaths) == 0 {
		m.Notice = "No selected file to remove."
		return
	}
	m.SelectedPaths = m.SelectedPaths[:len(m.SelectedPaths)-1]
	m.Prompting = false
	m.Notice = "Removed last selected file."
}

func (m FileSelectorModel) createFilteredSymlinkDir(srcDir string, query string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	lowerQuery := strings.ToLower(query)
	for _, file := range files {
		originalName := file.Name()
		lowerName := strings.ToLower(originalName)
		ext := strings.ToLower(filepath.Ext(originalName))
		if lowerQuery == "" || strings.Contains(lowerName, lowerQuery) ||
			strings.Contains(ext, lowerQuery) {

			srcPath := filepath.Join(srcDir, originalName)
			dstPath := filepath.Join(m.tempDir, originalName)

			if file.IsDir() {
				err := os.Symlink(srcPath, dstPath)
				if err != nil {
					log.Printf("error creating symlink: %v", err)
				}
				continue
			}

			err := os.Link(srcPath, dstPath)
			if err != nil {
				// Fallback to symlink if hard link is not supported (e.g., cross-device)
				err = os.Symlink(srcPath, dstPath)
			}

			if err != nil {
				log.Printf("error creating filtered entry: %v", err)
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

	if strings.TrimSpace(newSearch) == "" {
		m.UsingFilteredDir = false
		m.Picker.SetCurrentDirectory(m.Dir)
		return m, tea.Batch(m.Picker.Init(), cmd)
	}

	utils.ClearDir(m.tempDir)

	err := m.createFilteredSymlinkDir(m.Dir, newSearch)

	if err != nil {
		log.Printf("Couldn't create filteredDir, error: %v", err)
		return m, cmd
	}

	log.Printf("Switching to filtered dir: %s", m.tempDir)
	m.UsingFilteredDir = true
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

		if !m.Prompting && isSearchEditKey(msg) {
			m, cmd = m.handleSearch(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch strMsg {
		case "x":
			m.removeLastSelected()
			return m, tea.Batch(cmds...)
		case "y":
			m = InitialFilesSelectorModel(m.SelectedPaths, m.tempDir)
			m.Notice = ""
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
		if m.hasSelectedPath(path) {
			m.Notice = "That file is already selected."
			return m, tea.Batch(cmds...)
		}
		m.SelectedPaths = append(m.SelectedPaths, path)
		m.Notice = ""
		m.Prompting = true
	}

	return m, tea.Batch(cmds...)
}

func (m FileSelectorModel) View() string {
	var s strings.Builder

	if !m.Prompting {
		s.WriteString(styles.SecondaryStyle.Render("Search: "))
		s.WriteString(m.Search.View())
		s.WriteString("\n")
	}

	s.WriteString(styles.SecondaryStyle.Render("Selected:"))
	s.WriteString("\n")

	renders := []string{}
	for _, path := range m.SelectedPaths {
		renders = append(renders, m.Picker.Styles.Selected.Render(path))
	}
	s.WriteString(strings.Join(renders, "\n"))

	if len(renders) != 0 {
		s.WriteString("\n")
	}
	if m.Notice != "" {
		s.WriteString(styles.WarningStyle.Render(m.Notice))
	} else {
		s.WriteString(" ")
	}
	s.WriteString("\n")

	s.WriteString("\n")
	if m.UsingFilteredDir {
		s.WriteString(styles.WarningStyle.Render("(Filtered View)"))
	} else {
		s.WriteString(" ")
	}
	s.WriteString("\n")

	pickerContent := ""
	if m.Done {
		pickerContent = styles.SuccessStyle.Render("✓ File selection complete.")
	} else if m.Prompting {
		pickerContent = styles.HelpStyle.Render("Pick another file? (y/n)")
	} else {
		pickerContent = m.Picker.View()
	}

	viewportHeight := m.Picker.Height
	if viewportHeight <= 0 {
		viewportHeight = 5
	}

	s.WriteString(lipgloss.NewStyle().Height(viewportHeight).MaxHeight(viewportHeight).Render(pickerContent))

	if !m.Done {
		s.WriteString("\n")
		s.WriteString(styles.HelpStyle.Render(m.Help.View(m.helpBindings())))
	}

	return s.String()
}
