package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/hrt0x/nudgemug/internal/activity"
	"github.com/hrt0x/nudgemug/internal/keepawake"
)

const (
	defaultInterval = time.Minute
)

var version = "dev"

type keepMode string

const (
	modeCaffeine keepMode = "Caffeine"
	modeNudge    keepMode = "Nudge"
)

type sessionChoice struct {
	label    string
	tooltip  string
	duration time.Duration
}

var sessionChoices = []sessionChoice{
	{label: "Start for 15 min", tooltip: "Keep awake for 15 minutes", duration: 15 * time.Minute},
	{label: "Start for 1 hour", tooltip: "Keep awake for 1 hour", duration: time.Hour},
	{label: "Start for 2 hours", tooltip: "Keep awake for 2 hours", duration: 2 * time.Hour},
	{label: "Start until stopped / infinity", tooltip: "Keep awake until you stop it", duration: 0},
}

func main() {
	app := &trayApp{
		nudger:     activity.NewNudger(),
		keeper:     keepawake.NewKeeper(),
		mode:       modeNudge,
		interval:   defaultInterval,
		nudgeNowCh: make(chan struct{}, 1),
		resetCh:    make(chan struct{}, 1),
		done:       make(chan struct{}),
	}
	systray.Run(app.ready, app.exit)
}

type trayApp struct {
	mu         sync.Mutex
	nudgerMu   sync.Mutex
	nudger     activity.Nudger
	keeper     keepawake.Keeper
	mode       keepMode
	interval   time.Duration
	cancel     context.CancelFunc
	running    bool
	nextNudge  time.Time
	sessionEnd time.Time
	lastNudge  time.Time
	lastErr    error

	nudgeNowCh chan struct{}
	resetCh    chan struct{}
	done       chan struct{}
	doneOnce   sync.Once

	caffeineItem   *systray.MenuItem
	nudgeItem      *systray.MenuItem
	quickStartItem *systray.MenuItem
	startItems     []*systray.MenuItem
	stopItem       *systray.MenuItem
	nudgeNowItem   *systray.MenuItem
	statusItem     *systray.MenuItem
}

func (app *trayApp) ready() {
	systray.SetIcon(iconICO(false))
	systray.SetTitle("NudgeMug")

	app.statusItem = systray.AddMenuItem("", "Current NudgeMug status")
	app.statusItem.Disable()
	systray.AddSeparator()
	app.nudgeNowItem = systray.AddMenuItem("Nudge now", "Send one visible mouse nudge plus harmless F15 key")
	systray.AddSeparator()
	app.quickStartItem = systray.AddMenuItem("Start 1h · Nudge", "Keep your machine active for one hour")
	app.stopItem = systray.AddMenuItem("Stop", "Stop the active keep-awake session")
	app.stopItem.Disable()
	systray.AddSeparator()
	app.caffeineItem = systray.AddMenuItem("Mode: Caffeine", "Use OS keep-awake APIs; safest default")
	app.nudgeItem = systray.AddMenuItem("Mode: Nudge", "Use visible mouse nudges for idle-sensitive apps")
	systray.AddSeparator()
	for _, choice := range sessionChoices {
		app.startItems = append(app.startItems, systray.AddMenuItem(choice.label, choice.tooltip))
	}
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Quit NudgeMug")

	app.applyModeControls()
	app.updateStatus()
	go app.refreshLoop()
	go app.handleClicks(quitItem)
}

func (app *trayApp) handleClicks(quitItem *systray.MenuItem) {
	for {
		select {
		case <-app.nudgeNowItem.ClickedCh:
			app.nudgeNow()
		case <-app.quickStartItem.ClickedCh:
			app.startNudge(time.Hour)
		case <-app.startItems[0].ClickedCh:
			app.start(sessionChoices[0].duration)
		case <-app.startItems[1].ClickedCh:
			app.start(sessionChoices[1].duration)
		case <-app.startItems[2].ClickedCh:
			app.start(sessionChoices[2].duration)
		case <-app.startItems[3].ClickedCh:
			app.start(sessionChoices[3].duration)
		case <-app.caffeineItem.ClickedCh:
			app.selectMode(modeCaffeine)
		case <-app.nudgeItem.ClickedCh:
			app.selectMode(modeNudge)
		case <-app.stopItem.ClickedCh:
			app.stop()
		case <-quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (app *trayApp) startNudge(duration time.Duration) {
	app.selectMode(modeNudge)
	app.start(duration)
}

func (app *trayApp) start(duration time.Duration) {
	drainNudgeRequests(app.nudgeNowCh)

	var ctx context.Context
	startLoop := false

	app.mu.Lock()
	now := time.Now()
	mode := app.mode
	if !app.running {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		app.cancel = cancel
		app.running = true
		startLoop = true
	}
	if duration > 0 {
		app.sessionEnd = now.Add(duration)
	} else {
		app.sessionEnd = time.Time{}
	}
	app.nextNudge = now.Add(app.interval)
	app.lastErr = nil
	app.mu.Unlock()

	if startLoop {
		if err := app.startMechanism(mode); err != nil {
			app.fail(err)
			return
		}
	}
	app.applyRunningControls()
	app.updateStatus()
	app.requestTimerReset()
	if startLoop {
		go app.loop(ctx)
	}
}

func (app *trayApp) stop() {
	app.mu.Lock()
	mode := app.mode
	app.stopLocked()
	app.mu.Unlock()
	app.stopMechanism(mode)
	app.applyStoppedControls()
	app.updateStatus()
}

func (app *trayApp) stopLocked() {
	if app.cancel != nil {
		app.cancel()
		app.cancel = nil
	}
	drainNudgeRequests(app.nudgeNowCh)
	app.running = false
	app.nextNudge = time.Time{}
	app.sessionEnd = time.Time{}
}

func (app *trayApp) loop(ctx context.Context) {
	timer := time.NewTimer(app.timeUntilNextEvent())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if app.expireIfDue() {
				return
			}
			if app.currentMode() == modeNudge {
				if !app.performNudge() {
					return
				}
			}
			resetTimer(timer, app.timeUntilNextEvent())
		case <-app.nudgeNowCh:
			if app.currentMode() != modeNudge {
				resetTimer(timer, app.timeUntilNextEvent())
				continue
			}
			if !app.performNudge() {
				return
			}
			resetTimer(timer, app.timeUntilNextEvent())
		case <-app.resetCh:
			resetTimer(timer, app.timeUntilNextEvent())
		}
	}
}

func (app *trayApp) nudgeNow() {
	if app.currentMode() == modeNudge {
		select {
		case app.nudgeNowCh <- struct{}{}:
		default:
		}
		return
	}

	app.performNudge()
}

func (app *trayApp) performNudge() bool {
	if app.expireIfDue() {
		return false
	}

	app.nudgerMu.Lock()
	err := app.nudger.Nudge()
	app.nudgerMu.Unlock()
	now := time.Now()

	app.mu.Lock()
	if !app.running {
		app.mu.Unlock()
		return false
	}
	if err != nil {
		app.lastErr = err
		app.stopLocked()
		app.mu.Unlock()
		app.applyStoppedControls()
		app.updateStatus()
		return false
	}

	app.lastErr = nil
	app.lastNudge = now
	app.nextNudge = now.Add(app.interval)
	app.mu.Unlock()

	app.updateStatus()
	return true
}

func (app *trayApp) selectMode(mode keepMode) {
	app.mu.Lock()
	oldMode := app.mode
	if oldMode == mode {
		app.mu.Unlock()
		return
	}

	running := app.running
	if running {
		app.stopMechanism(oldMode)
		if err := app.startMechanism(mode); err != nil {
			app.mode = mode
			app.lastErr = err
			app.stopLocked()
			app.mu.Unlock()
			app.stopMechanism(mode)
			app.applyStoppedControls()
			app.updateStatus()
			return
		}
	}

	app.mode = mode
	if running && mode == modeNudge {
		app.nextNudge = time.Now().Add(app.interval)
	}
	if mode == modeCaffeine {
		drainNudgeRequests(app.nudgeNowCh)
	}
	app.lastErr = nil
	app.mu.Unlock()

	app.applyModeControls()
	app.updateStatus()
	app.requestTimerReset()
}

func (app *trayApp) currentMode() keepMode {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.mode
}

func (app *trayApp) startMechanism(mode keepMode) error {
	if mode == modeCaffeine || mode == modeNudge {
		return app.keeper.Start()
	}
	return nil
}

func (app *trayApp) stopMechanism(mode keepMode) {
	if mode == modeCaffeine || mode == modeNudge {
		_ = app.keeper.Stop()
	}
}

func (app *trayApp) requestTimerReset() {
	select {
	case app.resetCh <- struct{}{}:
	default:
	}
}

func (app *trayApp) exit() {
	app.mu.Lock()
	mode := app.mode
	app.stopLocked()
	app.mu.Unlock()
	app.stopMechanism(mode)
	app.doneOnce.Do(func() {
		close(app.done)
	})
}

func (app *trayApp) applyRunningControls() {
	systray.SetIcon(iconICO(true))
	app.stopItem.Enable()
	app.applyModeControls()
}

func (app *trayApp) applyStoppedControls() {
	systray.SetIcon(iconICO(false))
	app.stopItem.Disable()
	app.applyModeControls()
}

func (app *trayApp) applyModeControls() {
	app.mu.Lock()
	mode := app.mode
	running := app.running
	app.mu.Unlock()

	if mode == modeCaffeine {
		app.caffeineItem.Check()
		app.nudgeItem.Uncheck()
	} else {
		app.caffeineItem.Uncheck()
		app.nudgeItem.Check()
	}

	_ = running
}

func (app *trayApp) refreshLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.done:
			return
		case <-ticker.C:
			app.updateStatus()
		}
	}
}

func (app *trayApp) updateStatus() {
	app.mu.Lock()
	running := app.running
	mode := app.mode
	nextNudge := app.nextNudge
	sessionEnd := app.sessionEnd
	lastNudge := app.lastNudge
	lastErr := app.lastErr
	app.mu.Unlock()

	state := "Stopped"
	if running {
		state = "Active"
	}

	session := sessionSummary(time.Now(), running, sessionEnd)
	last := "last nudge pending"
	if !lastNudge.IsZero() {
		last = "last nudge " + lastNudge.Format("15:04:05")
	}

	header := buildStatusHeader(state, mode, session, nudgeSummary(running, mode, nextNudge, last))
	if lastErr != nil {
		header = fmt.Sprintf("%s · %s · error: %s · %s", state, mode, lastErr.Error(), last)
	}

	if app.statusItem != nil {
		app.statusItem.SetTitle(header)
		systray.SetTooltip(buildTooltip(state, mode, session))
	}
}

func (app *trayApp) timeUntilNextEvent() time.Duration {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.mode == modeCaffeine {
		if app.sessionEnd.IsZero() {
			return app.interval
		}
		return time.Until(app.sessionEnd)
	}

	next := app.nextNudge
	if next.IsZero() {
		next = time.Now().Add(app.interval)
	}
	if !app.sessionEnd.IsZero() && app.sessionEnd.Before(next) {
		next = app.sessionEnd
	}
	return time.Until(next)
}

func (app *trayApp) expireIfDue() bool {
	now := time.Now()

	app.mu.Lock()
	if !app.running || app.sessionEnd.IsZero() || now.Before(app.sessionEnd) {
		app.mu.Unlock()
		return false
	}
	mode := app.mode
	app.stopLocked()
	app.mu.Unlock()

	app.stopMechanism(mode)
	app.applyStoppedControls()
	app.updateStatus()
	return true
}

func (app *trayApp) fail(err error) {
	app.mu.Lock()
	mode := app.mode
	app.lastErr = err
	app.stopLocked()
	app.mu.Unlock()

	app.stopMechanism(mode)
	app.applyStoppedControls()
	app.updateStatus()
}

func (app *trayApp) mechanismNameLocked() string {
	if app.mode == modeCaffeine {
		return app.keeper.Name()
	}
	return app.nudger.Name()
}
