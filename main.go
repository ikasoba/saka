package main

import (
	"context"
	"debug/buildinfo"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ikasoba/saka/components"
	"github.com/ikasoba/saka/config"
	"github.com/ikasoba/saka/manager"
	"github.com/ikasoba/saka/rope"
	"github.com/ikasoba/saka/treesitter"
	"github.com/ikasoba/saka/util"
	"github.com/rivo/tview"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

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
	conf, theme := config.Load()

	fileName := ""

	if len(os.Args) == 2 {
		if os.Args[1] == "-e" || os.Args[1] == "--show-environment" {
			fmt.Println("SAKA_CONFIG_DIR=" + config.GetConfigDir())

			return
		} else if os.Args[1] == "--install-grammar" {
			for _, g := range conf.Grammars {
				fmt.Println("Install", "repo:", g.Repo, "rev:", g.Rev)

				err := manager.InstallGrammar(g.Repo, g.Rev)

				if err != nil {
					panic(err)
				}
			}

			return
		} else if os.Args[1] == "--fetch-default-config" {
			conf, theme, err := manager.FetchDefaultConfigs()
			if err != nil {
				panic(err)
			}

			d := config.GetConfigDir()

			confFd, err := os.Create(filepath.Join(d, "config.toml"))
			if err != nil {
				panic(err)
			}

			defer confFd.Close()

			confFd.ReadFrom(conf)

			themeFd, err := os.Create(filepath.Join(d, "theme.toml"))
			if err != nil {
				panic(err)
			}

			defer themeFd.Close()

			themeFd.ReadFrom(theme)

			return
		} else {
			fileName = os.Args[1]
		}
	} else {
		exefile, _ := os.Executable()
		info, _ := buildinfo.ReadFile(exefile)

		fmt.Println("Saka (" + info.Main.Version + ")\n")
		fmt.Println("Usage:", os.Args[0], "<filename>")
		fmt.Println("  Open file\n")
		fmt.Println("Usage:", os.Args[0], "(-e|--show-environment)")
		fmt.Println("  Show environment variables\n")
		fmt.Println("Usage:", os.Args[0], "--install-grammar")
		fmt.Println("  Install grammar from config\n")
		fmt.Println("Usage:", os.Args[0], "--fetch-default-config")
		fmt.Println("  Fetch default config")

		os.Exit(1)

		return
	}

	app := tview.NewApplication()

	style := tcell.StyleDefault.Background(theme.Background.Color()).Foreground(theme.Foreground.Color())

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
		configPath := config.GetConfigDir()

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
