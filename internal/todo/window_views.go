package todo

import (
	"encoding/json"
	"strings"
)

// ViewWindowState keeps normal bounds even while a view is maximised.
// It shares the backed-up window_states table without changing global display settings.
type ViewWindowState struct {
	View      string `json:"view"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Maximised bool   `json:"maximised"`
}

func (e *Engine) SaveViewWindow(state ViewWindowState) error {
	if !contains([]string{"day", "inbox", "calendar", "focus"}, state.View) || state.Width <= 0 || state.Height <= 0 {
		return fail("VALIDATION_ERROR", "无效视图窗口尺寸")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = e.db.Exec("INSERT INTO window_states(window_key,state_json) VALUES(?,?) ON CONFLICT(window_key) DO UPDATE SET state_json=excluded.state_json,version=version+1", "view:"+state.View, string(b))
	return err
}

func (e *Engine) ViewWindows() (map[string]ViewWindowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT window_key,state_json FROM window_states WHERE window_key LIKE 'view:%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]ViewWindowState{}
	for rows.Next() {
		var key, raw string
		var state ViewWindowState
		if err = rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &state); err != nil {
			return nil, err
		}
		if state.View == strings.TrimPrefix(key, "view:") && state.Width > 0 && state.Height > 0 {
			out[state.View] = state
		}
	}
	return out, rows.Err()
}
