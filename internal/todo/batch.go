package todo

import "time"

func (e *Engine) PreviewBatch(ids []string, action, revision string) (Preview, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !contains([]string{"complete", "pending", "delete"}, action) || len(ids) == 0 || len(ids) > 10000 {
		return Preview{}, fail("VALIDATION_ERROR", "请选择任务及有效批量操作")
	}
	var rev string
	if err := e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&rev); err != nil {
		return Preview{}, err
	}
	if rev != revision {
		return Preview{}, fail("CONFLICT", "列表已更新")
	}
	selected := map[string]Task{}
	trees := map[string]map[string]Task{}
	for _, id := range ids {
		if _, ok := selected[id]; ok {
			continue
		}
		t, err := scanTask(e.db.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", id))
		if err != nil {
			return Preview{}, fail("NOT_FOUND", "所选任务不存在，请重新选择")
		}
		selected[id] = t
		if trees[t.RootID] == nil {
			rows, er := e.db.Query("SELECT "+columns+" FROM tasks WHERE root_id=?", t.RootID)
			if er != nil {
				return Preview{}, er
			}
			ts, er := readTasks(rows)
			if er != nil {
				return Preview{}, er
			}
			m := map[string]Task{}
			for _, n := range ts {
				m[n.ID] = n
			}
			trees[t.RootID] = m
		}
	}
	anchors := []string{}
	tasks := []Task{}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		t := selected[id]
		m := trees[t.RootID]
		covered := false
		for p := t.ParentID; p != nil; {
			if _, ok := selected[*p]; ok {
				covered = true
				break
			}
			n, ok := m[*p]
			if !ok {
				break
			}
			p = n.ParentID
		}
		if covered {
			continue
		}
		anchors = append(anchors, id)
		sub := map[string]Task{}
		for key := range descendants(m, id) {
			sub[key] = m[key]
		}
		// sortedTree expects a root; retain actual parent IDs in the preview payload.
		ordered := sortedTree(m)
		for _, n := range ordered {
			if _, ok := sub[n.ID]; ok {
				tasks = append(tasks, n)
			}
		}
	}
	token := ID()
	e.previews[token] = pendingImport{Data: Transfer{Tasks: tasks, AppVersion: action}, BatchIDs: anchors, Expires: e.now().Add(5 * time.Minute), Revision: rev}
	return Preview{Token: token, Tasks: tasks, Roots: len(anchors), Nodes: len(tasks), Warnings: []string{"仅影响所选任务及其后代；父子同时选择时只处理一次，重复任务仅影响本次"}}, nil
}

func (e *Engine) CommitBatch(token string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.previews[token]
	if !ok || len(p.BatchIDs) == 0 || e.now().After(p.Expires) {
		return fail("PREVIEW_EXPIRED", "请重新预览")
	}
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var rev string
	if err = tx.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&rev); err != nil {
		return err
	}
	if rev != p.Revision {
		return fail("CONFLICT", "数据已经变化")
	}
	trees := map[string]map[string]Task{}
	for _, n := range p.Data.Tasks {
		if trees[n.RootID] == nil {
			m, er := treeTx(tx, n.RootID)
			if er != nil {
				return er
			}
			trees[n.RootID] = m
		}
	}
	now := e.now().UnixMilli()
	for rootID, m := range trees {
		previous := map[string]Task{}
		for id, t := range m {
			previous[id] = t
		}
		for _, id := range p.BatchIDs {
			t, exists := m[id]
			if !exists {
				continue
			}
			if p.Data.AppVersion == "delete" {
				branch := descendants(m, id)
				if t.ParentID == nil && t.SeriesID != nil {
					t.Status = "skipped"
					t.CompletedAt = nil
					t.RemindAt = nil
					t.Exception = true
					m[id] = t
					delete(branch, id)
				}
				for key := range branch {
					delete(m, key)
				}
			} else {
				for key := range descendants(m, id) {
					m[key] = setState(m[key], p.Data.AppVersion == "complete", now)
				}
			}
		}
		if len(m) > 0 {
			aggregate(m, now)
			markException(m, rootID)
			if err = validateTree(m, previous); err != nil {
				return err
			}
		}
		for id := range previous {
			if _, exists := m[id]; !exists {
				if _, err = tx.Exec("DELETE FROM tasks WHERE id=?", id); err != nil {
					return err
				}
			}
		}
		for _, t := range m {
			t.Version++
			t.UpdatedAt = now
			if err = putTask(tx, t); err != nil {
				return err
			}
		}
		if root, exists := m[rootID]; exists {
			if err = syncReminder(tx, root, now); err != nil {
				return err
			}
		} else if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending'", rootID); err != nil {
			return err
		}
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		delete(e.previews, token)
		e.signal()
	}
	return err
}
