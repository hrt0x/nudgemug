package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hrt0x/nudgemug/internal/jiggle"
)

var version = "dev"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: nudgemug [flags]\n\nKeep your laptop politely alive with tiny reversible mouse nudges.\n\nFlags:\n")
		flag.PrintDefaults()
	}

	interval := flag.Duration("interval", 30*time.Second, "time between mouse nudges")
	distance := flag.Int("distance", 1, "pixels to move each way")
	count := flag.Int("count", 0, "stop after N nudges; 0 means forever")
	dryRun := flag.Bool("dry-run", false, "print actions without moving the mouse")
	quiet := flag.Bool("quiet", false, "reduce output")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *interval <= 0 {
		fatal("--interval must be positive")
	}
	if *distance <= 0 {
		fatal("--distance must be positive")
	}
	if *count < 0 {
		fatal("--count cannot be negative")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	mover := jiggle.NewMover(*dryRun)
	if !*quiet {
		fmt.Printf("nudgemug: keeping laptop politely alive via %s every %s\n", mover.Name(), interval.String())
		fmt.Println("nudgemug: Ctrl+C to stop")
	}

	run(ctx, mover, *interval, *distance, *count, *quiet)
}

func run(ctx context.Context, mover jiggle.Mover, interval time.Duration, distance int, count int, quiet bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	moves := 0
	for {
		select {
		case <-ctx.Done():
			if !quiet {
				fmt.Println("nudgemug: nap time")
			}
			return
		case <-ticker.C:
			moves++
			if err := nudge(mover, distance); err != nil {
				fatal(err.Error())
			}
			if !quiet {
				fmt.Printf("nudgemug: nudge %d\n", moves)
			}
			if count > 0 && moves >= count {
				if !quiet {
					fmt.Println("nudgemug: done")
				}
				return
			}
		}
	}
}

func nudge(mover jiggle.Mover, distance int) error {
	if err := mover.MoveRelative(distance, 0); err != nil {
		return err
	}
	return mover.MoveRelative(-distance, 0)
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "nudgemug:", message)
	os.Exit(1)
}
