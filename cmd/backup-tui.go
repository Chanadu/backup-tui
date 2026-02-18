package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/checkserver"
	"github.com/Chanadu/backup-tui/cmd/createbackups"
	"github.com/Chanadu/backup-tui/cmd/getfiles"
	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/stage"
	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/Chanadu/backup-tui/cmd/uploadbackups"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	stage stage.Stage
	width int

	inputsModel parameters.InputModel
	paramsData  parameters.InputData

	checkModel checkserver.CheckServerModel

	filesModel    getfiles.FileSelectorModel
	filesSelected []string

	createBackupsModel createbackups.CreateBackupsModel

	uploadBackupsModel uploadbackups.UploadBackupsModel

	tempDir string
}

func (m model) stageTitle() string {
	switch m.stage {
	case stage.Input:
		return "Parameters"
	case stage.Check:
		return "Check Server"
	case stage.Files:
		return "Select Files"
	case stage.Create:
		return "Create Backups"
	case stage.Upload:
		return "Upload Backups"
	default:
		return "Backup TUI"
	}
}

// Paramters -> check server, create backups, upload to remote server, delete local backups
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) cleanUp() tea.Cmd {
	err := os.RemoveAll(m.tempDir)
	log.Printf("Cleaning up temp dir: %s", m.tempDir)
	if err != nil {
		log.Printf("Couldn't remove temp dir %s, error: %v", m.tempDir, err)
	}

	if m.stage == stage.Create {
		m.createBackupsModel.KillProcess()
		log.Printf("Killed create backups process")
	}

	log.Printf("Exiting")
	return tea.Quit
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		strMsg := msg.String()
		switch strMsg {
		case "ctrl+c":
			log.Printf("User initiated quit")

			return m, m.cleanUp()
		}
	case parameters.InputDataMessage:
		m.stage++
		m.paramsData = msg.Data
		m.checkModel = checkserver.InitialCheckServerModel(m.paramsData)
		log.Printf("Input Data Collected: %v, %s", m.paramsData, m.stage)
		m.inputsModel.SetCurrentIndex(0)
		return m, m.checkModel.Init()
	case checkserver.CheckServerMessage:
		if msg.Ok {
			m.stage++
			m.filesModel = getfiles.InitialFilesSelectorModel([]string{}, m.tempDir)
			return m, m.filesModel.Init()
		}
	case checkserver.TryAgainMessage:
		m.stage = stage.Input

	case getfiles.FilesSelectedMsg:
		if len(msg.Paths) == 0 {
			log.Println("No files selected, exiting")
			return m, m.cleanUp()
		}

		m.filesSelected = msg.Paths
		m.stage++
		m.createBackupsModel = createbackups.InitialCreateBackupsModel(m.paramsData, m.filesSelected, m.tempDir)
		return m, m.createBackupsModel.Init()
	case createbackups.CreateBackupsMessage:
		if !msg.Ok {
			for _, err := range msg.Errs {
				log.Printf("Error during backup creation: %v", err)
				return m, m.cleanUp()
			}
		}
		log.Printf("Created backups:")
		m.stage++
		m.uploadBackupsModel = uploadbackups.InitialUploadBackupsModel(m.paramsData, m.tempDir, msg.BackupTimes, msg.Paths)
		return m, m.uploadBackupsModel.Init()

	case uploadbackups.UploadBackupsMessage:
		if !msg.Ok {
			for _, err := range msg.Errs {
				log.Printf("Error during backup upload: %v", err)
			}
			return m, m.cleanUp()
		}
		log.Printf("Created uploads")
		return m, m.cleanUp()
	}

	var cmd tea.Cmd
	switch m.stage {
	case stage.Input:
		m.inputsModel, cmd = m.inputsModel.Update(msg)
	case stage.Check:
		m.checkModel, cmd = m.checkModel.Update(msg)
	case stage.Files:
		m.filesModel, cmd = m.filesModel.Update(msg)
	case stage.Create:
		m.createBackupsModel, cmd = m.createBackupsModel.Update(msg)
	case stage.Upload:
		m.uploadBackupsModel, cmd = m.uploadBackupsModel.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(styles.AppTitleStyle.Render("BETTER-TUI"))
	s.WriteString("\n")
	s.WriteString(styles.TitleStyle.Render(m.stageTitle()))
	s.WriteString("\n")
	switch m.stage {
	case stage.Input:
		s.WriteString(m.inputsModel.View())
	case stage.Check:
		s.WriteString(m.checkModel.View())
	case stage.Files:
		s.WriteString(m.filesModel.View())
	case stage.Create:
		s.WriteString(m.createBackupsModel.View())
	case stage.Upload:
		s.WriteString(m.uploadBackupsModel.View())
	}

	s.WriteString("\n")
	s.WriteString(styles.HelpStyle.Render("Press Ctrl+C to quit."))

	box := styles.BoxStyle
	if m.width > 6 {
		box = box.Width(m.width - 6)
	}

	return "\n" + box.Render(s.String())
}

func initialModel(tempDir string) model {

	return model{
		stage:       stage.Input,
		inputsModel: parameters.InitialParametersInputs(),
		tempDir:     tempDir,
	}
}

func Start() {
	log.Println("=========================BACKUP-TUI=======================================")

	tempDir, err := os.MkdirTemp("", "backup-tui-*")

	if err != nil {
		log.Fatalf("Couldn't create temp dir, error: %v", err)
	}

	m := initialModel(tempDir)
	p := tea.NewProgram(m)

	defer m.cleanUp()

	if _, err := p.Run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
