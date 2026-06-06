//go:build !linux && !windows

package jiggle

func platformMover() Mover {
	return nil
}
