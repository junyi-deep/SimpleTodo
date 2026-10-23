package todo

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// Nest moves a whole branch atomically. The snapshot revision protects both
// roots, and the receipt makes retries safe even after the source changes root.
func (e *Engine) Nest(c Command, revision string) ([]Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	raw, _ := json.Marshal(struct {
		C        Command
		Revision string
	}{c, revision})
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
	var current string
	if err = tx.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&current); err != nil {
		return nil, err
	}
	if current != revision {
		return nil, fail("CONFLICT", "列表已变化，请刷新后重试")
	}
	if c.ParentID == nil || *c.ParentID == c.ID {
		return nil, fail("VALIDATION_ERROR", "不能移动到自身")
	}
	source, err := treeTx(tx, c.RootID)
	if err != nil {
		return nil, err
	}
	node, ok := source[c.ID]
	if !ok {
		return nil, fail("NOT_FOUND", "任务不存在")
	}
	if fmt.Sprint(source[c.RootID].Version) != c.ExpectedVersion {
		return nil, fail("CONFLICT", "任务已变化，请刷新后重试")
	}
	var targetRoot string
	if err = tx.QueryRow("SELECT root_id FROM tasks WHERE id=?", *c.ParentID).Scan(&targetRoot); err != nil {
		return nil, fail("NOT_FOUND", "目标任务不存在")
	}
	target := source
	if targetRoot != c.RootID {
		target, err = treeTx(tx, targetRoot)
		if err != nil {
			return nil, err
		}
	}
	previous := map[string]Task{}
	for id, task := range source {
		previous[id] = task
	}
	for id, task := range target {
		previous[id] = task
	}
	branch := descendants(source, c.ID)
	if branch[*c.ParentID] {
		return nil, fail("VALIDATION_ERROR", "不能移动到自身或后代")
	}
	now := e.now().UnixMilli()
	var marker *Task
	if targetRoot != c.RootID {
		if node.ParentID == nil && node.Status == "skipped" {
			return nil, fail("VALIDATION_ERROR", "已跳过的重复任务不能移动")
		}
		if node.ParentID == nil && node.SeriesID != nil {
			tomb := node
			tomb.ID = ID()
			tomb.RootID = tomb.ID
			tomb.Status = "skipped"
			tomb.CompletedAt = nil
			tomb.Exception = true
			tomb.RemindAt = nil
			tomb.ReminderTime = nil
			tomb.ReminderZone = nil
			tomb.Version = 0
			marker = &tomb
		}
		if text := strings.TrimSpace(val(node.Description)); text != "" {
			root := target[targetRoot]
			root.Description = ptr(strings.TrimSpace(val(root.Description) + "\n\n" + node.Title + "\n" + text))
			target[targetRoot] = root
		}
		for id := range branch {
			t := source[id]
			delete(source, id)
			t.RootID = targetRoot
			if id == c.ID {
				t.ParentID = c.ParentID
				t.Description = nil
				t.Date = nil
				t.OriginalDate = nil
				t.Color = nil
				t.Postpone = nil
				t.SeriesID = nil
				t.RuleID = nil
				t.OccurrenceDate = nil
				t.Exception = false
				t.RemindAt = nil
				t.ReminderTime = nil
				t.ReminderZone = nil
				t.ReminderGeneration = 0
			}
			target[id] = t
		}
	} else {
		node.ParentID = c.ParentID
		target[c.ID] = node
	}
	reorder(target, c.ID, c.BeforeID)
	aggregate(source, now)
	aggregate(target, now)
	markException(source, c.RootID)
	markException(target, targetRoot)
	if err = validateTree(source, previous); err != nil {
		return nil, err
	}
	if err = validateTree(target, previous); err != nil {
		return nil, err
	}
	out := []Task{}
	trees := []map[string]Task{target}
	if targetRoot != c.RootID {
		trees = append(trees, source)
	}
	if marker != nil {
		trees = append(trees, map[string]Task{marker.ID: *marker})
	}
	for _, tree := range trees {
		for _, t := range tree {
			t.Version++
			t.UpdatedAt = now
			if err = putTask(tx, t); err != nil {
				return nil, err
			}
			out = append(out, t)
		}
	}
	if node.ParentID == nil && targetRoot != c.RootID {
		if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending'", node.ID); err != nil {
			return nil, err
		}
	}
	if root, ok := source[c.RootID]; ok && targetRoot != c.RootID {
		if err = syncReminder(tx, root, now); err != nil {
			return nil, err
		}
	}
	if err = syncReminder(tx, target[targetRoot], now); err != nil {
		return nil, err
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
		e.signal()
	}
	return out, err
}
