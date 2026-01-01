//go:build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func setupThemeReload(p *tea.Program) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR2)
	go func() {
		for range sigChan {
			p.Send(constants.ThemeReloadMsg{})
		}
	}()
}
