package issuesidebar

import (
	"github.com/dlvhdr/gh-dash/ui/styles"
)

var (
	pillStyle = styles.MainTextStyle.Copy().
		Foreground(styles.DefaultTheme.SubleMainText).
		PaddingLeft(1).
		PaddingRight(1)
)
