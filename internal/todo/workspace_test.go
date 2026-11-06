package todo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupRetentionAppliesToEveryKindAndRestorePinsSource(t *testing.T) {
	e := testEngine(t)
	mustRoot(t, e, "retain", nil)
	names := []string{}
	for i := 0; i < 6; i++ {
		b, err := e.CreateBackup()
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, b.Name)
	}
	bs, err := e.Backups()
	if err != nil || len(bs) != 3 {
		t.Fatal(bs, err)
	}
	for _, b := range bs {
		for _, old := range names[:3] {
			if b.Name == old {
				t.Fatal("old manual backup survived")
			}
		}
	}
	source := bs[len(bs)-1]
	if err = e.Restore(source.Name); err != nil {
		t.Fatal("oldest restore source was aged before copy", err)
	}
	bs, _ = e.Backups()
	if len(bs) != 3 {
		t.Fatal(len(bs))
	}
	if _, err = os.Stat(filepath.Join(e.Dir, "backups", source.Name)); !os.IsNotExist(err) {
		t.Fatal("old restore source should now be aged", err)
	}
	unknown := filepath.Join(e.Dir, "backups", "user-copy.db")
	os.WriteFile(unknown, []byte("user data"), 0600)
	if err = e.DailyBackup(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(unknown); err != nil {
		t.Fatal("user-named file removed", err)
	}
}
func TestContentTimestampsDoNotChangeForCompletionOrOtherNodes(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "root", nil)
	c := child(t, e, r, "child")
	before := getTree(t, e, r.ID)
	e.now = func() time.Time { return time.UnixMilli(before[r.ID].UpdatedAt).Add(time.Hour) }
	u := cmd(t, e, c)
	u.Title = "child changed"
	if _, err := e.Update(u); err != nil {
		t.Fatal(err)
	}
	after := getTree(t, e, r.ID)
	if after[r.ID].UpdatedAt != before[r.ID].UpdatedAt || after[c.ID].UpdatedAt == before[c.ID].UpdatedAt {
		t.Fatal(before, after)
	}
	c = after[c.ID]
	u = cmd(t, e, c)
	u.Completed = true
	if _, err := e.Complete(u); err != nil {
		t.Fatal(err)
	}
	after = getTree(t, e, r.ID)
	if after[c.ID].UpdatedAt != c.UpdatedAt || after[c.ID].CompletedAt == nil {
		t.Fatal(after[c.ID])
	}
	u = cmd(t, e, after[c.ID])
	u.Title = after[c.ID].Title
	if _, err := e.Update(u); err != nil {
		t.Fatal(err)
	}
	if getTree(t, e, r.ID)[c.ID].UpdatedAt != c.UpdatedAt {
		t.Fatal("no-op changed content time")
	}
}
func TestDailySpanSnapshotsTagsAndBounds(t *testing.T) {
	e := testEngine(t)
	now := e.localNow()
	date := today(now)
	a := now.UnixMilli()
	end := now.AddDate(0, 0, 1).UnixMilli()
	out, err := e.Create(Command{CommandID: ID(), Title: "daily span", Date: &date, RepeatRule: &Rule{Frequency: "daily", Interval: 1}, Span: &TaskSpan{Start: a, End: &end}})
	if err != nil {
		t.Fatal(err)
	}
	r := out[0]
	leaf := child(t, e, r, "daily child")
	if _, err = e.SetTaskTags(cmd(t, e, leaf), []string{"important", "urgent"}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.SetTaskSpan(cmd(t, e, r), a, &end); err != nil {
		t.Fatal(err)
	}
	u := cmd(t, e, r)
	u.Completed = true
	if _, err = e.Complete(u); err != nil {
		t.Fatal(err)
	}
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	q, err := e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || len(q.Tasks) != 2 {
		t.Fatal(q, err)
	}
	for _, n := range q.Tasks {
		if n.Status != "pending" || n.CompletedAt != nil {
			t.Fatal("completion leaked", n)
		}
		if n.ParentID != nil && len(n.TagIDs) != 2 {
			t.Fatal("tags not inherited", n)
		}
	}
	afterEnd := now.AddDate(0, 0, 2).Format("2006-01-02")
	q, err = e.Query(Query{View: "day", Date: afterEnd})
	if err != nil || len(q.Tasks) != 0 {
		t.Fatal(q, err)
	}
}
func TestFiniteDailySpanAutoExtendsEndWithoutMovingHistory(t *testing.T) {
	e := testEngine(t)
	base := e.localNow()
	date := today(base)
	end := base.UnixMilli()
	rows, err := e.Create(Command{CommandID: ID(), Title: "unfinished", Date: &date, Postpone: "inherit", RepeatRule: &Rule{Frequency: "daily", Interval: 1, Until: &date}, Span: &TaskSpan{Start: base.UnixMilli(), End: &end}})
	if err != nil {
		t.Fatal(err)
	}
	r := rows[0]
	e.now = func() time.Time { return base.AddDate(0, 0, 1) }
	if err = e.Tick(); err != nil {
		t.Fatal(err)
	}
	old := getTree(t, e, r.ID)[r.ID]
	if val(old.Date) != date || old.StartAt != r.StartAt || old.EndAt == nil || *old.EndAt <= end {
		t.Fatal(old)
	}
	q, err := e.Query(Query{View: "day", Date: today(e.localNow())})
	if err != nil || len(q.Tasks) != 1 {
		t.Fatal(q, err)
	}
}
func TestNotesCASAndTagOrderPersistInFullBackup(t *testing.T) {
	e := testEngine(t)
	if err := e.SaveTags([]Tag{{ID: "urgent", Name: "紧急"}, {ID: "important", Name: "重要"}}); err != nil {
		t.Fatal(err)
	}
	n, err := e.SaveNote(Note{Title: "复盘", Content: "# hello\n\n- note"})
	if err != nil {
		t.Fatal(err)
	}
	old := n
	e.now = func() time.Time { return time.UnixMilli(n.UpdatedAt).Add(time.Hour) }
	same, err := e.SaveNote(n)
	if err != nil || same.UpdatedAt != n.UpdatedAt {
		t.Fatal(same, err)
	}
	if _, err = e.SaveNote(old); err == nil {
		t.Fatal("stale note saved")
	}
	same.Content = "changed"
	updated, err := e.SaveNote(same)
	if err != nil || updated.UpdatedAt == same.UpdatedAt {
		t.Fatal(updated, err)
	}
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	if err = e.DeleteNote(updated.ID, updated.Version); err != nil {
		t.Fatal(err)
	}
	if err = e.Restore(b.Name); err != nil {
		t.Fatal(err)
	}
	notes, err := e.Notes("changed")
	if err != nil || len(notes) != 1 {
		t.Fatal(notes, err)
	}
	tags, err := e.Tags()
	if err != nil || len(tags) != 2 || tags[0].ID != "urgent" {
		t.Fatal(tags, err)
	}
	if _, err = e.SaveNote(Note{Content: "missing title"}); err == nil {
		t.Fatal("empty title saved")
	}
}

func TestPeriodUpgradePreservesHistoryAndExplicitPolicy(t *testing.T) {
	e := testEngine(t)
	day := today(e.localNow())
	r := mustRoot(t, e, "legacy daily", &day)
	u := cmd(t, e, r)
	if err := e.Repeat(RepeatRequest{Command: u, Rule: Rule{Frequency: "daily", Interval: 1}}); err != nil {
		t.Fatal(err)
	}
	// Simulate a pre-schema-3 JSON template that has no start/end metadata.
	rules, _ := e.Rules()
	e.mu.Lock()
	tx, _ := e.db.Begin()
	for _, rule := range rules {
		for i := range rule.Template {
			rule.Template[i].StartAt = 0
		}
		if err := putRule(tx, rule); err != nil {
			t.Fatal(err)
		}
	}
	tx.Commit()
	e.mu.Unlock()
	off := mustRoot(t, e, "keep disabled", &day)
	c := cmd(t, e, off)
	c.Title = off.Title
	c.Date = off.Date
	c.Color = "none"
	c.Postpone = "off"
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	done := mustRoot(t, e, "completed history", &day)
	c = cmd(t, e, done)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	before := getTree(t, e, done.ID)[done.ID]
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	after := getTree(t, e, done.ID)[done.ID]
	if after.Status != "completed" || val(after.CompletedAt) != val(before.CompletedAt) || after.CreatedAt != before.CreatedAt || after.UpdatedAt != before.UpdatedAt {
		t.Fatal(after, before)
	}
	if val(getTree(t, e, off.ID)[off.ID].Postpone) != "off" {
		t.Fatal("explicit disabled policy lost")
	}
	rules, _ = e.Rules()
	for _, rule := range rules {
		if rule.Frequency != "daily" || rule.Interval != 1 {
			t.Fatal(rule)
		}
		for _, task := range rule.Template {
			if task.StartAt == 0 {
				t.Fatal("uninitialized legacy template", task)
			}
		}
	}
}
func TestChildSpanBoundsAndCompletionApplyToActiveChildren(t *testing.T) {
	e := testEngine(t)
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	day := today(e.localNow())
	out, err := e.Create(Command{CommandID: ID(), Title: "parent", Date: &day, RepeatRule: &Rule{Frequency: "daily", Interval: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r := out[0]
	a := child(t, e, r, "active")
	b := child(t, e, r, "tomorrow")
	future := e.localNow().AddDate(0, 0, 1).UnixMilli()
	if _, err = e.SetTaskSpan(cmd(t, e, b), future, nil); err != nil {
		t.Fatal(err)
	}
	c := cmd(t, e, a)
	c.Completed = true
	if _, err = e.Complete(c); err != nil {
		t.Fatal(err)
	}
	if getTree(t, e, r.ID)[r.ID].Status != "pending" {
		t.Fatal("future pending child must keep the period active")
	}
	q, err := e.Query(Query{View: "day", Date: day})
	if err != nil || len(q.Tasks) != 2 {
		t.Fatal(q, err)
	}
	q, err = e.Query(Query{View: "day", Date: e.localNow().AddDate(0, 0, 1).Format("2006-01-02")})
	if err != nil || len(q.Tasks) != 3 {
		t.Fatal(q, err)
	}
	for _, n := range q.Tasks {
		if (n.Title == "active") != (n.Status == "completed") {
			t.Fatal("later day must retain the completed child", n)
		}
	}
}

func TestDeletedTagIsRemovedFromSnapshotsTemplatesAndFilters(t *testing.T) {
	e := testEngine(t)
	day := today(e.localNow())
	out, err := e.Create(Command{CommandID: ID(), Title: "tag cleanup", Date: &day, RepeatRule: &Rule{Frequency: "daily", Interval: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r := out[0]
	a := child(t, e, r, "tagged child")
	if _, err = e.SetTaskTags(cmd(t, e, a), []string{"important", "urgent"}); err != nil {
		t.Fatal(err)
	}
	if err = e.SaveSettings(map[string]string{"dayFilterTags": "[\"important\"]"}); err != nil {
		t.Fatal(err)
	}
	if err = e.SaveTags([]Tag{{ID: "urgent", Name: "紧急"}}); err != nil {
		t.Fatal(err)
	}
	n := getTree(t, e, r.ID)[a.ID]
	if len(n.TagIDs) != 1 || n.TagIDs[0] != "urgent" {
		t.Fatal(n)
	}
	rules, _ := e.Rules()
	for _, rule := range rules {
		for _, task := range rule.Template {
			for _, id := range task.TagIDs {
				if id == "important" {
					t.Fatal("deleted tag returned through future template")
				}
			}
		}
	}
	settings, _ := e.Settings()
	if settings["dayFilterTags"] != "[]" {
		t.Fatal(settings["dayFilterTags"])
	}
	if _, err = e.SetTaskTags(cmd(t, e, n), []string{"urgent"}); err != nil {
		t.Fatal(err)
	}
}
func TestStoppingPeriodPreservesCompletedHistoryAndPreventsRegeneration(t *testing.T) {
	e := testEngine(t)
	day := today(e.localNow())
	out, err := e.Create(Command{CommandID: ID(), Title: "stop", Date: &day, RepeatRule: &Rule{Frequency: "daily", Interval: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r := out[0]
	tomorrow := e.localNow().AddDate(0, 0, 1).Format("2006-01-02")
	q, err := e.Query(Query{View: "day", Date: tomorrow})
	if err != nil {
		t.Fatal(err)
	}
	future := q.Tasks[0]
	c := cmd(t, e, r)
	c.StopPeriod = true
	if _, err = e.Delete(c); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || len(q.Tasks) != 0 {
		t.Fatal(q, err)
	}
	if _, err = e.FindTask(future.ID); err == nil {
		t.Fatal("future pending snapshot survived deletion")
	}
}

func TestPeriodNestAndPromotePersistFutureHierarchy(t *testing.T) {
	e := testEngine(t)
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	day := today(e.localNow())
	makeRoot := func(title string) Task {
		out, err := e.Create(Command{CommandID: ID(), Title: title, Date: &day, RepeatRule: &Rule{Frequency: "daily", Interval: 1}})
		if err != nil {
			t.Fatal(err)
		}
		return out[0]
	}
	a, b := makeRoot("target"), makeRoot("move branch")
	leaf := child(t, e, b, "leaf")
	u := cmd(t, e, b)
	u.ParentID = &a.ID
	if _, err := e.Nest(u, e.Revision()); err != nil {
		t.Fatal(err)
	}
	tree := getTree(t, e, a.ID)
	if tree[b.ID].ParentID == nil || tree[leaf.ID].RootID != a.ID {
		t.Fatal(tree)
	}
	next := e.localNow().AddDate(0, 0, 1).Format("2006-01-02")
	q, err := e.Query(Query{View: "day", Date: next})
	if err != nil || len(q.Tasks) != 3 {
		t.Fatal(q, err)
	}
	roots := 0
	for _, n := range q.Tasks {
		if n.ParentID == nil {
			roots++
		}
	}
	if roots != 1 {
		t.Fatal("source daily root returned", q)
	}
	u = cmd(t, e, tree[b.ID])
	if _, err = e.Promote(u, e.Revision()); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: next})
	if err != nil || len(q.Tasks) != 3 {
		t.Fatal(q, err)
	}
	roots = 0
	for _, n := range q.Tasks {
		if n.ParentID == nil {
			roots++
		}
	}
	if roots != 2 {
		t.Fatal("promoted branch did not persist", q)
	}
}
func TestRootReorderingDoesNotChangeContentTimestamp(t *testing.T) {
	e := testEngine(t)
	a, b := mustRoot(t, e, "a", nil), mustRoot(t, e, "b", nil)
	oldA, oldB := a.UpdatedAt, b.UpdatedAt
	e.now = func() time.Time { return time.UnixMilli(oldA).Add(time.Hour) }
	if err := e.ReorderRoots([]string{b.ID, a.ID}, e.Revision()); err != nil {
		t.Fatal(err)
	}
	a, _ = e.FindTask(a.ID)
	b, _ = e.FindTask(b.ID)
	if a.UpdatedAt != oldA || b.UpdatedAt != oldB {
		t.Fatal(a, b)
	}
}
func TestTaskJSONCarriesCustomTagDefinitionsAndTimestamps(t *testing.T) {
	e := testEngine(t)
	if err := e.SaveTags([]Tag{{ID: "custom", Name: "项目"}}); err != nil {
		t.Fatal(err)
	}
	r := mustRoot(t, e, "portable", nil)
	if _, err := e.SetTaskTags(cmd(t, e, r), []string{"custom"}); err != nil {
		t.Fatal(err)
	}
	text, err := e.Export("json", Query{View: "all"}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	target := testEngine(t)
	p, err := target.PreviewImport(text, "json", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = target.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	tags, _ := target.Tags()
	found := false
	for _, tag := range tags {
		if tag.Name == "项目" {
			found = true
		}
	}
	if !found {
		t.Fatal(tags)
	}
	q, _ := target.Query(Query{View: "all"})
	if len(q.Tasks) != 1 || q.Tasks[0].TagIDs[0] != "custom" || q.Tasks[0].CreatedAt != r.CreatedAt {
		t.Fatal(q)
	}
}

func TestNestOnlyConflictsOnAffectedTrees(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "target", ptr("2026-10-20"))
	b := mustRoot(t, e, "source", ptr("2026-10-20"))
	revision := e.Revision()
	mustRoot(t, e, "unrelated", ptr("2026-10-21"))
	c := Command{CommandID: ID(), ID: b.ID, RootID: b.ID, ParentID: &a.ID, ExpectedVersion: fmt.Sprint(b.Version), TargetVersion: fmt.Sprint(a.Version)}
	if _, err := e.Nest(c, revision); err != nil {
		t.Fatal("unrelated changes blocked nesting", err)
	}
	d := mustRoot(t, e, "source2", ptr("2026-10-20"))
	latest := getTree(t, e, a.ID)[a.ID]
	c = Command{CommandID: ID(), ID: d.ID, RootID: d.ID, ParentID: &a.ID, ExpectedVersion: fmt.Sprint(d.Version), TargetVersion: fmt.Sprint(latest.Version)}
	update := editable(t, e, latest)
	update.Title = "edited target"
	if _, err := e.Update(update); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Nest(c, e.Revision()); err == nil {
		t.Fatal("changed target version was accepted")
	}
}

func TestFuturePeriodRangePreservesTaskIdentityAndCanStartEarlier(t *testing.T) {
	e := testEngine(t)
	e.SaveSettings(map[string]string{"autoPostpone": "false"})
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	day := e.localNow().AddDate(0, 0, 2).Format("2006-01-02")
	at, _ := wallTime(day, "12:00", e.location())
	rows, err := e.Create(Command{CommandID: ID(), Title: "future identity", Date: &day, Span: &TaskSpan{Start: at.UnixMilli()}, RepeatRule: &Rule{Frequency: "daily", Interval: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r := rootOf(rows)
	q, err := e.Query(Query{View: "day", Date: day})
	if err != nil || len(q.Tasks) != 1 || q.Tasks[0].ID != r.ID {
		t.Fatal("future creation replaced original identity", q.Tasks, err)
	}
	earlier := at.AddDate(0, 0, -1)
	c := editable(t, e, r)
	if _, err = e.SetTaskSpan(c, earlier.UnixMilli(), nil); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: earlier.Format("2006-01-02")})
	if err != nil || len(q.Tasks) != 1 {
		t.Fatal("moving start earlier did not generate date", q.Tasks, err)
	}
	q, err = e.Query(Query{View: "day", Date: day})
	if err != nil || len(q.Tasks) != 1 || q.Tasks[0].ID != r.ID {
		t.Fatal("time edit replaced future identity", q.Tasks, err)
	}
}
