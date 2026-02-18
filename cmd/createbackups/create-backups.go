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

func (m *CreateBackupsModel) createBackupsMessageCmd(err error) tea.Cmd {
	if err != nil {
		m.errs = append(m.errs, err)
	}
	m.done = true
	m.success = len(m.errs) == 0
	return func() tea.Msg {
		return CreateBackupsMessage{Ok: m.success, Errs: m.errs}
	}
}

type BackupOutputMsg struct {
	Err error
}
type SetCurrentFileMsg struct {
	File string
}

func setCurrentFileCmd(file string) tea.Cmd {
	return func() tea.Msg {
		return SetCurrentFileMsg{File: file}
	}
}

type CreateBackupsModel struct {
	data        parameters.InputData
	tempDir     string
	paths       []string
	done        bool
	success     bool
	errs        []error
	currentFile string
}

var runningCmd *exec.Cmd

func (m *CreateBackupsModel) KillProcess() {
	log.Printf("KillProcess called: runningCmd=%v", runningCmd)
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

func (m *CreateBackupsModel) backupCmd(filePath string) tea.Cmd {
	return func() tea.Msg {
		baseName := filepath.Base(filePath)
		archiveName := baseName + "-backup.7z"
		archivePath := filepath.Join(m.tempDir, archiveName)
		log.Printf("Creating archive for %s at %s", filePath, archivePath)

		cmd := exec.Command("7z", "a", "-mx=9", archivePath, filePath)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		log.Printf("Executing command: %s", strings.Join(cmd.Args, " "))

		if err := cmd.Start(); err != nil {
			return m.createBackupsMessageCmd(fmt.Errorf("failed to start 7z: %v", err))
		}

		log.Printf("Started 7z process with PID %d", cmd.Process.Pid)
		runningCmd = cmd
		log.Printf("Waiting for process to finish for %s", filePath)

		if err := cmd.Wait(); err != nil {
			log.Printf("7z command errored for %s: %v", filePath, err)
			return nil
		}

		return nil
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	log.Printf("Starting backup creation for %d files.", len(m.paths))

	var cmds []tea.Cmd
	for _, path := range m.paths {
		cmds = append(cmds, setCurrentFileCmd(path))
		cmds = append(cmds, m.backupCmd(path))
	}

	cmds = append(cmds, m.createBackupsMessageCmd(nil))

	return tea.Sequence(cmds...)
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SetCurrentFileMsg:
		m.currentFile = msg.File
		return m, nil

	case BackupOutputMsg:
		if msg.Err != nil {
			m.errs = append(m.errs, msg.Err)
		}
		return m, nil

	case CreateBackupsMessage:
		m.done = true
		m.success = msg.Ok
		m.errs = msg.Errs
		return m, nil
	}
	return m, nil
}

func (m CreateBackupsModel) View() string {
	var s strings.Builder
	s.WriteString("\n")
	if !m.done {
		currentIdx := 0
		total := len(m.paths)
		for i, f := range m.paths {
			if f == m.currentFile {
				currentIdx = i + 1
				break
			}
		}
		fmt.Fprintf(&s, "[%d/%d] Creating backup for: %s\n", currentIdx, total, m.currentFile)

		if m.data.Commands && m.currentFile != "" {
			baseName := filepath.Base(m.currentFile)
			archiveName := baseName + "-backup.7z"
			archivePath := filepath.Join(m.tempDir, archiveName)
			cmdStr := fmt.Sprintf("7z a -mx=9 %s %s", archivePath, m.currentFile)
			fmt.Fprintf(&s, "Command: %s\n", cmdStr)
		}

		return s.String()
	}
	if m.success {
		s.WriteString("All backups created successfully!")
	} else {
		fmt.Fprintf(&s, "Backups finished with %d errors.", len(m.errs))
		for i, err := range m.errs {
			fmt.Fprintf(&s, "[%d] %s: %v", i, m.paths[i], err)
		}
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
