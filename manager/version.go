package manager

import (
	"debug/buildinfo"
	"os"
)

var version string

func GetVersion() string {
	exefile, _ := os.Executable()
	info, _ := buildinfo.ReadFile(exefile)

	v := info.Main.Version
	if v == "(devel)" && version != "" {
		v = version
	}

	return v
}
