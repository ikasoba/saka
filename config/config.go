package config

type Config struct {
	Languages map[string]LanguageConfig `toml:"languages"`
}

type LanguageConfig struct {
	ServerCommand *string  `toml:"server_command"`
	Arguments     []string `toml:"arguments"`
	Patterns      []string `toml:"patterns"`
}
