package todo

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Promote detaches a complete child branch without changing its IDs or date.
// Both the old parent's aggregate and the new root order commit together.
func (e *Engine) Promote(c Command, revision string) ([]Task, error) {
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
	tree, err := treeTx(tx, c.RootID)
	if err != nil {
		return nil, err
	}
	node, ok := tree[c.ID]
	if !ok || node.ParentID == nil {
		return nil, fail("VALIDATION_ERROR", "仅子任务可以提升为主任务")
	}
	root := tree[c.RootID]
	if fmt.Sprint(root.Version) != c.ExpectedVersion {
		return nil, fail("CONFLICT", "任务已变化，请刷新后重试")
	}
	branch := descendants(tree, node.ID)
	previous := map[string]Task{}
	for id, t := range tree {
		previous[id] = t
	}
	promoted := map[string]Task{}
	for id := range branch {
		t := tree[id]
		delete(tree, id)
		t.RootID = node.ID
		if id == node.ID {
			t.ParentID = nil
			t.Date = root.Date
			t.Color = ptr("none")
			t.Description = ptr("")
			t.Postpone = root.Postpone
			t.SeriesID = nil
			t.RuleID = nil
			t.OccurrenceDate = nil
			t.Exception = false
			t.RemindAt = nil
			t.ReminderTime = nil
			t.ReminderZone = nil
			t.ReminderGeneration = 0
		}
		promoted[id] = t
	}
	now := e.now().UnixMilli()
	aggregate(tree, now)
	aggregate(promoted, now)
	markException(tree, c.RootID)
	if err = validateTree(tree, previous); err != nil {
		return nil, err
	}
	if err = validateTree(promoted); err != nil {
		return nil, err
	}
	rows, err := tx.Query("SELECT id FROM tasks WHERE parent_id IS NULL AND status!='skipped' ORDER BY sort_order,id")
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	at := len(ids)
	if c.BeforeID != "" {
		at = -1
		for i, id := range ids {
			if id == c.BeforeID {
				at = i
				break
			}
		}
		if at < 0 {
			return nil, fail("CONFLICT", "目标主任务已变化")
		}
	}
	ids = append(ids, "")
	copy(ids[at+1:], ids[at:])
	ids[at] = node.ID
	out := []Task{}
	for _, m := range []map[string]Task{tree, promoted} {
		for _, t := range m {
			t.Version++
			t.UpdatedAt = now
			if err = putTask(tx, t); err != nil {
				return nil, err
			}
			out = append(out, t)
		}
	}
	for i, id := range ids {
		if _, err = tx.Exec("UPDATE tasks SET sort_order=?,version=version+1,updated_at=? WHERE id=?", i, now, id); err != nil {
			return nil, err
		}
	}
	if err = syncReminder(tx, tree[c.RootID], now); err != nil {
		return nil, err
	}
	if err = revisionTx(tx); err != nil {
		return nil, err
	}
	// Read the committed shape into the receipt, including final root versions/order.
	out = []Task{}
	for _, m := range []map[string]Task{tree, promoted} {
		for id := range m {
			var t Task
			t, err = scanTask(tx.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", id))
			if err != nil {
				return nil, err
			}
			out = append(out, t)
		}
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
