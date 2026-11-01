package main

import "testing"

func TestRestartCheckWithoutReadyWindowsDoesNotQuit(t *testing.T) {
	s := &TodoService{}
	allow, err := s.approveRestart()
	if err != nil || !allow {
		t.Fatal(allow, err)
	}
	if s.quitActive || s.quitCheck != nil {
		t.Fatal("restart guard was not released")
	}
}
func TestRestartCancellationKeepsApplicationRunning(t *testing.T) {
	ch := make(chan bool, 1)
	s := &TodoService{quitActive: true, quitToken: 7, quitWaiting: "review", quitQueue: []string{"main"}, quitCheck: ch}
	if err := s.RespondQuit("review", 7, false); err != nil {
		t.Fatal(err)
	}
	if <-ch {
		t.Fatal("cancelled draft check approved restart")
	}
	if s.quitActive || s.quitCheck != nil || s.quitWaiting != "" || len(s.quitQueue) != 0 {
		t.Fatal("cancelled check leaked pending state")
	}
	if s.forceQuit.Load() {
		t.Fatal("cancelled restart quit application")
	}
}
