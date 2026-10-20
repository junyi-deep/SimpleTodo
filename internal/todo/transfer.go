package todo

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gofrs/flock"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (e *Engine) exportLocked() (Transfer, error) {
	d := Transfer{Version: 1, AppVersion: "0.1.0", ExportedAt: e.now().UnixMilli(), Tasks: []Task{}, Focus: []Focus{}}
	rows, err := e.db.Query("SELECT " + columns + " FROM tasks ORDER BY planned_date,sort_order,id")
	if err != nil {
		return d, err
	}
	d.Tasks, err = readTasks(rows)
	if err != nil {
		return d, err
	}
	d.Rules, err = allRules(e.db)
	if err != nil {
		return d, err
	}
	d.Countdowns, err = countdowns(e.db)
	if err != nil {
		return d, err
	}
	seriesRows, er := e.db.Query("SELECT id,stopped_from,version FROM repeat_series")
	if er != nil {
		return d, er
	}
	for seriesRows.Next() {
		var item Series
		if er = seriesRows.Scan(&item.ID, &item.StoppedFrom, &item.Version); er != nil {
			seriesRows.Close()
			return d, er
		}
		d.Series = append(d.Series, item)
	}
	if er = seriesRows.Close(); er != nil {
		return d, er
	}
	jobRows, er := e.db.Query("SELECT id,source_type,source_id,generation,cycle_key,fire_at,state,title FROM notification_jobs")
	if er != nil {
		return d, er
	}
	for jobRows.Next() {
		var item Job
		if er = jobRows.Scan(&item.ID, &item.SourceType, &item.SourceID, &item.Generation, &item.CycleKey, &item.FireAt, &item.State, &item.Title); er != nil {
			jobRows.Close()
			return d, er
		}
		d.Jobs = append(d.Jobs, item)
	}
	if er = jobRows.Close(); er != nil {
		return d, er
	}
	d.Settings, err = e.settingsLocked()
	for _, key := range []string{"shortcutMain", "shortcutMini", "shortcutQuick"} {
		delete(d.Settings, key)
	}
	if err != nil {
		return d, err
	}
	rows, err = e.db.Query("SELECT id,task_id,task_title_snapshot,duration_seconds,remaining_seconds,deadline_at,state,started_at,finished_at FROM focus_sessions")
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var f Focus
		if err = rows.Scan(&f.ID, &f.TaskID, &f.Title, &f.Duration, &f.Remaining, &f.Deadline, &f.State, &f.StartedAt, &f.FinishedAt); err != nil {
			return d, err
		}
		d.Focus = append(d.Focus, f)
	}
	return d, rows.Err()
}
func (e *Engine) Export(format string, q Query, includeInbox bool, ids []string) (string, error) {
	if !contains([]string{"json", "csv", "markdown", "text", "html"}, format) {
		return "", fail("VALIDATION_ERROR", "不支持的导出格式")
	}
	if q.View == "day" || q.View == "month" {
		if err := validateDate(&q.Date); err != nil {
			return "", err
		}
		if q.End != "" {
			if err := validateDate(&q.End); err != nil {
				return "", err
			}
			if q.End < q.Date {
				return "", fail("VALIDATION_ERROR", "结束日期早于开始日期")
			}
		}
	}

	e.mu.Lock()
	d, err := e.exportLocked()
	e.mu.Unlock()
	if err != nil {
		return "", err
	}
	if format == "json" {
		b, er := json.MarshalIndent(d, "", "  ")
		return string(b), er
	}
	roots := []Task{}
	selected := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	trees := map[string]map[string]Task{}
	for _, t := range d.Tasks {
		if trees[t.RootID] == nil {
			trees[t.RootID] = map[string]Task{}
		}
		trees[t.RootID][t.ID] = t
		if t.ParentID == nil && t.Status != "skipped" {
			if len(selected) > 0 && !selected[t.ID] {
				continue
			}
			if len(selected) == 0 {
				if t.Date == nil && !includeInbox && q.View != "inbox" && q.View != "all" {
					continue
				}
				switch q.View {
				case "inbox":
					if t.Date != nil {
						continue
					}
				case "day", "month":
					end := q.End
					if end == "" {
						end = q.Date
					}
					if t.Date != nil && (*t.Date < q.Date || *t.Date > end) {
						continue
					}
				case "pending":
					if t.Status != "pending" {
						continue
					}
				}
			}
			roots = append(roots, t)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		if val(roots[i].Date) != val(roots[j].Date) {
			return val(roots[i].Date) < val(roots[j].Date)
		}
		return roots[i].Order < roots[j].Order
	})
	var b strings.Builder
	var csvw *csv.Writer
	if format == "csv" {
		b.WriteRune('\uFEFF')
		csvw = csv.NewWriter(&b)
		_ = csvw.Write([]string{"id", "root_task_id", "parent_id", "depth", "sort_order", "date", "title", "description", "status", "color", "completed_at"})
	}
	if format == "html" {
		b.WriteString("<!doctype html><html lang=\"zh-CN\"><meta charset=\"utf-8\"><title>SimpleTodo</title><style>body{font:15px system-ui;max-width:760px;margin:40px auto;padding:20px;color:#28362e}li{margin:8px 0;white-space:pre-wrap}.done{text-decoration:line-through;color:#68736d}</style><h1>SimpleTodo</h1>")
	}
	lastDate := "\x00"
	for _, root := range roots {
		date := val(root.Date)
		label := date
		if label == "" {
			label = "收集箱"
		}
		if date != lastDate {
			switch format {
			case "markdown":
				b.WriteString("\n## " + label + "\n\n")
			case "text":
				b.WriteString("\n" + label + "\n")
			case "html":
				b.WriteString("<h2>" + html.EscapeString(label) + "</h2>")
			}
			lastDate = date
		}
		m := trees[root.ID]
		list := sortedTree(m)
		for _, t := range list {
			depth := 0
			cur := t
			for cur.ParentID != nil {
				depth++
				cur = m[*cur.ParentID]
			}
			mark := " "
			if t.Status == "completed" {
				mark = "x"
			}
			switch format {
			case "markdown", "text":
				prefix := "- [" + mark + "] "
				if format == "text" {
					prefix = "[" + mark + "] "
				}
				b.WriteString(strings.Repeat("  ", depth) + prefix + t.Title + "\n")
				if t.ParentID == nil && val(t.Description) != "" {
					for _, line := range strings.Split(*t.Description, "\n") {
						b.WriteString("  > " + line + "\n")
					}
				}
			case "csv":
				record := []string{t.ID, t.RootID, val(t.ParentID), strconv.Itoa(depth), strconv.Itoa(t.Order), date, t.Title, val(t.Description), t.Status, val(root.Color), fmt.Sprint(val(t.CompletedAt))}
				for i, s := range record {
					if len(s) > 0 && strings.ContainsRune("=+-@", rune(s[0])) {
						record[i] = "'" + s
					}
				}
				if er := csvw.Write(record); er != nil {
					return "", er
				}
			case "html":
				class := ""
				if t.Status == "completed" {
					class = "done"
				}
				b.WriteString(fmt.Sprintf("<div style=\"margin-left:%dpx\" class=\"%s\">[%s] %s</div>", depth*20, class, mark, html.EscapeString(t.Title)))
				if t.ParentID == nil && val(t.Description) != "" {
					b.WriteString("<p>" + html.EscapeString(*t.Description) + "</p>")
				}
			default:
				return "", fail("VALIDATION_ERROR", "不支持的导出格式")
			}
		}
	}
	if csvw != nil {
		csvw.Flush()
		if err = csvw.Error(); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

var numberPrefix = regexp.MustCompile(`^\s*\d+[.)、]\s*`)

func parseList(content, format string, date *string) ([]Task, error) {
	out := []Task{}
	stack := []string{}
	now := time.Now().UnixMilli()
	lastRoot := -1
	lineOrder := map[string]int{}
	currentDate := date
	for lineNo, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if format == "markdown" && strings.HasPrefix(line, "## ") {
			label := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if label == "收集箱" {
				currentDate = nil
			} else if err := validateDate(&label); err == nil {
				currentDate = ptr(label)
			} else {
				return nil, fail("VALIDATION_ERROR", fmt.Sprintf("第 %d 行的日期标题无效", lineNo+1))
			}
			stack = nil
			continue
		}
		if format == "markdown" && strings.HasPrefix(strings.TrimSpace(line), ">") {
			if lastRoot < 0 {
				return nil, fail("VALIDATION_ERROR", "描述前缺少根任务")
			}
			desc := val(out[lastRoot].Description)
			if desc != "" {
				desc += "\n"
			}
			out[lastRoot].Description = ptr(desc + strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">")))
			continue
		}
		depth := 0
		trim := strings.TrimLeft(line, " \t")
		if format == "markdown" {
			spaces := 0
			for _, c := range line {
				if c == ' ' {
					spaces++
				} else if c == '\t' {
					spaces += 2
				} else {
					break
				}
			}
			if spaces%2 != 0 {
				return nil, fail("VALIDATION_ERROR", fmt.Sprintf("第 %d 行缩进必须为 2 个空格的倍数", lineNo+1))
			}
			depth = spaces / 2
			if depth > len(stack) {
				return nil, fail("VALIDATION_ERROR", fmt.Sprintf("第 %d 行不能跳级缩进", lineNo+1))
			}
		}
		done := false
		for _, prefix := range []string{"- [x] ", "- [X] ", "* [x] ", "[x] "} {
			if strings.HasPrefix(trim, prefix) {
				done = true
				trim = strings.TrimPrefix(trim, prefix)
				break
			}
		}
		for _, prefix := range []string{"- [ ] ", "* [ ] ", "[ ] ", "- ", "* ", "• "} {
			trim = strings.TrimPrefix(trim, prefix)
		}
		trim = numberPrefix.ReplaceAllString(trim, "")
		trim = strings.TrimSpace(trim)
		if trim == "" {
			continue
		}
		t := Task{ID: ID(), Title: trim, Status: "pending", Order: lineNo, Version: 1, CreatedAt: now, UpdatedAt: now}
		if done {
			t = setState(t, true, now)
		}
		if depth == 0 {
			t.RootID = t.ID
			t.Date = currentDate
			t.Color = ptr("none")
			t.Description = ptr("")
			t.Postpone = ptr("inherit")
			lastRoot = len(out)
		} else {
			t.ParentID = ptr(stack[depth-1])
			t.RootID = stack[0]
		}
		t.Order = lineOrder[val(t.ParentID)]
		lineOrder[val(t.ParentID)]++
		if len(stack) > depth {
			stack = stack[:depth]
		}
		stack = append(stack, t.ID)
		out = append(out, t)
	}
	return out, nil
}
func validateTransfer(d *Transfer) error {
	if d.Version != 1 {
		return fail("VALIDATION_ERROR", "不支持的文件格式版本")
	}
	if len(d.Tasks) > 100000 {
		return fail("VALIDATION_ERROR", "最多导入 100,000 个节点")
	}
	trees := map[string]map[string]Task{}
	ids := map[string]bool{}
	roots := 0
	for _, t := range d.Tasks {
		if ids[t.ID] || t.ID == "" {
			return fail("VALIDATION_ERROR", "节点 ID 重复或为空")
		}
		ids[t.ID] = true
		if trees[t.RootID] == nil {
			trees[t.RootID] = map[string]Task{}
		}
		trees[t.RootID][t.ID] = t
		if t.ParentID == nil {
			roots++
		}
		if t.ParentID != nil && (t.Description != nil || t.Date != nil || t.Color != nil || t.Postpone != nil || t.SeriesID != nil || t.RuleID != nil || t.RemindAt != nil) {
			return fail("VALIDATION_ERROR", "子任务不能保存根日程字段")
		}
	}
	if roots > 10000 {
		return fail("VALIDATION_ERROR", "最多导入 10,000 个根任务")
	}
	for _, m := range trees {
		if err := validateTree(m); err != nil {
			return err
		}
	}
	ruleIDs := map[string]bool{}
	for _, r := range d.Rules {
		if ruleIDs[r.ID] || r.ID == "" {
			return fail("VALIDATION_ERROR", "重复规则 ID 无效")
		}
		ruleIDs[r.ID] = true
		if r.Interval < 1 || r.Interval > 99 || !contains([]string{"daily", "weekly", "monthly"}, r.Frequency) {
			return fail("VALIDATION_ERROR", "重复规则无效")
		}
		m := map[string]Task{}
		for _, t := range r.Template {
			m[t.ID] = t
		}
		if err := validateTree(m); err != nil {
			return err
		}
	}
	for _, t := range d.Tasks {
		if t.RuleID != nil && !ruleIDs[*t.RuleID] {
			return fail("VALIDATION_ERROR", "任务引用不存在的规则")
		}
	}
	return validateExchange(*d)
}
func (e *Engine) PreviewImport(content, format string, date *string, replace bool) (Preview, error) {
	if len(content) > 20*1024*1024 {
		return Preview{}, fail("VALIDATION_ERROR", "文件最多 20 MiB")
	}
	if !contains([]string{"json", "markdown", "text"}, format) {
		return Preview{}, fail("VALIDATION_ERROR", "不支持的导入格式")
	}
	var d Transfer
	var err error
	if format == "json" {
		err = json.Unmarshal([]byte(content), &d)
	} else {
		d = Transfer{Version: 1}
		d.Tasks, err = parseList(content, format, date)
	}
	if err != nil {
		return Preview{}, err
	}
	if err = validateTransfer(&d); err != nil {
		return Preview{}, err
	}
	if len(d.Tasks) == 0 && format != "json" {
		return Preview{}, fail("VALIDATION_ERROR", "没有可导入的任务")
	}
	warnings := []string{}
	if format != "json" {
		trees := map[string]map[string]Task{}
		for _, t := range d.Tasks {
			if trees[t.RootID] == nil {
				trees[t.RootID] = map[string]Task{}
			}
			trees[t.RootID][t.ID] = t
		}
		d.Tasks = nil
		for _, m := range trees {
			aggregate(m, e.now().UnixMilli())
			for _, t := range sortedTree(m) {
				d.Tasks = append(d.Tasks, t)
			}
		}
		warnings = append(warnings, "完成状态将根据叶节点向上汇总")
	}
	if !replace {
		warnings = append(warnings, "将创建新副本；提醒关闭，活动专注转为停止记录")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	var rev string
	if err = e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&rev); err != nil {
		return Preview{}, err
	}
	if !replace {
		rows, er := e.db.Query("SELECT title,planned_date FROM tasks WHERE parent_id IS NULL AND status<>'skipped'")
		if er != nil {
			return Preview{}, er
		}
		keys := map[string]bool{}
		for rows.Next() {
			var title string
			var date *string
			if er = rows.Scan(&title, &date); er != nil {
				rows.Close()
				return Preview{}, er
			}
			keys[title+"/"+val(date)] = true
		}
		rows.Close()
		similar := 0
		for _, t := range d.Tasks {
			if t.ParentID == nil {
				key := t.Title + "/" + val(t.Date)
				if keys[key] {
					similar++
				}
				keys[key] = true
			}
		}
		if similar > 0 {
			warnings = append(warnings, fmt.Sprintf("发现 %d 条同标题、同计划日期的相似任务；确认后仍创建新副本，不自动去重", similar))
		}
	}
	for key, old := range e.previews {
		if e.now().After(old.Expires) {
			delete(e.previews, key)
		}
	}
	token := ID()
	e.previews[token] = pendingImport{Data: d, Expires: e.now().Add(5 * time.Minute), Revision: rev, Replace: replace}
	p := Preview{Token: token, Tasks: d.Tasks, Nodes: len(d.Tasks), Warnings: warnings, Replacing: replace}
	for _, t := range d.Tasks {
		if t.ParentID == nil {
			p.Roots++
		}
	}
	return p, nil
}
func remapTransfer(d Transfer, now int64) Transfer {
	ids := map[string]string{}
	for _, t := range d.Tasks {
		ids[t.ID] = ID()
	}
	rules := map[string]string{}
	series := map[string]string{}
	for _, r := range d.Rules {
		rules[r.ID] = ID()
		if series[r.SeriesID] == "" {
			series[r.SeriesID] = ID()
		}
	}
	for i, t := range d.Tasks {
		t.ID = ids[t.ID]
		t.RootID = ids[t.RootID]
		if t.ParentID != nil {
			t.ParentID = ptr(ids[*t.ParentID])
		}
		if t.SeriesID != nil {
			t.SeriesID = ptr(series[*t.SeriesID])
			t.RuleID = ptr(rules[*t.RuleID])
		}
		t.RemindAt = nil
		t.ReminderTime = nil
		t.ReminderZone = nil
		t.ReminderGeneration = 0
		t.CreatedAt = now
		t.UpdatedAt = now
		t.Version = 1
		d.Tasks[i] = t
	}
	for i, item := range d.Series {
		if series[item.ID] == "" {
			series[item.ID] = ID()
		}
		item.ID = series[item.ID]
		d.Series[i] = item
	}
	d.Jobs = nil
	for i, r := range d.Rules {
		r.ID = rules[r.ID]
		r.SeriesID = series[r.SeriesID]
		for j, t := range r.Template {
			t.RemindAt = nil
			t.ReminderTime = nil
			t.ReminderZone = nil
			r.Template[j] = t
		}
		d.Rules[i] = r
	}
	for i, c := range d.Countdowns {
		c.ID = ID()
		c.Enabled = false
		d.Countdowns[i] = c
	}
	for i, f := range d.Focus {
		f.ID = ID()
		if f.TaskID != nil {
			if id := ids[*f.TaskID]; id != "" {
				f.TaskID = ptr(id)
			} else {
				f.TaskID = nil
			}
		}
		if f.State == "running" || f.State == "paused" {
			f.State = "stopped"
			f.FinishedAt = ptr(now)
		}
		d.Focus[i] = f
	}
	return d
}
func (e *Engine) CommitImport(token string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.previews[token]
	if !ok || e.now().After(p.Expires) {
		return fail("PREVIEW_EXPIRED", "预览已过期，请重新导入")
	}
	var rev string
	if err := e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&rev); err != nil {
		return err
	}
	if rev != p.Revision {
		return fail("CONFLICT", "数据已变化，请重新预览")
	}
	if p.Replace {
		if _, err := e.backupLocked("protect"); err != nil {
			return err
		}
	}
	d := p.Data
	if !p.Replace {
		d = remapTransfer(d, e.now().UnixMilli())
	}
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if p.Replace {
		for _, table := range []string{"notification_jobs", "focus_sessions", "tasks", "repeat_rules", "repeat_series", "countdowns"} {
			if _, err = tx.Exec("DELETE FROM " + table); err != nil {
				return err
			}
		}
	}
	series := map[string]bool{}
	for _, item := range d.Series {
		if _, err = tx.Exec("INSERT INTO repeat_series(id,stopped_from,version) VALUES(?,?,?)", item.ID, item.StoppedFrom, item.Version); err != nil {
			return err
		}
		series[item.ID] = true
	}
	for _, r := range d.Rules {
		if !series[r.SeriesID] {
			if _, err = tx.Exec("INSERT INTO repeat_series(id) VALUES(?)", r.SeriesID); err != nil {
				return err
			}
			series[r.SeriesID] = true
		}
		if err = putRule(tx, r); err != nil {
			return err
		}
	}
	for _, t := range d.Tasks {
		if err = putTask(tx, t); err != nil {
			return err
		}
		if t.ParentID == nil && len(d.Jobs) == 0 {
			if err = syncReminder(tx, t, e.now().UnixMilli()); err != nil {
				return err
			}
		}
	}

	for _, c := range d.Countdowns {
		if err = saveCountdown(tx, c, e.now().UnixMilli()); err != nil {
			return err
		}
	}
	for _, f := range d.Focus {
		if _, err = tx.Exec("INSERT INTO focus_sessions VALUES(?,?,?,?,?,?,?,?,?)", f.ID, f.TaskID, f.Title, f.Duration, f.Remaining, f.Deadline, f.State, f.StartedAt, f.FinishedAt); err != nil {
			return err
		}
	}
	if p.Replace {
		for _, job := range d.Jobs {
			state := job.State
			if state == "claimed" {
				state = "uncertain"
			}
			if _, err = tx.Exec("INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(source_type,source_id,generation,cycle_key) DO UPDATE SET id=excluded.id,state=excluded.state,fire_at=excluded.fire_at,title=excluded.title", job.ID, job.SourceType, job.SourceID, job.Generation, job.CycleKey, job.FireAt, state, job.Title); err != nil {
				return err
			}
		}
	}
	for _, f := range d.Focus {
		if f.State == "running" {
			_, err = tx.Exec("INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES(?,'focus',?,0,?,?,'pending',?) ON CONFLICT(source_type,source_id,generation,cycle_key) DO NOTHING", ID(), f.ID, fmt.Sprint(f.Deadline), f.Deadline, "专注完成 · "+f.Title)
			if err != nil {
				return err
			}
		}
	}
	if p.Replace {
		for k, v := range d.Settings {
			if _, ok := defaults[k]; ok && !strings.HasPrefix(k, "shortcut") {
				if _, err = tx.Exec("UPDATE settings SET value_json=? WHERE key=?", v, k); err != nil {
					return err
				}
			}
		}
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	delete(e.previews, token)
	e.lastDay = ""
	e.signal()
	return nil
}
func (e *Engine) PreviewBatch(ids []string, action, revision string) (Preview, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !contains([]string{"complete", "pending", "delete"}, action) {
		return Preview{}, fail("VALIDATION_ERROR", "未知批量操作")
	}
	var rev string
	if err := e.db.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&rev); err != nil {
		return Preview{}, err
	}
	if rev != revision {
		return Preview{}, fail("CONFLICT", "列表已更新")
	}
	set := map[string]bool{}
	tasks := []Task{}
	for _, id := range ids {
		if set[id] {
			continue
		}
		set[id] = true
		rows, err := e.db.Query("SELECT "+columns+" FROM tasks WHERE root_id=?", id)
		if err != nil {
			return Preview{}, err
		}
		ts, err := readTasks(rows)
		if err != nil {
			return Preview{}, err
		}
		if len(ts) == 0 {
			return Preview{}, fail("NOT_FOUND", "所选根任务不存在，请重新选择")
		}
		m := map[string]Task{}
		for _, t := range ts {
			m[t.ID] = t
		}
		tasks = append(tasks, sortedTree(m)...)
	}
	if len(tasks) == 0 {
		return Preview{}, fail("VALIDATION_ERROR", "请先选择任务")
	}
	token := ID()
	e.previews[token] = pendingImport{Data: Transfer{Tasks: tasks, AppVersion: action}, Expires: e.now().Add(5 * time.Minute), Revision: rev}
	return Preview{Token: token, Tasks: tasks, Roots: len(set), Nodes: len(tasks), Warnings: []string{"批量操作将影响所有层级子任务；重复任务仅影响本次"}}, nil
}
func (e *Engine) CommitBatch(token string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.previews[token]
	if !ok || e.now().After(p.Expires) {
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
	for _, t := range p.Data.Tasks {
		if p.Data.AppVersion == "delete" {
			if t.ParentID != nil {
				continue
			}
			if t.SeriesID == nil {
				if _, err = tx.Exec("DELETE FROM tasks WHERE id=?", t.ID); err != nil {
					return err
				}
			} else {
				t.Status = "skipped"
				t.CompletedAt = nil
				t.RemindAt = nil
				t.Version++
				if err = putTask(tx, t); err != nil {
					return err
				}
				if _, err = tx.Exec("DELETE FROM tasks WHERE root_id=? AND parent_id IS NOT NULL", t.ID); err != nil {
					return err
				}
			}
			if _, err = tx.Exec("UPDATE notification_jobs SET state='cancelled' WHERE source_type='task' AND source_id=? AND state='pending'", t.ID); err != nil {
				return err
			}
		} else {
			t = setState(t, p.Data.AppVersion == "complete", e.now().UnixMilli())
			t.Version++
			if t.ParentID == nil && t.SeriesID != nil {
				t.Exception = true
			}
			if err = putTask(tx, t); err != nil {
				return err
			}
			if t.ParentID == nil {
				if err = syncReminder(tx, t, e.now().UnixMilli()); err != nil {
					return err
				}
			}
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
func (e *Engine) SaveHolidayData(content string) error {
	if len(content) > 2*1024*1024 {
		return fail("VALIDATION_ERROR", "节假日文件过大")
	}
	var days []Holiday
	if err := json.Unmarshal([]byte(content), &days); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, d := range days {
		if err := validateDate(&d.Date); err != nil {
			return err
		}
		if !contains([]string{"holiday", "workday"}, d.Kind) || strings.TrimSpace(d.Name) == "" || seen[d.Date] {
			return fail("VALIDATION_ERROR", "节假日数据包含无效或重复日期")
		}
		seen[d.Date] = true
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, d := range days {
		if _, err = tx.Exec("INSERT INTO holiday_days VALUES(?,?,?,?,?) ON CONFLICT(date) DO UPDATE SET name=excluded.name,kind=excluded.kind,dataset_version=excluded.dataset_version,source_label=excluded.source_label", d.Date, d.Name, d.Kind, d.Version, d.Source); err != nil {
			return err
		}
	}
	if err = revisionTx(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err == nil {
		e.signal()
	}
	return err
}
func (e *Engine) Holidays() ([]Holiday, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT date,name,kind,dataset_version,source_label FROM holiday_days ORDER BY date")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Holiday{}
	for rows.Next() {
		var d Holiday
		if err = rows.Scan(&d.Date, &d.Name, &d.Kind, &d.Version, &d.Source); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (e *Engine) SaveWindow(w WindowState) error {
	if !contains([]string{"main", "mini", "overview", "quick"}, w.Key) {
		return fail("VALIDATION_ERROR", "未知窗口")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	b, _ := json.Marshal(w)
	_, err := e.db.Exec("INSERT INTO window_states(window_key,state_json) VALUES(?,?) ON CONFLICT(window_key) DO UPDATE SET state_json=excluded.state_json,version=version+1", w.Key, string(b))
	return err
}

// Native move/resize events update geometry without overwriting a concurrent
// mode, opacity or collapse change from the settings UI.
func (e *Engine) SaveWindowBounds(bounds WindowState) error {
	if !contains([]string{"main", "mini", "overview", "quick"}, bounds.Key) {
		return fail("VALIDATION_ERROR", "未知窗口")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	state := WindowState{Key: bounds.Key, Mode: "normal", Opacity: 1}
	var raw string
	err := e.db.QueryRow("SELECT state_json FROM window_states WHERE window_key=?", bounds.Key).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		if err = json.Unmarshal([]byte(raw), &state); err != nil {
			return err
		}
	}
	state.X, state.Y, state.Width, state.Height = bounds.X, bounds.Y, bounds.Width, bounds.Height
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = e.db.Exec("INSERT INTO window_states(window_key,state_json) VALUES(?,?) ON CONFLICT(window_key) DO UPDATE SET state_json=excluded.state_json,version=version+1", bounds.Key, string(b))
	return err
}
func (e *Engine) Windows() ([]WindowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rows, err := e.db.Query("SELECT state_json FROM window_states")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WindowState{}
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, err
		}
		var w WindowState
		if err = json.Unmarshal([]byte(s), &w); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (e *Engine) backupLocked(kind string) (Backup, error) {
	dir := filepath.Join(e.Dir, "backups")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Backup{}, err
	}
	name := fmt.Sprintf("%s-%s-v1.db", kind, e.now().UTC().Format("20060102T150405.000000000"))
	tmp := filepath.Join(dir, name+".tmp")
	if _, err := e.db.Exec("VACUUM INTO ?", tmp); err != nil {
		return Backup{}, err
	}
	if err := checkBackup(tmp); err != nil {
		os.Remove(tmp)
		return Backup{}, err
	}
	if err := os.Rename(tmp, filepath.Join(dir, name)); err != nil {
		return Backup{}, err
	}
	info, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return Backup{}, err
	}
	return Backup{Name: name, Size: info.Size(), CreatedAt: info.ModTime().UnixMilli()}, nil
}
func checkBackup(path string) error {
	db, err := sql.Open("sqlite", sqliteURI(path)+"?mode=ro&_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}
	defer db.Close()
	var result string
	if err = db.QueryRow("PRAGMA quick_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fail("IO_ERROR", "备份完整性校验失败")
	}
	var version int
	if err = db.QueryRow("SELECT max(version) FROM schema_migrations").Scan(&version); err != nil || version != 1 {
		return fail("SCHEMA_ERROR", "不支持的备份版本")
	}
	rows, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return fail("IO_ERROR", "备份包含损坏的关联")
	}
	return rows.Err()
}
func (e *Engine) CreateBackup() (Backup, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.backupLocked("manual")
}
func (e *Engine) Backups() ([]Backup, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.backupsLocked()
}
func (e *Engine) backupsLocked() ([]Backup, error) {
	entries, err := os.ReadDir(filepath.Join(e.Dir, "backups"))
	if os.IsNotExist(err) {
		return []Backup{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []Backup{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		i, err := entry.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, Backup{Name: entry.Name(), Size: i.Size(), CreatedAt: i.ModTime().UnixMilli()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out, nil
}
func (e *Engine) DailyBackup() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	var n int
	if err := e.db.QueryRow("SELECT count(*) FROM tasks").Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	bs, err := e.backupsLocked()
	if err != nil {
		return err
	}
	prefix := "daily-"
	for _, b := range bs {
		if strings.HasPrefix(b.Name, prefix) && today(time.UnixMilli(b.CreatedAt).In(e.location())) == today(e.localNow()) {
			return nil
		}
	}
	if _, err = e.backupLocked("daily"); err != nil {
		return err
	}
	bs, err = e.backupsLocked()
	if err != nil {
		return err
	}
	kept := 0
	for _, b := range bs {
		if strings.HasPrefix(b.Name, "daily-") {
			kept++
			if kept > 7 {
				if err = os.Remove(filepath.Join(e.Dir, "backups", b.Name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (e *Engine) Restore(name string) error {
	if filepath.Base(name) != name {
		return fail("VALIDATION_ERROR", "无效备份名称")
	}
	return e.RestoreFile(filepath.Join(e.DataDir(), "backups", name))
}
func (e *Engine) RestoreFile(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := checkBackup(path); err != nil {
		return err
	}
	if _, err := e.backupLocked("protect"); err != nil {
		return err
	}
	tmp := filepath.Join(e.Dir, "restore.tmp")
	_ = os.Remove(tmp)
	source, err := sql.Open("sqlite", sqliteURI(path)+"?mode=ro")
	if err != nil {
		return err
	}
	_, err = source.Exec("VACUUM INTO ?", tmp)
	closeErr := source.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = checkBackup(tmp); err != nil {
		return err
	}
	if err = e.db.Close(); err != nil {
		return err
	}
	dbpath := filepath.Join(e.Dir, "simpletodo.db")
	old := dbpath + ".previous"
	_ = os.Remove(old)
	if err = os.Rename(dbpath, old); err != nil {
		_ = e.openDB()
		return err
	}
	_ = os.Remove(dbpath + "-wal")
	_ = os.Remove(dbpath + "-shm")
	if err = os.Rename(tmp, dbpath); err == nil {
		err = e.openDB()
	}
	if err != nil {
		_ = os.Remove(dbpath)
		_ = os.Rename(old, dbpath)
		_ = e.openDB()
		return err
	}
	_, err = e.db.Exec("UPDATE app_meta SET value=CAST(value AS INTEGER)+1 WHERE key='revision'; UPDATE notification_jobs SET state='uncertain' WHERE state IN ('claimed','pending') AND fire_at < ?", e.now().UnixMilli())
	e.lastDay = ""
	e.previews = map[string]pendingImport{}
	e.signal()
	return err
}
func (e *Engine) Migrate(dir string) error { return e.MigrateWithCommit(dir, nil) }
func (e *Engine) MigrateWithCommit(dir string, commit func(string) error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if abs == e.Dir {
		return nil
	}
	if _, err = os.Stat(filepath.Join(abs, "simpletodo.db")); err == nil {
		return fail("CONFLICT", "目标目录已有数据库，不能覆盖")
	}
	if err = os.MkdirAll(abs, 0700); err != nil {
		return err
	}
	nextLock := flock.New(filepath.Join(abs, ".simpletodo.lock"))
	locked, err := nextLock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return fail("DB_BUSY", "目标目录正由另一进程使用")
	}
	success := false
	defer func() {
		if !success {
			_ = nextLock.Unlock()
		}
	}()
	if _, err = e.backupLocked("protect"); err != nil {
		return err
	}
	target := filepath.Join(abs, "simpletodo.db")
	if _, err = e.db.Exec("VACUUM INTO ?", target); err != nil {
		return err
	}
	if err = checkBackup(target); err != nil {
		return err
	}
	backups, err := e.backupsLocked()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(abs, "backups"), 0700); err != nil {
		return err
	}
	for _, b := range backups {
		raw, er := os.ReadFile(filepath.Join(e.Dir, "backups", b.Name))
		if er != nil {
			return er
		}
		if er = os.WriteFile(filepath.Join(abs, "backups", b.Name), raw, 0600); er != nil {
			return er
		}
	}
	old := e.Dir
	if commit != nil {
		if err = commit(abs); err != nil {
			_ = os.Remove(target)
			return err
		}
	}
	if err = e.db.Close(); err != nil {
		if commit != nil {
			_ = commit(old)
		}
		return err
	}
	e.Dir = abs
	if err = e.openDB(); err != nil {
		e.Dir = old
		_ = e.openDB()
		if commit != nil {
			_ = commit(old)
		}
		return err
	}
	oldLock := e.lock
	e.lock = nextLock
	success = true
	if oldLock != nil {
		_ = oldLock.Unlock()
	}
	e.lastDay = ""
	e.previews = map[string]pendingImport{}
	e.signal()
	return nil
}

// RecoverDatabase never opens a broken database as an empty store. Existing files
// are retained together so that an interrupted WAL can still be recovered manually.
func RecoverDatabase(dir, sourcePath string) error {
	if err := checkBackup(sourcePath); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	lock := flock.New(filepath.Join(dir, ".simpletodo.lock"))
	ok, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !ok {
		return fail("DB_BUSY", "数据目录正由另一进程使用，请先退出该进程")
	}
	defer lock.Unlock()
	tmp := filepath.Join(dir, "recover-"+ID()+".tmp")
	source, err := sql.Open("sqlite", sqliteURI(sourcePath)+"?mode=ro")
	if err != nil {
		return err
	}
	_, err = source.Exec("VACUUM INTO ?", tmp)
	closeErr := source.Close()
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	if closeErr != nil {
		return closeErr
	}
	if err = checkBackup(tmp); err != nil {
		return err
	}
	preserve := filepath.Join(dir, "recovery-preserved-"+time.Now().UTC().Format("20060102T150405")+"-"+ID()[:8])
	if err = os.MkdirAll(preserve, 0700); err != nil {
		return err
	}
	moved := []string{}
	rollback := func() {
		for _, name := range moved {
			_ = os.Rename(filepath.Join(preserve, name), filepath.Join(dir, name))
		}
	}
	for _, name := range []string{"simpletodo.db", "simpletodo.db-wal", "simpletodo.db-shm"} {
		old := filepath.Join(dir, name)
		if _, er := os.Stat(old); os.IsNotExist(er) {
			continue
		} else if er != nil {
			rollback()
			return er
		}
		if err = os.Rename(old, filepath.Join(preserve, name)); err != nil {
			rollback()
			return err
		}
		moved = append(moved, name)
	}
	if err = os.Rename(tmp, filepath.Join(dir, "simpletodo.db")); err != nil {
		rollback()
		return err
	}
	return nil
}
