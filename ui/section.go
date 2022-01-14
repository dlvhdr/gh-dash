package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/data"
)

type section struct {
	Id        int
	Config    config.PRSectionConfig
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
			return sectionPullRequestsFetchedMsg{
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

		return sectionPullRequestsFetchedMsg{
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
		lipgloss.NewStyle().Italic(true).Render(section.Config.Filters),
	))
	return fmt.Sprintf(emptyState + "\n")
}

func getTitleWidth(viewportWidth int) int {
	return viewportWidth - usedWidth
}

func (m *Model) renderTableHeader() string {
	reviewCell := singleRuneTitleCellStyle.Copy().Width(reviewCellWidth).Render("")
	mergeableCell := singleRuneTitleCellStyle.Copy().Width(mergeableCellWidth).Render("")
	ciCell := titleCellStyle.Copy().Width(ciCellWidth).Render("CI")
	linesCell := titleCellStyle.Copy().Width(linesCellWidth).Render("Lines")
	prAuthorCell := titleCellStyle.Copy().Width(prAuthorCellWidth).Render("Author")
	prRepoCell := titleCellStyle.Copy().Width(prRepoCellWidth).Render("Repo")
	updatedAtCell := titleCellStyle.Copy().Width(updatedAtCellWidth).Render(" Updated")

	prTitleCell := titleCellStyle.
		Copy().
		Width(getTitleWidth(m.mainViewport.model.Width)).
		MaxWidth(getTitleWidth(m.mainViewport.model.Width)).
		Render("Title")

	return headerStyle.
		Width(m.mainViewport.model.Width).
		MaxWidth(m.mainViewport.model.Width).
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
	if section == nil {
		return ""
	}
	if len(section.Prs) == 0 {
		return fmt.Sprintf("%s\n", section.renderEmptyState())
	}

	var renderedPRs []string
	for prId, pr := range section.Prs {
		isSelected := m.cursor.currSectionId == section.Id && m.cursor.currPrId == prId
		renderedPRs = append(renderedPRs, pr.render(isSelected, m.mainViewport.model.Width))
	}

	return lipgloss.NewStyle().Render(lipgloss.JoinVertical(lipgloss.Left, renderedPRs...))
}

func (m *Model) renderCurrentSection() string {
	section := m.getCurrSection()
	if section == nil {
		return ""
	}
	if section.IsLoading {
		return lipgloss.NewStyle().
			Height(m.mainViewport.model.Height + pagerHeight).
			Render(section.renderLoadingState())
	}

	return lipgloss.NewStyle().
		PaddingLeft(mainContentPadding).
		PaddingRight(mainContentPadding).
		MaxWidth(m.mainViewport.model.Width).
		Render(m.RenderMainViewPort())
}

func (section section) numPrs() int {
	return len(section.Prs)
}
