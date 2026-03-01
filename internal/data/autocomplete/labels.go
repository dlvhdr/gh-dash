package autocomplete

import "strings"

// LabelInfo contains information about a label at a specific cursor position
// in a comma-separated list of labels.
// StartIdx and EndIdx are rune indices (not byte indices).
type LabelInfo struct {
	Label    string
	StartIdx int
	EndIdx   int
	IsFirst  bool
	IsLast   bool
}

type LabelSource struct{}

func (LabelSource) ExtractContext(input string, cursorPos int) Context {
	info := ExtractLabelAtCursor(input, cursorPos)
	return Context{
		Start:   info.StartIdx,
		End:     info.EndIdx,
		Content: info.Label,
	}
}

func (LabelSource) InsertSuggestion(input string, suggestion string, contextStart int, contextEnd int) (newInput string, newCursorPos int) {
	labelInfo := ExtractLabelAtCursor(input, contextStart)
	runes := []rune(input)

	var replacement string
	if labelInfo.IsFirst {
		replacement = suggestion + ", "
	} else {
		replacement = " " + suggestion + ", "
	}

	remainingInput := string(runes[labelInfo.EndIdx:])
	remainingInput = strings.TrimPrefix(remainingInput, ",")
	remainingInput = strings.TrimLeft(remainingInput, " \t")

	newValue := string(runes[:labelInfo.StartIdx]) + replacement + remainingInput
	newCursorPos = labelInfo.StartIdx + len([]rune(replacement))

	return newValue, newCursorPos
}

func (LabelSource) ItemsToExclude(input string, cursorPos int) []string {
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
func ExtractLabelAtCursor(input string, cursorPos int) LabelInfo {
	if input == "" {
		return LabelInfo{
			Label:    "",
			StartIdx: 0,
			EndIdx:   0,
			IsFirst:  true,
			IsLast:   true,
		}
	}

	runes := []rune(input)

	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	startIdx := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if runes[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	endIdx := len(runes)
	for i := cursorPos; i < len(runes); i++ {
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
		StartIdx: startIdx,
		EndIdx:   endIdx,
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
