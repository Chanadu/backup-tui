package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Chanadu/backup-tui/cmd"
	"github.com/Chanadu/backup-tui/cmd/config"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Println("DEBUG MODE")
		logPath, err := config.GetLogFilePath("debug.log")
		if err != nil {
			fmt.Printf("Error getting log path: %v\n", err)
			os.Exit(1)
		}

		f, err := tea.LogToFile(logPath, "")
		if err != nil {
			fmt.Printf("fatal: %v\n", err)
			os.Exit(1)
		}

		defer func() {
			_ = f.Close()
		}()
	} else {
		logPath, err := config.GetLogFilePath(fmt.Sprintf("%s.log", time.Now().Format("2006-01-02_15-04-05")))
		if err != nil {
			fmt.Printf("Error getting log path: %v\n", err)
			os.Exit(1)
		}

		f, err := tea.LogToFile(logPath, "debug: ")
		if err != nil {
			fmt.Printf("fatal: %v\n", err)
			os.Exit(1)
		}

		defer func() {
			_ = f.Close()
		}()
	}

	cmd.Start()
}
