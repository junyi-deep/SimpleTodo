package todo

import (
	"database/sql"
	"fmt"
	"time"
)

// UpgradeTaskPeriods runs only at application upgrade/import boundaries. History
// and completed snapshots stay intact; pending unscheduled tasks gain daily spans.
func (e *Engine) UpgradeTaskPeriods() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var modelReady int
	if err = tx.QueryRow("SELECT count(*) FROM app_meta WHERE key='period_model_ready'").Scan(&modelReady); err != nil {
		return err
	}
	legacyOff := map[string]bool{}
	if modelReady == 0 {
		rs, er := tx.Query("SELECT DISTINCT series_id FROM tasks WHERE parent_id IS NULL AND series_id IS NOT NULL AND auto_postpone_policy='off'")
		if er != nil {
			return er
		}
		for rs.Next() {
			var id string
			if er = rs.Scan(&id); er != nil {
				rs.Close()
				return er
			}
			legacyOff[id] = true
		}
		rs.Close()
		if _, err = tx.Exec("UPDATE tasks SET auto_postpone_policy='inherit' WHERE parent_id IS NULL AND series_id IS NOT NULL AND auto_postpone_policy='off'"); err != nil {
			return err
		}
	}
	rows, err := tx.Query("SELECT " + columns + " FROM tasks WHERE parent_id IS NULL AND series_id IS NULL AND status='pending'")
	if err != nil {
		return err
	}
	roots, err := readTasks(rows)
	if err != nil {
		return err
	}
	for _, root := range roots {
		m, er := treeTx(tx, root.ID)
		if er != nil {
			return er
		}
		if root.Date == nil {
			root.Date = ptr(time.UnixMilli(root.CreatedAt).In(e.location()).Format("2006-01-02"))
			m[root.ID] = root
		}
		var until *string
		if root.EndAt != nil {
			d := time.UnixMilli(*root.EndAt).In(e.location()).Format("2006-01-02")
			until = &d
		}
		c := Command{ID: root.ID, RootID: root.ID, ExpectedVersion: fmt.Sprint(root.Version)}
		if er = e.repeatTx(tx, RepeatRequest{Command: c, Rule: Rule{Frequency: "daily", Interval: 1, Until: until}}, m); er != nil {
			return er
		}
	}
	rules, err := allRules(tx)
	if err != nil {
		return err
	}
	for _, r := range rules {
		changed := r.Frequency != "daily" || r.Interval != 1
		r.Frequency = "daily"
		r.Interval = 1
		for i := range r.Template {
			if r.Template[i].StartAt == 0 {
				at, er := wallTime(r.From, "00:00", e.location())
				if er != nil {
					return er
				}
				r.Template[i].StartAt = at.UnixMilli()
				changed = true
			}
			if r.Until != nil && r.Template[i].EndAt == nil {
				end, er := wallTime(*r.Until, "23:59", e.location())
				if er != nil {
					return er
				}
				r.Template[i].EndAt = ptr(end.UnixMilli() + 59999)
				changed = true
			}
			if r.Template[i].ParentID == nil && legacyOff[r.SeriesID] {
				r.Template[i].Postpone = ptr("inherit")
				changed = true
			}
		}
		if changed {
			r.GeneratedThrough = e.localNow().AddDate(0, 0, -1).Format("2006-01-02")
			if _, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND rule_id=? AND planned_date>? AND status='pending' AND is_exception=0", r.ID, today(e.localNow())); err != nil {
				return err
			}
			if err = putRule(tx, r); err != nil {
				return err
			}
		}
	}
	if _, err = tx.Exec("UPDATE tasks SET ends_at=completed_at WHERE status='completed' AND series_id IS NULL AND ends_at IS NULL"); err != nil {
		return err
	}
	if _, err = tx.Exec("INSERT INTO app_meta(key,value) VALUES('period_model_ready','1') ON CONFLICT(key) DO NOTHING"); err != nil {
		return err
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		e.periodModel = true
		e.signal()
	}
	return err
}

// Extend the end bound rather than moving any historical daily occurrence.
func (e *Engine) extendDailySpansTx(tx *sql.Tx, global string) (bool, error) {
	rules, err := allRules(tx)
	if err != nil {
		return false, err
	}
	changed := false
	day := today(e.localNow())
	for _, r := range rules {
		if r.Frequency != "daily" || r.Until == nil || *r.Until >= day {
			continue
		}
		var n int
		err = tx.QueryRow("SELECT count(*) FROM tasks WHERE series_id=? AND parent_id IS NULL AND status='pending' AND (auto_postpone_policy='on' OR (auto_postpone_policy='inherit' AND ?='true'))", r.SeriesID, global).Scan(&n)
		if err != nil {
			return false, err
		}
		if n == 0 {
			continue
		}
		var stopped *string
		if err = tx.QueryRow("SELECT stopped_from FROM repeat_series WHERE id=?", r.SeriesID).Scan(&stopped); err != nil {
			return false, err
		}
		if stopped != nil {
			continue
		}
		end, er := wallTime(day, "23:59", e.location())
		if er != nil {
			return false, er
		}
		endAt := end.UnixMilli() + 59999
		r.Until = ptr(day)
		for i := range r.Template {
			r.Template[i].EndAt = ptr(endAt)
		}
		if err = putRule(tx, r); err != nil {
			return false, err
		}
		if _, err = tx.Exec("UPDATE tasks SET ends_at=?,updated_at=? WHERE (series_id=? OR root_id IN (SELECT id FROM tasks WHERE series_id=?)) AND status='pending' AND (ends_at IS NULL OR ends_at<?)", endAt, e.now().UnixMilli(), r.SeriesID, r.SeriesID, endAt); err != nil {
			return false, err
		}
		changed = true
	}
	return changed, nil
}

func activeOnDay(t Task, day string, loc *time.Location) bool {
	if day == "" {
		return true
	}
	return (t.StartAt == 0 || time.UnixMilli(t.StartAt).In(loc).Format("2006-01-02") <= day) && (t.EndAt == nil || time.UnixMilli(*t.EndAt).In(loc).Format("2006-01-02") >= day)
}
func (e *Engine) aggregate(m map[string]Task, now int64) {
	if len(m) == 0 {
		return
	}
	if !e.periodModel {
		aggregate(m, now)
		return
	}
	visible := map[string]Task{}
	var root Task
	for _, t := range m {
		if t.ParentID == nil {
			root = t
		}
	}
	var add func(string)
	add = func(id string) {
		for _, t := range m {
			if val(t.ParentID) == id && activeOnDay(t, val(root.Date), e.location()) {
				visible[t.ID] = t
				add(t.ID)
			}
		}
	}
	if root.Exception && val(root.Date) != val(root.OccurrenceDate) {
		aggregate(m, now)
		return
	}
	visible[root.ID] = root
	add(root.ID)
	aggregate(visible, now)
	for id, t := range visible {
		m[id] = t
	}
}
