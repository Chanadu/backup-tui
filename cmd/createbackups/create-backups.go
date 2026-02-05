package createbackups

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
)

type CreateBackupsMessage struct {
	Ok   bool
	Errs []error
}

type BackupOutputMsg struct {
	Done bool
	Err  error
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
}

var runningCmd *exec.Cmd

func (m *CreateBackupsModel) KillProcess() {
	log.Printf("KillProcess called: m.cmd=%v", runningCmd)
	if runningCmd != nil && runningCmd.Process != nil {
		pgid, err := syscall.Getpgid(runningCmd.Process.Pid)
		if err == nil {
			log.Printf("Killing process group %d", pgid)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			log.Printf("Killing process PID %d", runningCmd.Process.Pid)
			_ = runningCmd.Process.Kill()
		}
	}
}

func (m *CreateBackupsModel) stream7zOutput() tea.Msg {
	filePath := m.paths[m.current]
	baseName := filepath.Base(filePath)
	archiveName := baseName + "-backup.7z"
	archivePath := filepath.Join(m.tempDir, archiveName)
	log.Printf("Creating archive for %s at %s", filePath, archivePath)

	m.currentFile = filePath

	cmd := exec.Command("7z", "a", "-mx=9", archivePath, filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	log.Printf("Executing command: %s", strings.Join(cmd.Args, " "))

	if err := cmd.Start(); err != nil {
		return BackupOutputMsg{
			Done: true,
			Err:  err,
		}
	}

	log.Printf("Started 7z process with PID %d", cmd.Process.Pid)
	runningCmd = cmd
	log.Printf("Waiting for process to finish for %s", filePath)

	if err := cmd.Wait(); err != nil {
		log.Printf("7z command failed for %s: %v", filePath, err)
		return BackupOutputMsg{
			Done: true,
			Err:  err,
		}
	}

	return BackupOutputMsg{
		Done: true,
		Err:  nil,
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	log.Printf("Starting backup creation for %d files.", len(m.paths))
	return m.stream7zOutput
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {

	case BackupOutputMsg:
		if msg.Err != nil {
			m.errs = append(m.errs, msg.Err)
		}
		m.current++
		if m.current < len(m.paths) {
			return m, m.stream7zOutput
		}

		m.done = true
		m.success = len(m.errs) == 0
		log.Printf("Backup creation done. Success: %v, Errors: %d\n", m.success, len(m.errs))
		return m, func() tea.Msg {
			return CreateBackupsMessage{
				Ok:   m.success,
				Errs: m.errs,
			}
		}
	}

	return m, nil
}

func (m CreateBackupsModel) View() string {
	var s strings.Builder
	s.WriteString("\n")
	if !m.done {
		fmt.Fprintf(&s, "Creating backup for: %s\n",
			m.currentFile)
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
