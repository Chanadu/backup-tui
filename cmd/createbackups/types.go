package createbackups

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// CreateBackupsMessage is sent when backup creation completes
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

// StartNextBackupMsg triggers the next backup in the queue
type StartNextBackupMsg struct{}

// BackupTickMsg is sent periodically to update progress
type BackupTickMsg struct{}
