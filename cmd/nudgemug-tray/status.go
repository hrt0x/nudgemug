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
	remaining := sessionEnd.Sub(now)
	if remaining < 0 {
		remaining = 0
	}
	return fmt.Sprintf("%s left - stops %s", formatDuration(remaining), sessionEnd.Format("15:04"))
}

func nudgeSummary(running bool, mode keepMode, nextNudge time.Time, last string) string {
	if mode != modeNudge {
		return last
	}
	next := "next nudge in --"
	if running && !nextNudge.IsZero() {
		next = fmt.Sprintf("next nudge in %ds", secondsUntil(nextNudge))
	}
	return fmt.Sprintf("%s - %s", next, last)
}

func buildStatusHeader(state string, mode keepMode, session, detail string) string {
	if session == "" {
		return fmt.Sprintf("%s - %s - %s", state, mode, detail)
	}
	return fmt.Sprintf("%s - %s - %s - %s", state, mode, session, detail)
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
