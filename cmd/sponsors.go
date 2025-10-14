/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// sponsorsCmd represents the sponsors command
var sponsorsCmd = &cobra.Command{
	Use:   "sponsors",
	Short: "Show the list of current sponsors for gh-dash",
	Long: `Show the list of current sponsors for gh-dash from GitHub Sponsors under https://github.com/sponsors/dlvhdr.
If you enjoy dash and want to help, consider supporting the project with a donation!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sponsors, err := data.FetchSponsors()
		if err != nil {
			return err
		}

		fmt.Print("\n")
		fmt.Print(
			lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("Thank you ❤️ "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Render(
					"to all the current (and past!) sponsors - you rock! 🤘🏽"),
			))
		fmt.Print("\n")
		fmt.Print("To help this project with a donation go to https://github.com/sponsors/dlvhdr\n")
		fmt.Print("\n")
		for _, sponsor := range sponsors.User.Sponsors.Nodes {
			if sponsor.Typename == "User" {
				fmt.Printf("  • %s (%s)\n", lipgloss.NewStyle().Bold(true).Render(
					fmt.Sprintf("@%s", sponsor.User.Login)), sponsor.User.Url)
			} else {
				fmt.Printf("  • %s (%s)\n", lipgloss.NewStyle().Bold(true).Render(
					sponsor.Organization.Name), sponsor.Organization.Url)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sponsorsCmd)
}
