package testutils

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/exp/teatest"
)

func bytesContains(t *testing.T, bts []byte, str string) bool {
	t.Helper()
	return bytes.Contains(bts, []byte(str))
}

func WaitForText(t *testing.T, tm *teatest.TestModel, text string, options ...teatest.WaitForOption) {
	teatest.WaitFor(t,
		tm.Output(),
		func(bts []byte) bool {
			contains := bytesContains(t, bts, text)
			if _, debug := os.LookupEnv("DEBUG"); debug {
				if contains {
					f, _ := os.CreateTemp("", "gh-dash-debug")
					defer f.Close()
					fmt.Fprintf(f, "%s", string(bts))
					log.Error("✅ wrote to file while looking for text", "file", f.Name(), "text", text)
				} else {
					f, _ := os.CreateTemp("", "not-found-gh-dash-debug")
					defer f.Close()
					fmt.Fprintf(f, "%s", string(bts))
					log.Error("❌ text not found", "file", f.Name(), "text", text)
				}
			}
			return contains
		},
		options...,
	)
}
