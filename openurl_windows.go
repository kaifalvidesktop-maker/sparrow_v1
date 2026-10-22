//go:build windows

package main

import "os/exec"

func openURLCmd(u string) *exec.Cmd {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
}
