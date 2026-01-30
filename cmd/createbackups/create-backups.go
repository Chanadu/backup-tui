package createbackups

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
)

type CreateBackupsMessage struct {
	Oks  []bool
	Errs []error
}

type BackupOutputMsg struct {
	CurrentFile string
	Done        bool
	Err         error
}

type CreateBackupsModel struct {
	data    parameters.InputData
	tempDir string
	paths   []string
	done    bool
	success bool
	errs    []error

	current     int
	currentFile string
	output      string
	CurrentCmd  *exec.Cmd
}

func (m *CreateBackupsModel) stream7zOutput() tea.Msg {
	filePath := m.paths[m.current]
	baseName := filepath.Base(filePath)
	archiveName := baseName + "-backup.7z"
	archivePath := filepath.Join(m.tempDir, archiveName)
	log.Printf("Creating archive for %s at %s", filePath, archivePath)

	cmd := exec.Command("7z", "a", "-mx=9", archivePath, filePath)
	m.CurrentCmd = cmd

	log.Printf("Executing command: %s", strings.Join(cmd.Args, " "))

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		m.output += "Failed to start 7z: " + err.Error() + "\n"
		m.CurrentCmd = nil
		return BackupOutputMsg{
			CurrentFile: filePath,
			Done:        true,
			Err:         err,
		}
	}

	log.Printf("Reading output for %s", filePath)
	reader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		m.output += line + "\n"
		log.Printf("7z output: %s", line)
	}

	log.Printf("Finished reading output for %s", filePath)
	err := cmd.Wait()
	if err != nil {
		log.Printf("7z command failed for %s: %v", filePath, err)
		return BackupOutputMsg{
			CurrentFile: filePath,
			Done:        true,
			Err:         err,
		}
	}

	m.CurrentCmd = nil
	return BackupOutputMsg{
		CurrentFile: filePath,
		Done:        true,
		Err:         err,
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	log.Printf("Starting backup creation for %d files.", len(m.paths))
	return m.stream7zOutput
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {

	case BackupOutputMsg:
		m.currentFile = msg.CurrentFile
		if msg.Err != nil {
			m.errs = append(m.errs, msg.Err)
		}
		m.current++
		if m.current < len(m.paths) {
			return m, m.stream7zOutput
		}

		m.done = true
		m.success = len(m.errs) == 0
		return m, nil
	}

	return m, nil
}

func (m CreateBackupsModel) View() string {
	var s strings.Builder
	if !m.done {
		fmt.Fprintf(&s, "Creating backup for: %s\n\n%s",
			m.currentFile,
			m.output)
		return s.String()
	}
	if m.success {
		s.WriteString("All backups created successfully!")
	} else {
		fmt.Fprintf(&s, "Backups finished with %d errors.", len(m.errs))
	}
	return s.String()
}

func InitialCreateBackupsModel(data parameters.InputData, paths []string, tempDir string) CreateBackupsModel {
	model := CreateBackupsModel{
		data:    data,
		tempDir: tempDir,
		paths:   paths,
	}
	return model
}
