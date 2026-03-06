//go:build !windows

package git

import "os/exec"

func hideConsole(cmd *exec.Cmd) {}
