package highlighter

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ikasoba/saka/rangemap"
	"github.com/ikasoba/saka/rope"
	"github.com/ikasoba/saka/theme"
)

type Highlights = rangemap.RangeMap[tcell.Style]

type Highlighter interface {
	Initialize(theme theme.EditorTheme)
	Highlight(offset int, text rope.RopeNode) Highlights
}
