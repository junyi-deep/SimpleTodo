package main

import (
	"changeme/internal/localcalendar"
	"changeme/internal/todo"
	"fmt"
	"time"
)

func calendarInterval(from, to string, loc *time.Location) (int64, int64, error) {
	a, err := time.ParseInLocation("2006-01-02", from, loc)
	if err != nil {
		return 0, 0, fmt.Errorf("开始日期无效")
	}
	b, err := time.ParseInLocation("2006-01-02", to, loc)
	if err != nil || b.Before(a) || b.Sub(a) > 370*24*time.Hour {
		return 0, 0, fmt.Errorf("读取日程的范围应在一年内")
	}
	return a.UnixMilli(), b.AddDate(0, 0, 1).UnixMilli(), nil
}
func (s *TodoService) LocalCalendarSources() ([]localcalendar.Calendar, error) {
	if s.core() == nil {
		return nil, fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	s.calendarMu.Lock()
	defer s.calendarMu.Unlock()
	r, err := localcalendar.Read(nil, 0, 0, true)
	return r.Calendars, err
}
func (s *TodoService) LocalCalendarState(from, to string) (todo.LocalCalendarState, error) {
	if s.core() == nil {
		return todo.LocalCalendarState{}, fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	a, b, err := calendarInterval(from, to, mustLocation(s.core().Timezone()))
	if err != nil {
		return todo.LocalCalendarState{}, err
	}
	return s.core().LocalCalendarState(a, b)
}
func (s *TodoService) SyncLocalCalendar(from, to string) (todo.LocalCalendarState, error) {
	if s.core() == nil {
		return todo.LocalCalendarState{}, fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	s.calendarMu.Lock()
	defer s.calendarMu.Unlock()
	a, b, err := calendarInterval(from, to, mustLocation(s.core().Timezone()))
	if err != nil {
		return todo.LocalCalendarState{}, err
	}
	state, err := s.core().LocalCalendarState(a, b)
	if err != nil || len(state.Sources) == 0 {
		return state, err
	}
	return s.readLocalCalendar(state.Sources, a, b)
}
func (s *TodoService) ConfigureLocalCalendar(ids []string, from, to string) (todo.LocalCalendarState, error) {
	if s.core() == nil {
		return todo.LocalCalendarState{}, fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	s.calendarMu.Lock()
	defer s.calendarMu.Unlock()
	a, b, err := calendarInterval(from, to, mustLocation(s.core().Timezone()))
	if err != nil {
		return todo.LocalCalendarState{}, err
	}
	if len(ids) == 0 {
		err = s.core().SaveLocalCalendar(nil, nil)
		if err != nil {
			return todo.LocalCalendarState{}, err
		}
		return s.core().LocalCalendarState(a, b)
	}
	return s.readLocalCalendar(ids, a, b)
}
func (s *TodoService) readLocalCalendar(ids []string, a, b int64) (todo.LocalCalendarState, error) {
	if len(ids) > 50 {
		return todo.LocalCalendarState{}, fmt.Errorf("一次最多选择 50 个日历")
	}
	r, err := localcalendar.Read(ids, a, b, false)
	if err != nil {
		return todo.LocalCalendarState{}, err
	}
	known := map[string]bool{}
	for _, c := range r.Calendars {
		known[c.ID] = true
	}
	for _, id := range ids {
		if !known[id] {
			return todo.LocalCalendarState{}, fmt.Errorf("原日历已不可用，请在设置中重新选择本机日历")
		}
	}
	if err = s.core().SaveLocalCalendar(ids, r.Events); err != nil {
		return todo.LocalCalendarState{}, err
	}
	return s.core().LocalCalendarState(a, b)
}
