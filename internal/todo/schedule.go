package todo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func syncReminder(tx *sql.Tx, t Task, now int64) error {
	if _, err := tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending' AND (generation<>? OR ?<>'pending' OR ? IS NULL)", t.ID, t.ReminderGeneration, t.Status, t.RemindAt); err != nil {
		return err
	}
	if t.RemindAt == nil || t.Status != "pending" {
		return nil
	}
	state := "pending"
	if *t.RemindAt < now-24*int64(time.Hour/time.Millisecond) {
		state = "expired"
	}
	_, err := tx.Exec(`INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES(?,'task',?,?,?,?,?,?) ON CONFLICT(source_type,source_id,generation,cycle_key) DO UPDATE SET title=excluded.title,state=CASE WHEN notification_jobs.state='cancelled' AND excluded.fire_at>? THEN excluded.state ELSE notification_jobs.state END`, ID(), t.ID, t.ReminderGeneration, "once", *t.RemindAt, state, t.Title, now)
	return err
}
func ruleScan(s scanner) (r Rule, err error) {
	var weekdays, template string
	err = s.Scan(&r.ID, &r.SeriesID, &r.From, &r.Until, &r.Anchor, &r.Frequency, &r.Interval, &weekdays, &r.MonthDay, &template, &r.GeneratedThrough)
	if err == nil {
		err = json.Unmarshal([]byte(weekdays), &r.Weekdays)
	}
	if err == nil {
		err = json.Unmarshal([]byte(template), &r.Template)
	}
	return
}

const ruleColumns = "id,series_id,effective_from,effective_to,anchor_date,frequency,interval_n,weekdays,month_day,template_tree_json,generated_through"

func allRules(db interface {
	Query(string, ...any) (*sql.Rows, error)
}) ([]Rule, error) {
	rows, err := db.Query("SELECT " + ruleColumns + " FROM repeat_rules ORDER BY effective_from")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Rule{}
	for rows.Next() {
		r, er := ruleScan(rows)
		if er != nil {
			return nil, er
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func putRule(tx *sql.Tx, r Rule) error {
	w, _ := json.Marshal(r.Weekdays)
	t, _ := json.Marshal(r.Template)
	_, err := tx.Exec("INSERT INTO repeat_rules("+ruleColumns+") VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET effective_to=excluded.effective_to,frequency=excluded.frequency,interval_n=excluded.interval_n,weekdays=excluded.weekdays,month_day=excluded.month_day,template_tree_json=excluded.template_tree_json,generated_through=excluded.generated_through", r.ID, r.SeriesID, r.From, r.Until, r.Anchor, r.Frequency, r.Interval, string(w), r.MonthDay, string(t), r.GeneratedThrough)
	return err
}
func (e *Engine) Rules() ([]Rule, error) { e.mu.Lock(); defer e.mu.Unlock(); return allRules(e.db) }
func ruleMatches(r Rule, d time.Time) bool {
	day := d.Format("2006-01-02")
	if day < r.From || (r.Until != nil && day > *r.Until) {
		return false
	}
	anchor, err := time.Parse("2006-01-02", r.Anchor)
	if err != nil || d.Before(anchor) {
		return false
	}
	n := r.Interval
	if n < 1 {
		n = 1
	}
	switch r.Frequency {
	case "daily":
		return int(d.Sub(anchor).Hours()/24)%n == 0
	case "weekly":
		startWeek := anchor.AddDate(0, 0, -(int(anchor.Weekday())+6)%7)
		dateWeek := d.AddDate(0, 0, -(int(d.Weekday())+6)%7)
		if int(dateWeek.Sub(startWeek).Hours()/24/7)%n != 0 {
			return false
		}
		for _, w := range r.Weekdays {
			if w == int(d.Weekday()) {
				return true
			}
		}
		return false
	case "monthly":
		months := (d.Year()-anchor.Year())*12 + int(d.Month()-anchor.Month())
		last := time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		want := r.MonthDay
		if want > last {
			want = last
		}
		return months%n == 0 && d.Day() == want
	}
	return false
}
func cloneTree(template []Task, date, series, rule string, now int64, keepStatus bool, loc *time.Location) map[string]Task {
	ids := map[string]string{}
	for _, t := range template {
		ids[t.ID] = ID()
	}
	var rootID string
	for _, t := range template {
		if t.ParentID == nil {
			rootID = ids[t.ID]
		}
	}
	out := map[string]Task{}
	for _, old := range template {
		t := old
		t.ID = ids[old.ID]
		t.RootID = rootID
		t.Version = 1
		t.CreatedAt = now
		t.UpdatedAt = now
		t.Exception = false
		t.ReminderGeneration = 0
		if !keepStatus {
			t.Status = "pending"
			t.CompletedAt = nil
		}
		if old.ParentID != nil {
			t.ParentID = ptr(ids[*old.ParentID])
		} else {
			t.ParentID = nil
			if date != "" {
				t.Date = ptr(date)
			} else {
				t.Date = nil
			}
			t.OriginalDate = nil
			if series != "" {
				t.SeriesID = ptr(series)
				t.RuleID = ptr(rule)
				t.OccurrenceDate = ptr(date)
				t.Postpone = ptr("off")
			} else {
				t.SeriesID = nil
				t.RuleID = nil
				t.OccurrenceDate = nil
				t.RemindAt = nil
				t.ReminderTime = nil
			}
			if series != "" && t.ReminderTime != nil {
				at, err := wallTime(date, *t.ReminderTime, loc)
				if err == nil {
					t.RemindAt = ptr(at.UnixMilli())
					t.ReminderZone = ptr(loc.String())
				}
			}
		}
		out[t.ID] = t
	}
	return out
}
func (e *Engine) Repeat(req RepeatRequest) error {
	c := req.Command
	c.RepeatRule = &req.Rule
	c.Following = req.Following
	_, err := e.mutate(c, "repeat", func(map[string]Task, int64) error { return nil })
	return err
}

// Called inside the same transaction that saves the edited tree.
func (e *Engine) repeatTx(tx *sql.Tx, req RepeatRequest, m map[string]Task) error {
	c := req.Command
	r := req.Rule
	if !contains([]string{"daily", "weekly", "monthly"}, r.Frequency) || r.Interval < 1 || r.Interval > 99 {
		return fail("VALIDATION_ERROR", "请选择正确的重复周期和间隔")
	}
	if r.Frequency == "weekly" && len(r.Weekdays) == 0 {
		return fail("VALIDATION_ERROR", "至少选择一个星期")
	}
	for _, w := range r.Weekdays {
		if w < 0 || w > 6 {
			return fail("VALIDATION_ERROR", "星期不正确")
		}
	}
	root, ok := m[c.RootID]
	if !ok || root.ParentID != nil || root.Date == nil {
		return fail("VALIDATION_ERROR", "请先安排根任务日期")
	}
	if root.SeriesID != nil && !req.Following {
		return fail("VALIDATION_ERROR", "修改系列须显式选择本次及以后")
	}
	var err error
	r.From = *root.Date
	r.Anchor = r.From
	if err = validateDate(r.Until); err != nil {
		return err
	}
	if r.Until != nil && *r.Until < r.From {
		return fail("VALIDATION_ERROR", "结束日不能早于开始日")
	}
	if root.SeriesID != nil && r.From < today(e.localNow()) {
		return fail("VALIDATION_ERROR", "历史实例只能编辑本次")
	}
	r.MonthDay, _ = func() (int, error) { t, er := time.Parse("2006-01-02", r.From); return t.Day(), er }()
	r.ID = ID()
	r.SeriesID = val(root.SeriesID)
	if r.SeriesID == "" {
		r.SeriesID = ID()
		if _, err = tx.Exec("INSERT INTO repeat_series(id) VALUES(?)", r.SeriesID); err != nil {
			return err
		}
	} else {
		before, _ := time.Parse("2006-01-02", r.From)
		cut := before.AddDate(0, 0, -1).Format("2006-01-02")
		if _, err = tx.Exec("UPDATE repeat_rules SET effective_to=? WHERE series_id=? AND effective_from<? AND (effective_to IS NULL OR effective_to>=?)", cut, r.SeriesID, r.From, r.From); err != nil {
			return err
		}
		// Re-editing the same boundary replaces its future template; occurrence snapshots stay authoritative.
		var boundary string
		er := tx.QueryRow("SELECT id FROM repeat_rules WHERE series_id=? AND effective_from=?", r.SeriesID, r.From).Scan(&boundary)
		if er == nil {
			r.ID = boundary
		} else if er != sql.ErrNoRows {
			return er
		}
		if _, err = tx.Exec("UPDATE repeat_rules SET effective_to=? WHERE series_id=? AND effective_from>?", cut, r.SeriesID, r.From); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND state='pending' AND source_id IN (SELECT id FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND status='pending' AND is_exception=0 AND id<>?)", r.SeriesID, r.From, root.ID); err != nil {
			return err
		}
		if _, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND status='pending' AND is_exception=0 AND id<>?", r.SeriesID, r.From, root.ID); err != nil {
			return err
		}
	}
	r.Template = []Task{}
	for _, t := range m {
		r.Template = append(r.Template, t)
	}
	d, _ := time.Parse("2006-01-02", r.From)
	r.GeneratedThrough = d.AddDate(0, 0, -1).Format("2006-01-02")
	if err = putRule(tx, r); err != nil {
		return err
	}
	root.SeriesID = ptr(r.SeriesID)
	root.RuleID = ptr(r.ID)
	root.OccurrenceDate = ptr(r.From)
	root.Postpone = ptr("off")
	root.Exception = false
	m[root.ID] = root
	root.UpdatedAt = e.now().UnixMilli()
	if err = putTask(tx, root); err != nil {
		return err
	}
	return nil
}
func (e *Engine) StopFollowing(c Command) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	root, err := scanTask(tx.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", c.RootID))
	if err != nil {
		return err
	}
	if fmt.Sprint(root.Version) != c.ExpectedVersion {
		return fail("CONFLICT", "任务已更新")
	}
	if root.SeriesID == nil {
		return fail("VALIDATION_ERROR", "不是重复任务")
	}
	cut := val(root.OccurrenceDate)
	if cut < today(e.localNow()) {
		return fail("VALIDATION_ERROR", "历史实例只允许删除本次")
	}
	if _, err = tx.Exec("UPDATE repeat_series SET stopped_from=?,version=version+1 WHERE id=?", cut, *root.SeriesID); err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND state='pending' AND source_id IN(SELECT id FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND status='pending')", *root.SeriesID, cut); err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND status='pending'", *root.SeriesID, cut); err != nil {
		return err
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		e.signal()
	}
	return err
}
func (e *Engine) EnsureThrough(end string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := validateDate(&end); err != nil {
		return err
	}
	rules, err := allRules(e.db)
	if err != nil {
		return err
	}
	changed := false
	for _, r := range rules {
		target := end
		if r.Until != nil && target > *r.Until {
			target = *r.Until
		}
		var stop *string
		if err = e.db.QueryRow("SELECT stopped_from FROM repeat_series WHERE id=?", r.SeriesID).Scan(&stop); err != nil {
			return err
		}
		if stop != nil && target >= *stop {
			d, _ := time.Parse("2006-01-02", *stop)
			target = d.AddDate(0, 0, -1).Format("2006-01-02")
		}
		start := r.GeneratedThrough
		if start >= target {
			continue
		}
		d, er := time.Parse("2006-01-02", start)
		if er != nil {
			return er
		}
		for d = d.AddDate(0, 0, 1); d.Format("2006-01-02") <= target; {
			tx, er := e.db.Begin()
			if er != nil {
				return er
			}
			nodes := 0
			last := ""
			for d.Format("2006-01-02") <= target && nodes < 1800 {
				date := d.Format("2006-01-02")
				last = date
				if ruleMatches(r, d) {
					var count int
					er = tx.QueryRow("SELECT count(*) FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date=?", r.SeriesID, date).Scan(&count)
					if er == nil && count == 0 {
						m := cloneTree(r.Template, date, r.SeriesID, r.ID, e.now().UnixMilli(), false, e.location())
						er = validateTree(m)
						if er == nil {
							for _, t := range m {
								if er = putTask(tx, t); er != nil {
									break
								}
							}
						}
						if er == nil {
							for _, t := range m {
								if t.ParentID == nil {
									er = syncReminder(tx, t, e.now().UnixMilli())
								}
							}
						}
						nodes += len(m)
						changed = true
					}
					if er != nil {
						break
					}
				}
				d = d.AddDate(0, 0, 1)
			}
			if er == nil {
				_, er = tx.Exec("UPDATE repeat_rules SET generated_through=? WHERE id=?", last, r.ID)
			}
			if er == nil {
				er = revisionTx(tx)
			}
			if er != nil {
				tx.Rollback()
				return er
			}
			if er = tx.Commit(); er != nil {
				return er
			}
		}
	}
	if changed {
		e.signal()
	}
	return nil
}
func (e *Engine) postpone() error {
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var global string
	_ = tx.QueryRow("SELECT value_json FROM settings WHERE key='autoPostpone'").Scan(&global)
	rows, err := tx.Query("SELECT "+columns+" FROM tasks WHERE parent_id IS NULL AND status='pending' AND planned_date<? AND series_id IS NULL AND (auto_postpone_policy='on' OR (auto_postpone_policy='inherit' AND ?='true')) ORDER BY planned_date,sort_order,id", today(e.localNow()), global)
	if err != nil {
		return err
	}
	ts, err := readTasks(rows)
	if err != nil {
		return err
	}
	var lastOrder int
	if err = tx.QueryRow("SELECT COALESCE(MAX(sort_order),-1) FROM tasks WHERE parent_id IS NULL AND planned_date=?", today(e.localNow())).Scan(&lastOrder); err != nil {
		return err
	}
	for i, t := range ts {
		if t.OriginalDate == nil {
			t.OriginalDate = t.Date
		}
		t.Date = ptr(today(e.localNow()))
		t.Order = lastOrder + 1 + i
		t.Version++
		t.UpdatedAt = e.now().UnixMilli()
		if err = putTask(tx, t); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending'", t.ID); err != nil {
			return err
		}
	}
	if len(ts) > 0 {
		if err = revisionTx(tx); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lastError := ""
	for {
		if err := e.Tick(); err != nil {
			if e.Report != nil && err.Error() != lastError {
				e.Report(err)
			}
			lastError = err.Error()
		} else {
			lastError = ""
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (e *Engine) Tick() error {
	now := e.localNow()
	day := today(now)
	e.mu.Lock()
	newDay := e.lastDay != day
	e.mu.Unlock()
	if newDay {
		if err := e.EnsureThrough(now.AddDate(0, 0, 7).Format("2006-01-02")); err != nil {
			return err
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if newDay {
		if err := e.postpone(); err != nil {
			return err
		}
		e.lastDay = day
		e.signal()
	}
	ms := now.UnixMilli()
	if _, err := e.db.Exec("UPDATE focus_sessions SET state='completed',finished_at=? WHERE state='running' AND deadline_at<=?", ms, ms); err != nil {
		return err
	}
	rows, err := e.db.Query("SELECT id,source_type,source_id,state,fire_at,title FROM notification_jobs WHERE state='pending' AND fire_at<=? ORDER BY fire_at", ms)
	if err != nil {
		return err
	}
	jobs := []Notice{}
	for rows.Next() {
		var n Notice
		if err = rows.Scan(&n.ID, &n.SourceType, &n.SourceID, &n.State, &n.FireAt, &n.Title); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, n)
	}
	rows.Close()
	missed := []Notice{}
	for _, n := range jobs {
		valid := 0
		switch n.SourceType {
		case "task":
			err = e.db.QueryRow("SELECT count(*) FROM tasks t JOIN notification_jobs j ON j.source_id=t.id WHERE j.id=? AND t.parent_id IS NULL AND t.status='pending' AND j.generation=t.reminder_generation AND t.remind_at=j.fire_at", n.ID).Scan(&valid)
		case "focus":
			err = e.db.QueryRow("SELECT count(*) FROM focus_sessions f JOIN notification_jobs j ON j.source_id=f.id WHERE j.id=? AND f.state IN ('running','completed') AND f.deadline_at=j.fire_at", n.ID).Scan(&valid)
		case "countdown":
			err = e.db.QueryRow("SELECT count(*) FROM countdowns c JOIN notification_jobs j ON j.source_id=c.id WHERE j.id=? AND c.enabled=1 AND c.generation=j.generation", n.ID).Scan(&valid)
		}
		if err != nil {
			return err
		}
		if valid == 0 {
			if _, err = e.db.Exec("UPDATE notification_jobs SET state='cancelled' WHERE id=? AND state='pending'", n.ID); err != nil {
				return err
			}
			continue
		}

		if n.FireAt < ms-24*int64(time.Hour/time.Millisecond) {
			_, err = e.db.Exec("UPDATE notification_jobs SET state='expired' WHERE id=? AND state='pending'", n.ID)
			if err != nil {
				return err
			}
			continue
		}
		if ms-n.FireAt > 60000 {
			missed = append(missed, n)
			continue
		}
		res, er := e.db.Exec("UPDATE notification_jobs SET state='claimed',claimed_at=? WHERE id=? AND state='pending'", ms, n.ID)
		if er != nil {
			return er
		}
		cnt, _ := res.RowsAffected()
		if cnt == 0 {
			continue
		}
		var sound string
		_ = e.db.QueryRow("SELECT value_json FROM settings WHERE key='sound'").Scan(&sound)
		n.Silent = sound == "false"
		state := "submitted"
		if e.Notify == nil || e.Notify(n) != nil {
			state = "failed"
		}
		if _, err = e.db.Exec("UPDATE notification_jobs SET state=?,submitted_at=? WHERE id=?", state, ms, n.ID); err != nil {
			return err
		}
		e.signal()
	}
	if len(missed) > 0 {
		for _, n := range missed {
			if _, err = e.db.Exec("UPDATE notification_jobs SET state='missed' WHERE id=? AND state='pending'", n.ID); err != nil {
				return err
			}
		}
		if e.Notify != nil {
			_ = e.Notify(Notice{ID: "missed-" + fmt.Sprint(ms), Title: fmt.Sprintf("错过 %d 条提醒，请在 SimpleTodo 中查看", len(missed))})
		}
		e.signal()
	}
	if err = e.advanceCountdowns(now); err != nil {
		return err
	}
	_, err = e.db.Exec("DELETE FROM command_receipts WHERE created_at<?", ms-7*24*int64(time.Hour/time.Millisecond))
	return err
}
func (e *Engine) Notices() ([]Notice, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT id,source_type,source_id,state,fire_at,title FROM notification_jobs WHERE state IN ('missed','expired','failed','uncertain') ORDER BY fire_at DESC LIMIT 200")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Notice{}
	for rows.Next() {
		var n Notice
		if err = rows.Scan(&n.ID, &n.SourceType, &n.SourceID, &n.State, &n.FireAt, &n.Title); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
func (e *Engine) DismissNotices() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Exec("UPDATE notification_jobs SET state='dismissed' WHERE state IN ('missed','expired','failed','uncertain')")
	return err
}
func (e *Engine) ActiveFocus() (*Focus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return activeFocus(e.db)
}
func activeFocus(db interface{ QueryRow(string, ...any) *sql.Row }) (*Focus, error) {
	var f Focus
	err := db.QueryRow("SELECT id,task_id,task_title_snapshot,duration_seconds,remaining_seconds,deadline_at,state,started_at,finished_at FROM focus_sessions ORDER BY CASE WHEN state IN ('running','paused') THEN 0 ELSE 1 END,started_at DESC,rowid DESC LIMIT 1").Scan(&f.ID, &f.TaskID, &f.Title, &f.Duration, &f.Remaining, &f.Deadline, &f.State, &f.StartedAt, &f.FinishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &f, err
}
func (e *Engine) FocusAction(action string, minutes int, taskID *string) (*Focus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	f, err := activeFocus(tx)
	if err != nil {
		return nil, err
	}
	now := e.now().UnixMilli()
	switch action {
	case "start":
		if minutes < 1 || minutes > 180 {
			return nil, fail("VALIDATION_ERROR", "专注时长为 1—180 分钟")
		}
		if f != nil && (f.State == "running" || f.State == "paused") {
			return nil, fail("CONFLICT", "已有专注，请先继续或停止")
		}
		title := "自由专注"
		if taskID != nil {
			if err = tx.QueryRow("SELECT title FROM tasks WHERE id=? AND parent_id IS NULL AND status<>'skipped'", *taskID).Scan(&title); err != nil {
				return nil, fail("NOT_FOUND", "关联任务不存在")
			}
		}
		f = &Focus{ID: ID(), TaskID: taskID, Title: title, Duration: minutes * 60, Remaining: minutes * 60, Deadline: now + int64(minutes)*60000, State: "running", StartedAt: now}
	case "pause":
		if f == nil || f.State != "running" {
			return nil, fail("CONFLICT", "没有运行中的专注")
		}
		if f.Deadline <= now {
			f.State = "completed"
			f.Remaining = 0
			f.FinishedAt = ptr(now)
			break
		}
		f.Remaining = int((f.Deadline - now + 999) / 1000)
		if f.Remaining < 0 {
			f.Remaining = 0
		}
		f.State = "paused"
	case "resume":
		if f == nil || f.State != "paused" {
			return nil, fail("CONFLICT", "没有暂停的专注")
		}
		f.Deadline = now + int64(f.Remaining)*1000
		f.State = "running"
	case "stop":
		if f == nil {
			return nil, fail("NOT_FOUND", "没有专注")
		}
		f.State = "stopped"
		f.FinishedAt = ptr(now)
	default:
		return nil, fail("VALIDATION_ERROR", "未知操作")
	}
	_, err = tx.Exec("INSERT INTO focus_sessions VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET remaining_seconds=excluded.remaining_seconds,deadline_at=excluded.deadline_at,state=excluded.state,finished_at=excluded.finished_at", f.ID, f.TaskID, f.Title, f.Duration, f.Remaining, f.Deadline, f.State, f.StartedAt, f.FinishedAt)
	if err != nil {
		return nil, err
	}
	if f.State != "completed" {
		if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='focus' AND source_id=? AND state='pending'", f.ID); err != nil {
			return nil, err
		}
	}
	if f.State == "running" {
		_, err = tx.Exec("INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES(?,'focus',?,?,?,?,'pending',?) ON CONFLICT(source_type,source_id,generation,cycle_key) DO UPDATE SET state='pending',fire_at=excluded.fire_at", ID(), f.ID, 0, fmt.Sprint(f.Deadline), f.Deadline, "专注完成 · "+f.Title)
		if err != nil {
			return nil, err
		}
	}
	if err = revisionTx(tx); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err == nil {
		e.signal()
	}
	return f, err
}
func countdowns(db interface {
	Query(string, ...any) (*sql.Rows, error)
}) ([]Countdown, error) {
	rows, err := db.Query("SELECT * FROM countdowns ORDER BY target_at,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Countdown{}
	for rows.Next() {
		var c Countdown
		if err = rows.Scan(&c.ID, &c.Name, &c.AllDay, &c.AnchorDate, &c.AnchorTime, &c.Target, &c.Zone, &c.Frequency, &c.Enabled, &c.Offset, &c.Generation); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (e *Engine) Countdowns() ([]Countdown, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return countdowns(e.db)
}
func saveCountdown(tx *sql.Tx, c Countdown, now int64) error {
	_, err := tx.Exec("INSERT INTO countdowns VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,all_day=excluded.all_day,anchor_local_date=excluded.anchor_local_date,anchor_local_time=excluded.anchor_local_time,target_at=excluded.target_at,timezone=excluded.timezone,repeat_frequency=excluded.repeat_frequency,enabled=excluded.enabled,reminder_offset_seconds=excluded.reminder_offset_seconds,generation=excluded.generation", c.ID, c.Name, c.AllDay, c.AnchorDate, c.AnchorTime, c.Target, c.Zone, c.Frequency, c.Enabled, c.Offset, c.Generation)
	if err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='countdown' AND source_id=? AND state='pending'", c.ID); err != nil {
		return err
	}
	if !c.Enabled {
		return nil
	}
	_, err = tx.Exec("INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES(?,'countdown',?,?,?,?,'pending',?) ON CONFLICT(source_type,source_id,generation,cycle_key) DO UPDATE SET state=CASE WHEN notification_jobs.state='cancelled' AND excluded.fire_at>? THEN 'pending' ELSE notification_jobs.state END", ID(), c.ID, c.Generation, fmt.Sprint(c.Target), c.Target-int64(c.Offset)*1000, "事项提醒 · "+c.Name, now)
	return err
}
func (e *Engine) SaveCountdown(c Countdown) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" || len([]rune(c.Name)) > 500 {
		return fail("VALIDATION_ERROR", "请填写事项名称")
	}
	if !contains([]string{"none", "daily", "weekly", "monthly"}, c.Frequency) || c.Offset < 0 || c.Offset > 86400 {
		return fail("VALIDATION_ERROR", "倒计时配置不正确")
	}
	if c.Target <= e.now().UnixMilli() {
		return fail("VALIDATION_ERROR", "目标时间必须在未来")
	}
	if c.ID == "" {
		c.ID = ID()
	}
	var gen int
	_ = e.db.QueryRow("SELECT generation FROM countdowns WHERE id=?", c.ID).Scan(&gen)
	c.Generation = gen + 1
	t := time.UnixMilli(c.Target).In(e.location())
	c.AnchorDate = t.Format("2006-01-02")
	c.AnchorTime = t.Format("15:04")
	c.Zone = e.Timezone()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = saveCountdown(tx, c, e.now().UnixMilli()); err != nil {
		return err
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		e.signal()
	}
	return err
}
func (e *Engine) DeleteCountdown(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='countdown' AND source_id=? AND state='pending'", id); err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM countdowns WHERE id=?", id); err != nil {
		return err
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		e.signal()
	}
	return err
}
func (e *Engine) advanceCountdowns(now time.Time) error {
	cs, err := countdowns(e.db)
	if err != nil {
		return err
	}
	for _, c := range cs {
		if !c.Enabled || c.Target > now.UnixMilli() || c.Frequency == "none" {
			continue
		}
		anchor, er := wallTime(c.AnchorDate, c.AnchorTime, e.location())
		if er != nil {
			return er
		}
		next := anchor
		for n := 1; next.UnixMilli() <= now.UnixMilli(); n++ {
			switch c.Frequency {
			case "daily":
				next = anchor.AddDate(0, 0, n)
			case "weekly":
				next = anchor.AddDate(0, 0, n*7)
			case "monthly":
				base := time.Date(anchor.Year(), anchor.Month()+time.Month(n), 1, anchor.Hour(), anchor.Minute(), 0, 0, e.location())
				last := base.AddDate(0, 1, -1).Day()
				day := anchor.Day()
				if day > last {
					day = last
				}
				next = time.Date(base.Year(), base.Month(), day, base.Hour(), base.Minute(), 0, 0, e.location())
			}
		}
		next, er = wallTime(next.Format("2006-01-02"), c.AnchorTime, e.location())
		if er != nil {
			return er
		}
		c.Zone = e.Timezone()
		c.Target = next.UnixMilli()
		c.Generation++
		tx, er := e.db.Begin()
		if er != nil {
			return er
		}
		if er = saveCountdown(tx, c, e.now().UnixMilli()); er == nil {
			er = revisionTx(tx)
		}
		if er != nil {
			tx.Rollback()
			return er
		}
		if er = tx.Commit(); er != nil {
			return er
		}
		e.signal()
	}
	return nil
}
func sortedTree(m map[string]Task) []Task {
	out := []Task{}
	var visit func(*string)
	visit = func(parent *string) {
		ss := []Task{}
		for _, t := range m {
			if (parent == nil && t.ParentID == nil) || (parent != nil && t.ParentID != nil && *parent == *t.ParentID) {
				ss = append(ss, t)
			}
		}
		sort.Slice(ss, func(i, j int) bool {
			if ss[i].Order == ss[j].Order {
				return ss[i].ID < ss[j].ID
			}
			return ss[i].Order < ss[j].Order
		})
		for _, t := range ss {
			out = append(out, t)
			visit(ptr(t.ID))
		}
	}
	visit(nil)
	return out
}

type SeriesImpact struct {
	Updated   int `json:"updated"`
	Preserved int `json:"preserved"`
	Removed   int `json:"removed"`
	Nodes     int `json:"nodes"`
}

func (e *Engine) SeriesImpact(c Command) (SeriesImpact, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out SeriesImpact
	root, err := scanTask(e.db.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", c.RootID))
	if err != nil {
		return out, err
	}
	if fmt.Sprint(root.Version) != c.ExpectedVersion {
		return out, fail("CONFLICT", "任务已在其他窗口更新")
	}
	if root.SeriesID == nil {
		return out, nil
	}
	err = e.db.QueryRow("SELECT COALESCE(SUM(status='pending' AND is_exception=0),0),COALESCE(SUM(status<>'pending' OR is_exception=1),0),COALESCE(SUM(status='pending'),0) FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND id<>?", *root.SeriesID, val(root.OccurrenceDate), root.ID).Scan(&out.Updated, &out.Preserved, &out.Removed)
	if err == nil {
		err = e.db.QueryRow("SELECT count(DISTINCT root_id),count(*) FROM tasks WHERE root_id IN(SELECT id FROM tasks WHERE parent_id IS NULL AND series_id=? AND occurrence_date>=? AND status='pending')", *root.SeriesID, val(root.OccurrenceDate)).Scan(&out.Removed, &out.Nodes)
	}
	return out, err
}
