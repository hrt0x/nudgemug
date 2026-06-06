//go:build !windows

package keepawake

import "errors"

func platformKeeper() Keeper {
	return UnsupportedKeeper{}
}

type UnsupportedKeeper struct{}

func (UnsupportedKeeper) Name() string { return "unsupported-caffeine" }
func (UnsupportedKeeper) Start() error {
	return errors.New("caffeine mode is only implemented for Windows right now")
}
func (UnsupportedKeeper) Stop() error { return nil }
