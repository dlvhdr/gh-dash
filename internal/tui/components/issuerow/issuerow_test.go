package issuerow

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestRenderType(t *testing.T) {
	tests := []struct {
		name      string
		issueType *data.IssueType
		expected  string
	}{
		{name: "no type renders nothing", issueType: nil, expected: ""},
		{name: "Bug", issueType: &data.IssueType{Name: "Bug", Color: "RED"}, expected: "Bug"},
		{
			name:      "Feature",
			issueType: &data.IssueType{Name: "Feature", Color: "BLUE"},
			expected:  "Feature",
		},
		{
			name:      "unknown color",
			issueType: &data.IssueType{Name: "Epic", Color: "TEAL"},
			expected:  "Epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{
				Ctx:  &context.ProgramContext{},
				Data: data.IssueData{IssueType: tt.issueType},
			}
			require.Equal(t, tt.expected, ansi.Strip(issue.renderType()))
		})
	}
}
