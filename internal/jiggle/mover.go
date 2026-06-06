package jiggle

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Mover nudges the mouse by a relative delta.
type Mover interface {
	MoveRelative(dx, dy int) error
	Name() string
}

// NewMover selects the simplest available mover for the current machine.
func NewMover(dryRun bool) Mover {
	if dryRun {
		return DryRunMover{}
	}

	if isWSL() && commandExists(windowsPowerShellPath()) {
		return PowerShellMover{}
	}

	if mover := platformMover(); mover != nil {
		return mover
	}

	return UnsupportedMover{}
}

type DryRunMover struct{}

func (DryRunMover) Name() string { return "dry-run" }
func (DryRunMover) MoveRelative(dx, dy int) error {
	fmt.Printf("dry-run: move mouse %+d,%+d\n", dx, dy)
	return nil
}

type UnsupportedMover struct{}

func (UnsupportedMover) Name() string { return "unsupported" }
func (UnsupportedMover) MoveRelative(dx, dy int) error {
	return errors.New("no supported mouse mover found; try --dry-run, install xdotool on Linux, use native Windows, or run from WSL with powershell.exe available")
}

type PowerShellMover struct{}

func (PowerShellMover) Name() string { return "powershell-mouse_event" }
func (PowerShellMover) MoveRelative(dx, dy int) error {
	ps := commandName(windowsPowerShellPath(), "powershell")
	if ps == "" {
		return errors.New("powershell not found")
	}

	script := fmt.Sprintf(`Add-Type -TypeDefinition 'using System; using System.Runtime.InteropServices; public class Mouse { [DllImport("user32.dll")] public static extern void mouse_event(uint flags, int dx, int dy, uint data, UIntPtr extra); }'; [Mouse]::mouse_event(0x0001, %d, %d, 0, [UIntPtr]::Zero)`, dx, dy)
	cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run()
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func windowsPowerShellPath() string {
	return "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"
}

func commandName(names ...string) string {
	for _, name := range names {
		if commandExists(name) {
			return name
		}
	}
	return ""
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	data, err := os.ReadFile("/proc/version")
	return err == nil && strings.Contains(strings.ToLower(string(data)), "microsoft")
}
