package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dlvhdr/gh-prs/ui"
)

func createModel() (ui.Model, *os.File) {
	loggerFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Error setting up logger")
	}

	return ui.NewModel(loggerFile), loggerFile
}

func main() {
	model, logger := createModel()
	defer logger.Close()
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
