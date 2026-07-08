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
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/fang"
	"github.com/cli/go-gh/v2/pkg/repository"
	zone "github.com/lrstanley/bubblezone/v2"
	"github.com/spf13/cobra"

	gitm "github.com/aymanbagabas/git-module"
	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	dctx "github.com/dlvhdr/gh-dash/v4/internal/tui/context"
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
		Long: lipgloss.JoinVertical(
			lipgloss.Left,
			logo.Render(),
			"A rich terminal UI for GitHub that doesn't break your flow.",
			"",
			lipgloss.NewStyle().
				Faint(true).
				Italic(true).
				Render("Visit https://gh-dash.dev for the docs."),
		),
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
	themeFunc := fang.WithColorSchemeFunc(func(
		ld lipgloss.LightDarkFunc,
	) fang.ColorScheme {
		c := ld(lipgloss.Color("#00196F"), lipgloss.Color("#02F9FB"))
		def := fang.DefaultColorScheme(ld)
		def.DimmedArgument = ld(lipgloss.Black, lipgloss.White)
		def.Codeblock = lipgloss.Color("#1E1E2C")
		def.Title = c
		def.Flag = lipgloss.Color("#42A0FA")
		def.Command = c
		def.Program = c
		return def
	})
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		themeFunc,
		fang.WithVersion(rootCmd.Version),
		fang.WithoutCompletions(),
		fang.WithoutManpage(),
	); err != nil {
		os.Exit(1)
	}
}

func setDebugLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
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
		result = fmt.Sprintf(
			"%s\nmodule version: %s, checksum: %s",
			result,
			info.Main.Version,
			info.Main.Sum,
		)
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
	rootCmd.SetVersionTemplate(
		lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			logo.Render(),
			`gh-dash {{printf "version %s\n" .Version}}`,
		),
	)

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
		debug, err := rootCmd.Flags().GetBool("debug")
		var loggerFile *os.File
		if debug {
			var fileErr error
			loggerFile, fileErr = os.OpenFile("debug.log",
				os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
			if fileErr == nil {
				log.SetOutput(loggerFile)
				log.SetTimeFormat(time.Kitchen)
				log.SetReportCaller(true)
				setDebugLogLevel()
				log.Info("Logging to debug.log")
			} else {
				loggerFile, _ = tea.LogToFile("debug.log", "debug")
				slog.Print("Failed setting up logging", fileErr)
			}
		} else {
			log.SetOutput(os.Stderr)
			log.SetLevel(log.FatalLevel)
		}
		if loggerFile != nil {
			defer loggerFile.Close()
		}

		var gitRepoPath string
		gitRepo, ghRepo, err := getCurrentGitAndGitHubRepos()
		if err != nil {
			log.Error("error while determining git and github repos", "err", err)
		}

		if gitRepo != nil {
			log.Info("found git repo at path", "path", gitRepo.Path())
			gitRepoPath = gitRepo.Path()
		} else {
			log.Warn("did not find git repo at current path")
		}

		if ghRepo != (repository.Repository{}) {
			log.Info(
				"found github repo at current path",
				"host",
				ghRepo.Host,
				"owner",
				ghRepo.Owner,
				"name",
				ghRepo.Name,
			)
		} else {
			log.Warn("did not find github repo at current path")
		}

		if err != nil {
			log.Fatal("Cannot parse debug flag", err)
		}

		zone.NewGlobal()

		model := tui.NewModel(
			config.Location{RepoPath: gitRepoPath, ConfigFlag: cfgFlag},
			tui.Repositories{GitRepo: gitRepo, GHRepo: &ghRepo},
		)

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

		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			fmt.Printf("%+v\n", err)
			log.Fatal("fatal error during run", "err", err)
		}
	}
}

func getCurrentGitAndGitHubRepos() (*gitm.Repository, repository.Repository, error) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gitRepo *gitm.Repository
	var ghRepo repository.Repository
	var gitErr, ghErr error
	var wg sync.WaitGroup

	wg.Go(func() {
		gitRepo, gitErr = git.GetRepoInPwd()
		if gitErr != nil {
			cancel() // Abort the context, so the other function can abort early
		}
	})

	wg.Go(func() {
		ghRepo, ghErr = repository.Current()
		if ghErr != nil {
			cancel() // Abort the context, so the other function can abort early
		}
	})

	wg.Wait()

	if gitErr == context.Canceled || gitErr == nil {
		return gitRepo, ghRepo, ghErr
	}
	return gitRepo, ghRepo, gitErr
}
