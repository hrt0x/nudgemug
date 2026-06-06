//go:build windows

package activity

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	inputMouse    = 0
	inputKeyboard = 1

	mouseEventMove = 0x0001
	keyEventKeyUp  = 0x0002
	vkF15          = 0x7E
)

var (
	user32SendInput = syscall.NewLazyDLL("user32.dll").NewProc("SendInput")
	ErrUnsupported  = errors.New("activity nudger is unsupported on this platform")
)

func platformNudger() Nudger {
	return WindowsF15Nudger{}
}

type WindowsF15Nudger struct{}

func (WindowsF15Nudger) Name() string { return "sendinput-f15-mouse" }
func (WindowsF15Nudger) Nudge() error {
	if err := sendVisibleMouseSquare(10, 5*time.Millisecond); err != nil {
		return err
	}
	if err := sendKeyboardInput(vkF15, 0); err != nil {
		return err
	}
	return sendKeyboardInput(vkF15, keyEventKeyUp)
}

func sendVisibleMouseSquare(distance int, delay time.Duration) error {
	for _, move := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
		for step := 0; step < distance; step++ {
			if err := sendMouseMove(move[0], move[1]); err != nil {
				return err
			}
			time.Sleep(delay)
		}
	}
	return nil
}

type input struct {
	Type  uint32
	_     uint32
	Union [32]byte
}

type keybdInput struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type mouseInput struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

func sendMouseMove(dx, dy int) error {
	var event input
	event.Type = inputMouse
	*(*mouseInput)(unsafe.Pointer(&event.Union[0])) = mouseInput{
		Dx:    int32(dx),
		Dy:    int32(dy),
		Flags: mouseEventMove,
	}

	result, _, err := user32SendInput.Call(
		1,
		uintptr(unsafe.Pointer(&event)),
		unsafe.Sizeof(event),
	)
	if result != 1 {
		if err != nil && !errors.Is(err, syscall.Errno(0)) {
			return err
		}
		return fmt.Errorf("SendInput sent %d mouse events, want 1", result)
	}
	return nil
}

func sendKeyboardInput(vk uint16, flags uint32) error {
	var event input
	event.Type = inputKeyboard
	*(*keybdInput)(unsafe.Pointer(&event.Union[0])) = keybdInput{
		Vk:    vk,
		Flags: flags,
	}

	result, _, err := user32SendInput.Call(
		1,
		uintptr(unsafe.Pointer(&event)),
		unsafe.Sizeof(event),
	)
	if result != 1 {
		if err != nil && !errors.Is(err, syscall.Errno(0)) {
			return err
		}
		return fmt.Errorf("SendInput sent %d keyboard events, want 1", result)
	}
	return nil
}
