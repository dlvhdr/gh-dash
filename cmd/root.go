/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	slog "log"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	zone "github.com/lrstanley/bubblezone"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	dctx "github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
	BuiltBy = ""
)

var (
	cfgFlag string

	logo = lipgloss.NewStyle().Foreground(dctx.LogoColor).MarginBottom(1).SetString(constants.Logo)

	rootCmd = &cobra.Command{
		Use: "gh dash",
		Long: lipgloss.JoinVertical(lipgloss.Left, logo.Render(),
			"A rich terminal UI for GitHub that doesn't break your flow.",
			"Visit https://gh-dash.dev for the docs."),
		Short:   "A rich terminal UI for GitHub that doesn't break your flow.",
		Version: "",
		Example: `
# Running without arguments will either:
#   - Use the global configuration file
#   - Use a local .gh-dash.yml file if in a git repo
gh dash

# Run with a specific configuration file
gh dash --config /path/to/configuration/file.yml

# Run with debug logging to debug.log
gh dash --debug

# Print version
gh dash -v
	`,
		Args: cobra.MaximumNArgs(1),
	}
)

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(rootCmd.Version),
		fang.WithoutCompletions(), fang.WithoutManpage()); err != nil {
		os.Exit(1)
	}
}

func createModel(location config.Location, debug bool) (tui.Model, *os.File) {
	var loggerFile *os.File

	if debug {
		var fileErr error
		newConfigFile, fileErr := os.OpenFile("debug.log",
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if fileErr == nil {
			log.SetOutput(newConfigFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)
			log.Debug("Logging to debug.log")
			if location.RepoPath != "" {
				log.Debug("Running in repo", "repo", location.RepoPath)
			}
		} else {
			loggerFile, _ = tea.LogToFile("debug.log", "debug")
			slog.Print("Failed setting up logging", fileErr)
		}
	} else {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.FatalLevel)
	}

	return tui.NewModel(location), loggerFile
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
		&cfgFlag,
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

	rootCmd.Flags().String(
		"cpuprofile",
		"",
		"write cpu profile to file",
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

		zone.NewGlobal()

		// see https://github.com/charmbracelet/lipgloss/issues/73
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
		markdown.InitializeMarkdownStyle(termenv.HasDarkBackground())

		model, logger := createModel(config.Location{RepoPath: repo, ConfigFlag: cfgFlag}, debug)
		if logger != nil {
			defer logger.Close()
		}

		cpuprofile, err := rootCmd.Flags().GetString("cpuprofile")
		if err != nil {
			log.Fatal("Cannot parse cpuprofile flag", err)
		}
		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				log.Fatal(err)
			}
			_ = pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithReportFocus(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			log.Fatal("Failed starting the TUI", err)
		}
	}
}
