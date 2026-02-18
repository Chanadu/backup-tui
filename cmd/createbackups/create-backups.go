package createbackups

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

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

type StartNextBackupMsg struct{}

type BackupTickMsg struct{}

type backupRuntimeState struct {
	mu      sync.Mutex
	percent string
}

var percentRegex = regexp.MustCompile(`(\d+)%`)

func startNextBackupCmd() tea.Cmd {
	return func() tea.Msg {
		return StartNextBackupMsg{}
	}
}

func backupTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return BackupTickMsg{}
	})
}

func (s *backupRuntimeState) snapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.percent
}

func (s *backupRuntimeState) updateWithLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if match := percentRegex.FindStringSubmatch(line); len(match) > 1 {
		s.percent = match[1] + "%"
	}
}

func consume7zStream(reader io.Reader, state *backupRuntimeState) {
	bufReader := bufio.NewReader(reader)
	var builder strings.Builder

	flush := func() {
		line := builder.String()
		builder.Reset()
		state.updateWithLine(line)
	}

	updateFromBuffer := func() {
		if builder.Len() == 0 {
			return
		}
		state.updateWithLine(builder.String())
	}

	for {
		b, err := bufReader.ReadByte()
		if err != nil {
			if builder.Len() > 0 {
				flush()
			}
			return
		}

		if b == '\n' || b == '\r' || b == '\b' {
			flush()
			continue
		}

		_ = builder.WriteByte(b)
		if builder.Len() > 512 {
			buffer := builder.String()
			builder.Reset()
			if len(buffer) > 256 {
				buffer = buffer[len(buffer)-256:]
			}
			_, _ = builder.WriteString(buffer)
		}
		updateFromBuffer()
	}
}

func run7zBackup(filePath string, tempDir string, state *backupRuntimeState, doneCh chan error) {
	baseName := filepath.Base(filePath)
	archiveName := baseName + "-backup.7z"
	archivePath := filepath.Join(tempDir, archiveName)
	log.Printf("Creating archive for %s at %s", filePath, archivePath)

	cmd := exec.Command("7z", "a", "-mx=9", "-bsp1", archivePath, filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		doneCh <- fmt.Errorf("failed to setup 7z stdout: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		doneCh <- fmt.Errorf("failed to setup 7z stderr: %v", err)
		return
	}

	log.Printf("Executing command: %s", strings.Join(cmd.Args, " "))

	if err := cmd.Start(); err != nil {
		doneCh <- fmt.Errorf("failed to start 7z: %v", err)
		return
	}

	log.Printf("Started 7z process with PID %d", cmd.Process.Pid)
	runningCmd = cmd
	log.Printf("Waiting for process to finish for %s", filePath)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		consume7zStream(stdout, state)
	}()

	go func() {
		defer wg.Done()
		consume7zStream(stderr, state)
	}()

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		log.Printf("7z command returned non-zero status for %s (treated as warning): %v", filePath, err)
		doneCh <- nil
		return
	}

	doneCh <- err
}

type CreateBackupsModel struct {
	data        parameters.InputData
	tempDir     string
	paths       []string
	done        bool
	success     bool
	errs        []error
	currentFile string
	currentIdx  int

	progressPercent string

	runtimeState *backupRuntimeState
	doneCh       chan error
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
	state := &backupRuntimeState{
		percent: "0%",
	}
	doneCh := make(chan error, 1)
	m.runtimeState = state
	m.doneCh = doneCh
	m.progressPercent = "0%"

	return func() tea.Msg {
		go run7zBackup(filePath, m.tempDir, state, doneCh)
		return nil
	}
}

func (m CreateBackupsModel) Init() tea.Cmd {
	log.Printf("Starting backup creation for %d files.", len(m.paths))
	return startNextBackupCmd()
}

func (m CreateBackupsModel) Update(msg tea.Msg) (CreateBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case StartNextBackupMsg:
		m.currentIdx++
		if m.currentIdx >= len(m.paths) {
			return m, m.createBackupsMessageCmd(nil)
		}

		m.currentFile = m.paths[m.currentIdx]
		return m, tea.Batch(m.backupCmd(m.currentFile), backupTickCmd())

	case BackupTickMsg:
		if m.runtimeState != nil {
			m.progressPercent = m.runtimeState.snapshot()
		}

		if m.doneCh == nil {
			return m, nil
		}

		select {
		case err := <-m.doneCh:
			m.progressPercent = "0%"
			if err != nil {
				m.errs = append(m.errs, err)
			}
			return m, startNextBackupCmd()
		default:
			return m, backupTickCmd()
		}

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
		fmt.Fprintf(&s, "Progress: %s\n", m.progressPercent)

		if m.data.Commands && m.currentFile != "" {
			baseName := filepath.Base(m.currentFile)
			archiveName := baseName + "-backup.7z"
			archivePath := filepath.Join(m.tempDir, archiveName)
			cmdStr := fmt.Sprintf("7z a -mx=9 -bsp1 %s %s", archivePath, m.currentFile)
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
		data:       data,
		tempDir:    tempDir,
		paths:      paths,
		currentIdx: -1,
	}
	return model
}
