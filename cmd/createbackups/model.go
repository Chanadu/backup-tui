package createbackups

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// CreateBackupsModel manages the backup creation stage
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
		s.WriteString(styles.InfoStyle.Render(m.currentFile))
		s.WriteString("\n\n")

		if m.data.Progress {
			s.WriteString(m.spinner.View())
			s.WriteString(" ")
			s.WriteString(renderProgressBar(m.progressPercent))
			s.WriteString("\n")

			if !m.currentBackupStart.IsZero() {
				elapsed := time.Since(m.currentBackupStart)
				s.WriteString("\n")
				s.WriteString(styles.TimerLabelStyle.Render("Elapsed: "))
				s.WriteString(styles.TimerStyle.Render(formatDurationWithMinutes(elapsed)))
				s.WriteString("\n")
			}
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
		s.WriteString(styles.TimerStyle.Render(formatDurationWithMinutes(totalTime)))
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
					formatDurationWithMinutes(dur))
				s.WriteString(row)
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}
	}
	return s.String()
}

func formatDurationWithMinutes(d time.Duration) string {
	totalSeconds := d.Seconds()
	minutes := int(totalSeconds) / 60
	seconds := totalSeconds - float64(minutes*60)
	return fmt.Sprintf("%dm %.2fs", minutes, seconds)
}

// InitialCreateBackupsModel creates a new CreateBackupsModel with initial state
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
