package manager

import (
	"debug/buildinfo"
	"io"
	"net/http"
	"net/url"
	"os"
)

func FetchDefaultConfigs() (config io.ReadCloser, theme io.ReadCloser, err error) {
	exefile, _ := os.Executable()
	info, _ := buildinfo.ReadFile(exefile)

	base := ""

	v := GetVersion()

	if v == "(devel)" {
		base, err = url.JoinPath("https://raw.githubusercontent.com/ikasoba/saka/refs/heads/main")
		if err != nil {
			return nil, nil, err
		}
	} else {
		base, err = url.JoinPath("https://raw.githubusercontent.com/ikasoba/saka/refs/tags/", info.Main.Version)
		if err != nil {
			return nil, nil, err
		}
	}

	u, err := url.JoinPath(base, "/config.toml")
	if err != nil {
		return nil, nil, err
	}

	res, err := http.Get(u)
	if err != nil {
		return nil, nil, err
	}

	config = res.Body

	u, err = url.JoinPath(base, "/theme.toml")
	if err != nil {
		return nil, nil, err
	}

	res, err = http.Get(u)
	if err != nil {
		return nil, nil, err
	}

	theme = res.Body

	return config, theme, nil
}
