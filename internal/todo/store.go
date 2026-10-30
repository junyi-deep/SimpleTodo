package todo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gofrs/flock"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Engine struct {
	zone     atomic.Pointer[time.Location]
	lock     *flock.Flock
	mu       sync.Mutex
	db       *sql.DB
	Dir      string
	now      func() time.Time
	Changed  func()
	Notify   func(Notice) error
	Report   func(error)
	previews map[string]pendingImport
	lastDay  string
}

var defaults = map[string]string{"expandDepth": "2", "dayShowCompleted": "true", "inboxShowCompleted": "false", "dayShowRecurring": "true", "inboxShowRecurring": "true", "theme": "system", "autoPostpone": "true", "completedBottom": "true", "titleClick": "edit", "tabEntry": "false", "sound": "true", "focusMinutes": "25", "emoji": "☺", "miniView": "today", "miniPending": "true", "edgeHide": "false", "releaseURL": "", "shortcutMain": "", "shortcutMini": "", "shortcutQuick": "", "onboarding": "false", "animation": "true", "notificationPrompted": "false", "autostartConfigured": "false"}

func Open(dir string) (*Engine, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	e := &Engine{Dir: dir, now: time.Now, previews: map[string]pendingImport{}}
	e.lock = flock.New(filepath.Join(dir, ".simpletodo.lock"))
	locked, err := e.lock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, fail("DB_BUSY", "此数据目录正由另一个进程使用")
	}
	if err := e.openDB(); err != nil {
		_ = e.lock.Unlock()
		return nil, err
	}
	return e, nil
}
func (e *Engine) openDB() error {
	path := filepath.Join(e.Dir, "simpletodo.db")
	db, err := sql.Open("sqlite", sqliteURI(path)+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;"); err != nil {
		db.Close()
		return err
	}
	var n int
	err = db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE type='table' AND name='schema_migrations'").Scan(&n)
	if err != nil {
		db.Close()
		return err
	}
	if n == 0 {
		var existing int
		_ = db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&existing)
		if existing != 0 {
			db.Close()
			return fail("SCHEMA_ERROR", "数据库不是 SimpleTodo 数据库")
		}
		data, _ := migrations.ReadFile("migrations/001_init.sql")
		tx, er := db.Begin()
		if er != nil {
			db.Close()
			return er
		}
		if _, er = tx.Exec(string(data)); er == nil {
			sum := sha256.Sum256(data)
			_, er = tx.Exec("INSERT INTO schema_migrations VALUES(1,?,?)", hex.EncodeToString(sum[:]), e.now().UnixMilli())
		}
		if er != nil {
			tx.Rollback()
			db.Close()
			return er
		}
		if er = tx.Commit(); er != nil {
			db.Close()
			return er
		}
	} else {
		var version int
		if er := db.QueryRow("SELECT max(version) FROM schema_migrations").Scan(&version); er != nil || version < 1 || version > 2 {
			db.Close()
			return fail("SCHEMA_ERROR", "数据库版本不受当前版本支持，请保留原文件")
		}
		data, _ := migrations.ReadFile("migrations/001_init.sql")
		want := fmt.Sprintf("%x", sha256.Sum256(data))
		var checksum string
		if er := db.QueryRow("SELECT checksum FROM schema_migrations WHERE version=1").Scan(&checksum); er != nil || checksum != want {
			db.Close()
			return fail("SCHEMA_ERROR", "迁移校验不匹配，请保留数据库并使用匹配版本恢复")
		}
	}
	var version int
	if err = db.QueryRow("SELECT max(version) FROM schema_migrations").Scan(&version); err != nil {
		db.Close()
		return err
	}
	data, _ := migrations.ReadFile("migrations/002_interaction_calendar.sql")
	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	if version == 1 {
		if n != 0 {
			backupDir := filepath.Join(e.Dir, "backups")
			if err = os.MkdirAll(backupDir, 0700); err != nil {
				db.Close()
				return err
			}
			path := filepath.Join(backupDir, "pre-schema-v2-"+e.now().UTC().Format("20060102T150405.000000000")+".db")
			if _, err = db.Exec("VACUUM INTO ?", path); err != nil {
				db.Close()
				return fmt.Errorf("升级前备份失败：%w", err)
			}
		}
		tx, er := db.Begin()
		if er != nil {
			db.Close()
			return er
		}
		if _, er = tx.Exec(string(data)); er == nil {
			_, er = tx.Exec("INSERT INTO schema_migrations VALUES(2,?,?)", checksum, e.now().UnixMilli())
		}
		if er != nil {
			tx.Rollback()
			db.Close()
			return er
		}
		if er = tx.Commit(); er != nil {
			db.Close()
			return er
		}
	} else {
		var actual string
		if er := db.QueryRow("SELECT checksum FROM schema_migrations WHERE version=2").Scan(&actual); er != nil || actual != checksum {
			db.Close()
			return fail("SCHEMA_ERROR", "迁移校验不匹配，请保留数据库")
		}
	}
	e.db = db
	e.loadTimezone(db)
	for k, v := range defaults {
		if _, err = db.Exec("INSERT INTO settings(key,value_json) VALUES(?,?) ON CONFLICT(key) DO NOTHING", k, v); err != nil {
			db.Close()
			return err
		}
	}
	_, err = db.Exec("UPDATE notification_jobs SET state='uncertain' WHERE state='claimed'")
	return err
}
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	err := e.db.Close()
	if e.lock != nil {
		_ = e.lock.Unlock()
	}
	return err
}
func (e *Engine) signal() {
	if e.Changed != nil {
		e.Changed()
	}
}

type scanner interface{ Scan(...any) error }

const columns = "id,parent_id,root_id,title,description,planned_date,original_planned_date,status,color,sort_order,completed_at,auto_postpone_policy,series_id,rule_id,occurrence_date,is_exception,remind_at,reminder_local_time,reminder_timezone,reminder_generation,version,created_at,updated_at"

func scanTask(s scanner) (t Task, err error) {
	err = s.Scan(&t.ID, &t.ParentID, &t.RootID, &t.Title, &t.Description, &t.Date, &t.OriginalDate, &t.Status, &t.Color, &t.Order, &t.CompletedAt, &t.Postpone, &t.SeriesID, &t.RuleID, &t.OccurrenceDate, &t.Exception, &t.RemindAt, &t.ReminderTime, &t.ReminderZone, &t.ReminderGeneration, &t.Version, &t.CreatedAt, &t.UpdatedAt)
	return
}
func readTasks(rows *sql.Rows) ([]Task, error) {
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func putTask(tx *sql.Tx, t Task) error {
	parts := strings.Split(columns, ",")
	updates := []string{}
	for _, k := range parts[1:] {
		updates = append(updates, k+"=excluded."+k)
	}
	_, err := tx.Exec("INSERT INTO tasks("+columns+") VALUES("+strings.TrimSuffix(strings.Repeat("?,", len(parts)), ",")+") ON CONFLICT(id) DO UPDATE SET "+strings.Join(updates, ","), t.ID, t.ParentID, t.RootID, t.Title, t.Description, t.Date, t.OriginalDate, t.Status, t.Color, t.Order, t.CompletedAt, t.Postpone, t.SeriesID, t.RuleID, t.OccurrenceDate, t.Exception, t.RemindAt, t.ReminderTime, t.ReminderZone, t.ReminderGeneration, t.Version, t.CreatedAt, t.UpdatedAt)
	return err
}
func treeTx(tx *sql.Tx, id string) (map[string]Task, error) {
	rows, err := tx.Query("SELECT "+columns+" FROM tasks WHERE root_id=?", id)
	if err != nil {
		return nil, err
	}
	ts, err := readTasks(rows)
	m := map[string]Task{}
	for _, t := range ts {
		m[t.ID] = t
	}
	return m, err
}
func revisionTx(tx *sql.Tx) error {
	_, err := tx.Exec("UPDATE app_meta SET value=CAST(value AS INTEGER)+1 WHERE key='revision'")
	return err
}
func (e *Engine) Revision() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var s string
	_ = e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&s)
	return s
}
func validateDate(s *string) error {
	if s == nil {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil || t.Format("2006-01-02") != *s || *s < "2000-01-01" || *s > "2100-12-31" {
		return fail("VALIDATION_ERROR", "日期范围为 2000—2100 年")
	}
	return nil
}
func (e *Engine) Settings() (map[string]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.settingsLocked()
}
func (e *Engine) settingsLocked() (map[string]string, error) {
	rows, err := e.db.Query("SELECT key,value_json FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err = rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}
func (e *Engine) SaveSettings(m map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for k, v := range m {
		if _, ok := defaults[k]; !ok {
			return fail("VALIDATION_ERROR", "未知设置项")
		}
		switch k {
		case "expandDepth":
			n, er := strconv.Atoi(v)
			if er != nil || n < 1 || n > 5 {
				return fail("VALIDATION_ERROR", "展开层级为 1—5 层")
			}
		case "theme":
			if !contains([]string{"system", "light", "dark"}, v) {
				return fail("VALIDATION_ERROR", "无效主题")
			}
		case "titleClick":
			if !contains([]string{"edit", "complete"}, v) {
				return fail("VALIDATION_ERROR", "无效点击行为")
			}
		case "miniView":
			if !contains([]string{"today", "pending"}, v) {
				return fail("VALIDATION_ERROR", "无效浮窗模式")
			}
		case "focusMinutes":
			n, er := strconv.Atoi(v)
			if er != nil || n < 1 || n > 180 {
				return fail("VALIDATION_ERROR", "专注时长为 1—180 分钟")
			}
		case "dayShowCompleted", "inboxShowCompleted", "dayShowRecurring", "inboxShowRecurring", "autoPostpone", "completedBottom", "tabEntry", "sound", "miniPending", "edgeHide", "onboarding", "animation", "notificationPrompted", "autostartConfigured":
			if v != "true" && v != "false" {
				return fail("VALIDATION_ERROR", "无效开关值")
			}
		}
		if len(v) > 1000 {
			return fail("VALIDATION_ERROR", "设置值过长")
		}
		if _, err = tx.Exec("UPDATE settings SET value_json=?,version=version+1 WHERE key=?", v, k); err != nil {
			return err
		}
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		if _, ok := m["autoPostpone"]; ok {
			e.lastDay = ""
		}
		e.signal()
	}
	return err
}
func (e *Engine) Query(q Query) (Snapshot, error) {
	if q.View == "day" || q.View == "month" {
		if err := validateDate(&q.Date); err != nil {
			return Snapshot{}, err
		}
		if q.End == "" {
			q.End = q.Date
		}
		if err := validateDate(&q.End); err != nil {
			return Snapshot{}, err
		}
		if q.End < q.Date {
			return Snapshot{}, fail("VALIDATION_ERROR", "日期范围顺序不正确")
		}
		end, _ := time.Parse("2006-01-02", q.End)
		start, _ := time.Parse("2006-01-02", q.Date)
		if end.Sub(start) > 366*24*time.Hour {
			return Snapshot{}, fail("VALIDATION_ERROR", "日期范围最多 366 天")
		}
		if err := e.EnsureThrough(q.End); err != nil {
			return Snapshot{}, err
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	result := Snapshot{Tasks: []Task{}, Counts: map[string]int{}, Today: today(e.localNow())}
	where := "parent_id IS NULL AND status<>'skipped'"
	args := []any{}
	switch q.View {
	case "all":
	case "day", "month":
		where += " AND planned_date BETWEEN ? AND ?"
		args = append(args, q.Date, q.End)
	case "inbox":
		where += " AND planned_date IS NULL"
	case "pending":
		where = "parent_id IS NULL AND status='pending'"
	default:
		return result, fail("VALIDATION_ERROR", "未知视图")
	}
	if q.Limit < 1 {
		q.Limit = 500
	}
	if q.Limit > 10000 {
		q.Limit = 10000
	}
	rootSQL := "SELECT id FROM tasks WHERE " + where + " ORDER BY planned_date IS NULL,planned_date,sort_order,id LIMIT ?"
	args = append(args, q.Limit+1)
	rows, err := e.db.Query(rootSQL, args...)
	if err != nil {
		return result, err
	}
	ids := []any{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return result, err
		}
		ids = append(ids, id)
	}
	if err = rows.Close(); err != nil {
		return result, err
	}
	if len(ids) > q.Limit {
		result.Truncated = true
		ids = ids[:q.Limit]
	}
	if len(ids) > 0 {
		rows, err = e.db.Query("SELECT "+columns+" FROM tasks WHERE root_id IN ("+strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")+") ORDER BY sort_order,id", ids...)
		if err != nil {
			return result, err
		}
		result.Tasks, err = readTasks(rows)
		if err != nil {
			return result, err
		}
	}
	for k, sqlq := range map[string]string{"inbox": "SELECT count(*) FROM tasks WHERE parent_id IS NULL AND status='pending' AND planned_date IS NULL", "pending": "SELECT count(*) FROM tasks WHERE parent_id IS NULL AND status='pending'", "today": "SELECT count(*) FROM tasks WHERE parent_id IS NULL AND status='pending' AND planned_date=?"} {
		var n int
		if k == "today" {
			err = e.db.QueryRow(sqlq, result.Today).Scan(&n)
		} else {
			err = e.db.QueryRow(sqlq).Scan(&n)
		}
		if err != nil {
			return result, err
		}
		result.Counts[k] = n
	}
	err = e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&result.Revision)
	return result, err
}
func (e *Engine) mutate(c Command, op string, fn func(map[string]Task, int64) error) ([]Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	raw, _ := json.Marshal(struct {
		Op string
		C  Command
	}{op, c})
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))
	if c.CommandID != "" {
		var h, result string
		er := tx.QueryRow("SELECT request_hash,result_json FROM command_receipts WHERE command_id=?", c.CommandID).Scan(&h, &result)
		if er == nil {
			if h != hash {
				return nil, fail("CONFLICT", "请求编号已用于其他操作")
			}
			var out []Task
			er = json.Unmarshal([]byte(result), &out)
			return out, er
		}
		if er != sql.ErrNoRows {
			return nil, er
		}
	}
	m, err := treeTx(tx, c.RootID)
	if err != nil {
		return nil, err
	}
	old := map[string]Task{}
	for k, v := range m {
		old[k] = v
	}
	if len(m) > 0 {
		root := m[c.RootID]
		if c.ExpectedVersion == "" || fmt.Sprint(root.Version) != c.ExpectedVersion {
			return nil, fail("CONFLICT", "任务已在其他窗口更新，请刷新后重试")
		}
	} else if op != "create" {
		return nil, fail("NOT_FOUND", "任务不存在")
	}
	now := e.now().UnixMilli()
	if err = fn(m, now); err != nil {
		return nil, err
	}
	// Insert the new sibling in the same transaction as creation, including root peers.
	if c.AfterID != "" && (op == "create" || op == "create-node") {
		created := m[c.ID]
		if op == "create-node" {
			after, ok := m[c.AfterID]
			if !ok || val(after.ParentID) != val(created.ParentID) {
				return nil, fail("CONFLICT", "同级任务已变化，请刷新后重试")
			}
			for id, t := range m {
				if id != created.ID && val(t.ParentID) == val(created.ParentID) && t.Order > after.Order {
					t.Order++
					m[id] = t
				}
			}
			created.Order = after.Order + 1
		} else {
			rows, er := tx.Query("SELECT "+columns+" FROM tasks WHERE parent_id IS NULL AND planned_date IS ? ORDER BY sort_order,id", created.Date)
			if er != nil {
				return nil, er
			}
			peers, er := readTasks(rows)
			if er != nil {
				return nil, er
			}
			found, order := false, 0
			for _, t := range peers {
				t.Order = order
				order++
				if t.ID == c.AfterID {
					found = true
					created.Order = order
					order++
				}
				t.Version++
				t.UpdatedAt = now
				if er = putTask(tx, t); er != nil {
					return nil, er
				}
			}
			if !found {
				return nil, fail("CONFLICT", "同级任务已变化，请刷新后重试")
			}
		}
		m[c.ID] = created
	}
	if err = validateTree(m, old); err != nil {
		return nil, err
	}
	for id := range old {
		if _, ok := m[id]; !ok {
			if _, err = tx.Exec("DELETE FROM tasks WHERE id=?", id); err != nil {
				return nil, err
			}
		}
	}
	for id, t := range m {
		t.Version++
		t.UpdatedAt = now
		if t.CreatedAt == 0 {
			t.CreatedAt = now
		}
		m[id] = t
		if err = putTask(tx, t); err != nil {
			return nil, err
		}
	}
	if c.RepeatRule != nil {
		if c.ID != c.RootID {
			return nil, fail("VALIDATION_ERROR", "只能为根任务设置重复")
		}
		if err = e.repeatTx(tx, RepeatRequest{Command: c, Rule: *c.RepeatRule, Following: c.Following}, m); err != nil {
			return nil, err
		}
	}
	if root, ok := m[c.RootID]; ok {
		if err = syncReminder(tx, root, now); err != nil {
			return nil, err
		}
	} else {
		if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending'", c.RootID); err != nil {
			return nil, err
		}
	}
	out := []Task{}
	for _, t := range m {
		out = append(out, t)
	}
	if err = revisionTx(tx); err != nil {
		return nil, err
	}
	if c.CommandID != "" {
		b, _ := json.Marshal(out)
		if _, err = tx.Exec("INSERT INTO command_receipts VALUES(?,?,?,?)", c.CommandID, hash, string(b), now); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err == nil {
		if c.RepeatRule != nil {
			e.lastDay = ""
		}
		e.signal()
	}
	return out, err
}

func (e *Engine) DataDir() string { e.mu.Lock(); defer e.mu.Unlock(); return e.Dir }

func sqliteURI(path string) string {
	absolute, err := filepath.Abs(path)
	if err == nil {
		path = absolute
	}
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}
