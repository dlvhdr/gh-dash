package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"

	config "dlvhdr/gh-prs/config"
	"dlvhdr/gh-prs/models"
	"dlvhdr/gh-prs/msgs"
	"dlvhdr/gh-prs/ui"
	utils "dlvhdr/gh-prs/utils"
)

var (
	emptyStateStyle = lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(2).
			MarginBottom(1)
	pullRequestStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true)
	selectedPullRequestStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(ui.NoColor.Light)).
					Foreground(lipgloss.Color(ui.SubtleIndigo.Light)).
					Inherit(pullRequestStyle)
	cellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
)

type cursor struct {
	currSectionId int
	currPrId      int
}

type model struct {
	keys            utils.KeyMap
	res             string
	err             error
	sectionsConfigs []config.Section
	data            *[]models.SectionData
	help            help.Model
	cursor          cursor
	ready           bool
	logger          *os.File
}

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Error setting up logger")
	}
	defer f.Close()

	p := tea.NewProgram(
		model{
			keys: utils.Keys,
			help: help.NewModel(),
			cursor: cursor{
				currSectionId: 0,
				currPrId:      0,
			},
			logger: f,
		},
		tea.WithAltScreen(),
	)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(initScreen, tea.EnterAltScreen)
}

func initScreen() tea.Msg {
	sections, err := config.ParseSectionsConfig()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return msgs.InitMsg{Config: sections}
}

const PerPage = 5

func (m model) getCurrSection() *models.SectionData {
	if m.data == nil || len(*m.data) == 0 {
		return nil
	}
	return &(*m.data)[m.cursor.currSectionId]
}

func (m *model) prevPr() {
	currSection := m.getCurrSection()
	if currSection == nil || (m.cursor.currPrId == 0 && m.cursor.currSectionId == 0) {
		return
	}

	newPrId := utils.Max(m.cursor.currPrId-1, 0)
	if newPrId%PerPage == PerPage-1 {
		currSection.Paginator.PrevPage()
	}

	if m.cursor.currPrId == 0 {
		prevSectionId := m.getPrevSectionId()
		m.cursor = cursor{
			currPrId:      utils.Max(0, m.getSectionAt(prevSectionId).NumPrs()-1),
			currSectionId: m.getPrevSectionId(),
		}
		return
	}

	m.cursor.currPrId = newPrId
}

func (m *model) nextPr() {
	currSection := m.getCurrSection()
	if currSection == nil ||
		(m.cursor.currSectionId == m.numSections()-1 && m.cursor.currPrId == currSection.NumPrs()-1) {
		return
	}

	newPrId := utils.Min(m.cursor.currPrId+1, currSection.NumPrs()-1)
	if newPrId != 0 && newPrId%PerPage == 0 {
		currSection.Paginator.NextPage()
	}

	if m.cursor.currPrId == currSection.NumPrs()-1 {
		m.cursor = cursor{
			currPrId:      0,
			currSectionId: m.getNextSectionId(),
		}
		return
	}

	m.cursor.currPrId = newPrId
}

func (m model) numSections() int {
	return len(m.sectionsConfigs)
}

func (m model) getSectionAt(id int) *models.SectionData {
	return &(*m.data)[id]
}

func (m model) getPrevSectionId() int {
	for sId := m.cursor.currSectionId - 1; sId >= 0; sId-- {
		if m.getSectionAt(sId).NumPrs() > 0 {
			return sId
		}
	}
	return m.cursor.currSectionId
}

func (m model) getNextSectionId() int {
	for sId := m.cursor.currSectionId + 1; sId < m.numSections(); sId++ {
		if m.getSectionAt(sId).NumPrs() > 0 {
			return sId
		}
	}
	return m.cursor.currSectionId
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "h":
			prevSection := m.getSectionAt(m.getPrevSectionId())
			newCursor := cursor{
				currSectionId: prevSection.Id,
				currPrId:      prevSection.Paginator.Page * PerPage,
			}

			m.cursor = newCursor
			return m, nil
		case "l":
			nextSection := m.getSectionAt(m.getNextSectionId())
			newCursor := cursor{
				currSectionId: nextSection.Id,
				currPrId:      nextSection.Paginator.Page * PerPage,
			}
			m.cursor = newCursor
			return m, nil
		case "j":
			m.nextPr()
			return m, nil
		case "k":
			m.prevPr()
			return m, nil
		case "o":
			currPR := func() models.PullRequest {
				var prs []models.PullRequest
				for _, section := range *m.data {
					prs = append(prs, section.Prs...)
				}
				return prs[m.cursor.currPrId]
			}()
			utils.OpenBrowser(currPR.Url)
			return m, nil
		default:
			return m, cmd
		}
	case msgs.InitMsg:
		m.sectionsConfigs = msg.Config
		var data []models.SectionData
		for i, sectionConfig := range m.sectionsConfigs {
			p := paginator.NewModel()
			p.Type = paginator.Dots
			p.PerPage = 5
			p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
			p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
			p.UseJKKeys = true
			s := spinner.Model{Spinner: spinner.Dot}
			data = append(data, models.SectionData{
				Id:        i,
				Config:    sectionConfig,
				Spinner:   models.SectionSpinner{Model: s, NumReposFetched: 0},
				Paginator: p,
			})
		}
		m.data = &data
		return m, m.startFetchingSectionsData()
	case models.RepoPullRequestsFetched:
		section := (*m.data)[msg.SectionId]
		section.Prs = append(section.Prs, msg.Prs...)
		section.Paginator.SetTotalPages(len(section.Prs))
		sort.Slice(section.Prs, func(i, j int) bool {
			return section.Prs[i].UpdatedAt.After(section.Prs[j].UpdatedAt)
		})

		section.Spinner.NumReposFetched += 1
		if section.Spinner.NumReposFetched == len(section.Config.Repos) {
			section.Spinner.Model.Finish()
		}

		(*m.data)[msg.SectionId] = section
		for sectionId, section := range *m.data {
			if section.NumPrs() > 0 {
				m.cursor.currSectionId = sectionId
				break
			}
		}
		return m, nil
	case msgs.TickMsg:
		var internalCmd tea.Cmd
		section := (*m.data)[msg.SectionId]
		section.Spinner.Model, internalCmd = section.Spinner.Model.Update(msg.InternalTickMsg)
		(*m.data)[msg.SectionId] = section
		return m, section.Tick(internalCmd)
	case string:
		m.res = msg
		return m, nil
	case msgs.ErrMsg:
		m.err = msg
		return m, nil
	case *[]models.SectionData:
		m.data = msg
		return m, nil
	default:
		var cmd tea.Cmd
		return m, cmd
	}
}

func renderEmptyState() string {
	emptyState := emptyStateStyle.Render("No PRs were found that match the given filters...")
	return fmt.Sprintf(emptyState + "\n")
}

func renderReviewStatus(pr models.PullRequest) string {
	if pr.ReviewDecision == "APPROVED" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("")
	}

	if pr.ReviewDecision == "REJECTED" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("")
	}

	return lipgloss.NewStyle().Faint(true).Render("")
}

func renderMergeableStatus(pr models.PullRequest) string {
	if pr.Mergeable == "MERGEABLE" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("")
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("")
}

func renderCiStatus(pr models.PullRequest) string {
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
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("")
	}

	if accStatus == "PENDING" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("")
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("")
}

func renderLines(pr models.PullRequest) string {
	separator := lipgloss.NewStyle().Faint(true).MarginLeft(1).MarginRight(1).Render("/")
	added := lipgloss.NewStyle().Render(fmt.Sprintf("%5d", pr.Additions))
	removed := lipgloss.NewStyle().Render(
		fmt.Sprintf("-%-5d", pr.Deletions),
	)

	return lipgloss.JoinHorizontal(lipgloss.Center, added, separator, removed)

}

func (m model) renderPullRequest(sectionId int, prId int, pr *models.PullRequest) string {
	log.SetOutput(m.logger)
	log.Printf("%+v", *pr)
	isSelected := m.cursor.currSectionId == sectionId && m.cursor.currPrId == prId
	selectedCell := cellStyle.Render(func() string {
		if isSelected {
			return ""
		} else {
			return " "
		}
	}())
	//
	// 
	// 
	// 
	// 
	// 

	// 
	// 
	// 
	// 

	reviewCell := cellStyle.Render(renderReviewStatus(*pr))
	mergeableCell := cellStyle.Render(renderMergeableStatus(*pr))
	ciCell := cellStyle.Render(renderCiStatus(*pr))
	linesCell := cellStyle.Render(renderLines(*pr))
	prTitleCell := cellStyle.Render(
		fmt.Sprintf(
			"%-65s",
			utils.TruncateString(pr.Title, 50)+
				lipgloss.NewStyle().Faint(true).Render(
					fmt.Sprintf("#%-6s", fmt.Sprintf("%d", pr.Number)),
				),
		),
	)
	prAuthorCell := cellStyle.Render(fmt.Sprintf("%-15s", utils.TruncateString(pr.Author.Login, 15)))
	prRepoCell := cellStyle.Render(fmt.Sprintf("%-20s", utils.TruncateString(pr.HeadRepository.Name, 20)))
	updatedAtCell := cellStyle.Render(utils.TimeElapsed(pr.UpdatedAt))

	return pullRequestStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, selectedCell, reviewCell, prTitleCell, mergeableCell, ciCell, linesCell, prAuthorCell, prRepoCell, updatedAtCell))
}

func (m model) View() string {
	s := strings.Builder{}
	if m.sectionsConfigs == nil {
		s.WriteString("Reading config...\n")
	} else if m.data == nil {
	} else if m.err != nil {
		s.WriteString("Error!\n")
	} else if m.data != nil {
		for sectionId, section := range *m.data {
			if sectionId > 0 {
				s.WriteString("\n")
			}
			s.WriteString(section.RenderTitle() + "\n")
			isLoading := section.Spinner.NumReposFetched < len(section.Config.Repos)
			if isLoading {
				s.WriteString(section.RenderLoadingState() + "\n")
			} else if len(section.Prs) == 0 {
				s.WriteString(renderEmptyState() + "\n")
			} else {
				var renderedPRs []string
				for prId, pr := range section.Prs {
					renderedPRs = append(renderedPRs, m.renderPullRequest(sectionId, prId, &pr))
				}
				start, end := section.Paginator.GetSliceBounds(section.NumPrs())
				s.WriteString(lipgloss.JoinVertical(lipgloss.Left, renderedPRs[start:end]...))
				if len(renderedPRs) > 0 {
					s.WriteString(lipgloss.PlaceHorizontal(
						lipgloss.Width(renderedPRs[0]),
						lipgloss.Center,
						"\n"+section.Paginator.View(),
					))
				}
			}
		}
	}

	s.WriteString("\n" + m.help.View(m.keys))
	return s.String()
}

func (m model) getNumberOfPRs() int {
	sum := 0
	for _, section := range *m.data {
		sum += len(section.Prs)
	}
	return sum
}

func (m model) startFetchingSectionsData() tea.Cmd {
	var cmds []tea.Cmd
	for _, sectionData := range *m.data {
		sectionData := sectionData
		cmds = append(cmds, sectionData.FetchSectionPullRequests()...)
		cmds = append(cmds, sectionData.Tick(spinner.Tick))
	}
	return tea.Batch(cmds...)
}
