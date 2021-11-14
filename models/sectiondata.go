package models

import (
	"dlvhdr/gh-prs/config"
	"dlvhdr/gh-prs/msgs"
	"fmt"

	"dlvhdr/gh-prs/ui"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ui.Indigo.Dark))
)

type SectionData struct {
	Id        int
	Config    config.Section
	Prs       []PullRequest
	Spinner   SectionSpinner
	Paginator paginator.Model
}

type SectionSpinner struct {
	Model           spinner.Model
	NumReposFetched int
}

func (section *SectionData) NumPrs() int {
	return len(section.Prs)
}

func (section *SectionData) Tick(spinnerTickCmd tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return msgs.TickMsg{
			SectionId:       section.Id,
			InternalTickMsg: spinnerTickCmd(),
		}
	}
}

func (section *SectionData) FetchSectionPullRequests() []tea.Cmd {
	var cmds []tea.Cmd
	for _, repo := range section.Config.Repos {
		repo := repo
		cmds = append(cmds, func() tea.Msg {
			fetched, err := FetchRepoPullRequests(repo, section.Config.Filters)
			if err != nil {
				return RepoPullRequestsFetched{
					SectionId: section.Id,
					RepoName:  repo,
					Prs:       []PullRequest{},
				}
			}

			return RepoPullRequestsFetched{
				SectionId: section.Id,
				RepoName:  repo,
				Prs:       fetched,
			}
		})
	}

	return cmds
}

func (section *SectionData) RenderLoadingState() string {
	return fmt.Sprintf(
		"%s %d/%d fetched...\n",
		section.Spinner.Model.View(),
		section.Spinner.NumReposFetched,
		len(section.Config.Repos),
	)
}

func (section *SectionData) RenderTitle() string {
	sectionTitle := titleStyle.Render(section.Config.Title)
	return fmt.Sprintf(sectionTitle + "\n")
}
