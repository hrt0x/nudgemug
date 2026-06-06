# NudgeMug

Keep your laptop politely alive.

`NudgeMug` is a tiny CLI and Windows tray app for finite keep-awake sessions. It is for reading, demos, long builds, downloads, presentations, remote sessions, and those suspiciously human pauses where you are just thinking.

> Status: tray-first useful-market build. Small, visible, and intentionally not a platform.

## Install

Clone and run with Go 1.22+:

```bash
git clone https://github.com/hrt0x/nudgemug.git
cd nudgemug
go run ./cmd/nudgemug --dry-run --count 3
```

Build the CLI:

```bash
go build -o nudgemug ./cmd/nudgemug
./nudgemug --version
```

Cross-build Windows binaries from Linux/WSL:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/nudgemug-cli.exe ./cmd/nudgemug
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o /tmp/nudgemug-tray.exe ./cmd/nudgemug-tray
```

Tagged releases publish:

- `nudgemug-windows-amd64.exe`
- `nudgemug-tray-windows-amd64.exe`
- `nudgemug-windows-amd64.zip`
- `checksums.txt`

## CLI

Dry run:

```bash
go run ./cmd/nudgemug --dry-run --interval 10ms --count 2
```

Real nudge:

```bash
go run ./cmd/nudgemug --interval 30s --distance 1
```

Stop with `Ctrl+C`.

Flags:

- `--interval`: time between nudges, default `30s`
- `--distance`: pixels to move each way, default `1`
- `--count`: stop after N nudges, default `0` = forever
- `--dry-run`: print actions without moving the mouse
- `--quiet`: reduce output
- `--version`: print version and exit

## Windows Tray

Build or download `nudgemug-tray-windows-amd64.exe`, then run it on Windows. The tray menu has:

- a status header with `Active` / `Stopped`
- `Mode: Nudge` as the default
- `Mode: Caffeine` when you only need OS keep-awake behavior
- one-click quick start: `Start 1h · Nudge`
- finite duration sessions: `Start for 15 min`, `Start for 1 hour`, `Start for 2 hours`
- `Start until stopped / infinity`
- auto-stop when a finite session expires
- `next nudge in Ns` countdown while Nudge is active
- `last nudge HH:MM:SS` after the first nudge
- `Stop`: stop the active keep-awake session
- `Nudge now`: do one visible mouse nudge plus harmless F15 key
- `Quit`: exit

Defaults are now Nudge-first: finite sessions visible first, visible reversible SendInput mouse movement plus F15 activity signals every minute, and Windows keep-awake held while Nudge is active. The tray app uses a self-contained built-in icon and does not require admin rights.

Runtime smoke-test note: the tray app is built from this repo and cross-compiled in CI, but the actual tray menu must be manually checked on native Windows after download/build.

### Modes

**Nudge mode** uses Windows `SendInput` to move the cursor in a small visible reversible square and send a harmless virtual `F15` key down/up pulse every minute, plus `SetThreadExecutionState` to keep the system and display awake. This follows the same modern input path used by mature mouse movers, while `F15` adds a second activity signal without typing text or triggering normal shortcuts.

**Caffeine mode** uses only the Windows `SetThreadExecutionState` API to keep the system and display awake while the session is active. It covers downloads, builds, presentations, renders, uploads, and remote sessions, but it does not create user activity for apps that watch the OS idle timer.

Switching modes while a session is active keeps the same duration session, stops the old mechanism, starts the new mechanism, and updates the status.

### Tray feedback

The tray stays tray-only: no window, no settings panel, no custom popover. Start/Stop swap between active and stopped icons, and the tooltip mirrors the menu header so you can confirm what NudgeMug is doing without opening anything else.

Screenshot pending. Expected menu shape:

```text
Active - Nudge - 1:23:45 left - stops 17:00 - next nudge in 14s - last nudge 16:42:10
----------------
Nudge now
----------------
Start 1h · Nudge
Stop
----------------
Mode: Nudge
Mode: Caffeine
----------------
Start for 15 min
Start for 1 hour
Start for 2 hours
Start until stopped / infinity
----------------
Quit
```

This is still a native tray menu, not the custom concept UI. No custom window, popover, settings panel, CGO, service, scheduler, scripts, or macros was added. Any generated mockups are future north-star direction only.

## Platform Matrix

| Platform | CLI | Tray | Notes |
| --- | --- | --- | --- |
| Windows native | Yes | Yes | Tray supports Nudge + Caffeine |
| WSL | Yes | No | CLI nudge uses `powershell.exe` + `mouse_event` from WSL |
| Linux | Yes | No | CLI uses `xdotool` if installed |
| macOS | Dry-run only | No | Not polished for v0 |

## Responsible Use

Use this for your own machine and legitimate workflows. Respect your organization's policies and other people's systems.

`NudgeMug` does not persistently change power settings, install services, autostart itself, phone home, collect analytics, run scripts/macros, or need admin privileges.

## Release

Pushing a tag like `v1.0.0` runs the GitHub Actions release workflow on `windows-latest`. It builds and attaches the Windows CLI and tray binaries with the tag injected into `--version`.

Release validation from Linux/WSL:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/nudgemug-cli.exe ./cmd/nudgemug
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui -X main.version=v1.0.0-local" -o /tmp/nudgemug-tray.exe ./cmd/nudgemug-tray
```

Manual smoke-test on Windows before publishing a release:

- launch `nudgemug-tray-windows-amd64.exe`
- confirm default mode is Nudge
- start a 15 minute session and confirm status shows remaining time
- confirm `Nudge now` updates `last nudge`
- switch to Caffeine and confirm nudge countdown disappears
- stop/quit and confirm the session clears

## Roadmap

- [x] CLI skeleton
- [x] Dry-run mode
- [x] WSL/Windows CLI mouse nudge adapter
- [x] Native Windows tray app with Start/Stop/Quit
- [x] Tray confidence feedback and Nudge now
- [x] Menu-native duration sessions and auto-stop
- [x] Nudge default and Caffeine mode selection
- [x] Windows-first release workflow
- [ ] Native Windows tray smoke-test
- [ ] Optional micro-window decision
- [ ] Better cross-platform Caffeine support
- [ ] Start-at-login option
- [x] Icon polish
- [ ] Demo GIF

## Name

`NudgeMug` because it gives your machine a tiny nudge, then lets you get back to work.
