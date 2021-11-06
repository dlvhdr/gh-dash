package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	yaml "gopkg.in/yaml.v2"
)

type model struct {
	res string
	err error
	prs *[]PullRequest
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

type Section struct {
	Title   string
	Filters string
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(pwd, "sections.yml"))
	if err != nil {
		panic(err)
	}
	var sections []Section
	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%+v\n\n", sections)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchAllReposPullRequests, tea.EnterAltScreen)
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
	case *[]PullRequest:
		m.prs = msg
		return m, nil
	default:
		return m, nil
	}
}

func (m model) View() string {
	s := strings.Builder{}
	if m.prs == nil {
		s.WriteString("Waiting...\n")
	} else if m.err != nil {
		s.WriteString("Error!\n")
	} else if m.prs != nil {
		for _, pr := range *m.prs {
			s.WriteString(
				fmt.Sprintf(
					"#%-5d | %-20s | %-10s | %-10s | %s\n",
					pr.Number,
					truncateString(pr.Title, 20),
					truncateString(pr.Author.Login, 10),
					truncateString(pr.HeadRepository.Name, 10),
					timeElapsed(pr.UpdatedAt),
				),
			)
		}
	}

	return s.String()
}

func fetchAllReposPullRequests() tea.Msg {
	repositories := []string{
		"wix-private/wix-code-code-editor",
		"wix-private/wix-code-devex",
	}
	var prs []PullRequest

	for _, repo := range repositories {
		fetched, err := fetchRepoPullRequests(repo)
		if err != nil {
			continue
		}
		prs = append(prs, *fetched...)
	}

	sort.Slice(prs, func(i, j int) bool {
		return prs[i].UpdatedAt.After(prs[j].UpdatedAt)
	})

	return &prs
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

func truncateString(str string, num int) string {
	truncated := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncated = str[0:num] + "..."
	}
	return truncated
}

func s(x float64) string {
	if int(x) == 1 {
		return ""
	}
	return "s"
}

func timeElapsed(then time.Time) string {
	var parts []string
	var text string

	now := time.Now()
	year2, month2, day2 := now.Date()
	hour2, minute2, second2 := now.Clock()

	year1, month1, day1 := then.Date()
	hour1, minute1, second1 := then.Clock()

	year := math.Abs(float64(int(year2 - year1)))
	month := math.Abs(float64(int(month2 - month1)))
	day := math.Abs(float64(int(day2 - day1)))
	hour := math.Abs(float64(int(hour2 - hour1)))
	minute := math.Abs(float64(int(minute2 - minute1)))
	second := math.Abs(float64(int(second2 - second1)))

	week := math.Floor(day / 7)

	if year > 0 {
		parts = append(parts, strconv.Itoa(int(year))+" year"+s(year))
	}

	if month > 0 {
		parts = append(parts, strconv.Itoa(int(month))+" month"+s(month))
	}

	if week > 0 {
		parts = append(parts, strconv.Itoa(int(week))+" week"+s(week))
	}

	if day > 0 {
		parts = append(parts, strconv.Itoa(int(day))+" day"+s(day))
	}

	if hour > 0 {
		parts = append(parts, strconv.Itoa(int(hour))+" hour"+s(hour))
	}

	if minute > 0 {
		parts = append(parts, strconv.Itoa(int(minute))+" minute"+s(minute))
	}

	if second > 0 {
		parts = append(parts, strconv.Itoa(int(second))+" second"+s(second))
	}

	if now.After(then) {
		text = " ago"
	} else {
		text = " after"
	}

	if len(parts) == 0 {
		return "just now"
	}

	return parts[0] + text
}
