package ui

import (
	"dlvhdr/gh-prs/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type section struct {
	Id        int
	Config    config.SectionConfig
	Prs       []PullRequest
	Spinner   sectionSpinner
	Paginator paginator.Model
}

type sectionSpinner struct {
	Model           spinner.Model
	NumReposFetched int
}

type tickMsg struct {
	SectionId       int
	InternalTickMsg tea.Msg
}

func (section *section) Tick(spinnerTickCmd tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return tickMsg{
			SectionId:       section.Id,
			InternalTickMsg: spinnerTickCmd(),
		}
	}
}

func (section *section) fetchSectionPullRequests() []tea.Cmd {
	var cmds []tea.Cmd
	for _, repo := range section.Config.Repos {
		repo := repo
		cmds = append(cmds, func() tea.Msg {
			fetched, err := fetchRepoPullRequests(repo, section.Config.Filters)
			if err != nil {
				return repoPullRequestsFetchedMsg{
					SectionId: section.Id,
					RepoName:  repo,
					Prs:       []PullRequest{},
				}
			}

			return repoPullRequestsFetchedMsg{
				SectionId: section.Id,
				RepoName:  repo,
				Prs:       fetched,
			}
		})
	}

	return cmds
}

func (m Model) makeRenderPullRequestCmd(sectionId int) tea.Cmd {
	return func() tea.Msg {
		return pullRequestsRenderedMsg{
			sectionId: sectionId,
			content:   m.renderPullRequestList(),
		}
	}
}

func (section *section) renderLoadingState() string {
	return fmt.Sprintf(
		"%s %d/%d fetched...\n",
		section.Spinner.Model.View(),
		section.Spinner.NumReposFetched,
		len(section.Config.Repos),
	)
}

func renderEmptyState() string {
	emptyState := emptyStateStyle.Render("No PRs were found that match the given filters...")
	return fmt.Sprintf(emptyState + "\n")
}

func getTitleWidth(viewportWidth int) int {
	return viewportWidth - usedWidth - cellPadding - 5
}

func (m Model) renderTableHeader() string {
	emptyCell := singleRuneCellStyle.Copy().Bold(true).Width(emptyCellWidth).Render(" ")
	reviewCell := singleRuneCellStyle.Copy().Bold(true).Width(reviewCellWidth).Render("")
	mergeableCell := singleRuneCellStyle.Copy().Bold(true).Width(mergeableCellWidth).Render("")
	ciCell := cellStyle.Copy().Bold(true).Width(ciCellWidth + 1).Render("CI")
	linesCell := cellStyle.Copy().Bold(true).Width(linesCellWidth).Render("Lines")
	prAuthorCell := cellStyle.Copy().Bold(true).Width(prAuthorCellWidth).Render("Author")
	prRepoCell := cellStyle.Copy().Bold(true).Width(prRepoCellWidth).Render("Repo")
	updatedAtCell := cellStyle.Copy().Bold(true).Width(updatedAtCellWidth).Render("Updated At")

	prTitleCell := cellStyle.
		Copy().
		Bold(true).
		Width(getTitleWidth(m.viewport.Width)).
		MaxWidth(getTitleWidth(m.viewport.Width)).
		Render("Title")

	return headerStyle.
		Width(m.viewport.Width - mainContentPadding).
		MaxWidth(m.viewport.Width - mainContentPadding).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				emptyCell,
				reviewCell,
				prTitleCell,
				mergeableCell,
				ciCell,
				linesCell,
				prAuthorCell,
				prRepoCell,
				updatedAtCell,
			),
		)
}

func (m Model) renderPullRequestList() string {
	section := m.getCurrSection()
	isLoading := section.Spinner.NumReposFetched < len(section.Config.Repos)
	if isLoading {
		return section.renderLoadingState() + "\n"
	}

	if len(section.Prs) == 0 {
		return renderEmptyState() + "\n"
	}

	s := strings.Builder{}
	var renderedPRs []string
	for prId, pr := range section.Prs {
		isSelected := m.cursor.currSectionId == section.Id && m.cursor.currPrId == prId
		renderedPRs = append(renderedPRs, pr.render(isSelected, m.viewport.Width))
	}

	s.WriteString(lipgloss.JoinVertical(lipgloss.Left, renderedPRs...))
	return s.String()
}

func (section section) numPrs() int {
	return len(section.Prs)
}
