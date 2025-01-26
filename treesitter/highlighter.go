package treesitter

import (
	"context"
	"os"
	"regexp"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/gdamore/tcell/v2"
	"github.com/ikasoba/saka/highlighter"
	"github.com/ikasoba/saka/rope"
	"github.com/ikasoba/saka/theme"
	"github.com/k0kubun/pp"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Highlighter struct {
	lang       *sitter.Language
	parser     *sitter.Parser
	oldtree    *sitter.Tree
	highlights []byte
	theme      theme.EditorTheme
}

func NewHighlighter(name string, grammarPath string, highlightsPath string) *Highlighter {
	path := grammarPath
	lib, err := openLibrary(path)
	if err != nil {
		panic(err)
	}

	var language func() uintptr

	purego.RegisterLibFunc(&language, lib, "tree_sitter_"+name)

	lang := sitter.NewLanguage(unsafe.Pointer(language()))
	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	highlights, err := os.ReadFile(highlightsPath)
	if err != nil {
		panic(err)
	}

	return &Highlighter{
		lang:       lang,
		parser:     parser,
		oldtree:    nil,
		highlights: highlights,
	}
}

func (h *Highlighter) Initialize(theme theme.EditorTheme) {
	h.theme = theme
}

func (h *Highlighter) Highlight(offset int, buf rope.RopeNode) (highlights highlighter.Highlights) {
	src := []byte(buf.String())

	n := h.parser.ParseCtx(context.Background(), src, nil)
	h.oldtree = n

	q, err := sitter.NewQuery(h.lang, string(h.highlights))
	if err != nil {
		pp.Fprintln(os.Stderr, err)
		return
	}

	root := n.RootNode()

	qc := sitter.NewQueryCursor()
	captures := qc.Captures(q, root, src)
	names := q.CaptureNames()

	for {
		m, _ := captures.Next()
		if m == nil {
			break
		}

		for _, c := range m.Captures {
			style := tcell.StyleDefault.Background(h.theme.Background.Color()).Foreground(h.theme.Foreground.Color())

			name := names[c.Index]

			if _, ok := h.theme.Highlights[name]; !ok {
				name = regexp.MustCompile(`\..*$`).ReplaceAllString(name, "")
			}

			if color, ok := h.theme.Highlights[name]; ok {
				style = style.Foreground(color.Color())

				start := offset + len([]rune(string(src[:c.Node.StartByte()])))

				highlights.Put(start, start+len([]rune(string(src[c.Node.StartByte():c.Node.EndByte()])))+1, style)
			}
		}
	}

	return highlights
}
