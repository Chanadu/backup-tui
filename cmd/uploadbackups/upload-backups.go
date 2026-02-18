package uploadbackups

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Chanadu/backup-tui/cmd/parameters"
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

type UploadFinishedMsg struct{}

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

func (m *UploadBackupsModel) uploadSingleFile(fileName string) error {
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

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file %s: %w", fileName, err)
	}
	log.Printf("Successfully uploaded file: %s", fileName)
	return nil
}

func setCurrentFileCmd(file string, index int, total int) tea.Cmd {
	return func() tea.Msg {
		return SetCurrentFileMsg{File: file, Index: index, Total: total}
	}
}

func uploadFileCmd(m UploadBackupsModel, file string) tea.Cmd {
	return func() tea.Msg {
		err := m.uploadSingleFile(file)
		if err != nil {
			log.Printf("Error uploading file %s: %v", file, err)
		}
		return UploadFileResultMsg{Err: err}
	}
}

func (m UploadBackupsModel) startUploadCmd(files []string) tea.Msg {
	var cmds []tea.Cmd

	total := len(files)
	for i, file := range files {
		cmds = append(cmds, setCurrentFileCmd(file, i+1, total))
		cmds = append(cmds, uploadFileCmd(m, file))
	}
	cmds = append(cmds, func() tea.Msg {
		if m.sftpClient != nil {
			err := m.sftpClient.Close()
			if err != nil {
				log.Printf("Error closing SFTP client: %v", err)
			}
		}
		if m.sshClient != nil {
			err := m.sshClient.Close()
			if err != nil {
				log.Printf("Error closing SSH client: %v", err)
			}
		}

		return UploadFinishedMsg{}
	})
	log.Printf("Starting Uploads")
	return tea.Sequence(cmds...)()
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
		return model.startUploadCmd(files)
	}
}

func (m UploadBackupsModel) Update(msg tea.Msg) (UploadBackupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SetCurrentFileMsg:
		m.currentFile = msg.File
		m.currentIdx = msg.Index
		m.totalFiles = msg.Total
		return m, nil
	case UploadFileResultMsg:
		if msg.Err != nil {
			m.errs = append(m.errs, msg.Err)
		}
		return m, nil
	case UploadFinishedMsg:
		return m, m.uploadBackupsMessageCmd(nil)
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
