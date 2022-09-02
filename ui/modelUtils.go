package ui

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/constants"
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
	repoPath := m.ctx.Config.RepoPaths[repoName]

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

	// job := exec.Command("/bin/sh", "-c", buff.String())
	// out, err := job.CombinedOutput()
	// if err != nil {
	// 	log.Fatalf("Got an error while executing command: %s. \nError: %v\n%s", job, err, out)
	// }
}
