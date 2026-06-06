package main

import (
	"strings"
	"testing"
	"time"
)

func TestSessionSummaryStopped(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)

	if got := sessionSummary(now, false, time.Time{}); got != "" {
		t.Fatalf("sessionSummary stopped = %q, want empty", got)
	}
}

func TestSessionSummaryUntilStopped(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)

	if got := sessionSummary(now, true, time.Time{}); got != "until stopped" {
		t.Fatalf("sessionSummary infinite = %q", got)
	}
}

func TestSessionSummaryFinite(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)
	end := now.Add(time.Hour + 23*time.Minute + 45*time.Second)

	got := sessionSummary(now, true, end)
	want := "stops 13:23"
	if got != want {
		t.Fatalf("sessionSummary finite = %q, want %q", got, want)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{name: "minutes", duration: 14*time.Minute + 5*time.Second, want: "14:05"},
		{name: "hours", duration: 2*time.Hour + 3*time.Minute + 4*time.Second, want: "2:03:04"},
		{name: "negative", duration: -time.Second, want: "0:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.duration); got != tt.want {
				t.Fatalf("formatDuration = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildStatusHeader(t *testing.T) {
	got := buildStatusHeader("Active", modeCaffeine, "stops 13:23", "last nudge 12:00:00")
	want := "Active · Caffeine · stops 13:23 · last nudge 12:00:00"
	if got != want {
		t.Fatalf("buildStatusHeader = %q, want %q", got, want)
	}
}

func TestNudgeStatusIncludesNextNudge(t *testing.T) {
	next := time.Now().Add(14 * time.Second)

	got := buildStatusHeader("Active", modeNudge, "until stopped", nudgeSummary(true, modeNudge, next, "last nudge 12:00:00"))
	for _, want := range []string{"Active · Nudge · until stopped", "next nudge", "last nudge 12:00:00"} {
		if !strings.Contains(got, want) {
			t.Fatalf("buildStatusHeader = %q, want to contain %q", got, want)
		}
	}
}

func TestNudgeStatusOmitsNextWhenStopped(t *testing.T) {
	got := nudgeSummary(false, modeNudge, time.Time{}, "last nudge pending")
	want := "last nudge pending"
	if got != want {
		t.Fatalf("nudgeSummary stopped = %q, want %q", got, want)
	}
}

func TestBuildTooltipStaysShort(t *testing.T) {
	got := buildTooltip("Active", modeNudge, "stops 13:23")
	want := "NudgeMug dev · Active (Nudge) · stops 13:23"
	if got != want {
		t.Fatalf("buildTooltip = %q, want %q", got, want)
	}
	if len(got) > 80 {
		t.Fatalf("buildTooltip length = %d, want short enough for tray tooltip", len(got))
	}
}

func TestTimeUntilNextEventUsesSessionEnd(t *testing.T) {
	now := time.Now()
	app := &trayApp{
		interval:   defaultInterval,
		mode:       modeNudge,
		nextNudge:  now.Add(time.Hour),
		sessionEnd: now.Add(time.Minute),
	}

	got := app.timeUntilNextEvent()
	if got < 55*time.Second || got > 65*time.Second {
		t.Fatalf("timeUntilNextEvent = %s, want about 1m", got)
	}
}

func TestTimeUntilNextEventCaffeineUsesSessionEnd(t *testing.T) {
	now := time.Now()
	app := &trayApp{
		interval:   defaultInterval,
		mode:       modeCaffeine,
		nextNudge:  now.Add(5 * time.Second),
		sessionEnd: now.Add(time.Minute),
	}

	got := app.timeUntilNextEvent()
	if got < 55*time.Second || got > 65*time.Second {
		t.Fatalf("timeUntilNextEvent caffeine = %s, want about 1m", got)
	}
}

func TestCaffeineMechanismUsesKeeper(t *testing.T) {
	keeper := &fakeKeeper{}
	app := &trayApp{keeper: keeper}

	if err := app.startMechanism(modeCaffeine); err != nil {
		t.Fatal(err)
	}
	app.stopMechanism(modeCaffeine)

	if keeper.starts != 1 || keeper.stops != 1 {
		t.Fatalf("keeper starts/stops = %d/%d, want 1/1", keeper.starts, keeper.stops)
	}
}

func TestNudgeMechanismUsesKeeper(t *testing.T) {
	keeper := &fakeKeeper{}
	app := &trayApp{keeper: keeper}

	if err := app.startMechanism(modeNudge); err != nil {
		t.Fatal(err)
	}
	app.stopMechanism(modeNudge)

	if keeper.starts != 1 || keeper.stops != 1 {
		t.Fatalf("keeper starts/stops = %d/%d, want 1/1", keeper.starts, keeper.stops)
	}
}

func TestPerformNudgeUsesNudger(t *testing.T) {
	nudger := &fakeNudger{}
	app := &trayApp{
		nudger:   nudger,
		running:  true,
		interval: defaultInterval,
	}

	if !app.performNudge() {
		t.Fatal("performNudge returned false")
	}

	if nudger.nudges != 1 {
		t.Fatalf("nudger nudges = %d, want 1", nudger.nudges)
	}
}

type fakeKeeper struct {
	starts int
	stops  int
}

func (f *fakeKeeper) Name() string { return "fake-keeper" }
func (f *fakeKeeper) Start() error {
	f.starts++
	return nil
}
func (f *fakeKeeper) Stop() error {
	f.stops++
	return nil
}

type fakeNudger struct {
	nudges int
}

func (f *fakeNudger) Name() string { return "fake-nudger" }
func (f *fakeNudger) Nudge() error {
	f.nudges++
	return nil
}
