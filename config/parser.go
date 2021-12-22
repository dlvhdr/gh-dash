package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"strings"

	"gopkg.in/yaml.v2"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
)

type SectionConfig struct {
	Title   string
	Filters string
	Repos   []string
}

type notificationResponse struct {
	responses []map[string]interface{}
	subs      []string
	rr        []string
	author    []string
}

const PrsDir = "prs"
const SectionsFileName = "sections.yml"

const NOTIFICATIONENDPOINT = "notifications"
const AUTHOR = "author"
const SUBSCRIBED = "subscribed"
const REVIEW_REQUESTED = "Review Requested"

type configError struct {
	configDir string
	err       error
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a sections.yml configuration file or an authenticated user.
Please login to your profile through gh or Create a config file under: %s

Example of a sections.yml file:
  - title: My Pull Requests
    repos:
      - dlvhdr/gh-prs
    filters: author:@me
  - title: Needs My Review
    repos:
      - dlvhdr/gh-prs
    filters: review-requested:@me
  - title: Subscribed
    repos:
      - cli/cli
      - charmbracelet/glamour
      - charmbracelet/lipgloss
    filters: -author:@me

For more info, go to https://github.com/dlvhdr/gh-prs
press q to exit.

Original error: %v`,
		path.Join(e.configDir, PrsDir, SectionsFileName),
		e.err,
	)
}

func ParseSectionsConfig() ([]SectionConfig, error) {
	var sections []SectionConfig
	configDir := os.Getenv("XDG_CONFIG_HOME")
	var err error

	if configDir == "" {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return sections, configError{configDir: configDir, err: err}
		}
	}

	data, err := os.ReadFile(filepath.Join(configDir, PrsDir, SectionsFileName))
	if err != nil {
		if !os.IsNotExist(err) {
			return sections, configError{configDir: configDir, err: err}
		}

		data, err = sectionConfigGenerator(filepath.Join(configDir, PrsDir, SectionsFileName))
		if err != nil {
			return sections, configError{configDir: configDir, err: err}
		}
	}
	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		return sections, fmt.Errorf("Failed parsing sections.yml: %w", err)
	}

	return sections, nil
}

func (nr *notificationResponse) getter() error {
	var opt = api.ClientOptions{
		Host: "github.com",
	}
	client, err := gh.RESTClient(&opt)
	if err != nil {
		return err
	}

	err = client.Get(NOTIFICATIONENDPOINT, &nr.responses)
	if err != nil {
		return err
	}

	return err
}

func (nr *notificationResponse) setter() {
	var typ, url, r string

	for _, response := range nr.responses {
		i := response["subject"]
		if sub, ok := i.(map[string]interface{}); ok {
			if typ, ok = sub["type"].(string); ok && typ == "PullRequest" {
				if url, ok = sub["url"].(string); !ok {
					continue
				}
				if r, ok = response["reason"].(string); !ok {
					continue
				}

				switch r {
				case "subscribed":
					nr.subs = append(nr.subs, strings.Split(strings.SplitAfter(url, "repos/")[1], "/pulls")[0])
				case "author":
					nr.author = append(nr.author, strings.Split(strings.SplitAfter(url, "repos/")[1], "/pulls")[0])
				case "review_requested":
					nr.rr = append(nr.rr, strings.Split(strings.SplitAfter(url, "repos/")[1], "/pulls")[0])
				}
			}
		}
	}
}

func sectionConfigGenerator(configLocation string) ([]byte, error) {
	var sc = make(map[string][]string)
	var filters = map[string]string{
		AUTHOR:           "author:@me",
		SUBSCRIBED:       "author:@me",
		REVIEW_REQUESTED: "review-requested:@me",
	}

	nr := new(notificationResponse)
	err := nr.getter()
	if err != nil {
		return nil, err
	}

	nr.setter()

	sc[SUBSCRIBED] = nr.subs
	sc[AUTHOR] = nr.author
	sc[REVIEW_REQUESTED] = nr.rr

	var configs []SectionConfig
	configs = append(configs,
		SectionConfig{
			Title:   AUTHOR,
			Repos:   sc[AUTHOR],
			Filters: filters[AUTHOR],
		},
		SectionConfig{
			Title:   SUBSCRIBED,
			Repos:   sc[SUBSCRIBED],
			Filters: filters[SUBSCRIBED],
		},
		SectionConfig{
			Title:   REVIEW_REQUESTED,
			Repos:   sc[REVIEW_REQUESTED],
			Filters: filters[REVIEW_REQUESTED],
		})

	return yaml.Marshal(configs)
}
