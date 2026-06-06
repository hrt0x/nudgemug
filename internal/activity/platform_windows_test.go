//go:build windows

package activity

import (
	"testing"
	"unsafe"
)

func TestSendInputStructSizes(t *testing.T) {
	if got := unsafe.Sizeof(input{}); got != 40 {
		t.Fatalf("input size = %d, want 40", got)
	}
	if got := unsafe.Sizeof(keybdInput{}); got != 24 {
		t.Fatalf("keybdInput size = %d, want 24", got)
	}
	if got := unsafe.Sizeof(mouseInput{}); got != 32 {
		t.Fatalf("mouseInput size = %d, want 32", got)
	}
}
