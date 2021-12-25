package ui

import (
	"fmt"
	"strings"

	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/data"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	if len(section.Prs) == 0 {
		return fmt.Sprintf("%s\n", renderEmptyState())
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
