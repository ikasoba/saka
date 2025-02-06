package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ikasoba/saka/theme"
	"github.com/kirsle/configdir"
)

type Config struct {
	Languages map[string]LanguageConfig `toml:"languages"`
	Grammars  map[string]GrammarConfig  `toml:"grammars"`
}

type LanguageConfig struct {
	ServerCommand *string  `toml:"server_command"`
	Arguments     []string `toml:"arguments"`
	Patterns      []string `toml:"patterns"`
}

type GrammarConfig struct {
	Repo string `toml:"repo"`
	Rev  string `toml:"rev"`
}

func GetConfigDir() string {
	configPath := os.Getenv("SAKA_CONFIG_DIR")
	if configPath == "" {
		configPath = configdir.LocalConfig("saka")
	}

	return configPath
}

func Load() (conf Config, theme theme.EditorTheme) {
	configPath := GetConfigDir()

	toml.DecodeFile(filepath.Join(configPath, "config.toml"), &conf)
	toml.DecodeFile(filepath.Join(configPath, "theme.toml"), &theme)

	return conf, theme
}
