package common

import (
	"regexp"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yuin/goldmark-emoji/definition"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func renderLabelPill(label data.Label, pillStyle lipgloss.Style, suffix string) string {
	base := lipgloss.Color("#" + label.Color)
	c := lipgloss.Darken(base, 0.3)

	return pillStyle.
		BorderForeground(c).
		Foreground(lipgloss.Lighten(c, 0.5)).
		Background(c).
		Render(renderLabelName(label.Name)) + suffix
}

var (
	githubEmojis      = definition.Github()
	labelEmojiPattern = regexp.MustCompile(`:[[:alnum:]_+-]+:`)
)

type LabelOpts struct {
	Width     int
	MaxRows   int
	PillStyle lipgloss.Style
	RowStyle  lipgloss.Style
}

func RenderLabels(labels []data.Label, opts LabelOpts) string {
	if opts.Width <= 0 || len(labels) == 0 {
		return ""
	}

	if opts.MaxRows <= 0 {
		renderedRows := []string{}
		rowContentsWidth := 0
		currentRowLabels := []string{}

		for _, l := range labels {
			currentLabel := renderLabelPill(l, opts.PillStyle, "")
			currentLabelWidth := lipgloss.Width(currentLabel)

			if rowContentsWidth+currentLabelWidth <= opts.Width {
				currentRowLabels = append(
					currentRowLabels,
					currentLabel,
				)
				rowContentsWidth += currentLabelWidth
			} else {
				currentRowLabels = append(currentRowLabels, "\n")
				renderedRows = append(
					renderedRows,
					lipgloss.JoinHorizontal(lipgloss.Top, currentRowLabels...),
				)

				currentRowLabels = []string{currentLabel}
				rowContentsWidth = currentLabelWidth
			}

			// +1 for the space between labels
			currentRowLabels = append(currentRowLabels, " ")
			rowContentsWidth += 1
		}

		renderedRows = append(
			renderedRows,
			lipgloss.JoinHorizontal(lipgloss.Top, currentRowLabels...),
		)

		return lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	}

	rowStylePrefix := stylePrefix(opts.RowStyle)
	renderedRows := make([]string, 0, opts.MaxRows)
	rowContentsWidth := 0
	currentRowLabels := []string{}
	remainingLabels := 0
	rowCount := 0

	for idx, label := range labels {
		currentLabel := renderLabelPill(label, opts.PillStyle, rowStylePrefix)
		currentLabelWidth := lipgloss.Width(currentLabel)

		if rowContentsWidth > 0 && rowContentsWidth+1+currentLabelWidth > opts.Width {
			rowCount++
			if rowCount >= opts.MaxRows {
				remainingLabels = len(labels) - idx
				break
			}

			renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
			currentRowLabels = []string{}
			rowContentsWidth = 0
		}

		if rowContentsWidth == 0 && currentLabelWidth > opts.Width {
			rowCount++
			currentRowLabels = append(
				currentRowLabels,
				truncateStyled(currentLabel, opts.Width, rowStylePrefix),
			)
			renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
			currentRowLabels = []string{}
			rowContentsWidth = 0
			if rowCount >= opts.MaxRows {
				remainingLabels = len(labels) - idx - 1
				break
			}
			continue
		}

		if rowContentsWidth > 0 {
			rowContentsWidth++
		}
		currentRowLabels = append(currentRowLabels, currentLabel)
		rowContentsWidth += currentLabelWidth
	}

	if remainingLabels > 0 &&
		rowCount >= opts.MaxRows &&
		len(currentRowLabels) == 0 {
		return strings.Join(renderedRows, "\n")
	}

	if remainingLabels > 0 {
		overflowLabel := opts.PillStyle.Render("+"+strconv.Itoa(remainingLabels)) + rowStylePrefix
		overflowWidth := lipgloss.Width(overflowLabel)
		for {
			if rowContentsWidth == 0 {
				if overflowWidth > opts.Width {
					currentRowLabels = []string{
						truncateStyled(overflowLabel, opts.Width, rowStylePrefix),
					}
				} else {
					currentRowLabels = []string{overflowLabel}
				}
				break
			}

			if rowContentsWidth+1+overflowWidth <= opts.Width {
				currentRowLabels = append(currentRowLabels, overflowLabel)
				break
			}

			remainingLabels++
			lastItemIdx := len(currentRowLabels) - 1
			rowContentsWidth -= lipgloss.Width(currentRowLabels[lastItemIdx])
			currentRowLabels = currentRowLabels[:lastItemIdx]
			if len(currentRowLabels) > 0 {
				rowContentsWidth--
			} else {
				rowContentsWidth = 0
			}

			overflowLabel = opts.PillStyle.Render(
				"+"+strconv.Itoa(remainingLabels),
			) + rowStylePrefix
			overflowWidth = lipgloss.Width(overflowLabel)
		}
	}

	if len(currentRowLabels) > 0 {
		renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
	}

	return strings.Join(renderedRows, "\n")
}

func stylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	if rendered == "" {
		return ""
	}

	prefix, _, found := strings.Cut(rendered, "x")
	if !found {
		return ""
	}

	return prefix
}

func truncateStyled(value string, width int, stylePrefix string) string {
	return ansi.Truncate(value, width, constants.Ellipsis) + stylePrefix
}

func renderLabelName(name string) string {
	return labelEmojiPattern.ReplaceAllStringFunc(name, func(token string) string {
		emoji, ok := githubEmojis.Get(token[1 : len(token)-1])
		if !ok || !emoji.IsUnicode() {
			return token
		}

		return string(emoji.Unicode)
	})
}
