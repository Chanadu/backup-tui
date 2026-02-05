package uploadbackups

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Chanadu/backup-tui/cmd/parameters"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type UploadBackupsMessage struct {
	Ok   bool
	Errs []error
}

type UploadFileProgressMsg struct {
	Err  error
	Done bool
}

type UploadBackupsModel struct {
	data        parameters.InputData
	tempDir     string
	files       []string
	done        bool
	success     bool
	errs        []error
	current     int // 0-based index
	currentFile string
}

var runningCmd *os.Process

func (m *UploadBackupsModel) KillProcess() {
	log.Printf("KillProcess called: runningCmd=%v", runningCmd)
	if runningCmd != nil {
		pgid, err := syscall.Getpgid(runningCmd.Pid)
		if err == nil {
			log.Printf("Killing process group %d", pgid)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			log.Printf("Killing process PID %d", runningCmd.Pid)
			_ = runningCmd.Kill()
		}
	}
}

func (m *UploadBackupsModel) connectSSH() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: m.data.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(m.data.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * 1e9, // 5 seconds
	}
	return ssh.Dial("tcp", m.data.Server+":22", config)
}

func uploadSingleFile(data parameters.InputData, tempDir, fileName string) error {
	client, err := (&UploadBackupsModel{data: data}).connectSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to start SFTP: %w", err)
	}
	defer sftpClient.Close()

	localPath := filepath.Join(tempDir, fileName)
	remotePath := fileName

	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file %s: %w", fileName, err)
	}
	return nil
}

func (m UploadBackupsModel) Init() tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(m.tempDir)
		if err != nil {
			return UploadBackupsMessage{
				Ok:   false,
				Errs: []error{fmt.Errorf("failed to read tempDir: %w", err)},
			}
		}
		var files []string
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
		if len(files) == 0 {
			return UploadBackupsMessage{
				Ok:   false,
				Errs: []error{fmt.Errorf("no files to upload")},
			}
		}
		m.files = files
		m.current = 0
		m.currentFile = files[0]
		return UploadFileProgressMsg{}
	}
}

func (m UploadBackupsModel) Update(msg tea.Msg) (UploadBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadFileProgressMsg:
		if len(m.files) == 0 {
			entries, _ := os.ReadDir(m.tempDir)
			for _, entry := range entries {
				if !entry.IsDir() {
					m.files = append(m.files, entry.Name())
				}
			}
		}

		// Handle error from previous upload
		if msg.Err != nil {
			m.errs = append(m.errs, fmt.Errorf("file %s: %w", m.currentFile, msg.Err))
		}
		// If done, finish
		if msg.Done {
			m.done = true
			m.success = len(m.errs) == 0
			return m, func() tea.Msg {
				return UploadBackupsMessage{
					Ok:   m.success,
					Errs: m.errs,
				}
			}
		}
		// Upload next file
		if m.current < len(m.files) {
			fileName := m.files[m.current]
			m.currentFile = fileName
			return m, func() tea.Msg {
				err := uploadSingleFile(m.data, m.tempDir, fileName)
				m2 := m // copy for closure
				m2.current++
				done := m2.current >= len(m2.files)
				return UploadFileProgressMsg{
					Err:  err,
					Done: done,
				}
			}
		}
	case UploadBackupsMessage:
		m.done = true
		m.success = msg.Ok
		m.errs = msg.Errs
		m.current = len(m.files)
	}
	return m, nil
}

func (m UploadBackupsModel) View() string {
	var s strings.Builder
	s.WriteString("\nUpload Backups\n")
	if !m.done {
		fmt.Fprintf(&s, "Uploading file %d of %d: %s\n", m.current+1, len(m.files), m.currentFile)
	} else if m.success {
		s.WriteString("All files uploaded successfully!\n")
	} else {
		fmt.Fprintf(&s, "Upload finished with %d errors.\n", len(m.errs))
	}
	return s.String()
}

func InitialUploadBackupsModel(data parameters.InputData, tempDir string) UploadBackupsModel {
	return UploadBackupsModel{
		data:    data,
		tempDir: tempDir,
	}
}
