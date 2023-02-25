package section

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type SectionDataFetchedMsg struct {
	Id         int
	Config     config.SectionConfig
	Type       string
	Prs        []data.PullRequestData
	Issues     []data.IssueData
	TotalCount int
	PageInfo   data.PageInfo
}

func (m *BaseModel) FetchSectionRows(pageInfo *data.PageInfo) []tea.Cmd {
	if pageInfo != nil && !pageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	startCursor := time.Now().String()
	if pageInfo != nil {
		startCursor = pageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_prs_%d_%s", m.Id, startCursor)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf(`Fetching PRs for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(`PRs for "%s" have been fetched`, m.Config.Title),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}

		var err error
		var prsRes *data.PullRequestsResponse
		var issuesRes *data.IssuesResponse
		if strings.Contains(m.Config.Filters, "is:pr") {
			res, prsError := data.FetchPullRequests(m.Config.Filters, *limit, pageInfo)
			prsRes = &res
			err = prsError
		} else {
			res, issuesErr := data.FetchIssues(m.Config.Filters, *limit, pageInfo)
			issuesRes = &res
			err = issuesErr
		}

		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		totalCount, pageInfo := func() (int, data.PageInfo) {
			if prsRes != nil {
				return prsRes.TotalCount, prsRes.PageInfo
			}

			return issuesRes.TotalCount, issuesRes.PageInfo
		}()

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionDataFetchedMsg{
				Id:         m.Id,
				Type:       m.Type,
				Prs:        prsRes.Prs,
				TotalCount: totalCount,
				PageInfo:   pageInfo,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	return cmds
}
