/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui"
	"github.com/dlvhdr/gh-dash/ui/markdown"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
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
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func createModel(configPath string, debug bool) (ui.Model, *os.File) {
	var loggerFile *os.File
	var err error

	if debug {
		loggerFile, err = tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("Error setting up logger")
		}
	}

	return ui.NewModel(configPath), loggerFile
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
		"use this configuration file (default is $XDG_CONFIG_HOME/gh-dash/config.yml)",
	)
	rootCmd.MarkFlagFilename("config", "yaml", "yml")

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

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		debug, err := rootCmd.Flags().GetBool("debug")
		if err != nil {
			log.Fatal(err)
		}

		// see https://github.com/charmbracelet/lipgloss/issues/73
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
		markdown.InitializeMarkdownStyle(termenv.HasDarkBackground())

		model, logger := createModel(cfgFile, debug)
		if logger != nil {
			defer logger.Close()
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
		)
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
