package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	checkServer "github.com/Chanadu/backup-tui/cmd/checkserver"
	"github.com/Chanadu/backup-tui/cmd/createbackups"
	"github.com/Chanadu/backup-tui/cmd/getfiles"
	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/stage"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	stage stage.Stage

	inputsModel parameters.InputModel
	paramsData  parameters.InputData

	checkModel checkServer.CheckServerModel

	filesModel    getfiles.FileSelectorModel
	filesSelected []string

	createBackupsModel createbackups.CreateBackupsModel

	tempDir string
}

// Paramters -> check server, create backups, upload to remote server, delete local backups

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) cleanUp() tea.Msg {
	err := os.RemoveAll(m.tempDir)
	log.Printf("Cleaning up temp dir: %s", m.tempDir)
	if err != nil {
		log.Printf("Couldn't remove temp dir %s, error: %v", m.tempDir, err)
	}

	log.Printf("Killing any running backup processes.")
	if m.createBackupsModel.CurrentCmd != nil {
		log.Printf("Killing process with PID: %d", m.createBackupsModel.CurrentCmd.Process.Pid)
		_ = m.createBackupsModel.CurrentCmd.Process.Kill()
	}
	return tea.Quit
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		strMsg := msg.String()
		switch strMsg {
		case "ctrl+c":
			m.cleanUp()
			return m, tea.Quit
		}
	case parameters.InputDataMessage:
		m.stage++
		m.paramsData = msg.Data
		m.checkModel = checkServer.InitialCheckServerModel(m.paramsData)
		log.Printf("Input Data Collected: %v, %s", m.paramsData, m.stage)
		m.inputsModel.SetCurrentIndex(0)
		return m, m.checkModel.Init()
	case checkServer.CheckServerMessage:
		if msg.Ok {
			m.stage++
			m.filesModel = getfiles.InitialFilesSelectorModel([]string{}, m.tempDir)
			return m, m.filesModel.Init()
		}
	case checkServer.TryAgainMessage:
		m.stage = stage.Input

	case getfiles.FilesSelectedMsg:
		if len(msg.Paths) == 0 {
			log.Println("No files selected, exiting")
			return m, tea.Quit
		}

		m.filesSelected = msg.Paths
		m.stage++
		m.createBackupsModel = createbackups.InitialCreateBackupsModel(m.paramsData, m.filesSelected, m.tempDir)
		return m, m.createBackupsModel.Init()
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
	case stage.Delete:
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder
	switch m.stage {
	case stage.Input:
		s.WriteString(m.inputsModel.View())
	case stage.Check:
		s.WriteString(m.checkModel.View())
	case stage.Files:
		s.WriteString(m.filesModel.View())
	case stage.Create:
		s.WriteString(m.createBackupsModel.View())
	case stage.Delete:
	}

	s.WriteString("\nPress Ctrl+C to quit.")

	return s.String()
}

func initialModel(tempDir string) model {

	return model{
		stage:       stage.Input,
		inputsModel: parameters.InitialParametersInputs(),
		tempDir:     tempDir,
	}
}

func Start() {
	fmt.Println("BackupTui")
	log.Println("=========================BACKUP-TUI=======================================")

	tempDir, err := os.MkdirTemp("", "filepicker-filtered-*")

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
