package dashboard

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components"
	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/sectioncard"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	ctx            *context.ProgramContext
	boards         []board
	focusedCardId  int
	focusedBoardId int
}

type board struct {
	cards []*sectioncard.Model
}

func NewModel(ctx context.ProgramContext) Model {
	return Model{ctx: &ctx, boards: []board{}}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:

		board := m.boards[m.focusedBoardId]
		currCard := board.cards[m.focusedCardId]

		switch {

		case key.Matches(msg, keys.Keys.Down):
			if currCard.Section.AtLastItem() {
				m.GoToNextCard()
			} else {
				currCard.Section.NextRow()
			}

		case key.Matches(msg, keys.Keys.Up):
			if currCard.Section.AtFirstItem() {
				m.GoToPrevCard()
			} else {
				currCard.Section.PrevRow()
			}
		case key.Matches(msg, keys.Keys.NextCard):
			m.GoToNextCard()

		case key.Matches(msg, keys.Keys.PrevCard):
			m.GoToPrevCard()

		case key.Matches(msg, keys.Keys.PrevSection):
			m.previousBoard()

		case key.Matches(msg, keys.Keys.NextSection):
			m.nextBoard()
		}

	case common.ConfigReadMsg:
		fetchSectionsCmds := m.fetchAllDashboards()
		cmd = fetchSectionsCmds

	case section.SectionDataFetchedMsg:
		card := m.getCardById(msg.Id)
		if msg.Type == string(config.PRsView) {
			card.SetPrs(prssection.SectionPullRequestsFetchedMsg{
				Prs:        msg.Prs,
				TotalCount: msg.TotalCount,
				PageInfo:   msg.PageInfo,
			})
		}
	}
	return m, cmd
}

func (m Model) View() string {
	cards := m.boards[m.focusedBoardId].cards
	elements := make([]string, len(cards))
	for i := 0; i < len(cards); i++ {
		elements[i] = cards[i].View()
	}
	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	return lipgloss.PlaceVertical(
		m.ctx.MainContentHeight,
		lipgloss.Top,
		content,
	)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx

	for i := 0; i < len(m.boards); i++ {
		for j := 0; j < len(m.boards[i].cards); j++ {
			m.boards[i].cards[j].UpdateProgramContext(ctx)
		}
	}
}

func (m *Model) fetchAllDashboards() tea.Cmd {
	var cmds []tea.Cmd

	sectionsSoFar := 0
	for i := 0; i < len(m.ctx.Config.Dashboards); i++ {
		var cards []*sectioncard.Model
		numSections := len(m.ctx.Config.Dashboards[i].Sections)
		for j := 0; j < numSections; j++ {
			sectionConfig := m.ctx.Config.Dashboards[i].Sections[j]
			sectionModel := section.Model{
				BaseModel: section.BaseModel{
					Id:     sectionsSoFar + j,
					Ctx:    m.ctx,
					Type:   "prs",
					Config: sectionConfig,
				},
			}
			if strings.Contains(sectionModel.Config.Filters, "is:pr") {
				cmds = append(cmds, sectionModel.FetchSectionRows(nil)...)
			}

			prsSection := prssection.NewModel(
				sectionModel.Id,
				m.ctx,
				config.PrsSectionConfig{
					Title:   sectionModel.Config.Title,
					Filters: sectionModel.Config.Filters,
					Layout:  config.PrsLayoutConfig{},
				},
				time.Now(),
				true,
			)
			prsSection.Table.SetHeaderHidden(true)
			prsSection.Table.SetFooterHidden(true)
			prsSection.Table.EmptyState = nil
			prsSection.Table.SetSeparatorHidden(true)
			prsSection.Table.IsActive = j == 0
			sectionStyle := lipgloss.NewStyle().Padding(0)
			prsSection.Style = &sectionStyle
			cellStyle := m.ctx.Styles.Table.CellStyle.Copy().
				Background(m.ctx.Styles.Card.Content.GetBackground())
			prsSection.Table.CellStyle = &cellStyle

			card := sectioncard.NewModel(m.ctx)
			card.Title = sectionModel.Config.Title
			card.Subtitle = "(loading...)"
			card.Section = &prsSection

			card.UpdateProgramContext(m.ctx)

			cards = append(cards, &card)
		}
		m.boards = append(m.boards, board{
			cards: cards,
		})
		sectionsSoFar += numSections
	}

	return tea.Batch(cmds...)
}

func (m *Model) GoToNextCard() {
	cards := m.boards[m.focusedBoardId].cards
	currCard := cards[m.focusedCardId]
	currCard.Section.Unfocus()

	m.focusedCardId = utils.Min(len(cards)-1, m.focusedCardId+1)
	nextCard := cards[m.focusedCardId]
	nextCard.Section.Focus()
	nextCard.Section.FirstItem()
}

func (m *Model) GoToPrevCard() {
	cards := m.boards[m.focusedBoardId].cards
	currCard := cards[m.focusedCardId]
	currCard.Section.Unfocus()

	m.focusedCardId = utils.Max(0, m.focusedCardId-1)
	nextCard := cards[m.focusedCardId]
	nextCard.Section.LastItem()
	nextCard.Section.Focus()
}

func (m *Model) previousBoard() {
	m.focusedBoardId = components.GetPrevCyclicItem(
		m.focusedBoardId,
		len(m.boards),
	)
}

func (m *Model) nextBoard() {
	m.focusedBoardId = components.GetNextCyclicItem(
		m.focusedBoardId,
		len(m.boards),
	)
}

func (m *Model) getCardById(id int) (card *sectioncard.Model) {
	i := 0
	cardsSoFar := 0
	for i < len(m.boards) {
		numCards := len(m.boards[i].cards)
		if cardsSoFar+numCards <= id {
			i++
			cardsSoFar += numCards
			continue
		}
		break
	}

	return m.boards[i].cards[id-cardsSoFar]
}
