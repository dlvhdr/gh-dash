package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "net/http/pprof"

	"charm.land/log/v2"
	"github.com/dlvhdr/gh-dash/v4/cmd"
)

func main() {
	if os.Getenv("DASH_PROFILE") != "" {
		go func() {
			log.Info("Serving pprof at localhost:6060")
			slog.Info("Serving pprof at localhost:6060")
			if httpErr := http.ListenAndServe("localhost:6060", nil); httpErr != nil {
				slog.Error("Failed to pprof listen", "error", httpErr)
			}
		}()
	}

	cmd.Execute()
}
