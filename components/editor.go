package components

import (
	"context"
	"io"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/ikasoba/saka/highlighter"
	"github.com/ikasoba/saka/rangemap"
	"github.com/ikasoba/saka/rope"
	"github.com/ikasoba/saka/theme"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type CursorPos struct {
	Row int
	Col int
}

type RenderState struct {
	reader io.RuneReader
	index  int
	row    int
	col    int
	line   int
	isEOF  bool
}

type CursorState struct {
	hasSelection   bool
	startSelection int
	endSelection   int
}

type LanguageProtocol struct {
	Server protocol.Server
}

type LspClient struct {
	editor           *Editor
	workspace        uri.URI
	diagnostics      map[string]rangemap.RangeMap[protocol.Diagnostic]
	diagnosticsLine  rangemap.RangeMap[int]
	DiagnosticsCount int
}

func (c *LspClient) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	return nil
}

func (c *LspClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	return nil
}

func (c *LspClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	return nil
}

func (c *LspClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	c.diagnosticsLine = rangemap.RangeMap[int]{}

	diagnosticsCount := 0

	m := rangemap.RangeMap[protocol.Diagnostic]{}

	for _, d := range params.Diagnostics {
		start := rope.GetIndexFromRowCol(c.editor.Text, int(d.Range.Start.Line), int(d.Range.Start.Character))
		end := rope.GetIndexFromRowCol(c.editor.Text, int(d.Range.End.Line), int(d.Range.End.Character))

		m.Put(start, end, d)
		c.diagnosticsLine.Put(int(d.Range.Start.Line), int(d.Range.End.Line), int(d.Severity)-1)
		diagnosticsCount++
	}

	c.diagnostics[string(params.URI)] = m

	c.DiagnosticsCount = diagnosticsCount

	c.editor.CurrentDiagnostic <- struct {
		Diagnostics int
		Diagnostic  *protocol.Diagnostic
	}{
		c.DiagnosticsCount,
		c.editor.currentDiagnosticMessage,
	}

	c.editor.app.Draw()

	return nil
}

func (c *LspClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	return nil
}

func (c *LspClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	return nil, nil
}

func (c *LspClient) Telemetry(ctx context.Context, params interface{}) (err error) {
	return nil
}

func (c *LspClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) (err error) {
	return nil
}

func (c *LspClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	return nil
}

func (c *LspClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result bool, err error) {
	return false, nil
}

func (c *LspClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error) {
	return nil, nil
}

func (c *LspClient) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	return []protocol.WorkspaceFolder{
		protocol.WorkspaceFolder{
			URI:  string(c.workspace),
			Name: c.workspace.Filename(),
		},
	}, nil
}

type Editor struct {
	*tview.Box
	app         *tview.Application
	uri         uri.URI
	Text        rope.RopeNode
	LineIndexes map[int]struct {
		Start int
		End   int
	}
	CurrentLine         int
	ScrollX             int
	offsetY             int
	clipboard           string
	cursor              CursorState
	prevCursor          CursorState
	highlight           highlighter.Highlighter
	highlights          highlighter.Highlights
	theme               theme.EditorTheme
	isTextModified      bool
	isHighlightRequired bool
	serverContext       context.Context
	workspace           uri.URI
	server              protocol.Server
	client              *LspClient
	CurrentDiagnostic   chan struct {
		Diagnostics int
		Diagnostic  *protocol.Diagnostic
	}
	currentDiagnosticMessage *protocol.Diagnostic
	PreviousMaxLine          int
}

func NewEditor(theme theme.EditorTheme) *Editor {
	return &Editor{
		Box:  tview.NewBox(),
		Text: rope.New(""),
		LineIndexes: map[int]struct {
			Start int
			End   int
		}{},
		CurrentLine: 0,
		ScrollX:     0,
		offsetY:     0,
		cursor: CursorState{
			hasSelection: false,
		},
		isTextModified: true,
		theme:          theme,
		CurrentDiagnostic: make(chan struct {
			Diagnostics int
			Diagnostic  *protocol.Diagnostic
		}),
	}
}

func (e *Editor) SetApplication(app *tview.Application) {
	e.app = app
}

func (e *Editor) SetHighlighter(highlighter highlighter.Highlighter) {
	e.highlight = highlighter
}

func (e *Editor) SetURI(uri uri.URI) {
	e.uri = uri
}

func (e *Editor) SetWorkspace(uri uri.URI) {
	e.workspace = uri
}

func (e *Editor) NewLspClient() *LspClient {
	client := &LspClient{
		editor:      e,
		diagnostics: map[string]rangemap.RangeMap[protocol.Diagnostic]{},
		workspace:   e.workspace,
	}

	e.client = client

	return client
}

func (e *Editor) SetLanguageServer(ctx context.Context, server protocol.Server) {
	ln := rope.GetLineBreaks(e.Text)

	e.PreviousMaxLine = ln + 1

	e.serverContext = ctx
	e.server = server

	initParams := &protocol.InitializeParams{
		RootPath: e.workspace.Filename(),
		RootURI:  e.workspace,
		WorkspaceFolders: []protocol.WorkspaceFolder{
			protocol.WorkspaceFolder{
				URI:  string(e.workspace),
				Name: e.workspace.Filename(),
			},
		},
		Capabilities: protocol.ClientCapabilities{
			Workspace: &protocol.WorkspaceClientCapabilities{
				WorkspaceFolders: true,
			},
			TextDocument: &protocol.TextDocumentClientCapabilities{
				PublishDiagnostics: &protocol.PublishDiagnosticsClientCapabilities{
					CodeDescriptionSupport: true,
					RelatedInformation:     true,
					VersionSupport:         true,
					DataSupport:            true,
				},
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DidSave: true,
				},
			},
		},
	}

	server.Initialize(ctx, initParams)

	server.Initialized(ctx, &protocol.InitializedParams{})

	server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        e.uri,
			LanguageID: protocol.LanguageIdentifier("go"),
			Version:    0,
			Text:       e.Text.String(),
		},
	})
}

func (e *Editor) Draw(screen tcell.Screen) {
	if e.isTextModified {
		e.isHighlightRequired = true
	}

	//	e.Box.DrawForSubclass(screen, e)

	if e.server != nil {
		if e.isTextModified {
			change := protocol.TextDocumentContentChangeEvent{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 0,
					},
					End: protocol.Position{
						Line:      uint32(e.PreviousMaxLine),
						Character: 0,
					},
				},
				Text: e.Text.String(),
			}

			e.server.DidChange(e.serverContext, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: e.uri},
					Version:                0,
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					change,
				},
			})

			ln := rope.GetLineBreaks(e.Text)

			e.PreviousMaxLine = ln + 1
		}
	}

	// Prepare
	x, y, width, height := e.GetInnerRect()
	if width <= 0 || height <= 0 {
		return // We have no space for anything.
	}

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			style := tcell.StyleDefault.Background(e.theme.Background.Color()).Foreground(e.theme.Foreground.Color())

			screen.SetContent(x+col, y+row, ' ', []rune{}, style)
		}
	}

	offsetY := e.offsetY

	offsetX := 5

	e.LineIndexes = map[int]struct {
		Start int
		End   int
	}{}

	{
		i, row, col := e.Text.GetLineStart(offsetY), 0, 0

		buf := rope.Slice(e.Text, i, max(0, e.Text.GetLineEnd(offsetY+height+1)-i))

		if e.isHighlightRequired {
			e.isHighlightRequired = false

			if e.highlight != nil {
				e.highlights = e.highlight.Highlight(i, buf)
			}
		}

		r := rope.NewReader(buf)

		isEOF := false
		ln := offsetY

		numberLineStyle := tcell.StyleDefault.Background(e.theme.LineNumberBackground.Color()).Foreground(e.theme.LineNumberForeground.Color())
		activeNumberLineStyle := tcell.StyleDefault.Background(e.theme.ActiveLineNumberBackground.Color()).Foreground(e.theme.ActiveLineNumberForeground.Color())

		{
			lineStyle := numberLineStyle
			if e.client != nil {
				m := e.client.diagnostics[string(e.uri)]

				_, ok, _ := m.Get(i, i)
				if ok {
					lineStyle = lineStyle.Background(tcell.Color52)
				}
			}

			t := []rune(strconv.Itoa(ln + 1))

			for j := 0; j < offsetX-1; j++ {
				ch := ' '
				if j >= (offsetX-1)-len(t) {
					ch = t[j-((offsetX-1)-len(t))]
				}

				screen.SetContent(x+j, y+row, ch, []rune{}, lineStyle)
			}
		}

		isCurrentLineUpdated := false

		prevChar := rune(0)

		for !isEOF {
			ch, sz, err := r.ReadRune()
			if err != nil {
				if err == io.EOF {
					isEOF = true
				} else {
					panic(err)
				}
			}

			if sz == 0 {
				isEOF = true
			}

			style := tcell.StyleDefault.Background(e.theme.Background.Color()).Foreground(e.theme.Foreground.Color())
			if s, ok, _ := e.highlights.Get(i, i+1); ok {
				style = s
			}

			startSelection, endSelection := e.cursor.startSelection, e.cursor.endSelection
			if !e.cursor.hasSelection {
				endSelection = startSelection
			}

			if i == endSelection {
				e.CurrentLine = ln
				isCurrentLineUpdated = true
			}

			if endSelection < startSelection {
				endSelection, startSelection = startSelection, endSelection
			}

			comb := []rune{}

			if e.cursor.hasSelection && i >= startSelection && i <= endSelection {
				style = style.Background(e.theme.SelectionBackground.Color())
			}

			if i == endSelection {
				style = style.Background(e.theme.CursorColor.Color())
			}

			if x+offsetX+col-e.ScrollX < x+width && y+row < y+height && x+offsetX+col-e.ScrollX >= x+offsetX && y+row >= y {
				screen.SetContent(x+offsetX+col-e.ScrollX, y+row, ch, comb, style)
			}

			if y+row > y+height {
				break
			}

			if ch == '\r' || (prevChar != '\r' && ch == '\n') {
				idx := e.LineIndexes[ln]
				idx.End = i
				e.LineIndexes[ln] = idx

				idx = e.LineIndexes[ln+1]
				idx.Start = i + 1
				e.LineIndexes[ln+1] = idx

				ln += 1
				row += 1
				col = 0

				if y+row < y+height {
					lineStyle := numberLineStyle
					if e.client != nil {
						m := e.client.diagnosticsLine

						d, ok, _ := m.Get(ln, ln)
						if ok {
							lineStyle = lineStyle.Background(e.theme.DiagnosticLineNumberBackground[d].Color())
						}
					}

					t := []rune(strconv.Itoa(ln + 1))

					for j := 0; j < offsetX-1; j++ {
						ch := ' '
						if j >= (offsetX-1)-len(t) {
							ch = t[j-((offsetX-1)-len(t))]
						}

						screen.SetContent(x+j, y+row, ch, []rune{}, lineStyle)
					}
				}
			} else if prevChar == '\r' && ch == '\n' {
				idx := e.LineIndexes[ln]
				idx.Start += 1
				e.LineIndexes[ln] = idx
			} else if ch == '\t' {
				if x+offsetX+col+1 < x+width {
					screen.SetContent(x+offsetX+col+1, y+row, ' ', []rune{}, style)
				}

				col += 2
			} else {
				col += runewidth.RuneWidth(ch)
			}

			i += 1
			prevChar = ch
		}

		if y+(e.CurrentLine-offsetY) >= y && y+(e.CurrentLine-offsetY) < y+height {
			lineStyle := activeNumberLineStyle
			if e.client != nil {
				m := e.client.diagnosticsLine

				d, ok, _ := m.Get(e.CurrentLine, e.CurrentLine)
				if ok {
					lineStyle = lineStyle.Background(e.theme.DiagnosticLineNumberBackground[d].Color())
				}
			}

			t := []rune(strconv.Itoa(e.CurrentLine + 1))

			for j := 0; j < offsetX-1; j++ {
				ch := ' '
				if j >= (offsetX-1)-len(t) {
					ch = t[j-((offsetX-1)-len(t))]
				}

				screen.SetContent(x+j, y+(e.CurrentLine-offsetY), ch, []rune{}, lineStyle)
			}
		}

		idx := e.LineIndexes[ln]
		idx.End = i
		e.LineIndexes[ln] = idx

		e.prevCursor = e.cursor

		if !isCurrentLineUpdated {
			i := e.cursor.startSelection
			if e.cursor.hasSelection && i < e.cursor.endSelection {
				i = e.cursor.endSelection
			}

			e.CurrentLine = e.Text.GetLineNumber(min(e.Text.Length(), i))
		}
	}

	isRedrawRequired := false

	ln := e.CurrentLine - e.offsetY

	if e.offsetY > 0 && ln <= 1 {
		if ln < 0 {
			e.offsetY += ln
		} else {
			e.offsetY -= 1
		}
		e.isHighlightRequired = true

		isRedrawRequired = true
	} else if ln >= height-2 {
		e.offsetY += 1
		e.isHighlightRequired = true

		isRedrawRequired = true
	}

	cursorPosition := e.cursor.startSelection
	if e.cursor.hasSelection && cursorPosition < e.cursor.endSelection {
		cursorPosition = e.cursor.endSelection
	}

	if e.client != nil {
		m := e.client.diagnostics[string(e.uri)]
		d, ok, _ := m.Get(cursorPosition, cursorPosition)
		if ok {
			e.currentDiagnosticMessage = &d

			e.CurrentDiagnostic <- struct {
				Diagnostics int
				Diagnostic  *protocol.Diagnostic
			}{
				e.client.DiagnosticsCount,
				e.currentDiagnosticMessage,
			}
		} else {
			e.currentDiagnosticMessage = nil

			e.CurrentDiagnostic <- struct {
				Diagnostics int
				Diagnostic  *protocol.Diagnostic
			}{e.client.DiagnosticsCount, e.currentDiagnosticMessage}
		}
	}

	start := e.Text.GetLineStart(e.CurrentLine)

	rawCol := cursorPosition - start
	if rawCol > e.Text.GetLineEnd(e.CurrentLine)-start {
		rawCol = e.Text.GetLineEnd(e.CurrentLine) - start
	}

	if rawCol < 0 {
		rawCol = 0
	}

	col := 0

	r := rope.NewReader(rope.Slice(e.Text, e.Text.GetLineStart(e.CurrentLine), rawCol))
	for {
		ch, _, err := r.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		if col == '\t' {
			col += 2
		} else {
			col += runewidth.RuneWidth(ch)
		}
	}

	w := width - offsetX - 2

	scrollX := col - e.ScrollX
	if e.ScrollX > 0 && scrollX <= 1 {
		if scrollX < 0 {
			e.ScrollX += scrollX
		} else {
			e.ScrollX -= 1
		}

		isRedrawRequired = true
	} else if scrollX >= w-w/8 {
		e.ScrollX += scrollX - (w - w/8) + 1
		isRedrawRequired = true
	}

	if e.isTextModified {
		e.isTextModified = false
	}

	if isRedrawRequired {
		e.Draw(screen)
	}
}

func (e *Editor) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return e.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyLeft:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				if e.cursor.endSelection > 0 {
					if e.cursor.endSelection > 1 && rope.Slice(e.Text, e.cursor.endSelection-2, 2).String() == "\r\n" {
						e.cursor.endSelection -= 2
					} else {
						e.cursor.endSelection -= 1
					}
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				if e.cursor.startSelection > 0 {
					if e.cursor.startSelection > 1 && rope.Slice(e.Text, e.cursor.startSelection-2, 2).String() == "\r\n" {
						e.cursor.startSelection -= 2
					} else {
						e.cursor.startSelection -= 1
					}
				}

				e.cursor.hasSelection = false
			}

		case tcell.KeyRight:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				if e.cursor.endSelection < e.Text.Length() {
					if rope.Slice(e.Text, e.cursor.endSelection, 2).String() == "\r\n" {
						e.cursor.endSelection += 2
					} else {
						e.cursor.endSelection += 1
					}
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				if e.cursor.startSelection < e.Text.Length() {
					if rope.Slice(e.Text, e.cursor.startSelection, 2).String() == "\r\n" {
						e.cursor.startSelection += 2
					} else {
						e.cursor.startSelection += 1
					}
				}

				e.cursor.hasSelection = false
			}

		case tcell.KeyUp:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				prevLine := e.LineIndexes[e.CurrentLine]
				prev := e.cursor.endSelection - prevLine.Start

				i, ok := e.LineIndexes[e.CurrentLine-1]
				if ok {
					e.cursor.endSelection = i.Start + min(i.End-i.Start, prev)
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				prevLine := e.LineIndexes[e.CurrentLine]
				prev := e.cursor.startSelection - prevLine.Start

				i, ok := e.LineIndexes[e.CurrentLine-1]
				if ok {
					e.cursor.startSelection = i.Start + min(i.End-i.Start, prev)
				}

				e.cursor.hasSelection = false
			}

		case tcell.KeyDown:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				prevLine := e.LineIndexes[e.CurrentLine]
				prev := e.cursor.endSelection - prevLine.Start

				i, ok := e.LineIndexes[e.CurrentLine+1]
				if ok {
					e.cursor.endSelection = i.Start + min(i.End-i.Start, prev)

					if rope.Slice(e.Text, e.cursor.endSelection-1, 2).String() == "\r\n" {
						e.cursor.endSelection += 1
					}
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				prevLine := e.LineIndexes[e.CurrentLine]
				prev := e.cursor.startSelection - prevLine.Start

				i, ok := e.LineIndexes[e.CurrentLine+1]
				if ok {
					e.cursor.startSelection = i.Start + min(i.End-i.Start, prev)

					if rope.Slice(e.Text, e.cursor.startSelection-1, 2).String() == "\r\n" {
						e.cursor.startSelection += 1
					}
				}

				e.cursor.hasSelection = false
			}

		case tcell.KeyEnter:
			if !e.cursor.hasSelection {
				e.Text = e.Text.Insert(e.cursor.startSelection, "\n")
			} else {
				start, end := e.cursor.startSelection, e.cursor.endSelection
				if start > end {
					start, end = end, start
				}

				e.Text = rope.Replace(e.Text, start, end-start, "\n")
			}

			e.cursor.startSelection += 1
			e.isTextModified = true

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			offset := e.cursor.startSelection
			if e.cursor.hasSelection && offset < e.cursor.endSelection {
				offset = e.cursor.endSelection
			}

			if offset > 0 {
				if !e.cursor.hasSelection {
					if e.cursor.startSelection > 1 && rope.Slice(e.Text, e.cursor.startSelection-2, 2).String() == "\r\n" {
						e.Text = e.Text.Delete(e.cursor.startSelection-2, 2)
						e.cursor.startSelection -= 2
					} else {
						e.Text = e.Text.Delete(e.cursor.startSelection-1, 1)
						e.cursor.startSelection -= 1
					}

				} else {
					start, end := e.cursor.startSelection, e.cursor.endSelection
					if start > end {
						start, end = end, start
					}

					e.Text = e.Text.Delete(start, end-start)

					e.cursor.hasSelection = false
					e.cursor.startSelection = start
				}

				e.isTextModified = true
			}

		case tcell.KeyCtrlK:
			startSelection, endSelection := e.cursor.startSelection, e.cursor.endSelection
			if !e.cursor.hasSelection {
				endSelection = startSelection
			}

			if endSelection < startSelection {
				endSelection, startSelection = startSelection, endSelection
			}

			res := rope.Slice(e.Text, startSelection, endSelection-startSelection)

			e.clipboard = res.String()

		case tcell.KeyCtrlU:
			if !e.cursor.hasSelection {
				e.Text = e.Text.Insert(e.cursor.startSelection, e.clipboard)

				e.cursor.startSelection += len([]rune(e.clipboard))
			} else {
				start, end := e.cursor.startSelection, e.cursor.endSelection
				if start > end {
					start, end = end, start
				}

				e.Text = rope.Replace(e.Text, start, end-start, e.clipboard)

				if e.cursor.startSelection < e.cursor.endSelection {
					e.cursor.startSelection = start
					e.cursor.endSelection = start + len([]rune(e.clipboard))
				} else {
					e.cursor.startSelection = end
					e.cursor.endSelection = end - len([]rune(e.clipboard))
				}
			}

			e.isTextModified = true

		case tcell.KeyCtrlA:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				i, ok := e.LineIndexes[e.CurrentLine]
				if ok {
					e.cursor.endSelection = i.Start
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				i, ok := e.LineIndexes[e.CurrentLine]
				if ok {
					e.cursor.startSelection = i.Start
				}

				e.cursor.hasSelection = false
			}

		case tcell.KeyCtrlE:
			if event.Modifiers() == tcell.ModShift {
				if !e.cursor.hasSelection {
					e.cursor.endSelection = e.cursor.startSelection
				}

				i, ok := e.LineIndexes[e.CurrentLine]
				if ok {
					e.cursor.endSelection = i.End
				}

				e.cursor.hasSelection = true
			} else {
				if e.cursor.hasSelection {
					e.cursor.startSelection = e.cursor.endSelection
				}

				i, ok := e.LineIndexes[e.CurrentLine]
				if ok {
					e.cursor.startSelection = i.End
				}

				e.cursor.hasSelection = false
			}

		default:
			if event.Modifiers() & ^tcell.ModShift == 0 {
				if !e.cursor.hasSelection {
					e.Text = e.Text.Insert(e.cursor.startSelection, string(event.Rune()))

					e.cursor.startSelection += 1
				} else {
					start, end := e.cursor.startSelection, e.cursor.endSelection
					if start > end {
						start, end = end, start
					}

					e.Text = rope.Replace(e.Text, start, end-start, string(event.Rune()))

					e.cursor.hasSelection = false
					e.cursor.startSelection = start + 1
				}

				e.isTextModified = true
			}
		}
	})
}
