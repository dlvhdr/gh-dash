package issueview

import (
	"strings"
)

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

// extractLabelAtCursor extracts information about the label at the given cursor position
// in a comma-separated list. It considers the entire word containing the cursor as the
// current label. The cursor position and returned indices are rune-based (not byte-based)
// to correctly handle multi-byte Unicode characters.
func extractLabelAtCursor(input string, cursorPos int) LabelInfo {
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

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Find the comma before the cursor (or start of string)
	startIdx := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if runes[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	// Find the comma after the cursor (or end of string)
	endIdx := len(runes)
	for i := cursorPos; i < len(runes); i++ {
		if runes[i] == ',' {
			endIdx = i
			break
		}
	}

	// Extract and trim the label
	label := strings.TrimSpace(string(runes[startIdx:endIdx]))

	// Determine if this is the first or last label
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

// labelAtCursor returns the label text at the cursor position
func labelAtCursor(cursorPos int, currentValue string) string {
	labelInfo := extractLabelAtCursor(currentValue, cursorPos)
	return labelInfo.Label
}

// allLabels splits the input by commas and returns trimmed labels
func allLabels(value string) []string {
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

// handleLabelSelection handles autocomplete selection for comma-separated labels.
// It enforces consistent formatting:
// - Single space after comma for non-first labels
// - Always adds ", " after the label for easy continuation
// All indices are rune-based to correctly handle multi-byte Unicode characters.
func handleLabelSelection(selected string, cursorPos int, currentValue string) (string, int) {
	labelInfo := extractLabelAtCursor(currentValue, cursorPos)
	runes := []rune(currentValue)

	// Build replacement with consistent spacing and trailing comma
	var replacement string
	if labelInfo.IsFirst {
		replacement = selected + ", "
	} else {
		replacement = " " + selected + ", "
	}

	// Determine what comes after the current label
	// Skip existing comma and spaces if present to avoid duplication
	remainingInput := string(runes[labelInfo.EndIdx:])
	// Remove the comma
	remainingInput = strings.TrimPrefix(remainingInput, ",")
	// Skip any spaces after the comma
	remainingInput = strings.TrimLeft(remainingInput, " \t")

	// Build new input by replacing the label at cursor position
	newValue := string(runes[:labelInfo.StartIdx]) + replacement + remainingInput

	// Position cursor after the ", " we added
	newCursorPos := labelInfo.StartIdx + len([]rune(replacement))

	return newValue, newCursorPos
}
