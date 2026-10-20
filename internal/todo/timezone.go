package todo

import (
	"database/sql"
	"time"
	_ "time/tzdata"
)

func (e *Engine) location() *time.Location {
	if l := e.zone.Load(); l != nil {
		return l
	}
	return time.Local
}
func (e *Engine) localNow() time.Time { return e.now().In(e.location()) }
func (e *Engine) Timezone() string    { return e.location().String() }

// Resolve both folds by choosing the first occurrence. For a spring gap, advance
// the wall clock to the first representable minute rather than Go's implicit fold.
func wallTime(date, clock string, loc *time.Location) (time.Time, error) {
	wall, err := time.Parse("2006-01-02 15:04", date+" "+clock)
	if err != nil {
		return time.Time{}, fail("VALIDATION_ERROR", "提醒时分无效")
	}
	offsets := map[int]bool{}
	for _, hours := range []int{-24, 0, 24} {
		_, offset := time.Date(wall.Year(), wall.Month(), wall.Day(), 12, 0, 0, 0, loc).Add(time.Duration(hours) * time.Hour).Zone()
		offsets[offset] = true
	}
	for minute := 0; minute <= 180; minute++ {
		target := wall.Add(time.Duration(minute) * time.Minute)
		var earliest time.Time
		for offset := range offsets {
			candidate := time.Unix(target.Unix()-int64(offset), 0).In(loc)
			if candidate.Format("2006-01-02 15:04") == target.Format("2006-01-02 15:04") && (earliest.IsZero() || candidate.Before(earliest)) {
				earliest = candidate
			}
		}
		if !earliest.IsZero() {
			return earliest, nil
		}
	}
	return time.Time{}, fail("VALIDATION_ERROR", "该时区没有可用的提醒时刻")
}
func (e *Engine) SetTimezone(name string) (int, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return 0, fail("VALIDATION_ERROR", "未知系统时区")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Timezone() == name {
		return 0, nil
	}
	tx, err := e.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.Query("SELECT "+columns+" FROM tasks WHERE parent_id IS NULL AND series_id IS NOT NULL AND status='pending' AND remind_at>?", e.now().UnixMilli())
	if err != nil {
		return 0, err
	}
	tasks, err := readTasks(rows)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, t := range tasks {
		var sent int
		if err = tx.QueryRow("SELECT count(*) FROM notification_jobs WHERE source_type='task' AND source_id=? AND generation=? AND state IN ('submitted','claimed','uncertain')", t.ID, t.ReminderGeneration).Scan(&sent); err != nil {
			return 0, err
		}
		if sent > 0 || t.ReminderTime == nil {
			continue
		}
		at, er := wallTime(val(t.Date), *t.ReminderTime, loc)
		if er != nil {
			return 0, er
		}
		t.RemindAt = ptr(at.UnixMilli())
		t.ReminderZone = ptr(name)
		t.ReminderGeneration++
		t.Version++
		t.UpdatedAt = e.now().UnixMilli()
		if err = putTask(tx, t); err != nil {
			return 0, err
		}
		if err = syncReminder(tx, t, e.now().UnixMilli()); err != nil {
			return 0, err
		}
		changed++
	}
	if _, err = tx.Exec("INSERT INTO app_meta(key,value) VALUES('timezone',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", name); err != nil {
		return 0, err
	}
	if err = revisionTx(tx); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	e.zone.Store(loc)
	e.lastDay = ""
	e.signal()
	return changed, nil
}
func (e *Engine) loadTimezone(db *sql.DB) {
	var name string
	if db.QueryRow("SELECT value FROM app_meta WHERE key='timezone'").Scan(&name) == nil {
		if l, err := time.LoadLocation(name); err == nil {
			e.zone.Store(l)
		}
	}
}
