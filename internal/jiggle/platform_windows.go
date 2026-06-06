//go:build windows

package jiggle

import (
	"errors"
	"syscall"
)

func platformMover() Mover {
	return WindowsMover{}
}

type WindowsMover struct{}

func (WindowsMover) Name() string { return "win32-mouse_event" }
func (WindowsMover) MoveRelative(dx, dy int) error {
	user32 := syscall.NewLazyDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")

	const mouseEventMove = 0x0001
	_, _, err := mouseEvent.Call(
		uintptr(mouseEventMove),
		uintptr(int32(dx)),
		uintptr(int32(dy)),
		0,
		0,
	)
	return errOrZeroOK("mouse_event failed", err)
}

func errOrZeroOK(message string, err error) error {
	// mouse_event returns void; syscall reports Errno(0) on success.
	if err != nil && !errors.Is(err, syscall.Errno(0)) {
		return err
	}
	return nil
}

func errOr(message string, err error) error {
	if err != nil && !errors.Is(err, syscall.Errno(0)) {
		return err
	}
	return errors.New(message)
}
