package createbackups

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var runningCmd *exec.Cmd

// consume7zStream reads 7z output and updates progress state
func consume7zStream(reader io.Reader, state *backupRuntimeState) {
	bufReader := bufio.NewReader(reader)
	var builder strings.Builder

	flush := func() {
		line := builder.String()
		builder.Reset()
		state.updateWithLine(line)
	}

	updateFromBuffer := func() {
		if builder.Len() == 0 {
			return
		}
		state.updateWithLine(builder.String())
	}

	for {
		b, err := bufReader.ReadByte()
		if err != nil {
			if builder.Len() > 0 {
				flush()
			}
			return
		}

		if b == '\n' || b == '\r' || b == '\b' {
			flush()
			continue
		}

		_ = builder.WriteByte(b)
		if builder.Len() > 512 {
			buffer := builder.String()
			builder.Reset()
			if len(buffer) > 256 {
				buffer = buffer[len(buffer)-256:]
			}
			_, _ = builder.WriteString(buffer)
		}
		updateFromBuffer()
	}
}

// run7zBackup executes the 7z command to create a backup archive
func run7zBackup(filePath string, tempDir string, state *backupRuntimeState, doneCh chan error) {
	baseName := filepath.Base(filePath)
	archiveName := baseName + "-backup.7z"
	archivePath := filepath.Join(tempDir, archiveName)
	log.Printf("Creating archive for %s at %s", filePath, archivePath)

	cmd := exec.Command("7z", "a", "-mx=9", "-bsp1", archivePath, filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		doneCh <- fmt.Errorf("failed to setup 7z stdout: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		doneCh <- fmt.Errorf("failed to setup 7z stderr: %v", err)
		return
	}

	log.Printf("Executing command: %s", strings.Join(cmd.Args, " "))

	if err := cmd.Start(); err != nil {
		doneCh <- fmt.Errorf("failed to start 7z: %v", err)
		return
	}

	log.Printf("Started 7z process with PID %d", cmd.Process.Pid)
	runningCmd = cmd
	log.Printf("Waiting for process to finish for %s", filePath)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		consume7zStream(stdout, state)
	}()

	go func() {
		defer wg.Done()
		consume7zStream(stderr, state)
	}()

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		log.Printf("7z command returned non-zero status for %s (treated as warning): %v", filePath, err)
		doneCh <- nil
		return
	}

	doneCh <- err
}

// KillProcess terminates the currently running 7z process
func (m *CreateBackupsModel) KillProcess() {
	log.Printf("KillProcess called: runningCmd=%v", runningCmd)
	if runningCmd != nil && runningCmd.Process != nil {
		pgid, err := syscall.Getpgid(runningCmd.Process.Pid)
		if err == nil {
			log.Printf("Killing process group %d", pgid)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			log.Printf("Killing process PID %d", runningCmd.Process.Pid)
			_ = runningCmd.Process.Kill()
		}
	}
}
