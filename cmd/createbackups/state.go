package createbackups

import (
	"regexp"
	"strings"
	"sync"
)

// backupRuntimeState tracks the progress of the current backup operation
type backupRuntimeState struct {
	mu      sync.Mutex
	percent string
}

var percentRegex = regexp.MustCompile(`(\d+)%`)

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
