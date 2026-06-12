package fuzzyselect

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// SearchQuerySource implements the Source interface.
type SearchQuerySource struct {
	Labels    []data.Label
	LabelsErr error
	Users     []data.User
	UsersErr  error
}

func authorPrefix(info WordInfo) (string, bool) {
	if strings.HasPrefix(info.Word, "author:") {
		return "author:", true
	}

	if strings.HasPrefix(info.Word, "-author:") {
		return "-author:", true
	}

	return "", false
}

func labelPrefix(info WordInfo) (string, bool) {
	if strings.HasPrefix(info.Word, "label:") {
		return "label:", true
	}

	if strings.HasPrefix(info.Word, "-label:") {
		return "-label:", true
	}

	return "", false
}

func (*SearchQuerySource) ExtractContext(input string, cursorPos tea.Position) Context {
	info := ExtractWordAtCursor(input, cursorPos)
	if prefix, ok := authorPrefix(info); ok {
		c, _ := strings.CutPrefix(info.Word, prefix)
		return Context{
			Start:   tea.Position{X: info.StartIdx.X + len(prefix), Y: info.StartIdx.Y},
			End:     info.EndIdx,
			Content: c,
		}
	}
	if prefix, ok := labelPrefix(info); ok {
		c, _ := strings.CutPrefix(info.Word, prefix)
		return Context{
			Start:   tea.Position{X: info.StartIdx.X + len(prefix), Y: info.StartIdx.Y},
			End:     info.EndIdx,
			Content: c,
		}
	}
	return Context{
		Start:   info.StartIdx,
		End:     info.EndIdx,
		Content: info.Word,
	}
}

func (src *SearchQuerySource) Suggestions(input string, cursorPos tea.Position) []Suggestion {
	wordInfo := ExtractWordAtCursor(input, cursorPos)
	if _, ok := authorPrefix(wordInfo); ok {
		suggestions := make([]Suggestion, 0)
		suggestions = append(suggestions, Suggestion{
			Value:  "@me",
			Detail: "Signed-in user",
		})
		for _, user := range src.Users {
			if user.Login != "" {
				suggestions = append(suggestions, Suggestion{Value: user.Login, Detail: user.Name})
			}
		}

		return suggestions
	}

	if _, ok := labelPrefix(wordInfo); ok {
		suggestions := make([]Suggestion, 0, len(src.Labels))
		for _, label := range src.Labels {
			suggestions = append(suggestions, Suggestion{
				Value:  label.Name,
				Detail: strings.TrimSpace(label.Description),
			})
		}
		return suggestions
	}

	return nil
}

func (*SearchQuerySource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart tea.Position,
	contextEnd tea.Position,
) (newInput string, newCursorPos tea.Position) {
	lines := lines(input)
	runes := []rune(lines[contextStart.Y])
	replacement := suggestion + " "
	newLine := string(runes[:contextStart.X]) + replacement + string(runes[contextEnd.X:])
	newValue := joinLines(lines[:contextStart.Y]) + newLine + joinLines(lines[contextEnd.Y+1:])
	newCursorPos.X = contextStart.X + len([]rune(replacement))
	return newValue, newCursorPos
}

func (*SearchQuerySource) ItemsToExclude(input string, cursorPos tea.Position) []string {
	return []string{}
}

func (src *SearchQuerySource) LoadSuggestions(ctx LoaderContext) error {
	var wg sync.WaitGroup
	wg.Go(func() {
		users, err := data.FetchRepoUsers(ctx.RepoOwner, ctx.RepoName)
		src.Users = users
		src.UsersErr = err
	})

	wg.Go(func() {
		labels, err := data.FetchRepoLabels(fmt.Sprintf("%s/%s", ctx.RepoOwner, ctx.RepoName))
		src.Labels = labels
		src.LabelsErr = err
	})

	wg.Wait()

	return errors.Join(src.UsersErr, src.LabelsErr)
}
