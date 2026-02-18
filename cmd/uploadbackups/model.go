package uploadbackups

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// UploadBackupsModel manages the upload stage
type UploadBackupsModel struct {
	data        parameters.InputData
	tempDir     string
	files       []string
	done        bool
	success     bool
	errs        []error
	currentFile string
	currentIdx  int
	totalFiles  int
	uploadPct   string
	spinner     spinner.Model

	runtimeState *uploadRuntimeState
	doneCh       chan error

	sshClient  *ssh.Client
	sftpClient *sftp.Client

	// Timing
	currentUploadStart time.Time
	totalStartTime     time.Time
	uploadTimes        []time.Duration
	creationTimes      []time.Duration
	creationPaths      []string
}

func (m UploadBackupsModel) Init() tea.Cmd {
	files := m.files
	if len(files) == 0 {
		entries, err := os.ReadDir(m.tempDir)
		if err != nil {
			return m.uploadBackupsMessageCmd(fmt.Errorf("failed to read tempDir: %w", err))
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
		if len(files) == 0 {
			return m.uploadBackupsMessageCmd(fmt.Errorf("no files to upload"))
		}
	}

	return func() tea.Msg {
		model := m
		if err := model.connectSSH(); err != nil {
			return model.uploadBackupsMessageCmd(fmt.Errorf("failed to connect SSH: %w", err))()
		}
		return UploadReadyMsg{Files: files, SSHClient: model.sshClient, SFTPClient: model.sftpClient}
	}
}

func (m UploadBackupsModel) Update(msg tea.Msg) (UploadBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadReadyMsg:
		m.files = msg.Files
		m.totalFiles = len(msg.Files)
		m.sshClient = msg.SSHClient
		m.sftpClient = msg.SFTPClient
		m.currentIdx = 0
		m.uploadPct = "0%"
		m.totalStartTime = time.Now()
		return m, startNextUploadCmd()

	case StartNextUploadMsg:
		if m.currentIdx >= len(m.files) {
			m.closeConnections()
			return m, m.uploadBackupsMessageCmd(nil)
		}

		m.currentFile = m.files[m.currentIdx]
		m.currentUploadStart = time.Now()
		m.currentIdx++
		m.uploadPct = "0%"
		m.runtimeState = &uploadRuntimeState{}
		m.doneCh = make(chan error, 1)

		return m, tea.Batch(uploadFileCmd(m, m.currentFile, m.runtimeState, m.doneCh), uploadTickCmd(), m.spinner.Tick)

	case UploadTickMsg:
		if m.runtimeState != nil {
			m.uploadPct = m.runtimeState.percent()
		}

		if m.doneCh == nil {
			return m, nil
		}

		select {
		case err := <-m.doneCh:
			duration := time.Since(m.currentUploadStart)
			m.uploadTimes = append(m.uploadTimes, duration)
			if err != nil {
				m.errs = append(m.errs, err)
			}
			m.uploadPct = "0%"
			return m, startNextUploadCmd()
		default:
			return m, uploadTickCmd()
		}

	case UploadFileResultMsg:
		if msg.Err != nil {
			m.errs = append(m.errs, msg.Err)
		}
		return m, nil

	case UploadBackupsMessage:
		m.done = true
		m.success = msg.Ok
		m.errs = msg.Errs
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m UploadBackupsModel) View() string {
	var s strings.Builder
	if !m.done {
		label := fmt.Sprintf("[%d/%d] Uploading:", m.currentIdx, m.totalFiles)
		s.WriteString(styles.ProgressLabelStyle.Render(label))
		s.WriteString("  ")
		s.WriteString(m.currentFile)
		s.WriteString("\n")
		s.WriteString(m.spinner.View())
		s.WriteString(" ")
		s.WriteString(renderProgressBar(m.uploadPct))
		s.WriteString("\n")

		if !m.currentUploadStart.IsZero() {
			elapsed := time.Since(m.currentUploadStart)
			s.WriteString("\n")
			s.WriteString(styles.TimerLabelStyle.Render("Elapsed: "))
			s.WriteString(styles.TimerStyle.Render(fmt.Sprintf("%.2fs", elapsed.Seconds())))
			s.WriteString("\n")
		}

	} else if m.success {
		s.WriteString(styles.SuccessStyle.Render("✓ All files uploaded successfully!"))
		s.WriteString("\n")
	} else {
		s.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("✗ Upload finished with %d errors.", len(m.errs))))
		s.WriteString("\n")
		for i, err := range m.errs {
			s.WriteString(styles.MutedStyle.Render(fmt.Sprintf("[%d] ", i)))
			s.WriteString(styles.ErrorStyle.Render(err.Error()))
			s.WriteString("\n")
		}
	}

	if !m.totalStartTime.IsZero() && m.done {
		totalTime := time.Since(m.totalStartTime)
		s.WriteString("\n")
		s.WriteString(styles.TimerLabelStyle.Render("Total time: "))
		s.WriteString(styles.TimerStyle.Render(fmt.Sprintf("%.2fs", totalTime.Seconds())))
		s.WriteString("\n")

		// Combined timing table
		if len(m.uploadTimes) > 0 || len(m.creationTimes) > 0 {
			s.WriteString("\n")
			s.WriteString(styles.SecondaryStyle.Render("File Times"))
			s.WriteString("\n\n")

			// Header
			header := fmt.Sprintf("  %-4s  %-32s  %-13s  %s", "#", "File", "Creation", "Upload")
			s.WriteString(styles.PrimaryStyle.Bold(true).Render(header))
			s.WriteString("\n")
			s.WriteString(styles.SecondaryStyle.Render(strings.Repeat("─", 70)))
			s.WriteString("\n")

			// Determine the number of files to display
			maxFiles := len(m.files)
			if len(m.creationTimes) > maxFiles {
				maxFiles = len(m.creationTimes)
			}
			if len(m.uploadTimes) > maxFiles {
				maxFiles = len(m.uploadTimes)
			}

			// Rows
			for i := 0; i < maxFiles; i++ {
				var fileName string
				if i < len(m.creationPaths) {
					fileName = filepath.Base(m.creationPaths[i])
				} else if i < len(m.files) {
					fileName = filepath.Base(m.files[i])
					// Strip -backup.7z suffix if present
					fileName = strings.TrimSuffix(fileName, "-backup.7z")
				} else {
					fileName = fmt.Sprintf("File %d", i+1)
				}
				if len(fileName) > 32 {
					fileName = fileName[:29] + "..."
				}

				var creationTime, uploadTime string
				if i < len(m.creationTimes) {
					creationTime = fmt.Sprintf("%.2fs", m.creationTimes[i].Seconds())
				} else {
					creationTime = "-"
				}
				if i < len(m.uploadTimes) {
					uploadTime = fmt.Sprintf("%.2fs", m.uploadTimes[i].Seconds())
				} else {
					uploadTime = "-"
				}

				row := fmt.Sprintf("  %-4d  %-32s  %-13s  %s",
					i+1,
					fileName,
					creationTime,
					uploadTime)
				s.WriteString(row)
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}
	}
	return s.String()
}

// InitialUploadBackupsModel creates a new UploadBackupsModel with initial state
func InitialUploadBackupsModel(data parameters.InputData, tempDir string, creationTimes []time.Duration, creationPaths []string) UploadBackupsModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = styles.InfoStyle

	return UploadBackupsModel{
		data:          data,
		tempDir:       tempDir,
		creationTimes: creationTimes,
		creationPaths: creationPaths,
		spinner:       s,
	}
}
