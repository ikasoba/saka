package manager

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ikasoba/saka/config"
)

type TreeSitterJson struct {
	Grammars []TreeSitterGrammar `json:"grammars"`
}

type TreeSitterGrammar struct {
	Name string `json:"name"`
}

func InstallGrammar(repo string, ref string) error {
	work, err := os.MkdirTemp("", "saka-grammar")
	if err != nil {
		return err
	}

	args := []string{"clone", "--depth", "1"}

	if ref != "" {
		args = append(args, "-b", ref)
	}

	args = append(args, repo, work)

	cmd := exec.Command("git", args...)

	err = cmd.Run()
	if err != nil {
		return err
	}

	var ts TreeSitterJson

	fd, err := os.Open(filepath.Join(work, "tree-sitter.json"))
	if err != nil {
		return err
	}

	defer fd.Close()

	d := json.NewDecoder(fd)

	err = d.Decode(&ts)
	if err != nil {
		return err
	}

	name := ""
	for _, g := range ts.Grammars {
		name = g.Name
		break
	}

	dylibExtname := ".so"
	switch runtime.GOOS {
	case "darwin":
		dylibExtname = ".dylib"

	case "windows":
		dylibExtname = ".dll"
	}

	cmd = exec.Command("tree-sitter", "build", "-o", name+dylibExtname)
	cmd.Dir = work

	err = cmd.Run()
	if err != nil {
		return err
	}

	configPath := config.GetConfigDir()

	grammarPath := filepath.Join(configPath, "grammars", name+dylibExtname)
	highlightsPath := filepath.Join(configPath, "queries", name, "highlights.scm")

	os.MkdirAll(filepath.Dir(grammarPath), os.ModePerm)
	os.MkdirAll(filepath.Dir(highlightsPath), os.ModePerm)

	grammar, err := os.Open(filepath.Join(work, name+dylibExtname))
	if err != nil {
		return nil
	}

	defer grammar.Close()

	dest_grammar, err := os.Create(grammarPath)
	if err != nil {
		return nil
	}

	defer dest_grammar.Close()

	dest_grammar.ReadFrom(grammar)

	highlights, err := os.Open(filepath.Join(work, "queries", "highlights.scm"))
	if err == nil {
		defer highlights.Close()

		dest_highlights, err := os.Create(highlightsPath)
		if err != nil {
			return err
		}

		defer dest_highlights.Close()

		dest_highlights.ReadFrom(highlights)
	}

	return nil
}
