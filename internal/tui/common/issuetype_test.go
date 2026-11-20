package common

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestIssueTypeColorCoversEveryGitHubColor(t *testing.T) {
	for _, name := range []string{"GRAY", "BLUE", "GREEN", "YELLOW", "ORANGE", "RED", "PINK", "PURPLE"} {
		_, ok := issueTypeColors[name]
		require.True(t, ok, "no color for %s", name)
	}
}

func TestIssueTypeColorFallsBackToGray(t *testing.T) {
	require.Equal(t, issueTypeColors["GRAY"], IssueTypeColor("TEAL"))
	require.Equal(t, issueTypeColors["GRAY"], IssueTypeColor(""))
}

func TestRenderIssueTypePillShowsName(t *testing.T) {
	pill := RenderIssueTypePill(data.IssueType{Name: "Bug", Color: "RED"}, lipgloss.NewStyle())
	require.Contains(t, pill, "Bug")
}
