package todo

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

func commandNodeID(c Command) string {
	if c.CommandID == "" {
		return ID()
	}
	h := sha256.Sum256([]byte(c.CommandID))
	return fmt.Sprintf("%x", h[:16])
}

const MaxTaskDepth = 5

// Existing overdeep trees remain editable after upgrade, but no node may
// grow deeper and all newly created/copied/imported nodes must fit five levels.
func validateTree(m map[string]Task, previous ...map[string]Task) error {
	if len(m) == 0 {
		return nil
	}
	if len(m) > 201 {
		return fail("VALIDATION_ERROR", "每棵任务树最多 200 个子任务")
	}
	roots := 0
	var root Task
	for _, t := range m {
		if t.ParentID == nil {
			roots++
			root = t
		}
		if strings.TrimSpace(t.Title) == "" || utf8.RuneCountInString(t.Title) > 500 || strings.ContainsAny(t.Title, "\r\n") {
			return fail("VALIDATION_ERROR", "标题须为 1—500 字的单行文本")
		}
		if utf8.RuneCountInString(val(t.Description)) > 20000 {
			return fail("VALIDATION_ERROR", "描述最多 20,000 字")
		}
	}
	if roots != 1 || root.RootID != root.ID {
		return fail("VALIDATION_ERROR", "任务树必须有且只有一个根节点")
	}
	if err := validateDate(root.Date); err != nil {
		return err
	}
	for _, t := range m {
		if !contains([]string{"pending", "completed", "skipped"}, t.Status) {
			return fail("VALIDATION_ERROR", "无效完成状态")
		}
		if (t.Status == "completed") != (t.CompletedAt != nil) {
			return fail("VALIDATION_ERROR", "完成状态与时间不一致")
		}
		if t.ParentID != nil && (t.Description != nil || t.Date != nil || t.OriginalDate != nil || t.Color != nil || t.Postpone != nil || t.SeriesID != nil || t.RuleID != nil || t.OccurrenceDate != nil || t.RemindAt != nil || t.ReminderTime != nil || t.ReminderZone != nil || t.Exception || t.ReminderGeneration != 0 || t.Status == "skipped") {
			return fail("VALIDATION_ERROR", "子任务包含独立日程字段或无效状态")
		}
	}
	for _, t := range m {
		if t.RootID != root.ID {
			return fail("VALIDATION_ERROR", "不允许跨根任务移动")
		}
		limit := MaxTaskDepth
		if len(previous) > 0 {
			limit = max(limit, taskDepth(previous[0], t.ID))
		}
		seen := map[string]bool{}
		cur := t
		for cur.ParentID != nil {
			if seen[cur.ID] {
				return fail("VALIDATION_ERROR", "不能将任务移入自身或后代")
			}
			seen[cur.ID] = true
			if len(seen) >= limit {
				return fail("VALIDATION_ERROR", "任务最多支持 5 层（包含根任务）")
			}
			parent, ok := m[*cur.ParentID]
			if !ok {
				return fail("VALIDATION_ERROR", "父节点不存在")
			}
			cur = parent
		}
		if cur.ID != root.ID {
			return fail("VALIDATION_ERROR", "节点归属不正确")
		}
	}
	return nil
}

// Bounded traversal also tolerates malformed historical input without hanging.
func taskDepth(m map[string]Task, id string) int {
	seen := map[string]bool{}
	for id != "" {
		t, ok := m[id]
		if !ok || seen[id] {
			return 0
		}
		seen[id] = true
		id = val(t.ParentID)
	}
	return len(seen)
}

func descendants(m map[string]Task, id string) map[string]bool {
	out := map[string]bool{id: true}
	for changed := true; changed; {
		changed = false
		for k, t := range m {
			if t.ParentID != nil && out[*t.ParentID] && !out[k] {
				out[k] = true
				changed = true
			}
		}
	}
	return out
}
func setState(t Task, completed bool, now int64) Task {
	if completed {
		if t.Status != "completed" {
			t.CompletedAt = ptr(now)
		}
		t.Status = "completed"
	} else {
		t.Status = "pending"
		t.CompletedAt = nil
	}
	return t
}
func aggregate(m map[string]Task, now int64) {
	depth := func(t Task) int {
		n := 0
		for t.ParentID != nil && n <= 201 {
			p, ok := m[*t.ParentID]
			if !ok {
				break
			}
			t = p
			n++
		}
		return n
	}
	list := []Task{}
	for _, t := range m {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return depth(list[i]) > depth(list[j]) })
	for _, t := range list {
		if t.Status == "skipped" {
			continue
		}
		count, done := 0, 0
		for _, c := range m {
			if c.ParentID != nil && *c.ParentID == t.ID {
				count++
				if c.Status == "completed" {
					done++
				}
			}
		}
		if count > 0 {
			m[t.ID] = setState(m[t.ID], count == done, now)
		}
	}
}
func markException(m map[string]Task, rootID string) {
	r, ok := m[rootID]
	if ok && r.SeriesID != nil {
		r.Exception = true
		m[rootID] = r
	}
}
func (e *Engine) Create(c Command) ([]Task, error) {
	if c.ParentID == nil {
		c.RootID = commandNodeID(c)
		c.ID = c.RootID
		return e.mutate(c, "create", func(m map[string]Task, now int64) error {
			if len(m) > 0 {
				return fail("CONFLICT", "任务已存在")
			}
			p := c.Postpone
			if p == "" {
				p = "inherit"
			}
			color := c.Color
			if color == "" {
				color = "none"
			}
			t := Task{ID: c.ID, RootID: c.ID, Title: strings.TrimSpace(c.Title), Description: ptr(c.Description), Date: c.Date, Status: "pending", Color: ptr(color), Postpone: ptr(p), Order: int(now), RemindAt: c.RemindAt, ReminderTime: c.ReminderTime}
			if err := validateRoot(&t, e.localNow()); err != nil {
				return err
			}
			m[t.ID] = t
			return nil
		})
	}
	c.ID = commandNodeID(c)
	return e.mutate(c, "create-node", func(m map[string]Task, now int64) error {
		p, ok := m[val(c.ParentID)]
		if !ok {
			return fail("NOT_FOUND", "父任务不存在")
		}
		if p.Status == "skipped" {
			return fail("VALIDATION_ERROR", "已删除实例不能添加子任务")
		}
		order := 0
		for _, t := range m {
			if val(t.ParentID) == p.ID && t.Order >= order {
				order = t.Order + 1
			}
		}
		m[c.ID] = Task{ID: c.ID, RootID: c.RootID, ParentID: c.ParentID, Title: strings.TrimSpace(c.Title), Status: "pending", Order: order}
		aggregate(m, now)
		markException(m, c.RootID)
		return nil
	})
}
func validateRoot(t *Task, now time.Time) error {
	if err := validateDate(t.Date); err != nil {
		return err
	}
	if t.Color == nil {
		t.Color = ptr("none")
	}
	if !contains([]string{"none", "red", "orange", "yellow", "green", "blue", "purple"}, *t.Color) {
		return fail("VALIDATION_ERROR", "无效颜色")
	}
	if t.Postpone == nil {
		t.Postpone = ptr("inherit")
	}
	if !contains([]string{"inherit", "on", "off"}, *t.Postpone) {
		return fail("VALIDATION_ERROR", "无效延期策略")
	}
	if t.SeriesID != nil && *t.Postpone == "on" {
		return fail("VALIDATION_ERROR", "重复任务不能自动延期")
	}
	if t.ReminderTime != nil {
		if _, err := time.Parse("15:04", *t.ReminderTime); err != nil {
			return fail("VALIDATION_ERROR", "提醒时分无效")
		}
	}
	if t.RemindAt != nil {
		if t.Date == nil {
			return fail("VALIDATION_ERROR", "请先为任务安排日期")
		}
		if *t.RemindAt <= now.UnixMilli() {
			return fail("VALIDATION_ERROR", "请选择未来的提醒时间")
		}
		t.ReminderZone = ptr(now.Location().String())
		if t.ReminderTime == nil {
			t.ReminderTime = ptr(time.UnixMilli(*t.RemindAt).In(now.Location()).Format("15:04"))
		}
	}
	return nil
}
func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}
func (e *Engine) Update(c Command) ([]Task, error) {
	return e.mutate(c, "update", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务已删除")
		}
		t.Title = strings.TrimSpace(c.Title)
		if t.ParentID == nil {
			t.Description = ptr(c.Description)
			t.Color = ptr(c.Color)
			t.Date = c.Date
			t.Postpone = ptr(c.Postpone)
			if t.SeriesID != nil && val(t.Date) != val(t.OccurrenceDate) {
				return fail("VALIDATION_ERROR", "重复任务不能直接改期")
			}
			if val(t.RemindAt) != val(c.RemindAt) {
				t.ReminderGeneration++
			}
			t.RemindAt = c.RemindAt
			t.ReminderTime = c.ReminderTime
			if t.RemindAt == nil {
				t.ReminderTime = nil
				t.ReminderZone = nil
			}
			if t.RemindAt != nil && t.Date == nil {
				return fail("VALIDATION_ERROR", "提醒需要计划日期")
			}
			if t.RemindAt != nil && *t.RemindAt <= now && val(m[c.ID].RemindAt) == *t.RemindAt {
				check := t
				check.RemindAt = nil
				if err := validateRoot(&check, e.localNow()); err != nil {
					return err
				}
			} else if err := validateRoot(&t, e.localNow()); err != nil {
				return err
			}
		}
		m[t.ID] = t
		markException(m, c.RootID)
		return nil
	})
}
func (e *Engine) Complete(c Command) ([]Task, error) {
	return e.mutate(c, "complete", func(m map[string]Task, now int64) error {
		if _, ok := m[c.ID]; !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		for id := range descendants(m, c.ID) {
			m[id] = setState(m[id], c.Completed, now)
		}
		aggregate(m, now)
		markException(m, c.RootID)
		return nil
	})
}
func (e *Engine) Delete(c Command) ([]Task, error) {
	return e.mutate(c, "delete", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		if t.ParentID == nil && t.SeriesID != nil {
			t.Status = "skipped"
			t.CompletedAt = nil
			t.RemindAt = nil
			t.ReminderGeneration++
			m[t.ID] = t
			for id := range descendants(m, t.ID) {
				if id != t.ID {
					delete(m, id)
				}
			}
		} else {
			for id := range descendants(m, t.ID) {
				delete(m, id)
			}
		}
		aggregate(m, now)
		markException(m, c.RootID)
		return nil
	})
}
func (e *Engine) Move(c Command) ([]Task, error) {
	return e.mutate(c, "move", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		if t.ParentID != nil {
			return fail("VALIDATION_ERROR", "请移动根任务日期，或使用层级调整")
		}
		if t.SeriesID != nil {
			return fail("VALIDATION_ERROR", "重复实例不能直接改期，可复制为普通任务")
		}
		if err := validateDate(c.Date); err != nil {
			return err
		}
		t.Date = c.Date
		t.Order = int(now)
		t.ReminderGeneration++
		if t.RemindAt != nil {
			if c.Date == nil {
				t.RemindAt = nil
				t.ReminderTime = nil
				t.ReminderZone = nil
			} else {
				at, err := wallTime(*c.Date, val(t.ReminderTime), e.location())
				if err != nil || at.UnixMilli() <= now {
					return fail("VALIDATION_ERROR", "改期后的提醒已过期，请先取消或调整提醒")
				}
				t.RemindAt = ptr(at.UnixMilli())
			}
		}
		m[t.ID] = t
		return nil
	})
}
func (e *Engine) Reparent(c Command) ([]Task, error) {
	return e.mutate(c, "reparent", func(m map[string]Task, now int64) error {
		t, ok := m[c.ID]
		if !ok {
			return fail("NOT_FOUND", "任务不存在")
		}
		if t.ParentID == nil || c.ParentID == nil {
			return fail("VALIDATION_ERROR", "根任务不能变成子任务，子任务不能直接提升为根任务")
		}
		if _, ok = m[*c.ParentID]; !ok {
			return fail("VALIDATION_ERROR", "父任务不属于当前根任务")
		}
		if descendants(m, t.ID)[*c.ParentID] {
			return fail("VALIDATION_ERROR", "不能移动到自身或后代")
		}
		t.ParentID = c.ParentID
		m[t.ID] = t
		reorder(m, t.ID, c.BeforeID)
		aggregate(m, now)
		markException(m, c.RootID)
		return nil
	})
}
func reorder(m map[string]Task, id, before string) {
	t := m[id]
	ss := []Task{}
	for _, x := range m {
		if x.ID != id && x.ParentID != nil && val(x.ParentID) == val(t.ParentID) {
			ss = append(ss, x)
		}
	}
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Order == ss[j].Order {
			return ss[i].ID < ss[j].ID
		}
		return ss[i].Order < ss[j].Order
	})
	idx := len(ss)
	for i, x := range ss {
		if x.ID == before {
			idx = i
			break
		}
	}
	ss = append(ss, Task{})
	copy(ss[idx+1:], ss[idx:])
	ss[idx] = t
	for i, x := range ss {
		x.Order = i
		m[x.ID] = x
	}
}
func (e *Engine) ReorderRoots(ids []string, revision string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var r string
	if err = tx.QueryRow("SELECT value FROM app_meta WHERE key='revision'").Scan(&r); err != nil {
		return err
	}
	if r != revision {
		return fail("CONFLICT", "列表已变化，请刷新")
	}
	seen := make(map[string]bool, len(ids))
	for i, id := range ids {
		if seen[id] {
			return fail("VALIDATION_ERROR", "任务 ID 不能重复")
		}
		seen[id] = true
		var parent *string
		if err = tx.QueryRow("SELECT parent_id FROM tasks WHERE id=?", id).Scan(&parent); err != nil {
			return err
		}
		if parent != nil {
			return fail("VALIDATION_ERROR", "仅支持根任务重排")
		}
		if _, err = tx.Exec("UPDATE tasks SET sort_order=?,version=version+1,updated_at=? WHERE id=?", i, e.now().UnixMilli(), id); err != nil {
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

func (e *Engine) FindTask(id string) (Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	t, err := scanTask(e.db.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", id))
	if err != nil {
		return Task{}, fail("NOT_FOUND", "提醒对应的任务已经删除")
	}
	if t.ParentID != nil {
		t, err = scanTask(e.db.QueryRow("SELECT "+columns+" FROM tasks WHERE id=?", t.RootID))
	}
	return t, err
}
func (e *Engine) Duplicate(c Command, date *string) error {
	if err := validateDate(date); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	source, err := treeTx(tx, c.RootID)
	if err != nil {
		return err
	}
	root, ok := source[c.RootID]
	if !ok {
		return fail("NOT_FOUND", "任务不存在")
	}
	if fmt.Sprint(root.Version) != c.ExpectedVersion {
		return fail("CONFLICT", "任务已经变化")
	}
	template := []Task{}
	for _, t := range source {
		template = append(template, t)
	}
	copy := cloneTree(template, val(date), "", "", e.now().UnixMilli(), false, e.location())
	if err = validateTree(copy); err != nil {
		return err
	}
	for _, t := range copy {
		if err = putTask(tx, t); err != nil {
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

func (e *Engine) ReadTree(rootID string) ([]Task, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	m, err := treeTx(tx, rootID)
	if err != nil {
		return nil, err
	}
	root, ok := m[rootID]
	if !ok || root.ParentID != nil || root.Status == "skipped" {
		return nil, fail("NOT_FOUND", "根任务已经删除，请关闭面板并刷新列表")
	}
	return sortedTree(m), nil
}
