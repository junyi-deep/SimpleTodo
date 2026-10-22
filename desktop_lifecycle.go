package main

import "context"

// Native dispatch and notifications require Wails' ApplicationStarted event.
// A cancelled startup must not leave workers waiting or dispatch after shutdown.
func afterDesktopReady(ctx context.Context, ready <-chan struct{}, run func()) {
	select {
	case <-ctx.Done():
		return
	case <-ready:
	}
	if ctx.Err() == nil {
		run()
	}
}
