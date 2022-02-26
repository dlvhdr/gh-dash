package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/constants"
)

func isStatusWaiting(status string) bool {
	return status == "PENDING" ||
		status == "QUEUED" ||
		status == "IN_PROGRESS" ||
		status == "WAITING"
}

func renderCheckRunConclusion(checkRun data.CheckRun) string {
	conclusionStr := string(checkRun.Conclusion)
	if isStatusWaiting(string(checkRun.Status)) {
		return constants.WaitingGlyph
	}

	if isConclusionAFailure(conclusionStr) {
		return constants.FailureGlyph
	}

	return constants.SuccessGlyph
}

func renderStatusContextConclusion(statusContext data.StatusContext) string {
	if isConclusionAFailure(string(statusContext.State)) {
		return constants.FailureGlyph
	}

	return constants.SuccessGlyph
}

func isConclusionAFailure(conclusion string) bool {
	return conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE"
}

func renderCheckRunName(checkRun data.CheckRun) string {
	var parts []string
	creator := strings.TrimSpace(string(checkRun.CheckSuite.Creator.Login))
	if creator != "" {
		parts = append(parts, creator)
	}

	workflow := strings.TrimSpace(string(checkRun.CheckSuite.WorkflowRun.Workflow.Name))
	if workflow != "" {
		parts = append(parts, workflow)
	}

	name := strings.TrimSpace(string(checkRun.Name))
	if name != "" {
		parts = append(parts, name)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		strings.Join(parts, "/"),
	)
}

func renderStatusContextName(statusContext data.StatusContext) string {
	var parts []string
	creator := strings.TrimSpace(string(statusContext.Creator.Login))
	if creator != "" {
		parts = append(parts, creator)
	}

	context := strings.TrimSpace(string(statusContext.Context))
	if context != "" && context != "/" {
		parts = append(parts, context)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		strings.Join(parts, "/"),
	)
}
