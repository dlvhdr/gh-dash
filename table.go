package main

// import (
// 	"flag"
// 	"fmt"
// 	"log"
// 	"os"
//
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/dlvhdr/gh-prs/ui/components/table"
// )
//
// func createModel2(debug bool) (table.Model, *os.File) {
// 	var loggerFile *os.File
// 	var err error
//
// 	if debug {
// 		loggerFile, err = tea.LogToFile("debug.log", "debug")
// 		if err != nil {
// 			fmt.Println("Error setting up logger")
// 		}
// 	}
//
// 	emptyState := "No items!"
// 	return table.NewModel(100, []table.Column{{Title: "Wow"}}, nil, &emptyState), loggerFile
// }
//
// func main() {
// 	debug := flag.Bool(
// 		"debug",
// 		false,
// 		"passing this flag will allow writing debug output to debug.log",
// 	)
// 	flag.Parse()
//
// 	model, logger := createModel2(*debug)
// 	if logger != nil {
// 		defer logger.Close()
// 	}
//
// 	p := tea.NewProgram(
// 		model,
// 		tea.WithAltScreen(),
// 	)
// 	if err := p.Start(); err != nil {
// 		log.Fatal(err)
// 	}
// }
