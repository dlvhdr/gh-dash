package autocomplete

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// LabelInfo contains information about a label at a specific cursor position
// in a comma-separated list of labels.
// StartIdx and EndIdx are rune indices (not byte indices).
type LabelInfo struct {
	Label    string
	StartIdx tea.Position
	EndIdx   tea.Position
	IsFirst  bool
	IsLast   bool
}

type LabelSource struct{}

func (LabelSource) ExtractContext(input string, cursorPos tea.Position) Context {
	info := ExtractLabelAtCursor(input, cursorPos)
	return Context{
		Start:   info.StartIdx,
		End:     info.EndIdx,
		Content: info.Label,
	}
}

func (LabelSource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart tea.Position,
	contextEnd tea.Position,
) (newInput string, newCursorPos tea.Position) {
	labelInfo := ExtractLabelAtCursor(input, contextEnd)
	lines := lines(input)
	runes := []rune(lines[contextStart.Y])

	var replacement string
	if labelInfo.IsFirst {
		replacement = suggestion + ", "
	} else {
		replacement = " " + suggestion + ", "
	}

	remainingInput := string(runes[labelInfo.EndIdx.X:])
	remainingInput = strings.TrimPrefix(remainingInput, ",")
	remainingInput = strings.TrimLeft(remainingInput, " \t")

	newLine := string(runes[:labelInfo.StartIdx.X]) + replacement + remainingInput
	newValue := joinLines(lines[:contextStart.Y]) + newLine + joinLines(lines[contextEnd.Y+1:])
	newCursorPos.X = labelInfo.StartIdx.X + len([]rune(replacement))

	return newValue, newCursorPos
}

func (LabelSource) ItemsToExclude(input string, cursorPos tea.Position) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	currentLabel := ExtractLabelAtCursor(input, cursorPos).Label
	currentLabels := CurrentLabels(input)
	if currentLabels == nil {
		return nil
	}

	excluded := make([]string, 0, len(currentLabels))
	for _, label := range currentLabels {
		if label != currentLabel {
			excluded = append(excluded, label)
		}
	}

	return excluded
}

// ExtractLabelAtCursor extracts information about the label at the given cursor position
// in a comma-separated list. It considers the entire word containing the cursor as the
// current label. The cursor position and returned indices are rune-based (not byte-based)
// to correctly handle multi-byte Unicode characters.
func ExtractLabelAtCursor(input string, cursorPos tea.Position) LabelInfo {
	if input == "" {
		return LabelInfo{
			Label:    "",
			StartIdx: tea.Position{},
			EndIdx:   tea.Position{},
			IsFirst:  true,
			IsLast:   true,
		}
	}

	lines := strings.Split(input, "\n")
	if cursorPos.Y > len(lines) {
		return LabelInfo{
			Label:    "",
			StartIdx: tea.Position{},
			EndIdx:   tea.Position{},
			IsFirst:  true,
			IsLast:   true,
		}
	}

	line := lines[cursorPos.Y]
	runes := []rune(line)

	if cursorPos.X < 0 {
		cursorPos.X = 0
	}
	if cursorPos.X > len(runes) {
		cursorPos.X = len(runes)
	}

	startIdx := 0
	for i := cursorPos.X - 1; i >= 0; i-- {
		if runes[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	endIdx := len(runes)
	for i := cursorPos.X; i < len(runes); i++ {
		if runes[i] == ',' {
			endIdx = i
			break
		}
	}

	label := strings.TrimSpace(string(runes[startIdx:endIdx]))
	isFirst := startIdx == 0
	isLast := endIdx == len(runes)

	return LabelInfo{
		Label:    label,
		StartIdx: tea.Position{X: startIdx, Y: cursorPos.Y},
		EndIdx:   tea.Position{X: endIdx, Y: cursorPos.Y},
		IsFirst:  isFirst,
		IsLast:   isLast,
	}
}

// CurrentLabels splits the input by commas and returns all trimmed, non-empty labels.
func CurrentLabels(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}
