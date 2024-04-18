package ui

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/markdown"
)

func (m *Model) getCurrSection() section.Section {
	sections := m.getCurrentViewSections()
	if len(sections) == 0 || m.currSectionId >= len(sections) {
		return nil
	}
	return sections[m.currSectionId]
}

func (m *Model) getCurrRowData() data.RowData {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return section.GetCurrRow()
}

func (m *Model) getSectionAt(id int) section.Section {
	sections := m.getCurrentViewSections()
	if len(sections) <= id {
		return nil
	}
	return sections[id]
}

func (m *Model) getPrevSectionId() int {
	sectionsConfigs := m.ctx.GetViewSectionsConfig()
	m.currSectionId = (m.currSectionId - 1) % len(sectionsConfigs)
	if m.currSectionId < 0 {
		m.currSectionId += len(sectionsConfigs)
	}

	return m.currSectionId
}

func (m *Model) getNextSectionId() int {
	return (m.currSectionId + 1) % len(m.ctx.GetViewSectionsConfig())
}

type IssueCommandTemplateInput struct {
	RepoName    string
	RepoPath    string
	IssueNumber int
	HeadRefName string
}

func (m *Model) executeKeybinding(key string) tea.Cmd {
	currRowData := m.getCurrRowData()

	switch m.ctx.View {
	case config.IssuesView:
		for _, keybinding := range m.ctx.Config.Keybindings.Issues {
			if keybinding.Key != key {
				continue
			}

			switch data := currRowData.(type) {
			case *data.IssueData:
				return m.runCustomIssueCommand(keybinding.Command, data)
			}
		}
	case config.PRsView:
		for _, keybinding := range m.ctx.Config.Keybindings.Prs {
			if keybinding.Key != key {
				continue
			}

			switch data := currRowData.(type) {
			case *data.PullRequestData:
				return m.runCustomPRCommand(keybinding.Command, data)
			}
		}
	default:
		// Not a valid case - ignore it
	}

	return nil
}

func (m *Model) runCustomPRCommand(commandTemplate string, prData *data.PullRequestData) tea.Cmd {
	// A generic map is a pretty easy & flexible way to populate a template if there's no pressing need
	// for sructured data, existing structs, etc. Especially if holes in the data are expected.
	input := map[string]any {
		"RepoName":    prData.GetRepoNameWithOwner(),
		"PrNumber":    prData.Number,
		"HeadRefName": prData.HeadRefName,
		"BaseRefName": prData.BaseRefName,
	}

	// Append in the local RepoPath only if it can be found
	if repoPath, ok := common.GetRepoLocalPath(input["RepoName"].(string), m.ctx.Config.RepoPaths); ok {
		input["RepoPath"] = repoPath
	}

	cmd, err := template.New("keybinding_command").Parse(commandTemplate)
	if err != nil {
		log.Fatal("Failed parse keybinding template", err)
	}

	// Set the command to error out if required input (e.g. RepoPath) is missing
	cmd = cmd.Option("missingkey=error")

	var buff bytes.Buffer
	err = cmd.Execute(&buff, input)
	if err != nil {
		log.Fatal("Failed executing keybinding command", err)
	}
	return m.executeCustomCommand(buff.String())
}

func (m *Model) runCustomIssueCommand(commandTemplate string, issueData *data.IssueData) tea.Cmd {
	repoName := issueData.GetRepoNameWithOwner()
	repoPath, ok := common.GetRepoLocalPath(repoName, m.ctx.Config.RepoPaths)

	if !ok {
		return func() tea.Msg {
			return constants.ErrMsg{Err: fmt.Errorf("Failed to find local path for repo %s", repoName)}
		}
	}

	input := IssueCommandTemplateInput{
		RepoName:    repoName,
		RepoPath:    repoPath,
		IssueNumber: issueData.Number,
	}

	cmd, err := template.New("keybinding_command").Parse(commandTemplate)
	if err != nil {
		log.Fatal("Failed parse keybinding template", err)
	}

	var buff bytes.Buffer
	err = cmd.Execute(&buff, input)
	if err != nil {
		log.Fatal("Failed executing keybinding command", err)
	}
	return m.executeCustomCommand(buff.String())
}

func (m *Model) executeCustomCommand(cmd string) tea.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	c := exec.Command(shell, "-c", cmd)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			mdRenderer := markdown.GetMarkdownRenderer(m.ctx.ScreenWidth)
			md, mdErr := mdRenderer.Render(fmt.Sprintf("While running: `%s`", cmd))
			if mdErr != nil {
				return constants.ErrMsg{Err: mdErr}
			}
			return constants.ErrMsg{Err: fmt.Errorf(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("Whoops, got an error: %s", err),
					md,
				),
			)}
		}
		return nil
	})
}

func (m *Model) notify(text string) tea.Cmd {
	id := fmt.Sprint(time.Now().Unix())
	startCmd := m.ctx.StartTask(
		context.Task{
			Id:           id,
			StartText:    text,
			FinishedText: text,
			State:        context.TaskStart,
		})

	finishCmd := func() tea.Msg {
		return constants.TaskFinishedMsg{
			TaskId: id,
		}
	}

	return tea.Sequence(startCmd, finishCmd)
}
