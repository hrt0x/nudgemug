package main

import (
	"errors"
	"testing"
)

type fakeMover struct {
	moves [][2]int
	errAt int
}

func (f *fakeMover) Name() string {
	return "fake"
}

func (f *fakeMover) MoveRelative(dx, dy int) error {
	f.moves = append(f.moves, [2]int{dx, dy})
	if f.errAt > 0 && len(f.moves) == f.errAt {
		return errors.New("boom")
	}
	return nil
}

func TestNudgeMovesOutAndBack(t *testing.T) {
	mover := &fakeMover{}

	if err := nudge(mover, 3); err != nil {
		t.Fatal(err)
	}

	want := [][2]int{{3, 0}, {-3, 0}}
	if len(mover.moves) != len(want) {
		t.Fatalf("moves = %v, want %v", mover.moves, want)
	}
	for i := range want {
		if mover.moves[i] != want[i] {
			t.Fatalf("moves = %v, want %v", mover.moves, want)
		}
	}
}

func TestNudgeReturnsMoverError(t *testing.T) {
	mover := &fakeMover{errAt: 1}

	if err := nudge(mover, 1); err == nil {
		t.Fatal("expected error")
	}
}
