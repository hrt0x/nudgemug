//go:build windows

package keepawake

import (
	"errors"
	"runtime"
	"sync"
	"syscall"
)

const (
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
	esContinuous      = 0x80000000
)

func platformKeeper() Keeper {
	return NewWindowsKeeper()
}

type WindowsKeeper struct {
	mu     sync.Mutex
	active bool
	calls  chan executionStateCall
}

type executionStateCall struct {
	flags uintptr
	reply chan error
}

func NewWindowsKeeper() *WindowsKeeper {
	keeper := &WindowsKeeper{
		calls: make(chan executionStateCall),
	}
	go keeper.run()
	return keeper
}

func (keeper *WindowsKeeper) Name() string { return "caffeine-setthreadexecutionstate" }

func (keeper *WindowsKeeper) Start() error {
	keeper.mu.Lock()
	defer keeper.mu.Unlock()

	if err := keeper.setThreadExecutionState(esContinuous | esSystemRequired | esDisplayRequired); err != nil {
		return err
	}
	keeper.active = true
	return nil
}

func (keeper *WindowsKeeper) Stop() error {
	keeper.mu.Lock()
	defer keeper.mu.Unlock()

	if !keeper.active {
		return nil
	}
	if err := keeper.setThreadExecutionState(esContinuous); err != nil {
		return err
	}
	keeper.active = false
	return nil
}

func (keeper *WindowsKeeper) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for call := range keeper.calls {
		call.reply <- setThreadExecutionState(call.flags)
	}
}

func (keeper *WindowsKeeper) setThreadExecutionState(flags uintptr) error {
	reply := make(chan error)
	keeper.calls <- executionStateCall{flags: flags, reply: reply}
	return <-reply
}

func setThreadExecutionState(flags uintptr) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setThreadExecutionState := kernel32.NewProc("SetThreadExecutionState")

	result, _, err := setThreadExecutionState.Call(flags)
	if result == 0 {
		if err != nil && !errors.Is(err, syscall.Errno(0)) {
			return err
		}
		return errors.New("SetThreadExecutionState failed")
	}
	return nil
}
