package todo

import (
	"changeme/internal/localcalendar"
	"encoding/json"
	"fmt"
)

type LocalCalendarState struct {
	Sources  []string              `json:"sources"`
	LastSync int64                 `json:"lastSync"`
	Events   []localcalendar.Event `json:"events"`
}

func (e *Engine) LocalCalendarState(from, to int64) (LocalCalendarState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := LocalCalendarState{Sources: []string{}, Events: []localcalendar.Event{}}
	var raw string
	if err := e.db.QueryRow("SELECT sources_json,last_sync FROM local_calendar_config WHERE id=1").Scan(&raw, &out.LastSync); err != nil {
		return out, err
	}
	if err := json.Unmarshal([]byte(raw), &out.Sources); err != nil {
		return out, err
	}
	rows, err := e.db.Query("SELECT event_json FROM local_calendar_events WHERE start_at<? AND end_at>? ORDER BY start_at,id", to, from)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		var event localcalendar.Event
		if err = rows.Scan(&raw); err != nil {
			return out, err
		}
		if err = json.Unmarshal([]byte(raw), &event); err != nil {
			return out, err
		}
		out.Events = append(out.Events, event)
	}
	return out, rows.Err()
}
func (e *Engine) SaveLocalCalendar(ids []string, events []localcalendar.Event) error {
	if len(ids) > 50 || len(events) > 10000 {
		return fail("VALIDATION_ERROR", "本机日历数据过多")
	}
	selected := map[string]bool{}
	for _, id := range ids {
		if len(id) == 0 || len(id) > 4096 {
			return fail("VALIDATION_ERROR", "日历标识无效")
		}
		selected[id] = true
	}
	seen := map[string]bool{}
	events = append([]localcalendar.Event(nil), events...)
	for i, ev := range events {
		if !selected[ev.CalendarID] || ev.ID == "" || seen[ev.ID] || len(ev.Title) > 5000 || len(ev.Location) > 5000 || ev.End < ev.Start {
			return fail("VALIDATION_ERROR", "日历日程数据无效")
		}
		// Zero-duration appointments still belong to their start date.
		if ev.End == ev.Start {
			events[i].End = ev.Start + 1
		}
		seen[ev.ID] = true
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec("DELETE FROM local_calendar_events"); err != nil {
		return err
	}
	for _, ev := range events {
		b, er := json.Marshal(ev)
		if er != nil {
			return er
		}
		if _, err = tx.Exec("INSERT INTO local_calendar_events VALUES(?,?,?,?,?)", ev.ID, ev.CalendarID, ev.Start, ev.End, string(b)); err != nil {
			return fmt.Errorf("保存本机日程失败：%w", err)
		}
	}
	b, _ := json.Marshal(ids)
	if ids == nil {
		b = []byte("[]")
	}
	if _, err = tx.Exec("UPDATE local_calendar_config SET sources_json=?,last_sync=? WHERE id=1", string(b), e.now().UnixMilli()); err != nil {
		return err
	}
	return tx.Commit()
}
