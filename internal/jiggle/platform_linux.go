//go:build linux

package jiggle

import (
	"fmt"
	"os/exec"
)

func platformMover() Mover {
	if commandExists("xdotool") {
		return XDoToolMover{}
	}
	return nil
}

type XDoToolMover struct{}

func (XDoToolMover) Name() string { return "xdotool" }
func (XDoToolMover) MoveRelative(dx, dy int) error {
	return exec.Command("xdotool", "mousemove_relative", "--", fmt.Sprint(dx), fmt.Sprint(dy)).Run()
}
