package ui

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
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

// support [user|org]/* matching for repositories
// and local path mapping to [partial path prefix]/*
// prioritize full repo mapping if it exists
func getRepoLocalPath(repoName string, cfgPaths map[string]string) string {
	exactMatchPath, ok := cfgPaths[repoName]
	// prioritize full repo to path mapping in config
	if ok {
		return exactMatchPath
	}

	var repoPath string

	owner, repo, repoValid := func() (string, string, bool) {
		repoParts := strings.Split(repoName, "/")
		// return repo owner, repo, and indicate properly owner/repo format
		return repoParts[0], repoParts[len(repoParts)-1], len(repoParts) == 2
	}()

	if repoValid {
		// match config:repoPath values of {owner}/* as map key
		wildcardPath, wildcarded := cfgPaths[fmt.Sprintf("%s/*", owner)]

		if wildcarded {
			// adjust wildcard match to wildcard path - ~/somepath/* to ~/somepath/{repo}
			repoPath = fmt.Sprintf("%s/%s", strings.TrimSuffix(wildcardPath, "/*"), repo)
		}
	}

	return repoPath
}

type CommandTemplateInput struct {
	RepoName    string
	RepoPath    string
	PrNumber    int
	HeadRefName string
}

func (m *Model) executeKeybinding(key string) tea.Cmd {
	currRowData := m.getCurrRowData()
	for _, keybinding := range m.ctx.Config.Keybindings.Prs {
		if keybinding.Key != key {
			continue
		}

		switch data := currRowData.(type) {
		case *data.PullRequestData:
			return m.runCustomCommand(keybinding.Command, data)
		}
	}
	return nil
}

func (m *Model) runCustomCommand(commandTemplate string, prData *data.PullRequestData) tea.Cmd {
	cmd, err := template.New("keybinding_command").Parse(commandTemplate)
	if err != nil {
		log.Fatal(err)
	}
	repoName := prData.GetRepoNameWithOwner()
	repoPath := getRepoLocalPath(repoName, m.ctx.Config.RepoPaths)

	var buff bytes.Buffer
	err = cmd.Execute(&buff, CommandTemplateInput{
		RepoName:    repoName,
		RepoPath:    repoPath,
		PrNumber:    prData.Number,
		HeadRefName: prData.HeadRefName,
	})
	if err != nil {
		log.Fatal(err)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	c := exec.Command(shell, "-c", buff.String())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			mdRenderer := markdown.GetMarkdownRenderer(m.ctx.ScreenWidth)
			md, mdErr := mdRenderer.Render(fmt.Sprintf("While running: `%s`", buff.String()))
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
	m.tasks[id] = context.Task{
		Id:           id,
		FinishedText: text,
		State:        context.TaskFinished,
	}
	return func() tea.Msg {
		return constants.TaskFinishedMsg{
			SectionId:   m.getCurrSection().GetId(),
			SectionType: m.getCurrSection().GetType(),
			TaskId:      id,
			Err:         nil,
		}
	}
}
