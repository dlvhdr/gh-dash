package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"

	config "dlvhdr/gh-prs/config"
	utils "dlvhdr/gh-prs/utils"
)

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},   // first column
		{k.Help, k.Quit}, // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	keys   keyMap
	res    string
	err    error
	config []config.Section
	data   *[]SectionData
	help   help.Model
}

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

type SectionData struct {
	config  config.Section
	spinner spinner.Spinner
	prs     *[]PullRequest
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func main() {
	p := tea.NewProgram(
		model{
			keys: keys,
			help: help.NewModel(),
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

type initMsg struct {
	config []config.Section
}

func initScreen() tea.Msg {
	sections, err := config.ParseSectionsConfig()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return initMsg{config: sections}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	case initMsg:
		m.config = msg.config
    var data []SectionData
    for _, sectionConfig := range m.config {
      s := spinner.NewModel()
      s.Spinner = spinner.Dot

      data = append(data, SectionData{
        config: sectionConfig,
        spinner: s,
      })
    }
    m.data = 
		return m, m.fetchAllReposPullRequests
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case string:
		m.res = msg
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	case *[]SectionData:
		m.data = msg
		return m, nil
	default:
		return m, nil
	}
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true)

	emptyStateStyle = lipgloss.NewStyle().Faint(true).PaddingLeft(2).MarginBottom(1)
)

func renderEmptyState(s *strings.Builder) {
	emptyState := emptyStateStyle.Render("Nothing here...")
	s.WriteString(emptyState + "\n")
}

func renderTitle(s *strings.Builder, title string) {
	sectionTitle := titleStyle.Render(title)
	s.WriteString(sectionTitle + "\n")
}

func renderPullRequest(s *strings.Builder, pr *PullRequest) {
	s.WriteString(
		fmt.Sprintf(
			"#%-5d | %-30s | %-10s | %-10s | %s\n",
			pr.Number,
			utils.TruncateString(pr.Title, 30),
			utils.TruncateString(pr.Author.Login, 10),
			utils.TruncateStringTrailing(pr.HeadRepository.Name, 10),
			utils.TimeElapsed(pr.UpdatedAt),
		),
	)
}

func (m model) renderLoadingState(s *strings.Builder) {
	for _, section := range m.config {
		renderTitle(s, section.Title)

	}
}

func (m model) View() string {
	s := strings.Builder{}
	if m.config == nil {
		s.WriteString("Reading config...\n")
	} else if m.data == nil {
		m.renderLoadingState(&s)
	} else if m.err != nil {
		s.WriteString("Error!\n")
	} else if m.data != nil {
		for _, section := range *m.data {
			renderTitle(&s, section.config.Title)
			if len(section.prs) == 0 {
				renderEmptyState(&s)
			}
			for _, pr := range section.prs {
				renderPullRequest(&s, &pr)
			}
		}
	}

	s.WriteString("\n" + m.help.View(m.keys))
	return s.String()
}

func (m model) fetchAllReposPullRequests() tea.Msg {
	var data []SectionData
	for _, sectionConfig := range m.config {
		sectionData := SectionData{
			config: sectionConfig,
			prs:    []PullRequest{},
		}
		for _, repo := range sectionConfig.Repos {
			fetched, err := fetchRepoPullRequests(repo, sectionData.config.Filters)
			if err != nil {
				continue
			}
			sectionData.prs = append(sectionData.prs, *fetched...)
		}

		sort.Slice(sectionData.prs, func(i, j int) bool {
			return sectionData.prs[i].UpdatedAt.After(sectionData.prs[j].UpdatedAt)
		})
		data = append(data, sectionData)
	}

	return &data
}

const JsonFields = "title,mergeable,author,url,additions,headRefName,headRepository,isDraft,number,reviewDecision,state,statusCheckRollup,updatedAt"

func fetchRepoPullRequests(repo string, search string) (*[]PullRequest, error) {
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
		"5",
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

	return &prs, nil
}
