//go:build windows

package treesitter

import (
	"golang.org/x/sys/windows"
)

func openLibrary(name string) (uintptr, error) {
	handle := windows.NewLazyDLL(name)

	err := handle.Load()
	if err != nil {
		return 0, err
	}

	return handle.Handle(), nil
}
