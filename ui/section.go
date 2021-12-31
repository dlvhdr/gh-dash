package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/data"
)

type section struct {
	Id        int
	Config    config.SectionConfig
	Prs       []PullRequest
	Spinner   spinner.Model
	IsLoading bool
	Paginator paginator.Model
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

func (section *section) fetchSectionPullRequests() tea.Cmd {
	return func() tea.Msg {
		fetched, err := data.FetchRepoPullRequests(section.Config.Filters)
		if err != nil {
			return repoPullRequestsFetchedMsg{
				SectionId: section.Id,
				Prs:       []PullRequest{},
			}
		}

		prs := make([]PullRequest, 0, len(fetched))
		for _, prData := range fetched {
			prs = append(prs, PullRequest{
				Data: prData,
			})
		}

		return repoPullRequestsFetchedMsg{
			SectionId: section.Id,
			Prs:       prs,
		}
	}
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
	if !section.IsLoading {
		return ""
	}
	return spinnerStyle.Render(fmt.Sprintf("%s Fetching Pull Requests...", section.Spinner.View()))
}

func (section *section) renderEmptyState() string {
	emptyState := emptyStateStyle.Render(fmt.Sprintf(
		"No PRs were found that match the given filters: %s",
		section.Config.Filters,
	))
	return fmt.Sprintf(emptyState + "\n")
}

func getTitleWidth(viewportWidth int) int {
	return viewportWidth - usedWidth
}

func (m Model) renderTableHeader() string {
	reviewCell := singleRuneTitleCellStyle.Copy().Width(reviewCellWidth).Render("")
	mergeableCell := singleRuneTitleCellStyle.Copy().Width(mergeableCellWidth).Render("")
	ciCell := titleCellStyle.Copy().Width(ciCellWidth).Render("CI")
	linesCell := titleCellStyle.Copy().Width(linesCellWidth).Render("Lines")
	prAuthorCell := titleCellStyle.Copy().Width(prAuthorCellWidth).Render("Author")
	prRepoCell := titleCellStyle.Copy().Width(prRepoCellWidth).Render("Repo")
	updatedAtCell := titleCellStyle.Copy().Width(updatedAtCellWidth).Render(" Updated")

	prTitleCell := titleCellStyle.
		Copy().
		Width(getTitleWidth(m.viewport.Width)).
		MaxWidth(getTitleWidth(m.viewport.Width)).
		Render("Title")

	return headerStyle.
		Width(m.viewport.Width).
		MaxWidth(m.viewport.Width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				updatedAtCell,
				reviewCell,
				prRepoCell,
				prTitleCell,
				prAuthorCell,
				mergeableCell,
				ciCell,
				linesCell,
			),
		)
}

func (m Model) renderPullRequestList() string {
	section := m.getCurrSection()
	if len(section.Prs) == 0 {
		return fmt.Sprintf("%s\n", section.renderEmptyState())
	}

	s := strings.Builder{}
	var renderedPRs []string
	for prId, pr := range section.Prs {
		isSelected := m.cursor.currSectionId == section.Id && m.cursor.currPrId == prId
		renderedPRs = append(renderedPRs, pr.render(isSelected, m.viewport.Width))
	}

	s.WriteString(lipgloss.NewStyle().Height(m.viewport.Height).Render(lipgloss.JoinVertical(lipgloss.Left, renderedPRs...)))
	return s.String()
}

func (m Model) renderCurrentSection() string {
	section := m.getCurrSection()
	if section.IsLoading {
		return lipgloss.NewStyle().
			Height(m.viewport.Height).
			Render(section.renderLoadingState())
	}

	return lipgloss.NewStyle().
		PaddingLeft(mainContentPadding).
		PaddingRight(mainContentPadding).
		Render(m.viewport.View())
}

func (section section) numPrs() int {
	return len(section.Prs)
}
