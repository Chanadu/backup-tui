package createbackups

import (
	"strconv"
	"strings"
	"time"

	"github.com/Chanadu/backup-tui/cmd/styles"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// percentToFraction converts a percentage string to a float64 fraction
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

// renderProgressBar creates a styled progress bar from a percentage string
func renderProgressBar(percent string) string {
	bar := progress.New(
		progress.WithWidth(40),
		progress.WithGradient(string(styles.Secondary), string(styles.Primary)),
	)
	return bar.ViewAs(percentToFraction(percent))
}

// startNextBackupCmd returns a command to start the next backup
func startNextBackupCmd() tea.Cmd {
	return func() tea.Msg {
		return StartNextBackupMsg{}
	}
}

// backupTickCmd returns a command to tick the backup progress
func backupTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return BackupTickMsg{}
	})
}
