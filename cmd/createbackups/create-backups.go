package createbackups

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type CreateBackupsMessage struct {
	Ok          bool
	Errs        []error
	BackupTimes []time.Duration
	Paths       []string
}

func (m *CreateBackupsModel) createBackupsMessageCmd(err error) tea.Cmd {
	if err != nil {
		m.errs = append(m.errs, err)
	}
	m.done = true
	m.success = len(m.errs) == 0
	return func() tea.Msg {
		return CreateBackupsMessage{
			Ok:          m.success,
			Errs:        m.errs,
			BackupTimes: m.backupTimes,
			Paths:       m.paths,
		}
	}
}

type StartNextBackupMsg struct{}

type BackupTickMsg struct{}

type backupRuntimeState struct {
	mu      sync.Mutex
	percent string
}

var percentRegex = regexp.MustCompile(`(\d+)%`)

func percentToFraction(percent string) float64 {
	percent = strings.TrimSuffix(strings.TrimSpace(percent), "%")
	value, err := strconv.Atoi(percent)
	if err != nil {
		return 0
	}
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return float64(value) / 100
}

func renderProgressBar(percent string) string {
	bar := progress.New(
		progress.WithWidth(40),
		progress.WithGradient(string(styles.Secondary), string(styles.Primary)),
	)
	return bar.ViewAs(percentToFraction(percent))
}

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
	spinner         spinner.Model

	runtimeState *backupRuntimeState
	doneCh       chan error

	// Timing
	currentBackupStart time.Time
	totalStartTime     time.Time
	backupTimes        []time.Duration
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

		if m.currentIdx == 0 && m.totalStartTime.IsZero() {
			m.totalStartTime = time.Now()
		}

		m.currentFile = m.paths[m.currentIdx]
		m.currentBackupStart = time.Now()
		return m, tea.Batch(m.backupCmd(m.currentFile), backupTickCmd(), m.spinner.Tick)

	case BackupTickMsg:
		if m.runtimeState != nil {
			m.progressPercent = m.runtimeState.snapshot()
		}

		if m.doneCh == nil {
			return m, nil
		}

		select {
		case err := <-m.doneCh:
			duration := time.Since(m.currentBackupStart)
			m.backupTimes = append(m.backupTimes, duration)
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

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m CreateBackupsModel) View() string {
	var s strings.Builder
	if !m.done {
		currentIdx := 0
		total := len(m.paths)
		for i, f := range m.paths {
			if f == m.currentFile {
				currentIdx = i + 1
				break
			}
		}
		label := fmt.Sprintf("[%d/%d] Creating backup for:", currentIdx, total)
		s.WriteString(styles.ProgressLabelStyle.Render(label))
		s.WriteString(" ")
		s.WriteString(m.currentFile)
		s.WriteString("\n\n")
		s.WriteString(m.spinner.View())
		s.WriteString(" ")
		s.WriteString(renderProgressBar(m.progressPercent))
		s.WriteString("\n")

		if !m.currentBackupStart.IsZero() {
			elapsed := time.Since(m.currentBackupStart)
			s.WriteString("\n")
			s.WriteString(styles.TimerLabelStyle.Render("Elapsed: "))
			s.WriteString(styles.TimerStyle.Render(fmt.Sprintf("%.2fs", elapsed.Seconds())))
			s.WriteString("\n")
		}

		if m.data.Commands && m.currentFile != "" {
			baseName := filepath.Base(m.currentFile)
			archiveName := baseName + "-backup.7z"
			archivePath := filepath.Join(m.tempDir, archiveName)
			cmdStr := fmt.Sprintf("7z a -mx=9 -bsp1 %s %s", archivePath, m.currentFile)
			s.WriteString("\n")
			s.WriteString(styles.MutedStyle.Render("Command: "))
			s.WriteString(styles.InfoStyle.Render(cmdStr))
			s.WriteString("\n")
		}

		return s.String()
	}
	if m.success {
		s.WriteString(styles.SuccessStyle.Render("✓ All backups created successfully!"))
	} else {
		s.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("✗ Backups finished with %d errors.", len(m.errs))))
		s.WriteString("\n\n")
		for i, err := range m.errs {
			s.WriteString(styles.MutedStyle.Render(fmt.Sprintf("[%d] %s: ", i, m.paths[i])))
			s.WriteString(styles.ErrorStyle.Render(err.Error()))
			s.WriteString("\n")
		}
	}

	if !m.totalStartTime.IsZero() {
		totalTime := time.Since(m.totalStartTime)
		s.WriteString("\n")
		s.WriteString(styles.TimerLabelStyle.Render("Total time: "))
		s.WriteString(styles.TimerStyle.Render(fmt.Sprintf("%.2fs", totalTime.Seconds())))
		s.WriteString("\n")
		if len(m.backupTimes) > 0 {
			s.WriteString("\n")
			s.WriteString(styles.SecondaryStyle.Render("Backup Times"))
			s.WriteString("\n\n")

			// Header
			header := fmt.Sprintf("  %-4s  %-40s  %s", "#", "File", "Time")
			s.WriteString(styles.PrimaryStyle.Bold(true).Render(header))
			s.WriteString("\n")
			s.WriteString(styles.SecondaryStyle.Render(strings.Repeat("─", 60)))
			s.WriteString("\n")

			// Rows
			for i, dur := range m.backupTimes {
				var fileName string
				if i < len(m.paths) {
					fileName = filepath.Base(m.paths[i])
				} else {
					fileName = fmt.Sprintf("File %d", i+1)
				}
				if len(fileName) > 40 {
					fileName = fileName[:37] + "..."
				}
				row := fmt.Sprintf("  %-4d  %-40s  %s",
					i+1,
					fileName,
					fmt.Sprintf("%.2fs", dur.Seconds()))
				s.WriteString(row)
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}
	}
	return s.String()
}

func InitialCreateBackupsModel(data parameters.InputData, paths []string, tempDir string) CreateBackupsModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = styles.InfoStyle

	model := CreateBackupsModel{
		data:       data,
		tempDir:    tempDir,
		paths:      paths,
		currentIdx: -1,
		spinner:    s,
	}
	return model
}
