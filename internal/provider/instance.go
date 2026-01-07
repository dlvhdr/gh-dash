package provider

import (
	"sync"
)

var (
	currentProvider Provider
	providerMu      sync.RWMutex
)

// SetProvider sets the global provider instance
func SetProvider(p Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	currentProvider = p
}

// GetProvider returns the current provider instance
// Defaults to GitHub if not set
func GetProvider() Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	if currentProvider == nil {
		return NewGitHubProvider()
	}
	return currentProvider
}

// IsGitLab returns true if the current provider is GitLab
func IsGitLab() bool {
	p := GetProvider()
	return p.GetType() == GitLab
}

// IsGitHub returns true if the current provider is GitHub
func IsGitHub() bool {
	p := GetProvider()
	return p.GetType() == GitHub
}

// GetCLICommand returns the CLI command for the current provider
func GetCLICommand() string {
	return GetProvider().GetCLICommand()
}
