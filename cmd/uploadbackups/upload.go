package uploadbackups

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// connectSSH establishes SSH and SFTP connections to the remote server
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

// uploadSingleFile uploads a single file to the remote server via SFTP
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

// uploadFileCmd creates a command to upload a file
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

// closeConnections closes SSH and SFTP connections
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
