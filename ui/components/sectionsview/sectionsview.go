package sectionsview

import (
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	log "github.com/charmbracelet/log"
	"github.com/cli/go-gh/pkg/browser"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
)

type Model struct {
	ctx           *context.ProgramContext
	prs           []section.Section
	issues        []section.Section
	currSectionId int
	keys          keys.KeyMap
}

func NewModel(ctx *context.ProgramContext) Model {
	m := Model{
		ctx:           ctx,
		keys:          keys.Keys,
		currSectionId: 1,
	}

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd         tea.Cmd
		cmds        []tea.Cmd
		currSection = m.getCurrSection()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.ctx.Error = nil

		if currSection != nil && currSection.IsSearchFocused() {
			cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
			return m, cmd
		}

		switch {
		case m.isUserDefinedKeybinding(msg):
			cmd = m.executeKeybinding(msg.String())
			return m, cmd

		case key.Matches(msg, m.keys.PrevSection):
			prevSection := m.getSectionAt(m.getPrevSectionId())
			if prevSection != nil {
				m.setCurrSectionId(prevSection.GetId())
			}

		case key.Matches(msg, m.keys.NextSection):
			nextSectionId := m.getNextSectionId()
			nextSection := m.getSectionAt(nextSectionId)
			if nextSection != nil {
				m.setCurrSectionId(nextSection.GetId())
			}

		case key.Matches(msg, m.keys.Down):
			prevRow := currSection.CurrRow()
			nextRow := currSection.NextRow()
			if prevRow != nextRow && nextRow == currSection.NumRows()-1 {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}

		case key.Matches(msg, m.keys.Up):
			currSection.PrevRow()

		case key.Matches(msg, m.keys.FirstLine):
			currSection.FirstItem()

		case key.Matches(msg, m.keys.LastLine):
			if currSection.CurrRow()+1 < currSection.NumRows() {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}
			currSection.LastItem()

		case key.Matches(msg, m.keys.OpenGithub):
			var currRow = m.getCurrRowData()
			b := browser.New("", os.Stdout, os.Stdin)
			if currRow != nil {
				err := b.Browse(currRow.GetUrl())
				if err != nil {
					log.Fatal(err)
				}
			}

		case key.Matches(msg, m.keys.Refresh):
			currSection.ResetFilters()
			currSection.ResetRows()
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)

		case key.Matches(msg, m.keys.RefreshAll):
			fetchSectionsCmds := m.FetchAllViewSections(true)
			cmds = append(cmds, fetchSectionsCmds)

		case key.Matches(msg, m.keys.Search):
			if currSection != nil {
				cmd = currSection.SetIsSearching(true)
				return m, cmd
			}

		case key.Matches(msg, m.keys.CopyNumber):
			number := fmt.Sprint(m.getCurrRowData().GetNumber())
			clipboard.WriteAll(number)
			cmd := m.ctx.Notify(fmt.Sprintf("Copied %s to clipboard", number))
			return m, cmd

		case key.Matches(msg, m.keys.CopyUrl):
			url := m.getCurrRowData().GetUrl()
			clipboard.WriteAll(url)
			cmd := m.ctx.Notify(fmt.Sprintf("Copied %s to clipboard", url))
			return m, cmd

		}

	case common.ConfigReadMsg:
		fetchSectionsCmds := m.FetchAllViewSections(true)
		cmd = fetchSectionsCmds

	case section.SectionMsg:
		cmd = m.updateRelevantSection(msg)

	}

	sectionCmd := m.updateCurrentSection(msg)

	cmds = append(
		cmds,
		cmd,
		sectionCmd,
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.getCurrSection().View()
}

func (m *Model) setCurrSectionId(newSectionId int) {
	m.currSectionId = newSectionId
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	switch sType {
	case prssection.SectionType:
		updatedSection, cmd = m.prs[id].Update(msg)
		m.prs[id] = updatedSection
	case issuessection.SectionType:
		updatedSection, cmd = m.issues[id].Update(msg)
		m.issues[id] = updatedSection
	}

	return cmd
}

func (m *Model) updateRelevantSection(msg section.SectionMsg) (cmd tea.Cmd) {
	return m.updateSection(msg.Id, msg.Type, msg.InternalMsg)
}

func (m *Model) updateCurrentSection(msg tea.Msg) (cmd tea.Cmd) {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return m.updateSection(section.GetId(), section.GetType(), msg)
}

func (m *Model) FetchAllViewSections(force bool) tea.Cmd {
	var cmd tea.Cmd
	var sections []section.Section

	if !force && len(m.getCurrentViewSections()) > 0 {
		return nil
	}

	if m.ctx.View == config.PRsView {
		sections, cmd = prssection.FetchAllSections(*m.ctx)
	} else if m.ctx.View == config.IssuesView {
		sections, cmd = issuessection.FetchAllSections(*m.ctx)
	}
	if len(sections) > 0 {
		m.setCurrentViewSections(sections)
	}
	return cmd
}

func (m *Model) getCurrentViewSections() []section.Section {
	if m.ctx.View == config.PRsView {
		return m.prs
	} else {
		return m.issues
	}
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	if m.ctx.View == config.PRsView {
		search := prssection.NewModel(
			0,
			m.ctx,
			config.PrsSectionConfig{
				Title:   "",
				Filters: "archived:false",
			},
			time.Now(),
			false,
		)
		m.prs = append([]section.Section{&search}, newSections...)
	} else {
		search := issuessection.NewModel(
			0,
			m.ctx,
			config.IssuesSectionConfig{
				Title:   "",
				Filters: "",
			},
			time.Now(),
		)
		m.issues = append([]section.Section{&search}, newSections...)
	}
}

func (m *Model) isUserDefinedKeybinding(msg tea.KeyMsg) bool {
	if m.ctx.View == config.IssuesView {
		for _, keybinding := range m.ctx.Config.Keybindings.Issues {
			if keybinding.Key == msg.String() {
				return true
			}
		}
	}

	if m.ctx.View == config.PRsView {
		for _, keybinding := range m.ctx.Config.Keybindings.Prs {
			if keybinding.Key == msg.String() {
				return true
			}
		}
	}

	return false
}

func (m *Model) GetCurrRow() data.RowData {
	currSection := m.getCurrSection()
	if currSection != nil {
		return currSection.GetCurrRow()
	}
	return nil
}

func (m *Model) GetCurrSection() section.Section {
	return m.getCurrSection()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx

	width, height :=
		m.ctx.MainContentWidth-m.ctx.Styles.Section.ContainerStyle.GetHorizontalPadding(),
		m.ctx.MainContentHeight-common.SearchHeight
	for _, section := range m.prs {
		section.UpdateProgramContext(m.ctx)
		section.SetDimensions(width, height)
	}
}

func (m *Model) GoToFirstSection() {
	m.setCurrSectionId(1)
}

func (m *Model) GetItemSingularForm() string {
	return "Issue"
}

func (m *Model) GetItemPluralForm() string {
	return "Issues"
}
