package main

import (
	"fmt"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func (s *TodoService) ShowSettings(tab string) error {
	switch tab {
	case "general", "appearance", "keys", "data", "tags", "about":
	default:
		return fmt.Errorf("未知设置分类")
	}
	s.mu.Lock()
	s.settingsTab = tab
	s.mu.Unlock()
	if err := s.ShowWindow("settings"); err != nil {
		return err
	}
	s.app.Event.Emit("todo:settings-tab", tab)
	return nil
}
func (s *TodoService) InitialSettingsTab() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.settingsTab == "" {
		return "general"
	}
	return s.settingsTab
}
func (s *TodoService) WindowVisible(key string) bool {
	w := s.windows[key]
	if w == nil {
		return false
	}
	var visible bool
	application.InvokeSync(func() { visible = w.IsVisible() })
	return visible
}
