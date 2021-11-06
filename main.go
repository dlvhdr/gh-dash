package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	config "dlvhdr/gh-prs/config"
	utils "dlvhdr/gh-prs/utils"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type model struct {
	res    string
	err    error
	config []config.Section
	data   *[]SectionData
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
	config config.Section
	prs    []PullRequest
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func main() {
	sections, err := config.ParseSectionsConfig()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	p := tea.NewProgram(
		model{config: sections},
		tea.WithAltScreen(),
	)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.fetchAllReposPullRequests, tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
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

func (m model) View() string {
	s := strings.Builder{}
	if m.data == nil {
		s.WriteString("Waiting...\n")
	} else if m.err != nil {
		s.WriteString("Error!\n")
	} else if m.data != nil {
		for _, section := range *m.data {
			title := titleStyle.Render(section.config.Title)
			s.WriteString(title + "\n")
			if len(section.prs) == 0 {
				emptyState := emptyStateStyle.Render("Nothing here...")
				s.WriteString(emptyState + "\n")
			}
			for _, pr := range section.prs {
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
		}
	}

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
