//go:build !windows

package activity

import "errors"

var ErrUnsupported = errors.New("activity nudger is unsupported on this platform")

func platformNudger() Nudger {
	return nil
}
