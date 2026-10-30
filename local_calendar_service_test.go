package main

import (
	"testing"
	"time"
)

func TestLocalCalendarAndExpansionRecoveryGuards(t *testing.T) {
	s := &TodoService{}
	if _, err := s.TaskExpansion(); err == nil {
		t.Fatal("missing recovery error")
	}
	if err := s.SaveTaskExpansion(nil); err == nil {
		t.Fatal("missing recovery error")
	}
	if _, err := s.LocalCalendarSources(); err == nil {
		t.Fatal("permission requested without database")
	}
	if _, err := s.SyncLocalCalendar("2026-10-01", "2026-10-31"); err == nil {
		t.Fatal("missing recovery error")
	}
	if _, err := s.ConfigureLocalCalendar(nil, "2026-10-01", "2026-10-31"); err == nil {
		t.Fatal("missing recovery error")
	}
}
func TestCalendarRangeUsesLocalCalendarDays(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	a, b, err := calendarInterval("2026-03-08", "2026-03-08", loc)
	if err != nil || b-a != 23*int64(time.Hour/time.Millisecond) {
		t.Fatal(a, b, err)
	}
	for _, dates := range [][2]string{{"bad", "2026-01-01"}, {"2026-02-02", "2026-01-01"}, {"2025-01-01", "2027-01-01"}} {
		if _, _, err = calendarInterval(dates[0], dates[1], loc); err == nil {
			t.Fatal(dates)
		}
	}
}
