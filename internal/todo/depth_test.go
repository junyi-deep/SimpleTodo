package todo

import (
	"strings"
	"testing"
)

func TestFiveLevelsAndAtomicOverflow(t *testing.T) {
	e := testEngine(t)
	root := mustRoot(t, e, "root", nil)
	leaf := root
	for _, title := range []string{"two", "three", "four", "five"} {
		leaf = child(t, e, leaf, title)
	}
	rev := e.Revision()
	c := cmd(t, e, leaf)
	c.Title = "six"
	c.ParentID = ptr(leaf.ID)
	if _, err := e.Create(c); err == nil || !strings.Contains(err.Error(), "5 层") {
		t.Fatalf("sixth accepted: %v", err)
	}
	if e.Revision() != rev || len(getTree(t, e, root.ID)) != 5 {
		t.Fatal("failed create mutated tree")
	}
	other := mustRoot(t, e, "other", nil)
	rev = e.Revision()
	c = cmd(t, e, other)
	c.ParentID = ptr(leaf.ID)
	if _, err := e.Nest(c, rev); err == nil {
		t.Fatal("overdeep nest accepted")
	}
	if e.Revision() != rev || len(getTree(t, e, other.ID)) != 1 {
		t.Fatal("failed nest mutated source")
	}
	if _, err := e.PreviewImport("- a\n  - b\n    - c\n      - d\n        - e\n          - f", "markdown", nil, false); err == nil {
		t.Fatal("overdeep import accepted")
	}
}

func TestLegacyDeepTreeCanBeEditedAndShortened(t *testing.T) {
	e := testEngine(t)
	root := mustRoot(t, e, "legacy", nil)
	leaf := root
	for _, title := range []string{"two", "three", "four", "five"} {
		leaf = child(t, e, leaf, title)
	}
	// Simulate a pre-upgrade sixth-level row without changing the current API limit.
	if _, err := e.db.Exec(`INSERT INTO tasks SELECT 'legacy-six',?,root_id,title,description,planned_date,original_planned_date,status,color,sort_order,completed_at,auto_postpone_policy,series_id,rule_id,occurrence_date,is_exception,remind_at,reminder_local_time,reminder_timezone,reminder_generation,version,created_at,updated_at FROM tasks WHERE id=?`, leaf.ID, leaf.ID); err != nil {
		t.Fatal(err)
	}
	legacy := getTree(t, e, root.ID)["legacy-six"]
	c := cmd(t, e, legacy)
	c.Title = "still editable"
	if _, err := e.Update(c); err != nil {
		t.Fatal(err)
	}
	c = cmd(t, e, legacy)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	c = cmd(t, e, legacy)
	c.ParentID = ptr(root.ID)
	if _, err := e.Reparent(c); err != nil {
		t.Fatal(err)
	}
	if taskDepth(getTree(t, e, root.ID), legacy.ID) != 2 {
		t.Fatal("legacy branch could not be shortened")
	}
}
