package ui

import (
	"bytes"
	"log"
	"os/exec"
	"text/template"

	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/section"
)

func (m *Model) getCurrSection() section.Section {
	sections := m.getCurrentViewSections()
	if len(sections) == 0 {
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

func (m *Model) executeKeybinding(key string) {
	currRowData := m.getCurrRowData()
	for _, keybinding := range m.ctx.Config.Keybindings.Prs {
		if keybinding.Key != key {
			continue
		}

		switch data := currRowData.(type) {
		case *data.PullRequestData:
			m.runCustomCommand(keybinding.Command, data)
		}
	}
}

func (m *Model) runCustomCommand(commandTemplate string, prData *data.PullRequestData) {
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

	job := exec.Command("/bin/sh", "-c", buff.String())
	out, err := job.CombinedOutput()
	if err != nil {
		log.Fatalf("Got an error while executing command: %s. \nError: %v\n%s", job, err, out)
	}
}
