package main

import (
	"fmt"
	"time"
)

func sessionSummary(now time.Time, running bool, sessionEnd time.Time) string {
	if !running {
		return ""
	}
	if sessionEnd.IsZero() {
		return "until stopped"
	}
	return "stops " + sessionEnd.Format("15:04")
}

func nudgeSummary(running bool, mode keepMode, nextNudge time.Time, last string) string {
	if mode != modeNudge {
		return last
	}
	if !running {
		return last
	}
	next := "next nudge pending"
	if !nextNudge.IsZero() {
		next = fmt.Sprintf("next nudge %ds", secondsUntil(nextNudge))
	}
	return fmt.Sprintf("%s · %s", next, last)
}

func buildStatusHeader(state string, mode keepMode, session, detail string) string {
	if session == "" {
		return fmt.Sprintf("%s · %s · %s", state, mode, detail)
	}
	return fmt.Sprintf("%s · %s · %s · %s", state, mode, session, detail)
}

func buildTooltip(state string, mode keepMode, session string) string {
	if session == "" {
		return fmt.Sprintf("NudgeMug %s · %s (%s)", version, state, mode)
	}
	return fmt.Sprintf("NudgeMug %s · %s (%s) · %s", version, state, mode, session)
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	seconds := int(duration.Round(time.Second) / time.Second)
	hours := seconds / 3600
	minutes := seconds % 3600 / 60
	seconds = seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func secondsUntil(deadline time.Time) int {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 0
	}
	return int((remaining + time.Second - time.Nanosecond) / time.Second)
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	if duration < 0 {
		duration = 0
	}
	timer.Reset(duration)
}

func drainNudgeRequests(ch <-chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
