//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

func openURLCmd(u string) *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", u)
	}
	return exec.Command("xdg-open", u)
}
