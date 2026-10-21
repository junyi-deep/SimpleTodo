package todo

import (
	"strings"
	"testing"
)

func TestNestRootBranchAtomicAndIdempotent(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "源任务", ptr("2026-10-20"))
	leaf := child(t, e, a, "子任务")
	b := mustRoot(t, e, "目标", ptr("2026-10-21"))
	edit := editable(t, e, a)
	edit.Description = "原来的说明"
	if _, err := e.Update(edit); err != nil {
		t.Fatal(err)
	}
	c := cmd(t, e, a)
	c.ParentID = ptr(b.ID)
	revision := e.Revision()
	if _, err := e.Nest(c, revision); err != nil {
		t.Fatal(err)
	}
	m := getTree(t, e, b.ID)
	if len(m) != 3 || val(m[a.ID].ParentID) != b.ID || m[leaf.ID].RootID != b.ID || m[a.ID].Date != nil || !strings.Contains(val(m[b.ID].Description), "原来的说明") {
		t.Fatalf("branch not preserved: %+v", m)
	}
	after := e.Revision()
	if _, err := e.Nest(c, revision); err != nil || e.Revision() != after {
		t.Fatalf("retry not idempotent: %v", err)
	}
	cyc := cmd(t, e, b)
	cyc.ParentID = ptr(leaf.ID)
	if _, err := e.Nest(cyc, e.Revision()); err == nil {
		t.Fatal("accepted cycle")
	}
	if e.Revision() != after {
		t.Fatal("failed move changed revision")
	}
	c = cmd(t, e, a)
	c.ParentID = ptr(b.ID)
	if _, err := e.Nest(c, "stale"); err == nil {
		t.Fatal("stale snapshot accepted")
	}
}
func TestNestRecurringInstanceDoesNotRegenerate(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "每天", ptr("2026-10-20"))
	c := editable(t, e, a)
	c.RepeatRule = &Rule{Frequency: "daily", Interval: 1}
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	b := mustRoot(t, e, "项目", nil)
	move := cmd(t, e, a)
	move.ParentID = ptr(b.ID)
	if _, err := e.Nest(move, e.Revision()); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureThrough("2026-10-22"); err != nil {
		t.Fatal(err)
	}
	m := getTree(t, e, b.ID)
	if val(m[a.ID].ParentID) != b.ID || m[a.ID].SeriesID != nil {
		t.Fatal("instance not converted")
	}
	var count int
	if err := e.db.QueryRow("SELECT count(*) FROM tasks WHERE parent_id IS NULL AND occurrence_date='2026-10-20'").Scan(&count); err != nil || count != 1 {
		t.Fatal("occurrence regenerated", count, err)
	}
	var state string
	_ = e.db.QueryRow("SELECT status FROM tasks WHERE parent_id IS NULL AND occurrence_date='2026-10-20'").Scan(&state)
	if state != "skipped" {
		t.Fatal("missing skipped marker")
	}
}
func TestDefaultPostponeAndAutostartInitialization(t *testing.T) {
	e := testEngine(t)
	m, err := e.Settings()
	if err != nil || m["autoPostpone"] != "true" || m["autostartConfigured"] != "false" {
		t.Fatal(m, err)
	}
	if err = e.SaveSettings(map[string]string{"autoPostpone": "false", "autostartConfigured": "true"}); err != nil {
		t.Fatal(err)
	}
	m, _ = e.Settings()
	if m["autoPostpone"] != "false" || m["autostartConfigured"] != "true" {
		t.Fatal("preference lost")
	}
}

func TestNestDescriptionOverflowRollsBack(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "source", nil)
	b := mustRoot(t, e, "target", nil)
	c := editable(t, e, a)
	c.Description = strings.Repeat("x", 20000)
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	c = cmd(t, e, a)
	c.ParentID = ptr(b.ID)
	revision := e.Revision()
	if _, err := e.Nest(c, revision); err == nil {
		t.Fatal("accepted merged description overflow")
	}
	if e.Revision() != revision || getTree(t, e, a.ID)[a.ID].ParentID != nil || len(getTree(t, e, b.ID)) != 1 {
		t.Fatal("failed move did not roll back")
	}
}
