package uploadbackups

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type UploadBackupsMessage struct {
	Ok   bool
	Errs []error
}

func (m *UploadBackupsModel) uploadBackupsMessageCmd(err error) tea.Cmd {
	if err != nil {
		m.errs = append(m.errs, err)
	}

	m.done = true
	m.success = len(m.errs) == 0

	return func() tea.Msg {
		return UploadBackupsMessage{Ok: m.success, Errs: m.errs}
	}
}

type SetCurrentFileMsg struct {
	File  string
	Index int
	Total int
}

type UploadFileResultMsg struct {
	Err error
}

type UploadReadyMsg struct {
	Files      []string
	SSHClient  *ssh.Client
	SFTPClient *sftp.Client
}

type StartNextUploadMsg struct{}

type UploadTickMsg struct{}

type uploadRuntimeState struct {
	mu           sync.Mutex
	totalBytes   int64
	uploadedByte int64
}

type progressReader struct {
	reader io.Reader
	state  *uploadRuntimeState
}

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
	bar := progress.New(progress.WithWidth(40))
	return bar.ViewAs(percentToFraction(percent))
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.state.addUploaded(int64(n))
	}
	return n, err
}

func (s *uploadRuntimeState) setTotal(total int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalBytes = total
	s.uploadedByte = 0
}

func (s *uploadRuntimeState) addUploaded(n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadedByte += n
	if s.uploadedByte > s.totalBytes {
		s.uploadedByte = s.totalBytes
	}
}

func (s *uploadRuntimeState) percent() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.totalBytes <= 0 {
		return "0%"
	}
	percent := int((s.uploadedByte * 100) / s.totalBytes)
	if percent > 100 {
		percent = 100
	}
	return fmt.Sprintf("%d%%", percent)
}

func startNextUploadCmd() tea.Cmd {
	return func() tea.Msg {
		return StartNextUploadMsg{}
	}
}

func uploadTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return UploadTickMsg{}
	})
}

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

	runtimeState *uploadRuntimeState
	doneCh       chan error

	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (m *UploadBackupsModel) connectSSH() error {
	config := &ssh.ClientConfig{
		User: m.data.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(m.data.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * 1e9,
	}
	client, err := ssh.Dial("tcp", m.data.Server+":22", config)
	if err != nil {
		return err
	}
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		defer func() {
			err := client.Close()
			if err != nil {
				log.Printf("Error closing SSH client: %v", err)
			}
		}()
		return fmt.Errorf("failed to start SFTP: %v", err)
	}
	m.sshClient = client
	m.sftpClient = sftpClient
	return nil
}

func (m *UploadBackupsModel) uploadSingleFile(fileName string, state *uploadRuntimeState) error {
	if m.sftpClient == nil {
		return fmt.Errorf("SFTP client not initialized")
	}

	localPath := filepath.Join(m.tempDir, fileName)
	remotePath := path.Join(m.data.BackupPath, fileName)

	log.Printf("Uploading file: local=%s remote=%s", localPath, remotePath)

	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %v", localPath, err)
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			log.Printf("Error closing local file %s: %v", localPath, err)
		}
	}()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file %s: %v", localPath, err)
	}
	state.setTotal(fileInfo.Size())

	dstFile, err := m.sftpClient.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %v", remotePath, err)
	}
	defer func() {
		err := dstFile.Close()
		if err != nil {
			log.Printf("Error closing remote file %s: %v", remotePath, err)
		}
	}()

	_, err = io.Copy(dstFile, &progressReader{reader: srcFile, state: state})
	if err != nil {
		return fmt.Errorf("failed to copy file %s: %w", fileName, err)
	}
	log.Printf("Successfully uploaded file: %s", fileName)
	return nil
}

func uploadFileCmd(m UploadBackupsModel, file string, state *uploadRuntimeState, doneCh chan error) tea.Cmd {
	return func() tea.Msg {
		go func() {
			err := m.uploadSingleFile(file, state)
			if err != nil {
				log.Printf("Error uploading file %s: %v", file, err)
			}
			doneCh <- err
		}()
		return nil
	}
}

func (m *UploadBackupsModel) closeConnections() {
	if m.sftpClient != nil {
		err := m.sftpClient.Close()
		if err != nil {
			log.Printf("Error closing SFTP client: %v", err)
		}
		m.sftpClient = nil
	}
	if m.sshClient != nil {
		err := m.sshClient.Close()
		if err != nil {
			log.Printf("Error closing SSH client: %v", err)
		}
		m.sshClient = nil
	}
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
		return m, startNextUploadCmd()

	case StartNextUploadMsg:
		if m.currentIdx >= len(m.files) {
			m.closeConnections()
			return m, m.uploadBackupsMessageCmd(nil)
		}

		m.currentFile = m.files[m.currentIdx]
		m.currentIdx++
		m.uploadPct = "0%"
		m.runtimeState = &uploadRuntimeState{}
		m.doneCh = make(chan error, 1)

		return m, tea.Batch(uploadFileCmd(m, m.currentFile, m.runtimeState, m.doneCh), uploadTickCmd())

	case UploadTickMsg:
		if m.runtimeState != nil {
			m.uploadPct = m.runtimeState.percent()
		}

		if m.doneCh == nil {
			return m, nil
		}

		select {
		case err := <-m.doneCh:
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
	return m, nil
}

func (m UploadBackupsModel) View() string {
	var s strings.Builder
	s.WriteString("\nUpload Backups\n")
	if !m.done {
		fmt.Fprintf(&s, "[%d/%d] Uploading:  %s\n", m.currentIdx, m.totalFiles, m.currentFile)
		fmt.Fprintf(&s, "%s\n", renderProgressBar(m.uploadPct))

	} else if m.success {
		s.WriteString("All files uploaded successfully!\n")
	} else {
		fmt.Fprintf(&s, "Upload finished with %d errors.\n", len(m.errs))
		for i, err := range m.errs {
			fmt.Fprintf(&s, "[%d] %v\n", i, err)
		}
	}
	return s.String()
}

func InitialUploadBackupsModel(data parameters.InputData, tempDir string) UploadBackupsModel {
	return UploadBackupsModel{
		data:    data,
		tempDir: tempDir,
	}
}
