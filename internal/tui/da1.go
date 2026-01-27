package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func QueryDA1() bool {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set raw mode: %v\n", err)
		return false
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	fmt.Fprint(os.Stdout, "\x1b[c")

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		buf := make([]byte, 256)
		var response strings.Builder

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errChan <- err
				return
			}

			response.Write(buf[:n])
			responseStr := response.String()

			// DA1 response ends with 'c' character
			// Look for ESC [ ? ... c pattern or ESC [ ... c pattern
			if strings.Contains(responseStr, "c") {
				resultChan <- responseStr
				return
			}
		}
	}()

	select {
	case response := <-resultChan:
		// Check if response is a proper DA1 response
		// Format: ESC [ ? <params> c  or  ESC [ <params> c
		// If we get this, the terminal supports ANSI escape sequences
		supported := strings.Contains(response, "\x1b[")
		if supported {
			// Most modern terminals that respond to DA1 also support OSC52
			fmt.Fprintf(os.Stderr, "DA1 response: %q, assuming OSC52 support\n", response)
		}
		return supported
	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "DA1 error: %v\n", err)
		return false
	case <-time.After(100 * time.Millisecond):
		fmt.Fprintln(os.Stderr, "DA1 timeout - no response")
		return false
	}
}
