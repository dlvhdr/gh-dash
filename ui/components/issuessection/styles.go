package issuessection

import "github.com/charmbracelet/lipgloss"

var (
	ciCellWidth               = lipgloss.Width(" CI ")
	linesCellWidth            = lipgloss.Width(" 123450 / -123450 ")
	updatedAtCellWidth        = lipgloss.Width("2mo ago")
	issueRepoCellWidth        = 15
	issueAuthorCellWidth      = 15
	issueAssigneesCellWidth   = 20
	issueNumCommentsCellWidth = 6
	ContainerPadding          = 1

	containerStyle = lipgloss.NewStyle().
			Padding(0, ContainerPadding)

	spinnerStyle = lipgloss.NewStyle().Padding(0, 1)

	emptyStateStyle = lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(1).
			MarginBottom(1)
)
