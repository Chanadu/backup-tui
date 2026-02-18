package uploadbackups

import (
	"fmt"
	"io"
	"sync"
)

// uploadRuntimeState tracks the progress of the current upload operation
type uploadRuntimeState struct {
	mu           sync.Mutex
	totalBytes   int64
	uploadedByte int64
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

func (s *uploadRuntimeState) uploadedBytes() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uploadedByte
}

// progressReader wraps an io.Reader to track upload progress
type progressReader struct {
	reader io.Reader
	state  *uploadRuntimeState
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.state.addUploaded(int64(n))
	}
	return n, err
}
