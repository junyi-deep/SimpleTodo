package main

import (
	"context"
	"testing"
	"time"
)

func TestDesktopWorkersWaitForNativeStartup(t *testing.T) {
	ready, ran := make(chan struct{}), make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go afterDesktopReady(ctx, ready, func() { close(ran) })
	select {
	case <-ran:
		t.Fatal("worker ran before native startup")
	case <-time.After(20 * time.Millisecond):
	}
	close(ready)
	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("worker did not start after native startup")
	}
}

func TestCancelledDesktopStartupDoesNotRunWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ready, done := make(chan struct{}), make(chan struct{})
	cancel()
	go func() {
		afterDesktopReady(ctx, ready, func() { t.Error("worker ran after cancelled startup") })
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker kept waiting after startup cancellation")
	}
	close(ready)
	afterDesktopReady(ctx, ready, func() { t.Error("ready gate bypassed cancellation") })
}
