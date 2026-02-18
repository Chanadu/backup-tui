package uploadbackups

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// UploadBackupsMessage is sent when upload completes
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

// SetCurrentFileMsg sets the current file being uploaded
type SetCurrentFileMsg struct {
	File  string
	Index int
	Total int
}

// UploadFileResultMsg contains the result of a file upload
type UploadFileResultMsg struct {
	Err error
}

// UploadReadyMsg is sent when SSH connection is established and upload can begin
type UploadReadyMsg struct {
	Files      []string
	SSHClient  *ssh.Client
	SFTPClient *sftp.Client
}

// StartNextUploadMsg triggers the next file upload
type StartNextUploadMsg struct{}

// UploadTickMsg is sent periodically to update progress
type UploadTickMsg struct{}
