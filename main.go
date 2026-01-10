package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Chanadu/backup-tui/cmd"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Println("DEBUG MODE")
		f, err := tea.LogToFile("debug.log", "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		defer func() {
			_ = f.Close()
		}()
	} else {
		f, err := tea.LogToFile(fmt.Sprintf("%s.log", time.Now().Format("2006-01-02_15-04-05")), "debug: ")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		defer func() {
			_ = f.Close()
		}()
	}

	cmd.Start()
}
