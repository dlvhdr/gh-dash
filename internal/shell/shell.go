// Package shell resolves the platform-appropriate shell for executing
// user-supplied custom commands.
package shell

import (
	"os"
	"os/exec"
	"runtime"
)

// Command resolves a *exec.Cmd that runs `cmd` through a sensible shell:
//
//   - If the SHELL env var is set, run "$SHELL -c <cmd>". This honors the
//     POSIX convention on every platform, so Linux/macOS, WSL, and Windows
//     users with Git Bash / a custom shell continue to work as before.
//   - Otherwise on Windows, run "%COMSPEC% /C <cmd>" (or "cmd.exe /C <cmd>"
//     if COMSPEC is unset). Closes #686, where the previous fallback to
//     "sh" failed because Windows ships neither sh nor SHELL by default.
//   - Otherwise, run "sh -c <cmd>" (preserves the previous POSIX fallback).
func Command(cmd string) *exec.Cmd {
	if shell := os.Getenv("SHELL"); shell != "" {
		return exec.Command(shell, "-c", cmd)
	}
	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = "cmd.exe"
		}
		return exec.Command(comspec, "/C", cmd)
	}
	return exec.Command("sh", "-c", cmd)
}
