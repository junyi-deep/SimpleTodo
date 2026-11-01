package main

import (
	"changeme/internal/releases"
	"context"
	"time"
)

// A single process-wide scheduler prevents one request per independent window.
func (s *TodoService) runUpdateChecks(ctx context.Context) {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.checkAutomaticUpdate()
			timer.Reset(time.Hour)
		}
	}
}
func (s *TodoService) AutomaticUpdate() *releases.Info {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()
	if s.updateInfo == nil {
		return nil
	}
	info := *s.updateInfo
	return &info
}
func (s *TodoService) SkipUpdate(version string) error {
	if err := s.ready(); err != nil {
		return err
	}
	if err := s.core().SaveSettings(map[string]string{"skippedUpdateVersion": version}); err != nil {
		return err
	}
	s.updateMu.Lock()
	s.updateInfo = nil
	s.updateMu.Unlock()
	s.app.Event.Emit("todo:update-skipped", version)
	return nil
}

func (s *TodoService) checkAutomaticUpdate() {
	if s.core() == nil {
		return
	}
	s.updateMu.Lock()
	if s.updateChecking || time.Since(s.lastAutoCheck) < time.Minute {
		s.updateMu.Unlock()
		return
	}
	s.updateChecking = true
	s.lastAutoCheck = time.Now()
	s.updateMu.Unlock()
	defer func() { s.updateMu.Lock(); s.updateChecking = false; s.updateMu.Unlock() }()
	info, err := s.CheckUpdate()
	if err != nil {
		return
	}
	settings, err := s.core().Settings()
	if err != nil {
		return
	}
	s.updateMu.Lock()
	prompt := info.Available && settings["skippedUpdateVersion"] != info.Version
	if prompt {
		s.updateInfo = &info
	} else {
		s.updateInfo = nil
	}
	s.updateMu.Unlock()
	if prompt {
		s.app.Event.Emit("todo:update-available", info)
		_ = s.ShowWindow("main")
	}
}
