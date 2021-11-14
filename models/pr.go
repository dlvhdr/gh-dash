package models

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"
)

type RepoPullRequestsFetched struct {
	SectionId int
	RepoName  string
	Prs       []PullRequest
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

const JsonFields = "title,mergeable,author,url,additions,headRefName,headRepository,isDraft,number,reviewDecision,state,statusCheckRollup,updatedAt"

func FetchRepoPullRequests(repo string, search string) ([]PullRequest, error) {
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

	return prs, nil
}
