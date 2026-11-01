package main

import (
	"context"
	"fmt"
	"time"
)

// Visit ready webviews one by one so each can protect unsaved task/note drafts.
func (s *TodoService) RegisterWindow(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.windows[key] == nil {
		return fmt.Errorf("未知窗口")
	}
	if s.readyWindows == nil {
		s.readyWindows = map[string]bool{}
	}
	s.readyWindows[key] = true
	return nil
}
func (s *TodoService) RequestQuit() error {
	s.beginQuitChecks(nil)
	return nil
}
func (s *TodoService) beginQuitChecks(check chan bool) bool {
	s.mu.Lock()
	if s.quitActive {
		s.mu.Unlock()
		return false
	}
	s.quitActive = true
	s.quitCheck = check
	s.quitToken++
	s.quitQueue = nil
	for _, key := range []string{"review", "calendar", "focus", "settings", "main"} {
		if s.readyWindows[key] {
			s.quitQueue = append(s.quitQueue, key)
		}
	}
	s.mu.Unlock()
	s.nextQuitWindow()
	return true
}
func (s *TodoService) nextQuitWindow() {
	s.mu.Lock()
	if len(s.quitQueue) == 0 {
		s.quitWaiting = ""
		s.quitActive = false
		check := s.quitCheck
		s.quitCheck = nil
		s.mu.Unlock()
		if check != nil {
			check <- true
		} else {
			s.Quit()
		}
		return
	}
	key := s.quitQueue[0]
	s.quitQueue = s.quitQueue[1:]
	s.quitWaiting = key
	token := s.quitToken
	s.mu.Unlock()
	s.app.Event.Emit("todo:check-quit", map[string]any{"key": key, "token": token})
}
func (s *TodoService) RespondQuit(key string, token int, allow bool) error {
	s.mu.Lock()
	if token != s.quitToken || s.quitWaiting != key {
		s.mu.Unlock()
		return nil
	}
	s.quitWaiting = ""
	if !allow {
		s.quitQueue = nil
		s.quitActive = false
		check := s.quitCheck
		s.quitCheck = nil
		s.mu.Unlock()
		if check != nil {
			check <- false
		}
		return nil
	}
	s.mu.Unlock()
	s.nextQuitWindow()
	return nil
}

// Updating can restart only after every loaded window has protected its drafts.
func (s *TodoService) approveRestart() (bool, error) {
	check := make(chan bool, 1)
	if !s.beginQuitChecks(check) {
		return false, fmt.Errorf("正在确认退出，请完成后重试更新")
	}
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(2 * time.Minute)
	defer timer.Stop()
	select {
	case allow := <-check:
		return allow, nil
	case <-ctx.Done():
		s.cancelRestartCheck(check)
		return false, ctx.Err()
	case <-timer.C:
		s.cancelRestartCheck(check)
		return false, fmt.Errorf("窗口未响应更新确认，请保存内容后重试")
	}
}
func (s *TodoService) cancelRestartCheck(check chan bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.quitCheck == check {
		s.quitActive = false
		s.quitCheck = nil
		s.quitWaiting = ""
		s.quitQueue = nil
		s.quitToken++
	}
}
