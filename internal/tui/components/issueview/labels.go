package issueview

import (
	"strings"
)

// LabelInfo contains information about a label at a specific cursor position
// in a comma-separated list of labels.
type LabelInfo struct {
	Label    string
	StartIdx int
	EndIdx   int
	IsFirst  bool
	IsLast   bool
}

// extractLabelAtCursor extracts information about the label at the given cursor position
// in a comma-separated list. It considers the entire word containing the cursor as the
// current label.
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

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	// Find the comma before the cursor (or start of string)
	startIdx := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if input[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	// Find the comma after the cursor (or end of string)
	endIdx := len(input)
	for i := cursorPos; i < len(input); i++ {
		if input[i] == ',' {
			endIdx = i
			break
		}
	}

	// Extract and trim the label
	label := strings.TrimSpace(input[startIdx:endIdx])

	// Determine if this is the first or last label
	isFirst := startIdx == 0
	isLast := endIdx == len(input)

	return LabelInfo{
		Label:    label,
		StartIdx: startIdx,
		EndIdx:   endIdx,
		IsFirst:  isFirst,
		IsLast:   isLast,
	}
}

// currentLabel returns the label text at the cursor position
func currentLabel(cursorPos int, currentValue string) string {
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
func handleLabelSelection(selected string, cursorPos int, currentValue string) (string, int) {
	labelInfo := extractLabelAtCursor(currentValue, cursorPos)

	// Build replacement with consistent spacing and trailing comma
	var replacement string
	if labelInfo.IsFirst {
		replacement = selected + ", "
	} else {
		replacement = " " + selected + ", "
	}

	// Determine what comes after the current label
	// Skip existing comma and spaces if present to avoid duplication
	remainingInput := currentValue[labelInfo.EndIdx:]
	// Remove the comma
	remainingInput = strings.TrimPrefix(remainingInput, ",")
	// Skip any spaces after the comma
	remainingInput = strings.TrimLeft(remainingInput, " \t")

	// Build new input by replacing the label at cursor position
	newValue := currentValue[:labelInfo.StartIdx] + replacement + remainingInput

	// Position cursor after the ", " we added
	newCursorPos := labelInfo.StartIdx + len(replacement)

	return newValue, newCursorPos
}
