package data

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"

	"charm.land/log/v2"
)

var (
	repoLabelCache = make(map[string][]Label)
	labelCacheMu   sync.RWMutex
	// execCommand is injectable for testing; defaults to exec.Command
	execCommand = exec.Command
)

func CachedRepoLabels(repoNameWithOwner string) ([]Label, bool) {
	labelCacheMu.RLock()
	defer labelCacheMu.RUnlock()
	labels, ok := repoLabelCache[repoNameWithOwner]
	return labels, ok
}

func FetchRepoLabels(repoNameWithOwner string) ([]Label, error) {
	// Check cache first
	if cachedLabels, ok := CachedRepoLabels(repoNameWithOwner); ok {
		return cachedLabels, nil
	}

	log.Debug("Fetching repo labels", "repoNameWithOwner", repoNameWithOwner)

	cmd := execCommand(
		"gh",
		"label",
		"list",
		"-R",
		repoNameWithOwner,
		"--json",
		"name,description,color",
		"--limit",
		"300",
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var labels []Label
	if err := json.Unmarshal(output, &labels); err != nil {
		return nil, err
	}

	filteredLabels := make([]Label, 0, len(labels))
	for _, label := range labels {
		if strings.TrimSpace(label.Name) != "" {
			filteredLabels = append(filteredLabels, label)
		}
	}

	labelCacheMu.Lock()
	defer labelCacheMu.Unlock()

	if labels, ok := repoLabelCache[repoNameWithOwner]; ok {
		return labels, nil
	}

	repoLabelCache[repoNameWithOwner] = filteredLabels
	log.Debug(
		"Successfully fetched repo labels",
		"repoNameWithOwner",
		repoNameWithOwner,
		"len",
		len(filteredLabels),
	)
	return filteredLabels, nil
}

func ClearLabelCache() {
	labelCacheMu.Lock()
	defer labelCacheMu.Unlock()
	repoLabelCache = make(map[string][]Label)
}

func ClearRepoLabelCache(repoNameWithOwner string) {
	labelCacheMu.Lock()
	defer labelCacheMu.Unlock()
	delete(repoLabelCache, repoNameWithOwner)
}

func LabelNames(labels []Label) []string {
	names := make([]string, len(labels))
	for i, label := range labels {
		names[i] = label.Name
	}
	return names
}
