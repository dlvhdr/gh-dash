package config

import (
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/testutils"
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

const repoPath = "testdata"

func init() {
	log.SetLevel(log.ErrorLevel)
}

var keybindSorter = cmp.Transformer("Sort", func(in []Keybinding) []Keybinding {
	out := append([]Keybinding(nil), in...) // Copy input to avoid mutating it
	sort.Slice(out, func(i, j int) bool {
		return strings.Compare(out[i].Key, out[j].Key) == -1
	})
	return out
})

func Testwd(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func TestParser(t *testing.T) {
	t.Run("Should create default config", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "config")
		testutils.AssertNoError(t, err)
		defer os.RemoveAll(dir)

		os.Setenv("XDG_CONFIG_HOME", dir)
		defer func() {
			os.Unsetenv("XDG_CONFIG_HOME")
		}()

		parsed, err := ParseConfig(Location{})
		testutils.AssertNoError(t, err)
		require.Len(t, parsed.PRSections, 3)
	})

	t.Run("Should read config passed by flag with highest priority", func(t *testing.T) {
		clearXDGEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearXDGEnv()
		clearConfigEnv := setupConfigEnvVar(t)
		defer clearConfigEnv()

		cwd := Testwd(t)
		parsed, err := ParseConfig(Location{
			RepoPath:   path.Join(cwd, repoPath),
			ConfigFlag: path.Join(cwd, "testdata/test-config.yml"),
		})

		testutils.AssertNoError(t, err)
		require.Len(t, parsed.PRSections, 3)
		require.Equal(t, "#E2E1ED", parsed.Theme.Colors.Inline.Text.Primary.String())
	})

	t.Run("Should then try GH_DASH_CONFIG env var", func(t *testing.T) {
		clearXDGEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearXDGEnv()
		clearEnv := setupConfigEnvVar(t)
		defer clearEnv()

		cwd := Testwd(t)
		parsed, err := ParseConfig(Location{
			RepoPath:   path.Join(cwd, repoPath),
			ConfigFlag: "",
		})

		require.Equal(t, err, nil)
		require.Len(t, parsed.PRSections, 1)
	})

	t.Run("Should then try config in repo", func(t *testing.T) {
		clearXDGEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearXDGEnv()
		cwd := Testwd(t)
		parsed, err := ParseConfig(Location{
			RepoPath:   path.Join(cwd, repoPath),
			ConfigFlag: "",
		})

		testutils.AssertNoError(t, err)
		require.Len(t, parsed.PRSections, 2)
	})

	t.Run("Should then read global config", func(t *testing.T) {
		clearXDGEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearXDGEnv()
		clearEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearEnv()

		// parse config in ./testdata/gh-dash/config.yml
		actual, err := ParseConfig(Location{})
		testutils.AssertNoError(t, err)

		expected := loadExpected(t, "./testdata/global-config.golden.yml")
		assert.Empty(t, cmp.Diff(expected, actual, keybindSorter))
	})

	t.Run("Should merge global config with passed config", func(t *testing.T) {
		clearEnv := setXDGConfigHomeEnvVar(t, "testdata")
		defer clearEnv()

		// merge with config in ./testdata/gh-dash/config.yml
		cwd := Testwd(t)
		actual, err := ParseConfig(Location{
			ConfigFlag: path.Join(cwd, "./testdata/other-test-config.yml"),
		})
		testutils.AssertNoError(t, err)

		expected := loadExpected(t, "./testdata/merged-config.golden.yml")
		assert.Empty(t, cmp.Diff(expected, actual, keybindSorter))
	})
}

func loadExpected(t *testing.T, fpath string) Config {
	t.Helper()
	cwd := Testwd(t)
	k := koanf.NewWithConf(conf)
	err := k.Load(file.Provider(path.Join(cwd, fpath)), yaml.Parser())
	testutils.AssertNoError(t, err)

	expected := Config{}
	err = k.UnmarshalWithConf("", &expected, koanf.UnmarshalConf{Tag: "yaml"})
	testutils.AssertNoError(t, err)

	return expected
}

func setXDGConfigHomeEnvVar(t *testing.T, dir string) func() {
	t.Helper()
	cwd := Testwd(t)
	os.Setenv("XDG_CONFIG_HOME", path.Join(cwd, dir))
	return func() {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
}

func setupConfigEnvVar(t *testing.T) func() {
	t.Helper()
	cwd := Testwd(t)
	os.Setenv("GH_DASH_CONFIG", path.Join(cwd, "testdata/other-test-config.yml"))
	return func() {
		os.Unsetenv("GH_DASH_CONFIG")
	}
}
