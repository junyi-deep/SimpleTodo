package todo

func (e *Engine) TaskExpansion() (map[string]bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT task_id,expanded FROM task_expansion")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		var expanded bool
		if err = rows.Scan(&id, &expanded); err != nil {
			return nil, err
		}
		out[id] = expanded
	}
	return out, rows.Err()
}
func (e *Engine) SaveTaskExpansion(states map[string]bool) error {
	if len(states) > 10000 {
		return fail("VALIDATION_ERROR", "展开状态过多")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for id, expanded := range states {
		if _, err = tx.Exec("INSERT INTO task_expansion(task_id,expanded) SELECT id,? FROM tasks WHERE id=? ON CONFLICT(task_id) DO UPDATE SET expanded=excluded.expanded", expanded, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
