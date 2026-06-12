package shell

import (
	"runtime"
	"strings"
	"testing"
)

func TestCommand_HonorsShellEnv(t *testing.T) {
	// SHELL is the POSIX convention and overrides platform defaults whenever
	// set, so users with bash / zsh / fish on any OS keep their behavior.
	t.Setenv("SHELL", "/usr/local/bin/bash")
	t.Setenv("COMSPEC", "C:\\Windows\\System32\\cmd.exe")

	c := Command("echo hello")
	if c.Path != "/usr/local/bin/bash" {
		t.Fatalf("expected SHELL to be honored, got %q", c.Path)
	}
	if len(c.Args) < 3 || c.Args[1] != "-c" || c.Args[2] != "echo hello" {
		t.Fatalf("expected POSIX -c invocation, got %v", c.Args)
	}
}

func TestCommand_FallbackWithoutShell(t *testing.T) {
	t.Setenv("SHELL", "")
	c := Command("echo hello")

	if runtime.GOOS == "windows" {
		// Should fall back to COMSPEC (or cmd.exe) with /C — closes #686.
		if !strings.HasSuffix(strings.ToLower(c.Path), "cmd.exe") {
			t.Fatalf("on Windows expected cmd.exe fallback, got %q", c.Path)
		}
		if len(c.Args) < 3 || c.Args[1] != "/C" || c.Args[2] != "echo hello" {
			t.Fatalf("on Windows expected /C invocation, got %v", c.Args)
		}
	} else {
		// Pre-fix behavior preserved for POSIX: sh -c <cmd>.
		if c.Path != "sh" && !strings.HasSuffix(c.Path, "/sh") {
			t.Fatalf("on POSIX expected sh fallback, got %q", c.Path)
		}
		if len(c.Args) < 3 || c.Args[1] != "-c" || c.Args[2] != "echo hello" {
			t.Fatalf("on POSIX expected -c invocation, got %v", c.Args)
		}
	}
}

func TestCommand_HonorsComspecOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only behavior")
	}
	t.Setenv("SHELL", "")
	t.Setenv("COMSPEC", "C:\\custom\\path\\cmd.exe")

	c := Command("echo hello")
	if c.Path != "C:\\custom\\path\\cmd.exe" {
		t.Fatalf("expected COMSPEC override, got %q", c.Path)
	}
	if len(c.Args) < 3 || c.Args[1] != "/C" {
		t.Fatalf("expected /C, got %v", c.Args)
	}
}

func TestCommand_PassesCommandStringIntact(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")
	c := Command("echo 'hello world' && exit 0")
	if c.Args[2] != "echo 'hello world' && exit 0" {
		t.Fatalf("command string altered, got %q", c.Args[2])
	}
}
