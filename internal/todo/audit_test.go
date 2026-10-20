package todo

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"
)

func editable(t *testing.T, e *Engine, node Task) Command {
	c := cmd(t, e, node)
	r := getTree(t, e, node.RootID)[node.RootID]
	c.Title = node.Title
	c.Date = r.Date
	c.Color = val(r.Color)
	c.Description = val(r.Description)
	c.Postpone = val(r.Postpone)
	c.RemindAt = r.RemindAt
	c.ReminderTime = r.ReminderTime
	return c
}
func TestAtomicRepeatEditAndExceptionPreservation(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "original", ptr("2026-10-20"))
	leaf := child(t, e, r, "leaf")
	c := editable(t, e, r)
	c.Title = "must roll back"
	c.RepeatRule = &Rule{Frequency: "weekly", Interval: 1}
	before := e.Revision()
	if _, err := e.Update(c); err == nil {
		t.Fatal("accepted empty weekdays")
	}
	if getTree(t, e, r.ID)[r.ID].Title != "original" || e.Revision() != before {
		t.Fatal("partial task update escaped rollback")
	}
	c = editable(t, e, r)
	c.RepeatRule = &Rule{Frequency: "daily", Interval: 1}
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureThrough("2026-10-25"); err != nil {
		t.Fatal(err)
	}
	q, _ := e.Query(Query{View: "day", Date: "2026-10-21"})
	exception := rootOf(q.Tasks)
	c = editable(t, e, exception)
	c.Title = "my exception"
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-22"})
	complete := rootOf(q.Tasks)
	c = cmd(t, e, complete)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-23"})
	skipped := rootOf(q.Tasks)
	if _, err := e.Delete(cmd(t, e, skipped)); err != nil {
		t.Fatal(err)
	}
	c = editable(t, e, r)
	c.Title = "new template"
	c.Following = true
	c.RepeatRule = &Rule{Frequency: "daily", Interval: 1}
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureThrough("2026-10-25"); err != nil {
		t.Fatal(err)
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-21"})
	if rootOf(q.Tasks).Title != "my exception" {
		t.Fatal("exception overwritten")
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-22"})
	if rootOf(q.Tasks).Status != "completed" {
		t.Fatal("completed lost")
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-23"})
	if len(q.Tasks) != 0 {
		t.Fatal("skipped regenerated")
	}
	q, _ = e.Query(Query{View: "day", Date: "2026-10-24"})
	if rootOf(q.Tasks).Title != "new template" || len(q.Tasks) != 2 {
		t.Fatal("template tree not copied")
	}
	_ = leaf
}
func TestReminderCompletionAndClockRollbackDoNotReplay(t *testing.T) {
	e := testEngine(t)
	now := e.now()
	sent := 0
	e.Notify = func(Notice) error { sent++; return nil }
	r := mustRoot(t, e, "reminder", ptr("2026-10-20"))
	c := editable(t, e, r)
	c.RemindAt = ptr(now.Add(5 * time.Second).UnixMilli())
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	c = cmd(t, e, r)
	c.Completed = true
	e.Complete(c)
	c = cmd(t, e, r)
	e.Complete(c)
	e.now = func() time.Time { return now.Add(6 * time.Second) }
	if err := e.Tick(); err != nil {
		t.Fatal(err)
	}
	if sent != 1 {
		t.Fatalf("want one notification, got %d", sent)
	}
	c = cmd(t, e, r)
	c.Completed = true
	e.Complete(c)
	c = cmd(t, e, r)
	e.Complete(c)
	e.Tick()
	e.now = func() time.Time { return now }
	e.Tick()
	e.now = func() time.Time { return now.Add(6 * time.Second) }
	e.Tick()
	if sent != 1 {
		t.Fatal("notification replayed")
	}
}
func TestCountdownToggleAndDeadlinePause(t *testing.T) {
	e := testEngine(t)
	now := e.now()
	sent := 0
	e.Notify = func(Notice) error { sent++; return nil }
	if err := e.SaveCountdown(Countdown{Name: "count", Target: now.Add(time.Minute).UnixMilli(), Frequency: "none", Enabled: true, Offset: 10}); err != nil {
		t.Fatal(err)
	}
	e.now = func() time.Time { return now.Add(51 * time.Second) }
	e.Tick()
	cs, _ := e.Countdowns()
	e.ToggleCountdown(cs[0].ID, false)
	e.ToggleCountdown(cs[0].ID, true)
	e.Tick()
	if sent != 1 {
		t.Fatal("same cycle delivered twice", sent)
	}
	f, err := e.FocusAction("start", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	e.now = func() time.Time { return time.UnixMilli(f.Deadline) }
	f, err = e.FocusAction("pause", 0, nil)
	if err != nil || f.State != "completed" {
		t.Fatal("expired focus left paused", f, err)
	}
	e.Tick()
	if sent != 2 {
		t.Fatal("focus completion cancelled", sent)
	}
}
func TestTimezoneAndDSTWallTime(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	gap, err := wallTime("2026-03-08", "02:30", ny)
	if err != nil || gap.Format("15:04") != "03:00" {
		t.Fatal("spring gap", gap, err)
	}
	fold, err := wallTime("2026-11-01", "01:30", ny)
	if err != nil || fold.UTC().Format("15:04") != "05:30" {
		t.Fatal("fold not earliest", fold, err)
	}
	e := testEngine(t)
	e.SetTimezone("Asia/Shanghai")
	r := mustRoot(t, e, "regular", ptr("2026-10-21"))
	at, _ := wallTime("2026-10-21", "09:00", e.location())
	c := editable(t, e, r)
	c.RemindAt = ptr(at.UnixMilli())
	e.Update(c)
	repeat := mustRoot(t, e, "repeat", ptr("2026-10-21"))
	c = editable(t, e, repeat)
	c.RemindAt = ptr(at.UnixMilli())
	c.RepeatRule = &Rule{Frequency: "daily", Interval: 1}
	if _, err = e.Update(c); err != nil {
		t.Fatal(err)
	}
	count, err := e.SetTimezone("America/New_York")
	if err != nil || count != 1 {
		t.Fatal(count, err)
	}
	if val(getTree(t, e, r.ID)[r.ID].RemindAt) != at.UnixMilli() {
		t.Fatal("ordinary UTC changed")
	}
	root := getTree(t, e, repeat.ID)[repeat.ID]
	if time.UnixMilli(*root.RemindAt).In(ny).Format("15:04") != "09:00" {
		t.Fatal("repeated wall clock changed")
	}
}
func TestExportFormatsAndMarkdownRoundTrip(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "=SUM(1,2)<script>", ptr("2026-10-20"))
	a := child(t, e, r, "中文 😊")
	child(t, e, a, "deep")
	c := editable(t, e, r)
	c.Description = "first\nsecond"
	e.Update(c)
	for _, format := range []string{"csv", "html", "markdown", "text", "json"} {
		out, err := e.Export(format, Query{View: "day", Date: "2026-10-20"}, false, nil)
		if err != nil || out == "" {
			t.Fatal(format, err)
		}
		switch format {
		case "csv":
			if !strings.HasPrefix(out, "\uFEFF") || !strings.Contains(out, "'=SUM") {
				t.Fatal("unsafe csv")
			}
		case "html":
			if strings.Contains(out, "<script>") {
				t.Fatal("unescaped html")
			}
		case "markdown":
			p, err := e.PreviewImport(out, "markdown", nil, false)
			if err != nil || p.Nodes != 3 || val(rootOf(p.Tasks).Description) != "first\nsecond" {
				t.Fatal("markdown round trip", p, err)
			}
		}
	}
}
func TestEmptyTaskJSONAndInvalidExchangePreview(t *testing.T) {
	e := testEngine(t)
	if err := e.SaveCountdown(Countdown{Name: "only countdown", Target: e.now().Add(time.Hour).UnixMilli(), Frequency: "daily", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	raw, err := e.Export("json", Query{}, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.PreviewImport(raw, "json", nil, true)
	if err != nil {
		t.Fatal("countdown-only export cannot restore", err)
	}
	if err = e.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	cs, _ := e.Countdowns()
	if len(cs) != 1 {
		t.Fatal("countdown missing")
	}
	var d Transfer
	json.Unmarshal([]byte(raw), &d)
	d.Settings["theme"] = "invalid"
	b, _ := json.Marshal(d)
	if _, err = e.PreviewImport(string(b), "json", nil, true); err == nil {
		t.Fatal("invalid settings preview accepted")
	}
	d.Settings["theme"] = "light"
	d.Countdowns = append(d.Countdowns, d.Countdowns[0])
	b, _ = json.Marshal(d)
	if _, err = e.PreviewImport(string(b), "json", nil, true); err == nil {
		t.Fatal("duplicate countdown preview accepted")
	}
}
func TestDeepTreeLimitAndBatchIntegrity(t *testing.T) {
	e := testEngine(t)
	ts := []Task{{ID: "root", RootID: "root", Title: "root", Status: "pending", Color: ptr("none"), Postpone: ptr("off"), Version: 1}}
	parent := "root"
	for i := 0; i < 200; i++ {
		id := ID()
		ts = append(ts, Task{ID: id, RootID: "root", ParentID: ptr(parent), Title: "nested", Status: "pending", Version: 1})
		parent = id
	}
	raw, _ := json.Marshal(Transfer{Version: 1, Tasks: ts})
	p, err := e.PreviewImport(string(raw), "json", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = e.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	q, _ := e.Query(Query{View: "inbox"})
	r := rootOf(q.Tasks)
	c := cmd(t, e, r)
	c.ParentID = ptr(r.ID)
	c.Title = "overflow"
	if _, err = e.Create(c); err == nil {
		t.Fatal("201st descendant accepted")
	}
	if len(getTree(t, e, r.ID)) != 201 {
		t.Fatal("partial overflow insert")
	}
	p, err = e.PreviewBatch([]string{r.ID, r.ID}, "complete", e.Revision())
	if err != nil || p.Roots != 1 || p.Nodes != 201 {
		t.Fatal(p, err)
	}
	if err = e.CommitBatch(p.Token); err != nil {
		t.Fatal(err)
	}
	for _, task := range getTree(t, e, r.ID) {
		if task.Status != "completed" {
			t.Fatal("batch skipped descendant")
		}
	}
}
func TestCopyEmptySourceDayOffsetAndBackupMetadata(t *testing.T) {
	e := testEngine(t)
	mustRoot(t, e, "day two", ptr("2026-10-21"))
	raw, err := e.CopyTasks(Query{View: "month", Date: "2026-10-20", End: "2026-10-22"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.PreviewPaste(raw, "2026-11-01")
	if err != nil || val(rootOf(p.Tasks).Date) != "2026-11-02" {
		t.Fatal("empty first day offset lost", p, err)
	}
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	info, err := InspectBackup(e.DataDir() + "/backups/" + b.Name)
	if err != nil || info.Roots != 1 || info.Nodes != 1 {
		t.Fatal(info, err)
	}
}

func TestFailedWriteRollbackAndRestoredFocusJob(t *testing.T) {
	e := testEngine(t)
	rev := e.Revision()
	if _, err := e.db.Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Create(Command{CommandID: ID(), Title: "failed", Color: "none", Postpone: "off"}); err == nil {
		t.Fatal("read-only write accepted")
	}
	e.db.Exec("PRAGMA query_only=OFF")
	if e.Revision() != rev {
		t.Fatal("failed write changed revision")
	}
	f := Focus{ID: ID(), Title: "restored", Duration: 60, Remaining: 60, State: "running", StartedAt: e.now().UnixMilli(), Deadline: e.now().Add(time.Minute).UnixMilli()}
	raw, _ := json.Marshal(Transfer{Version: 1, Focus: []Focus{f}})
	p, err := e.PreviewImport(string(raw), "json", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = e.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	var n int
	e.db.QueryRow("SELECT count(*) FROM notification_jobs WHERE source_type='focus' AND state='pending'").Scan(&n)
	if n != 1 {
		t.Fatal("restored focus has no deadline job")
	}
}

func TestDataDirectoryWithURICharacters(t *testing.T) {
	name := "/中文 #计划"
	if runtime.GOOS != "windows" {
		name += "?"
	} // Windows forbids ? in directory names.
	dir := t.TempDir() + name
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	mustRoot(t, e, "persist", nil)
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = InspectBackup(dir + "/backups/" + b.Name); err != nil {
		t.Fatal(err)
	}
	if err = e.Restore(b.Name); err != nil {
		t.Fatal(err)
	}
}

func TestWindowGeometryPreservesDisplayConfiguration(t *testing.T) {
	e := testEngine(t)
	configured := WindowState{Key: "mini", X: 10, Y: 20, Width: 300, Height: 62, Mode: "top", Opacity: .65, Collapsed: true}
	if err := e.SaveWindow(configured); err != nil {
		t.Fatal(err)
	}
	// A native resize event carries only geometry, never a snapshot of settings.
	if err := e.SaveWindowBounds(WindowState{Key: "mini", X: 40, Y: 50, Width: 298, Height: 62}); err != nil {
		t.Fatal(err)
	}
	rows, err := e.Windows()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || !rows[0].Collapsed || rows[0].Mode != "top" || rows[0].Opacity != .65 || rows[0].X != 40 || rows[0].Width != 298 {
		t.Fatalf("resize overwrote display configuration: %+v", rows)
	}
	configured.Collapsed = false
	configured.Height = 420
	if err := e.SaveWindow(configured); err != nil {
		t.Fatal(err)
	}
	if err := e.SaveWindowBounds(WindowState{Key: "mini", X: 80, Y: 90, Width: 298, Height: 418}); err != nil {
		t.Fatal(err)
	}
	rows, _ = e.Windows()
	if rows[0].Collapsed || rows[0].Height != 418 || rows[0].Mode != "top" {
		t.Fatalf("expanded configuration lost: %+v", rows)
	}
	if err := e.SaveWindowBounds(WindowState{Key: "unknown"}); err == nil {
		t.Fatal("accepted invalid window key")
	}
}

func TestAllTasksViewIncludesScheduledInboxAndCompleted(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "planned", ptr("2026-10-21"))
	b := mustRoot(t, e, "inbox", nil)
	if _, err := e.Complete(Command{CommandID: "all-complete", ID: a.ID, RootID: a.RootID, ExpectedVersion: cmd(t, e, a).ExpectedVersion, Completed: true}); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "all"})
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, v := range q.Tasks {
		found[v.ID] = true
	}
	if !found[a.ID] || !found[b.ID] {
		t.Fatal("all view omitted dated/completed/inbox task")
	}
	text, err := e.Export("markdown", Query{View: "all"}, false, nil)
	if err != nil || !strings.Contains(text, "planned") || !strings.Contains(text, "inbox") {
		t.Fatalf("all view export mismatch: %q %v", text, err)
	}
}
