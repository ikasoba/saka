package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
	"github.com/ikasoba/saka/components"
	"github.com/ikasoba/saka/config"
	"github.com/ikasoba/saka/rope"
	"github.com/ikasoba/saka/theme"
	"github.com/ikasoba/saka/treesitter"
	"github.com/ikasoba/saka/util"
	"github.com/rivo/tview"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

func LoadConfig() (conf config.Config, theme theme.EditorTheme) {
	exefile, _ := os.Executable()

	configPath := os.Getenv("SAKA_CONFIG_DIR")
	if configPath == "" {
		configPath = filepath.Dir(exefile)
	}

	toml.DecodeFile(filepath.Join(configPath, "config.toml"), &conf)
	toml.DecodeFile(filepath.Join(configPath, "theme.toml"), &theme)

	return conf, theme
}

func FindMatchedLanguage(file string, langs map[string]config.LanguageConfig) (name string, conf config.LanguageConfig, ok bool) {
	for name, lang := range langs {
		for _, pattern := range lang.Patterns {
			r, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if r.MatchString(file) {
				return name, lang, true
			}
		}
	}

	return name, conf, false
}

func main() {
	conf, theme := LoadConfig()

	app := tview.NewApplication()

	style := tcell.StyleDefault.Background(theme.Background.Color()).Foreground(theme.Foreground.Color())

	fileName := ""
	if len(os.Args) == 2 {
		fileName = os.Args[1]
	} else {
		fmt.Fprintln(os.Stderr, "file name required.")

		os.Exit(1)

		return
	}

	statusText := tview.NewTextView().SetTextStyle(style.Reverse(true)).SetTextAlign(tview.AlignRight)

	statusLine := tview.NewFlex().
		AddItem(
			tview.NewTextView().SetTextStyle(style.Reverse(true)).SetText(" 🐈 Saka 0.1"), 0, 1, false,
		).
		AddItem(
			tview.NewTextView().SetTextStyle(style.Reverse(true)).SetText(fileName).SetTextAlign(tview.AlignCenter), 0, 1, false,
		).
		AddItem(
			statusText, 0, 1, false,
		)

	hintBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				AddItem(
					tview.NewTextView().
						SetTextStyle(style).
						SetRegions(true).
						SetText(`["hl"]^S[""] Save`).
						Highlight("hl"),
					0, 1, false,
				).
				AddItem(
					tview.NewTextView().
						SetTextStyle(style).
						SetRegions(true).
						SetText(`["hl"]^X[""] Close`).
						Highlight("hl"),
					0, 1, false,
				).
				AddItem(
					tview.NewTextView().
						SetTextStyle(style).
						SetRegions(true).
						SetText(`["hl"]^K[""] Cut`).
						Highlight("hl"),
					0, 1, false,
				).
				AddItem(
					tview.NewTextView().
						SetTextStyle(style).
						SetRegions(true).
						SetText(`["hl"]^U[""] Paste`).
						Highlight("hl"),
					0, 1, false,
				),
			0, 1, false,
		)

	hintBox.Box = hintBox.SetBackgroundColor(theme.Background.Color())

	textarea := components.NewEditor(theme)
	textarea.SetApplication(app)

	absPath, _ := filepath.Abs(fileName)
	cwd, _ := os.Getwd()

	absPath = strings.ReplaceAll(absPath, "\\", "/")
	cwd = strings.ReplaceAll(cwd, "\\", "/")

	documentUri, _ := uri.Parse("file://" + absPath)
	workspaceUri, _ := uri.Parse("file://" + cwd)

	textarea.SetURI(documentUri)
	textarea.SetWorkspace(workspaceUri)

	name, lang, ok := FindMatchedLanguage(fileName, conf.Languages)
	if ok {
		exefile, _ := os.Executable()

		configPath := os.Getenv("SAKA_CONFIG_DIR")
		if configPath == "" {
			configPath = filepath.Dir(exefile)
		}

		dylibExtname := ".so"
		switch runtime.GOOS {
		case "darwin":
			dylibExtname = ".dylib"

		case "windows":
			dylibExtname = ".dll"
		}

		grammarPath := filepath.Join(configPath, "grammars", name+dylibExtname)
		highlightsPath := filepath.Join(configPath, "queries", name, "highlights.scm")

		_, grammarErr := os.Stat(grammarPath)
		_, highlightsErr := os.Stat(highlightsPath)

		if grammarErr == nil && highlightsErr == nil {
			hl := treesitter.NewHighlighter(name, grammarPath, highlightsPath)
			hl.Initialize(theme)
			textarea.SetHighlighter(hl)
		}

		if lang.ServerCommand != nil {
			cmd := exec.Command(*lang.ServerCommand, lang.Arguments...)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				panic(err)
			}

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				panic(err)
			}

			cmd.Start()

			rwc := &util.ReadWriteCloser{
				stdout,
				stdin,
			}

			stream := jsonrpc2.NewStream(rwc)

			if len(os.Args) == 2 {
				buf, _ := os.ReadFile(os.Args[1])

				textarea.Text = rope.New(string(buf)).Rebalance()
			}

			logger, _ := zap.NewProduction()

			client := textarea.NewLspClient()

			go func() {
				for x := range textarea.CurrentDiagnostic {
					diagnostics := ""
					if x.Diagnostics > 0 {
						diagnostics = "🛑 " + strconv.Itoa(x.Diagnostics) + " "
					}

					if x.Diagnostic != nil {
						statusText.SetText(diagnostics + x.Diagnostic.Message)
					} else {
						statusText.SetText(diagnostics)
					}
				}
			}()

			ctx, _, server := protocol.NewClient(context.Background(), client, stream, logger)

			textarea.SetLanguageServer(ctx, server)
		}
	}

	box := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusLine, 1, 1, false).
		AddItem(
			textarea,
			0, 1, true,
		).
		AddItem(hintBox, 3, 1, false)

	box.Box = box.
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyCtrlX:
				app.Stop()
				return nil
			case tcell.KeyCtrlS:
				err := os.WriteFile(fileName, []byte(textarea.Text.String()), os.ModePerm)
				if err != nil {
					panic(err)
				}

				return nil
			}

			return event
		})

	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	screen.SetStyle(style)

	app.SetScreen(screen)

	if err := app.SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
}
