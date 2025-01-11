/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	slog "log"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui"
	"github.com/dlvhdr/gh-dash/v4/ui/markdown"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
	BuiltBy = ""
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:     "gh dash",
		Short:   "A gh extension that shows a configurable dashboard of pull requests and issues.",
		Version: "",
		Args:    cobra.MaximumNArgs(1),
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func createModel(repoPath string, configPath string, debug bool) (ui.Model, *os.File) {
	var loggerFile *os.File

	if debug {
		var fileErr error
		newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr == nil {
			log.SetOutput(newConfigFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)
			log.Debug("Logging to debug.log")
			if repoPath != "" {
				log.Debug("Running in repo", "repo", repoPath)
			}
		} else {
			loggerFile, _ = tea.LogToFile("debug.log", "debug")
			slog.Print("Failed setting up logging", fileErr)
		}
	} else {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.FatalLevel)
	}

	return ui.NewModel(repoPath, configPath), loggerFile
}

func buildVersion(version, commit, date, builtBy string) string {
	result := version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		`use this configuration file
(default lookup:
  1. a .gh-dash.yml file if inside a git repo
  2. $GH_DASH_CONFIG env var
  3. $XDG_CONFIG_HOME/gh-dash/config.yml
)`,
	)
	err := rootCmd.MarkPersistentFlagFilename("config", "yaml", "yml")
	if err != nil {
		log.Fatal("Cannot mark config flag as filename", err)
	}

	rootCmd.Version = buildVersion(Version, Commit, Date, BuiltBy)
	rootCmd.SetVersionTemplate(`gh-dash {{printf "version %s\n" .Version}}`)

	rootCmd.Flags().Bool(
		"debug",
		false,
		"passing this flag will allow writing debug output to debug.log",
	)

	rootCmd.Flags().BoolP(
		"help",
		"h",
		false,
		"help for gh-dash",
	)

	rootCmd.Run = func(_ *cobra.Command, args []string) {
		var repo string
		repos := config.IsFeatureEnabled(config.FF_REPO_VIEW)
		if repos && len(args) > 0 {
			repo = args[0]
		}

		if repo == "" {
			r, err := git.GetRepoInPwd()
			if err == nil && r != nil {
				repo = r.Path()
			}
		}
		debug, err := rootCmd.Flags().GetBool("debug")
		if err != nil {
			log.Fatal("Cannot parse debug flag", err)
		}

		// see https://github.com/charmbracelet/lipgloss/issues/73
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
		markdown.InitializeMarkdownStyle(termenv.HasDarkBackground())

		model, logger := createModel(repo, cfgFile, debug)
		if logger != nil {
			defer logger.Close()
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithReportFocus(),
		)
		if _, err := p.Run(); err != nil {
			log.Fatal("Failed starting the TUI", err)
		}
	}
}
