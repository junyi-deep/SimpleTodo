package todo

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

func sameTaskContent(a, b Task) bool {
	return a.Title == b.Title && val(a.Description) == val(b.Description) && a.StartAt == b.StartAt && val(a.EndAt) == val(b.EndAt) && reflect.DeepEqual(a.TagIDs, b.TagIDs) && val(a.RemindAt) == val(b.RemindAt)
}

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
	Color string `json:"color"`
}

const defaultTagColor = "#64748b"

var tagColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func tagColor(color string) string {
	if color == "" {
		return defaultTagColor
	}
	return strings.ToLower(color)
}

func (e *Engine) Tags() ([]Tag, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT id,name,sort_order,color FROM tags ORDER BY sort_order,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tag{}
	for rows.Next() {
		var t Tag
		if err = rows.Scan(&t.ID, &t.Name, &t.Order, &t.Color); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (e *Engine) SaveTags(tags []Tag) error {
	if len(tags) > 200 {
		return fail("VALIDATION_ERROR", "最多200个标签")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	names := map[string]bool{}
	for _, t := range tags {
		if t.Color != "" && !tagColorPattern.MatchString(t.Color) {
			return fail("VALIDATION_ERROR", "标签颜色须为六位十六进制颜色")
		}
		name := strings.TrimSpace(t.Name)
		if name == "" || len([]rune(name)) > 40 || names[name] {
			return fail("VALIDATION_ERROR", "标签名称须为1—40字且不能重复")
		}
		names[name] = true
	}
	if _, err = tx.Exec("UPDATE tags SET name=?||id", ID()); err != nil {
		return err
	}
	seen := map[string]bool{}
	for i, t := range tags {
		t.Name = strings.TrimSpace(t.Name)
		if t.Name == "" || len([]rune(t.Name)) > 40 || seen[t.ID] {
			return fail("VALIDATION_ERROR", "标签名称须为1—40字且不能重复")
		}
		if t.ID == "" {
			t.ID = ID()
		}
		seen[t.ID] = true
		if t.Color == "" {
			_ = tx.QueryRow("SELECT color FROM tags WHERE id=?", t.ID).Scan(&t.Color)
		}
		if _, err = tx.Exec("INSERT INTO tags(id,name,sort_order,color) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,sort_order=excluded.sort_order,color=excluded.color", t.ID, t.Name, i, tagColor(t.Color)); err != nil {
			return fail("VALIDATION_ERROR", "标签名称不能重复")
		}
	}
	rows, err := tx.Query("SELECT id FROM tags")
	if err != nil {
		return err
	}
	remove := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		if !seen[id] {
			remove = append(remove, id)
		}
	}
	rows.Close()
	for _, id := range remove {
		if _, err = tx.Exec("DELETE FROM tags WHERE id=?", id); err != nil {
			return err
		}
	}
	if len(remove) > 0 {
		rs, er := tx.Query("SELECT " + columns + " FROM tasks WHERE tags_json<>'[]'")
		if er != nil {
			return er
		}
		tasks, er := readTasks(rs)
		if er != nil {
			return er
		}
		roots := map[string]bool{}
		changed := map[string]bool{}
		keep := func(ids []string) []string {
			out := []string{}
			for _, id := range ids {
				if seen[id] {
					out = append(out, id)
				}
			}
			return out
		}
		for _, t := range tasks {
			ids := keep(t.TagIDs)
			if len(ids) != len(t.TagIDs) {
				t.TagIDs = ids
				t.UpdatedAt = e.now().UnixMilli()
				t.Version++
				if er = putTask(tx, t); er != nil {
					return er
				}
				changed[t.ID] = true
				roots[t.RootID] = true
			}
		}
		for root := range roots {
			if !changed[root] {
				if _, er = tx.Exec("UPDATE tasks SET version=version+1 WHERE id=?", root); er != nil {
					return er
				}
			}
		}
		rules, er := allRules(tx)
		if er != nil {
			return er
		}
		for _, r := range rules {
			dirty := false
			for i := range r.Template {
				ids := keep(r.Template[i].TagIDs)
				if len(ids) != len(r.Template[i].TagIDs) {
					r.Template[i].TagIDs = ids
					dirty = true
				}
			}
			if dirty {
				if er = putRule(tx, r); er != nil {
					return er
				}
			}
		}
		for _, key := range []string{"dayFilterTags", "inboxFilterTags"} {
			var raw string
			if er = tx.QueryRow("SELECT value_json FROM settings WHERE key=?", key).Scan(&raw); er == nil {
				var ids []string
				_ = json.Unmarshal([]byte(raw), &ids)
				data, _ := json.Marshal(keep(ids))
				if _, er = tx.Exec("UPDATE settings SET value_json=?,version=version+1 WHERE key=?", string(data), key); er != nil {
					return er
				}
			}
		}
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		e.signal()
	}
	return err
}
func (e *Engine) SetTaskTags(c Command, ids []string) ([]Task, error) {
	if len(ids) > 200 {
		return nil, fail("VALIDATION_ERROR", "标签过多")
	}
	tags, err := e.Tags()
	if err != nil {
		return nil, err
	}
	valid := map[string]bool{}
	for _, t := range tags {
		valid[t.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !valid[id] || seen[id] {
			return nil, fail("VALIDATION_ERROR", "标签已变化，请重新选择")
		}
		seen[id] = true
	}
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	return e.mutate(c, "tags", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		t.TagIDs = ids
		m[c.ID] = t
		markException(m, c.RootID)
		return nil
	})
}

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	Version   int64  `json:"version"`
}

func (e *Engine) Notes(search string) ([]Note, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT id,title,content,created_at,updated_at,version FROM notes WHERE instr(lower(title||' '||content),lower(?))>0 ORDER BY updated_at DESC,id LIMIT 1000", search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Note{}
	for rows.Next() {
		var n Note
		if err = rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt, &n.Version); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
func (e *Engine) SaveNote(n Note) (Note, error) {
	n.Title = strings.TrimSpace(n.Title)
	if n.Title == "" || len([]rune(n.Title)) > 200 || len(n.Content) > 1000000 {
		return n, fail("VALIDATION_ERROR", "笔记标题必填（最多200字），内容最多1MB")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now().UnixMilli()
	if n.ID == "" {
		n.ID = ID()
		n.CreatedAt = now
		n.UpdatedAt = now
		n.Version = 1
		_, err := e.db.Exec("INSERT INTO notes VALUES(?,?,?,?,?,?)", n.ID, n.Title, n.Content, n.CreatedAt, n.UpdatedAt, n.Version)
		if err == nil {
			e.signal()
		}
		return n, err
	}
	var old Note
	err := e.db.QueryRow("SELECT title,content,created_at,updated_at,version FROM notes WHERE id=?", n.ID).Scan(&old.Title, &old.Content, &old.CreatedAt, &old.UpdatedAt, &old.Version)
	if err != nil {
		return n, err
	}
	if old.Version != n.Version {
		return n, fail("CONFLICT", "笔记已在其他窗口更新，请保留草稿后重新打开")
	}
	n.CreatedAt = old.CreatedAt
	n.UpdatedAt = old.UpdatedAt
	if old.Title != n.Title || old.Content != n.Content {
		n.UpdatedAt = now
	}
	n.Version++
	_, err = e.db.Exec("UPDATE notes SET title=?,content=?,updated_at=?,version=? WHERE id=? AND version=?", n.Title, n.Content, n.UpdatedAt, n.Version, n.ID, old.Version)
	if err == nil {
		e.signal()
	}
	return n, err
}
func (e *Engine) DeleteNote(id string, version int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	r, err := e.db.Exec("DELETE FROM notes WHERE id=? AND version=?", id, version)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return fail("CONFLICT", "笔记已变化")
	}
	e.signal()
	return nil
}
func (e *Engine) TaskReport(q Query) ([]Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	query := "SELECT " + columns + " FROM tasks WHERE status<>'skipped' ORDER BY created_at,id"
	if q.View == "gantt" {
		query = "SELECT t." + strings.ReplaceAll(columns, ",", ",t.") + " FROM tasks t JOIN tasks r ON r.id=t.root_id LEFT JOIN repeat_series s ON s.id=r.series_id WHERE t.status<>'skipped' AND (s.stopped_from IS NULL OR r.status='completed') ORDER BY t.created_at,t.id"
	}
	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	return readTasks(rows)
}

// Time ranges control the days on which a task can appear.
func (e *Engine) SetTaskSpan(c Command, start int64, end *int64) ([]Task, error) {
	if start <= 0 || end != nil && *end < start {
		return nil, fail("VALIDATION_ERROR", "结束时间不能早于开始时间")
	}
	c.Span = &TaskSpan{Start: start, End: end}
	return e.mutate(c, "span", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		t.StartAt = start
		t.EndAt = end
		m[c.ID] = t
		markException(m, c.RootID)
		return nil
	})
}
func (e *Engine) spanRulesTx(tx *sql.Tx, c Command, m map[string]Task) error {
	root, ok := m[c.RootID]
	if !ok || c.ID != c.RootID || e.periodModel && root.Status == "completed" {
		return nil
	}
	if root.SeriesID == nil {
		if root.Date == nil {
			d := time.UnixMilli(c.Span.Start).In(e.location()).Format("2006-01-02")
			root.Date = &d
			m[root.ID] = root
		}
		var until *string
		if c.Span.End != nil {
			d := time.UnixMilli(*c.Span.End).In(e.location()).Format("2006-01-02")
			until = &d
		}
		if err := e.repeatTx(tx, RepeatRequest{Command: c, Rule: Rule{Frequency: "daily", Interval: 1, Until: until}}, m); err != nil {
			return err
		}
		root = m[c.RootID]
	}
	var stopped *string
	if er := tx.QueryRow("SELECT stopped_from FROM repeat_series WHERE id=?", *root.SeriesID).Scan(&stopped); er != nil {
		return er
	}
	if stopped != nil {
		return nil
	}
	rules, err := allRules(tx)
	if err != nil {
		return err
	}
	from := time.UnixMilli(c.Span.Start).In(e.location()).Format("2006-01-02")
	var until *string
	if c.Span.End != nil {
		s := time.UnixMilli(*c.Span.End).In(e.location()).Format("2006-01-02")
		until = &s
	}
	earliest := "9999-12-31"
	for _, r := range rules {
		if r.SeriesID == *root.SeriesID && r.From < earliest {
			earliest = r.From
		}
	}
	for _, r := range rules {
		if r.SeriesID != *root.SeriesID {
			continue
		}
		if from < r.From && r.From == earliest {
			r.From = from
		}
		r.Frequency = "daily"
		r.Interval = 1
		r.Until = until
		r.Anchor = from
		r.Template = sortedTree(m)
		r.GeneratedThrough = e.localNow().AddDate(0, 0, -1).Format("2006-01-02")
		if err = putRule(tx, r); err != nil {
			return err
		}
	}
	cutoff := today(e.localNow())
	_, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND planned_date>? AND status='pending' AND is_exception=0 AND id<>?", *root.SeriesID, cutoff, root.ID)
	return err
}
func (e *Engine) refreshDailyTemplateTx(tx *sql.Tx, m map[string]Task) error {
	var root Task
	for _, t := range m {
		if t.ParentID == nil {
			root = t
			break
		}
	}
	if root.SeriesID == nil || root.Status != "pending" || val(root.Date) < today(e.localNow()) {
		return nil
	}
	rules, err := allRules(tx)
	if err != nil {
		return err
	}
	for _, r := range rules {
		if r.ID != val(root.RuleID) || r.Frequency != "daily" {
			continue
		}
		r.Template = sortedTree(m)
		// A reopened series may already need to regenerate dates before this row.
		r.GeneratedThrough = min(r.GeneratedThrough, val(root.Date))
		if err = putRule(tx, r); err != nil {
			return err
		}
		_, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND planned_date>? AND status='pending' AND is_exception=0", *root.SeriesID, val(root.Date))
		return err
	}
	return nil
}

func (e *Engine) stopPeriodTx(tx *sql.Tx, root Task) error {
	cut := val(root.Date)
	if cut == "" {
		cut = today(e.localNow())
	}
	return e.stopPeriodFromTx(tx, root, cut)
}

func (e *Engine) stopPeriodFromTx(tx *sql.Tx, root Task, cut string) error {
	if _, err := tx.Exec("UPDATE repeat_series SET stopped_from=CASE WHEN stopped_from IS NULL OR stopped_from>? THEN ? ELSE stopped_from END,version=version+1 WHERE id=?", cut, cut, *root.SeriesID); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND state='pending' AND source_id IN (SELECT id FROM tasks WHERE series_id=? AND planned_date>=?)", *root.SeriesID, cut); err != nil {
		return err
	}
	_, err := tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND planned_date>=? AND id<>? AND status='pending'", *root.SeriesID, cut, root.ID)
	return err
}
