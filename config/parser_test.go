package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// See https://www.gh-dash.dev/configuration/#_top
//  1. get default config file or create it if it's missing
//     1.1. try GH_DASH_CONFIG
//     1.2. then check if we're in a git repo
//     1.2.1. try both `.yml` and `.yaml`
//     1.3. try to look under `XDG_CONFIG_HOME`
//     1.4. if not, try with `os.UserHomeDir()`
//     1.5. if still doesn't exist, create with defaults
//  2. read the config file
//     2.1. read the file at the path
//     2.2. validate the config

const repoPath = "./testdata/"

func TestParser(t *testing.T) {
	t.Run("Should read config passed by flag with highest priority", func(t *testing.T) {
		clearEnv := setupConfigEnvVar()
		defer clearEnv()

		parsed, err := ParseConfig(Location{
			RepoPath:   repoPath,
			ConfigFlag: "./testdata/test-config.yml",
		})

		require.Equal(t, err, nil)
		require.Len(t, parsed.PRSections, 3)
	})

	t.Run("Should then try GH_DASH_CONFIG env var", func(t *testing.T) {
		clearEnv := setupConfigEnvVar()
		defer clearEnv()

		parsed, err := ParseConfig(Location{
			RepoPath:   repoPath,
			ConfigFlag: "",
		})

		require.Equal(t, err, nil)
		require.Len(t, parsed.PRSections, 1)
	})

	t.Run("Should then try config in repo", func(t *testing.T) {
		parsed, err := ParseConfig(Location{
			RepoPath:   repoPath,
			ConfigFlag: "",
		})

		require.Equal(t, err, nil)
		require.Len(t, parsed.PRSections, 2)
	})
}

func setupConfigEnvVar() func() {
	os.Setenv("GH_DASH_CONFIG", "./testdata/other-test-config.yml")
	return func() {
		os.Unsetenv("GH_DASH_CONFIG")
	}
}
