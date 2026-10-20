package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.now = func() time.Time { return time.Date(2026, 10, 20, 10, 0, 0, 0, time.Local) }
	t.Cleanup(func() { _ = e.Close() })
	return e
}
func rootOf(ts []Task) Task {
	for _, t := range ts {
		if t.ParentID == nil {
			return t
		}
	}
	return Task{}
}
func mustRoot(t *testing.T, e *Engine, title string, date *string) Task {
	t.Helper()
	out, err := e.Create(Command{CommandID: ID(), Title: title, Date: date, Color: "green", Postpone: "inherit"})
	if err != nil {
		t.Fatal(err)
	}
	return rootOf(out)
}
func getTree(t *testing.T, e *Engine, id string) map[string]Task {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	tx, err := e.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	m, err := treeTx(tx, id)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func child(t *testing.T, e *Engine, parent Task, title string) Task {
	t.Helper()
	m := getTree(t, e, parent.RootID)
	out, err := e.Create(Command{CommandID: ID(), Title: title, RootID: parent.RootID, ParentID: ptr(parent.ID), ExpectedVersion: fmt.Sprint(m[parent.RootID].Version)})
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range out {
		if x.Title == title {
			return x
		}
	}
	t.Fatal("missing child")
	return Task{}
}
func cmd(t *testing.T, e *Engine, node Task) Command {
	m := getTree(t, e, node.RootID)
	return Command{CommandID: ID(), ID: node.ID, RootID: node.RootID, ExpectedVersion: fmt.Sprint(m[node.RootID].Version)}
}
func TestSingleTableTreeAndAncestorCompletion(t *testing.T) {
	e := testEngine(t)
	root := mustRoot(t, e, "项目", ptr("2026-10-20"))
	a := child(t, e, root, "分支")
	b := child(t, e, a, "叶节点")
	other := child(t, e, root, "另一分支")
	c := cmd(t, e, b)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	m := getTree(t, e, root.ID)
	if m[a.ID].Status != "completed" || m[root.ID].Status != "pending" || m[other.ID].Status != "pending" {
		t.Fatal("incorrect ancestor propagation")
	}
	c = cmd(t, e, root)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	c = cmd(t, e, b)
	c.Completed = false
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	m = getTree(t, e, root.ID)
	if m[root.ID].Status != "pending" || m[a.ID].Status != "pending" || m[other.ID].Status != "completed" {
		t.Fatal("siblings were modified")
	}
	var n int
	if err := e.db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name='subtasks'").Scan(&n); err != nil || n != 0 {
		t.Fatal("unexpected child table")
	}
}
func TestConflictAndCommandIdempotency(t *testing.T) {
	e := testEngine(t)
	c := Command{CommandID: ID(), Title: "only once", Color: "none", Postpone: "inherit"}
	one, err := e.Create(c)
	if err != nil {
		t.Fatal(err)
	}
	two, err := e.Create(c)
	if err != nil {
		t.Fatal(err)
	}
	if rootOf(one).ID != rootOf(two).ID {
		t.Fatal("duplicate")
	}
	r := rootOf(one)
	ch := child(t, e, r, "new")
	stale := Command{CommandID: ID(), ID: r.ID, RootID: r.ID, ExpectedVersion: rVersion(r), Completed: true}
	if _, err = e.Complete(stale); err == nil || !strings.Contains(err.Error(), "CONFLICT") {
		t.Fatal("stale write accepted", err)
	}
	_ = ch
}
func rVersion(t Task) string { return fmt.Sprint(t.Version) }
func TestCycleAndCrossRootRejected(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "r", nil)
	a := child(t, e, r, "a")
	b := child(t, e, a, "b")
	c := cmd(t, e, a)
	c.ParentID = ptr(b.ID)
	if _, err := e.Reparent(c); err == nil {
		t.Fatal("cycle accepted")
	}
	r2 := mustRoot(t, e, "r2", nil)
	c = cmd(t, e, a)
	c.ParentID = ptr(r2.ID)
	if _, err := e.Reparent(c); err == nil {
		t.Fatal("cross-root accepted")
	}
	m := getTree(t, e, r.ID)
	if val(m[a.ID].ParentID) != r.ID {
		t.Fatal("failed move mutated tree")
	}
}
func TestSubtreeDeleteAndRootCounts(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "inbox", nil)
	a := child(t, e, r, "a")
	_ = child(t, e, a, "b")
	q, err := e.Query(Query{View: "inbox"})
	if err != nil || q.Counts["inbox"] != 1 || len(q.Tasks) != 3 {
		t.Fatal(q, err)
	}
	if _, err = e.Delete(cmd(t, e, a)); err != nil {
		t.Fatal(err)
	}
	if len(getTree(t, e, r.ID)) != 1 {
		t.Fatal("orphan descendants")
	}
	if _, err = e.Delete(cmd(t, e, r)); err != nil {
		t.Fatal(err)
	}
	if len(getTree(t, e, r.ID)) != 0 {
		t.Fatal("root not deleted")
	}
}
func TestMonthlyRuleAndSkippedNotRegenerated(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "月底", ptr("2026-01-31"))
	_ = child(t, e, r, "树模板")
	r = getTree(t, e, r.ID)[r.ID]
	if err := e.Repeat(RepeatRequest{Command: cmd(t, e, r), Rule: Rule{Frequency: "monthly", Interval: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureThrough("2026-03-31"); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "month", Date: "2026-02-01", End: "2026-02-28"})
	if err != nil {
		t.Fatal(err)
	}
	feb := rootOf(q.Tasks)
	if val(feb.Date) != "2026-02-28" || len(q.Tasks) != 2 {
		t.Fatal(q)
	}
	if _, err = e.Delete(cmd(t, e, feb)); err != nil {
		t.Fatal(err)
	}
	if err = e.EnsureThrough("2026-03-31"); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: "2026-02-28"})
	if err != nil || len(q.Tasks) != 0 {
		t.Fatal("skipped regenerated", q, err)
	}
	q, err = e.Query(Query{View: "day", Date: "2026-03-31"})
	if err != nil || len(q.Tasks) != 2 {
		t.Fatal("month-end drift", q, err)
	}
}
func TestPostponeIsIdempotentAndDoesNotMoveInbox(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "old", ptr("2026-10-10"))
	_ = child(t, e, r, "nested")
	_ = mustRoot(t, e, "idea", nil)
	if err := e.SaveSettings(map[string]string{"autoPostpone": "true"}); err != nil {
		t.Fatal(err)
	}
	if err := e.Tick(); err != nil {
		t.Fatal(err)
	}
	m := getTree(t, e, r.ID)
	if val(m[r.ID].Date) != "2026-10-20" || val(m[r.ID].OriginalDate) != "2026-10-10" {
		t.Fatal(m)
	}
	v := m[r.ID].Version
	if err := e.Tick(); err != nil {
		t.Fatal(err)
	}
	if getTree(t, e, r.ID)[r.ID].Version != v {
		t.Fatal("postponed twice")
	}
}
func TestReminderDeliveredOnceAndFocusRestores(t *testing.T) {
	e := testEngine(t)
	now := e.now()
	r := mustRoot(t, e, "reminder", ptr(today(now)))
	c := cmd(t, e, r)
	c.Title = r.Title
	c.Color = "green"
	c.Postpone = "inherit"
	c.Date = r.Date
	c.RemindAt = ptr(now.Add(time.Minute).UnixMilli())
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	sent := 0
	e.Notify = func(n Notice) error { sent++; return nil }
	e.now = func() time.Time { return now.Add(time.Minute) }
	if err := e.Tick(); err != nil {
		t.Fatal(err)
	}
	if err := e.Tick(); err != nil {
		t.Fatal(err)
	}
	if sent != 1 {
		t.Fatal("duplicate notification", sent)
	}
	f, err := e.FocusAction("start", 25, nil)
	if err != nil {
		t.Fatal(err)
	}
	if f.State != "running" {
		t.Fatal(f)
	}
	paused, err := e.FocusAction("pause", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Remaining != 1500 {
		t.Fatal(paused)
	}
	if _, err = e.FocusAction("resume", 0, nil); err != nil {
		t.Fatal(err)
	}
}
func TestMarkdownRoundTripAndAtomicImport(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "根 😊", ptr("2026-10-20"))
	a := child(t, e, r, "一级")
	_ = child(t, e, a, "二级")
	md, err := e.Export("markdown", Query{View: "all"}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.PreviewImport(md, "markdown", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Nodes != 3 || p.Roots != 1 {
		t.Fatal(p)
	}
	if err = e.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "day", Date: "2026-10-20"})
	if err != nil || len(q.Tasks) != 6 {
		t.Fatal(q, err)
	}
	if _, err = e.PreviewImport("- [ ] root\n      - [ ] jumped", "markdown", nil, false); err == nil {
		t.Fatal("invalid indentation accepted")
	}
	var d Transfer
	js, err := e.Export("json", Query{}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(js), &d); err != nil || len(d.Tasks) != 6 {
		t.Fatal(err)
	}
}
func TestBackupRestoreAndIndexes(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "backup", nil)
	_ = child(t, e, r, "child")
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.Delete(cmd(t, e, r)); err != nil {
		t.Fatal(err)
	}
	if err = e.Restore(b.Name); err != nil {
		t.Fatal(err)
	}
	if len(getTree(t, e, r.ID)) != 2 {
		t.Fatal("backup not restored")
	}
	if err = checkBackup(filepath.Join(e.Dir, "backups", b.Name)); err != nil {
		t.Fatal(err)
	}
	var indexes int
	if err = e.db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE type='index' AND sql IS NOT NULL").Scan(&indexes); err != nil || indexes != 13 {
		t.Fatal(indexes, err)
	}
}
func TestCancelledRunExits(t *testing.T) {
	e := testEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan bool)
	go func() { e.Run(ctx); done <- true }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}
}

func TestJSONReplacementPreservesStoppedSeriesAndNotifications(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "repeat", ptr("2026-10-20"))
	if err := e.Repeat(RepeatRequest{Command: cmd(t, e, r), Rule: Rule{Frequency: "daily", Interval: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureThrough("2026-10-22"); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "day", Date: "2026-10-21"})
	if err != nil {
		t.Fatal(err)
	}
	cut := rootOf(q.Tasks)
	if err = e.StopFollowing(cmd(t, e, cut)); err != nil {
		t.Fatal(err)
	}
	if _, err = e.db.Exec("INSERT INTO notification_jobs(id,source_type,source_id,generation,cycle_key,fire_at,state,title) VALUES('saved','task',?,0,'once',?,'submitted','already sent')", r.ID, e.now().UnixMilli()); err != nil {
		t.Fatal(err)
	}
	raw, err := e.Export("json", Query{}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	other := testEngine(t)
	p, err := other.PreviewImport(raw, "json", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = other.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	if err = other.EnsureThrough("2026-10-30"); err != nil {
		t.Fatal(err)
	}
	q, err = other.Query(Query{View: "pending"})
	if err != nil || len(q.Tasks) != 1 {
		t.Fatal("stopped series resumed", q, err)
	}
	var state string
	if err = other.db.QueryRow("SELECT state FROM notification_jobs WHERE id='saved'").Scan(&state); err != nil || state != "submitted" {
		t.Fatal("notification dedup lost", state, err)
	}
}
func TestMigrationRollbackAndDirectoryLock(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "kept", nil)
	if second, err := Open(e.Dir); err == nil {
		second.Close()
		t.Fatal("same directory lock not enforced")
	}
	original := e.Dir
	target := t.TempDir()
	if err := e.MigrateWithCommit(target, func(string) error { return fmt.Errorf("simulated bootstrap failure") }); err == nil {
		t.Fatal("migration should fail")
	}
	if e.Dir != original || len(getTree(t, e, r.ID)) != 1 {
		t.Fatal("original database not preserved")
	}
	target = t.TempDir()
	if err := e.Migrate(target); err != nil {
		t.Fatal(err)
	}
	if e.Dir != target || len(getTree(t, e, r.ID)) != 1 {
		t.Fatal("migration lost data")
	}
	old, err := Open(original)
	if err != nil {
		t.Fatal("old lock not released", err)
	}
	defer old.Close()
}
func TestStructuredClipboardRetainsOffsetsColorsAndTree(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "first", ptr("2026-10-20"))
	_ = child(t, e, a, "nested")
	b := mustRoot(t, e, "second", ptr("2026-10-22"))
	text, err := e.CopyTasks(Query{View: "month", Date: "2026-10-20", End: "2026-10-22"}, []string{a.ID, b.ID})
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.PreviewPaste(text, "2026-11-01")
	if err != nil {
		t.Fatal(err)
	}
	if p.Roots != 2 || p.Nodes != 3 {
		t.Fatal(p)
	}
	dates := map[string]bool{}
	for _, task := range p.Tasks {
		if task.ParentID == nil {
			dates[val(task.Date)] = true
			if val(task.Color) != "green" || task.Status != "pending" {
				t.Fatal(task)
			}
		}
	}
	if !dates["2026-11-01"] || !dates["2026-11-03"] {
		t.Fatal(dates)
	}
	if err = e.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
}

func TestRecoverBrokenDatabaseRetainsOriginalFiles(t *testing.T) {
	source := testEngine(t)
	r := mustRoot(t, source, "recover me", nil)
	backup, err := source.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	broken := filepath.Join(dir, "simpletodo.db")
	original := []byte("not a sqlite database")
	if err = os.WriteFile(broken, original, 0600); err != nil {
		t.Fatal(err)
	}
	if invalid, err := Open(dir); err == nil {
		invalid.Close()
		t.Fatal("broken database accepted")
	}
	raw, err := os.ReadFile(broken)
	if err != nil || string(raw) != string(original) {
		t.Fatal("broken database overwritten")
	}
	if err = RecoverDatabase(dir, filepath.Join(source.Dir, "backups", backup.Name)); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if len(getTree(t, restored, r.ID)) != 1 {
		t.Fatal("recovery lost tasks")
	}
	preserved, err := filepath.Glob(filepath.Join(dir, "recovery-preserved-*", "simpletodo.db"))
	if err != nil || len(preserved) != 1 {
		t.Fatal("original files not retained", preserved, err)
	}
}
