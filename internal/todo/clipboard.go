package todo

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const copyMarker = "<!-- simpletodo:v1:"

func (e *Engine) CopyTasks(q Query, ids []string) (string, error) {
	q.Limit = 10000
	snap, err := e.Query(q)
	if err != nil {
		return "", err
	}
	selected := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	d := Transfer{Version: 1, Tasks: []Task{}}
	if q.View == "month" || q.View == "day" {
		d.SourceStart = ptr(q.Date)
	}
	for _, t := range snap.Tasks {
		if len(selected) > 0 && !selected[t.RootID] {
			continue
		}
		t = setState(t, false, e.now().UnixMilli())
		t.Exception = false
		if t.ParentID == nil {
			t.SeriesID = nil
			t.RuleID = nil
			t.OccurrenceDate = nil
			t.RemindAt = nil
			t.ReminderTime = nil
			t.ReminderZone = nil
			t.ReminderGeneration = 0
		}
		d.Tasks = append(d.Tasks, t)
	}
	if len(d.Tasks) == 0 {
		return "", fail("VALIDATION_ERROR", "没有可复制的任务")
	}
	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	text, err := e.Export("markdown", q, q.View == "inbox", ids)
	if err != nil {
		return "", err
	}
	return copyMarker + base64.StdEncoding.EncodeToString(b) + " -->\n" + text, nil
}
func (e *Engine) PreviewPaste(content, target string) (Preview, error) {
	if err := validateDate(&target); err != nil {
		return Preview{}, err
	}
	if len(content) > 20*1024*1024 {
		return Preview{}, fail("VALIDATION_ERROR", "剪贴板内容超过 20 MiB")
	}
	var d Transfer
	if strings.HasPrefix(content, copyMarker) {
		end := strings.Index(content, " -->")
		if end < 0 {
			return Preview{}, fail("VALIDATION_ERROR", "剪贴板结构损坏")
		}
		b, err := base64.StdEncoding.DecodeString(content[len(copyMarker):end])
		if err != nil {
			return Preview{}, err
		}
		if err = json.Unmarshal(b, &d); err != nil {
			return Preview{}, err
		}
		if err = validateTransfer(&d); err != nil {
			return Preview{}, err
		}
		origin := val(d.SourceStart)
		for _, t := range d.Tasks {
			if t.ParentID == nil && t.Date != nil && d.SourceStart == nil && (origin == "" || *t.Date < origin) {
				origin = *t.Date
			}
		}
		start, _ := time.Parse("2006-01-02", target)
		for i, t := range d.Tasks {
			t = setState(t, false, e.now().UnixMilli())
			if t.ParentID == nil {
				next := target
				if t.Date != nil && origin != "" {
					a, _ := time.Parse("2006-01-02", origin)
					b, _ := time.Parse("2006-01-02", *t.Date)
					next = start.AddDate(0, 0, int(b.Sub(a)/(24*time.Hour))).Format("2006-01-02")
				}
				t.Date = ptr(next)
			}
			d.Tasks[i] = t
		}
	} else {
		var err error
		d = Transfer{Version: 1}
		d.Tasks, err = parseList(content, "text", &target)
		if err != nil {
			return Preview{}, err
		}
	}
	raw, err := json.Marshal(d)
	if err != nil {
		return Preview{}, err
	}
	p, err := e.PreviewImport(string(raw), "json", &target, false)
	if err == nil {
		p.Warnings = append(p.Warnings, "按目标首日对齐源首日并保留相对日期偏移；实际日期以预览为准")
	}
	return p, err
}
func (e *Engine) ToggleCountdown(id string, enabled bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	cs, err := countdowns(e.db)
	if err != nil {
		return err
	}
	for _, c := range cs {
		if c.ID != id {
			continue
		}
		if enabled && c.Frequency == "none" && c.Target <= e.now().UnixMilli() {
			return fail("VALIDATION_ERROR", "事项已到期，请编辑目标时间后再启用")
		}
		c.Enabled = enabled
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
	return fail("NOT_FOUND", "倒计时不存在")
}
