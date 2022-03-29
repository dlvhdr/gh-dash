package issuesidebar

import (
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	pillStyle = styles.MainTextStyle.Copy().
		Foreground(styles.DefaultTheme.SubleMainText).
		PaddingLeft(1).
		PaddingRight(1)
)
