package footer

import (
	"fmt"
	"path"
	"strings"

	bbHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const viewSeparator = " │ "

type Model struct {
	ctx             *context.ProgramContext
	leftSection     *string
	rightSection    *string
	help            bbHelp.Model
	ShowAll         bool
	ShowConfirmQuit bool
}

func NewModel(ctx *context.ProgramContext) Model {
	help := bbHelp.New()
	help.ShowAll = true
	help.Styles = ctx.Styles.Help.BubbleStyles
	l := ""
	r := ""
	return Model{
		ctx:          ctx,
		help:         help,
		leftSection:  &l,
		rightSection: &r,
	}
}

func (m Model) View() string {
	var footer string

	if m.ShowConfirmQuit {
		footer = lipgloss.NewStyle().Render("Really quit? (Press y/enter to confirm, any other key to cancel)")
	} else {
		helpIndicator := lipgloss.NewStyle().
			Background(m.ctx.Theme.FaintText).
			Foreground(m.ctx.Theme.SelectedBackground).
			Padding(0, 1).
			Render("? help")
		donationIndicator := zone.Mark("donate", lipgloss.NewStyle().
			Background(m.ctx.Theme.SelectedBackground).
			Foreground(m.ctx.Theme.WarningText).
			Padding(0, 1).
			Underline(true).
			Render(fmt.Sprintf("%s donate", constants.DonateIcon)))
		viewSwitcher := m.renderViewSwitcher(m.ctx)
		leftSection := ""
		if m.leftSection != nil {
			leftSection = *m.leftSection
		}
		rightSection := ""
		if m.rightSection != nil {
			rightSection = *m.rightSection
		}
		spacing := lipgloss.NewStyle().
			Background(m.ctx.Theme.SelectedBackground).
			Render(
				strings.Repeat(
					" ",
					utils.Max(0,
						m.ctx.ScreenWidth-lipgloss.Width(
							viewSwitcher,
						)-lipgloss.Width(leftSection)-
							lipgloss.Width(rightSection)-
							lipgloss.Width(
								helpIndicator,
							)-lipgloss.Width(donationIndicator),
					)))

		footer = m.ctx.Styles.Common.FooterStyle.
			Render(lipgloss.JoinHorizontal(lipgloss.Top, viewSwitcher, leftSection, spacing,
				rightSection, donationIndicator, helpIndicator))
	}

	if m.ShowAll {
		keymap := keys.CreateKeyMapForView(m.ctx.View)
		fullHelp := m.help.View(keymap)
		return lipgloss.JoinVertical(lipgloss.Top, footer, fullHelp)
	}

	return footer
}

func (m *Model) SetShowConfirmQuit(val bool) {
	m.ShowConfirmQuit = val
}

func (m *Model) SetWidth(width int) {
	m.help.Width = width
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.help.Styles = ctx.Styles.Help.BubbleStyles
}

func (m *Model) renderViewButton(view config.ViewType) string {
	isActive := m.ctx.View == view

	// Define icons and labels for each view
	var icon, label string
	// Define icons - notifications has solid/outline variants
	solidBell := ""
	outlineBell := ""

	switch view {
	case config.NotificationsView:
		if m.ctx.View == config.NotificationsView {
			icon = solidBell
		} else {
			icon = outlineBell
		}
		label = ""
	case config.PRsView:
		icon = ""
		label = " PRs"
	case config.IssuesView:
		icon = ""
		label = " Issues"
	}

	if isActive {
		// Active: colored icon + prominent background
		// Use gold for notifications bell, green for others
		iconColor := m.ctx.Theme.SuccessText
		if view == config.NotificationsView {
			iconColor = lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#FFD700"} // Gold
		}
		activeStyle := lipgloss.NewStyle().
			Foreground(iconColor).
			Background(m.ctx.Styles.ViewSwitcher.ActiveView.GetBackground()).
			Bold(true)
		if label != "" {
			return activeStyle.Render(icon) + activeStyle.Render(label)
		}
		return activeStyle.Render(icon)
	}

	// Inactive: faint styling
	return m.ctx.Styles.ViewSwitcher.InactiveView.Render(icon + label)
}

func (m *Model) renderViewSwitcher(ctx *context.ProgramContext) string {
	var repo string
	if m.ctx.RepoPath != "" {
		name := path.Base(m.ctx.RepoPath)
		if m.ctx.RepoUrl != "" {
			name = git.GetRepoShortName(m.ctx.RepoUrl)
		}
		repo = ctx.Styles.Common.FooterStyle.Render(fmt.Sprintf(" %s", name))
	}

	var user string
	if ctx.User != "" {
		user = ctx.Styles.Common.FooterStyle.Render("@" + ctx.User)
	}

	view := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ctx.Styles.ViewSwitcher.ViewsSeparator.PaddingLeft(1).Render(m.renderViewButton(config.NotificationsView)),
		ctx.Styles.ViewSwitcher.ViewsSeparator.Render(viewSeparator),
		m.renderViewButton(config.PRsView),
		ctx.Styles.ViewSwitcher.ViewsSeparator.Render(viewSeparator),
		m.renderViewButton(config.IssuesView),
		lipgloss.NewStyle().Background(ctx.Styles.Common.FooterStyle.GetBackground()).Foreground(
			ctx.Styles.ViewSwitcher.ViewsSeparator.GetBackground()).Render(" "),
		repo,
		ctx.Styles.Common.FooterStyle.Foreground(m.ctx.Theme.FaintText).Render(" • "),
		user,
		ctx.Styles.Common.FooterStyle.Foreground(m.ctx.Theme.FaintBorder).Render(" │"),
	)

	return ctx.Styles.ViewSwitcher.Root.Render(view)
}

func (m *Model) SetLeftSection(leftSection string) {
	*m.leftSection = leftSection
}

func (m *Model) SetRightSection(rightSection string) {
	*m.rightSection = rightSection
}
