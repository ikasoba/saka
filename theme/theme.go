package theme

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type Color tcell.Color

func (c *Color) Color() tcell.Color {
	return tcell.Color(*c)
}

func (c *Color) UnmarshalTOML(data any) error {
	switch data := data.(type) {
	case string:
		if data[0] == '#' && len(data) == 4 {
			r, _ := strconv.ParseInt(string(data[1]), 16, 32)
			g, _ := strconv.ParseInt(string(data[2]), 16, 32)
			b, _ := strconv.ParseInt(string(data[3]), 16, 32)

			*c = Color(tcell.NewRGBColor(int32(r|r<<16), int32(g|g<<16), int32(b|b<<16)))

			return nil
		}

		if data[0] == '#' && len(data) == 7 {
			r, _ := strconv.ParseInt(string(data[1:3]), 16, 32)
			g, _ := strconv.ParseInt(string(data[3:5]), 16, 32)
			b, _ := strconv.ParseInt(string(data[5:7]), 16, 32)

			*c = Color(tcell.NewRGBColor(int32(r), int32(g), int32(b)))

			return nil
		}

		*c = Color(tcell.ColorDefault)

		return nil

	case int64:
		*c = Color(tcell.Color(uint64(tcell.ColorValid) + uint64(data)))

		return nil
	}

	return nil
}

type EditorTheme struct {
	Foreground                     Color            `toml:"foreground"`
	Background                     Color            `toml:"background"`
	SelectionBackground            Color            `toml:"selection_background"`
	CursorColor                    Color            `toml:"cursor_color"`
	LineNumberBackground           Color            `toml:"line_number_background"`
	LineNumberForeground           Color            `toml:"line_number_foreground"`
	ActiveLineNumberBackground     Color            `toml:"active_line_number_background"`
	ActiveLineNumberForeground     Color            `toml:"active_line_number_foreground"`
	Highlights                     map[string]Color `toml:"highlights"`
	DiagnosticLineNumberBackground [4]Color         `toml:"diagnostic_line_number_background"`
}

var DefaultTheme = EditorTheme{
	Foreground:                 Color(tcell.Color231.TrueColor()),
	Background:                 Color(tcell.Color235.TrueColor()),
	SelectionBackground:        Color(tcell.Color240.TrueColor()),
	CursorColor:                Color(tcell.Color245.TrueColor()),
	LineNumberBackground:       Color(tcell.Color235.TrueColor()),
	LineNumberForeground:       Color(tcell.Color245.TrueColor()),
	ActiveLineNumberBackground: Color(tcell.Color235.TrueColor()),
	ActiveLineNumberForeground: Color(tcell.Color250.TrueColor()),
	Highlights: map[string]Color{
		"constant":           Color(tcell.Color147.TrueColor()),
		"comment":            Color(tcell.Color245.TrueColor()),
		"keyword":            Color(tcell.Color153.TrueColor()),
		"string":             Color(tcell.Color216.TrueColor()),
		"function":           Color(tcell.Color115.TrueColor()),
		"type":               Color(tcell.Color115.TrueColor()),
		"variable.parameter": Color(tcell.Color115.TrueColor()),
	},
	DiagnosticLineNumberBackground: [4]Color{
		Color(tcell.Color52),
		Color(tcell.ColorDarkGoldenrod),
		Color(tcell.Color17),
		Color(tcell.Color53),
	},
}
