package ui

import (
	"dlvhdr/gh-prs/utils"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	JsonFields = "title,mergeable,author,url,additions,deletions,headRefName,headRepository,isDraft,number,reviewDecision,state,statusCheckRollup,updatedAt"
	Limit      = "20"
)

type PullRequest struct {
	Number            int
	Title             string
	Author            Author
	UpdatedAt         time.Time
	Url               string
	State             string
	Mergeable         string
	ReviewDecision    string
	Additions         int
	Deletions         int
	HeadRefName       string
	HeadRepository    Repository
	IsDraft           bool
	StatusCheckRollup []StatusCheck
}

type Author struct {
	Login string
}

type Repository struct {
	Id   string
	Name string
}

type StatusCheck struct {
	__typename  string
	Name        string
	Context     string
	State       string
	Status      string
	Conclusion  string
	StartedAt   string
	CompletedAt string
	DetailsUrl  string
	TargetUrl   string
}

type repoPullRequestsFetchedMsg struct {
	SectionId int
	RepoName  string
	Prs       []PullRequest
}

func fetchRepoPullRequests(repo string, search string) ([]PullRequest, error) {
	out, err := exec.Command(
		"gh",
		"pr",
		"list",
		"--repo",
		repo,
		"--json",
		JsonFields,
		"--search",
		search,
		"--limit",
		Limit,
	).Output()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	prs := []PullRequest{}
	if err := json.Unmarshal(out, &prs); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return prs, nil
}

func (pr PullRequest) renderReviewStatus() string {
	if pr.ReviewDecision == "APPROVED" {
		return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("42")).Render("")
	}

	if pr.ReviewDecision == "CHANGES_REQUESTED" {
		return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("196")).Render("")
	}

	return SingleRuneCellStyle.Copy().Faint(true).Render("")
}

func (pr PullRequest) renderMergeableStatus() string {
	if pr.Mergeable == "MERGEABLE" {
		return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("42")).Render("")
	}

	return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("196")).Render("")
}

func (pr PullRequest) renderCiStatus() string {
	accStatus := "SUCCESS"
	for _, statusCheck := range pr.StatusCheckRollup {
		if statusCheck.State == "FAILURE" {
			accStatus = "FAILURE"
			break
		}

		if statusCheck.State == "PENDING" {
			accStatus = "PENDING"
		}
	}
	if accStatus == "SUCCESS" {
		return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("42")).Render("")
	}

	if accStatus == "PENDING" {
		return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("214")).Render("")
	}

	return SingleRuneCellStyle.Copy().Foreground(lipgloss.Color("196")).Render("")
}

func (pr PullRequest) renderLines() string {
	separator := lipgloss.NewStyle().Faint(true).MarginLeft(1).MarginRight(1).Render("/")
	added := lipgloss.NewStyle().Render(fmt.Sprintf("%5d", pr.Additions))
	removed := lipgloss.NewStyle().Render(
		fmt.Sprintf("-%-5d", pr.Deletions),
	)

	return CellStyle.Copy().Width(13).Render(lipgloss.JoinHorizontal(lipgloss.Center, added, separator, removed))
}

func (pr PullRequest) renderTitle() string {
	title := lipgloss.NewStyle().Width(44).MaxWidth(44).Render(pr.Title)
	number := lipgloss.NewStyle().Width(6).Faint(true).Render(
		fmt.Sprintf("#%s", fmt.Sprintf("%d", pr.Number)),
	)

	return CellStyle.Copy().Width(50).MaxWidth(50).Render(lipgloss.JoinHorizontal(lipgloss.Left, title, number))
}

func (pr PullRequest) renderAuthor() string {
	return CellStyle.Render(fmt.Sprintf("%-15s", utils.TruncateString(pr.Author.Login, 15)))
}

func (pr PullRequest) renderRepoName() string {
	return CellStyle.Render(fmt.Sprintf("%-20s", utils.TruncateString(pr.HeadRepository.Name, 20)))
}

func (pr PullRequest) renderUpdateAt() string {
	return CellStyle.Render(utils.TimeElapsed(pr.UpdatedAt))
}

func renderSelectionPointer(isSelected bool) string {
	return SingleRuneCellStyle.Render(func() string {
		if isSelected {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Render("")
		} else {
			return " "
		}
	}())

}

func (pr PullRequest) render(isSelected bool) string {
	selectionPointerCell := renderSelectionPointer(isSelected)
	reviewCell := pr.renderReviewStatus()
	mergeableCell := pr.renderMergeableStatus()
	ciCell := pr.renderCiStatus()
	linesCell := pr.renderLines()
	prTitleCell := pr.renderTitle()
	prAuthorCell := pr.renderAuthor()
	prRepoCell := pr.renderRepoName()
	updatedAtCell := pr.renderUpdateAt()

	return PullRequestStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left,
		selectionPointerCell,
		reviewCell,
		prTitleCell,
		mergeableCell,
		ciCell,
		linesCell,
		prAuthorCell,
		prRepoCell,
		updatedAtCell,
	))
}
