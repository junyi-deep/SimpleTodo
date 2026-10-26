package todo

import (
	"fmt"
	"testing"
)

func TestPromoteBranchAtomicIdempotentAndDate(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "parent", ptr("2026-10-20"))
	b := child(t, e, a, "branch")
	c := child(t, e, b, "leaf")
	other := mustRoot(t, e, "other", nil)
	cmd := cmd(t, e, b)
	cmd.BeforeID = other.ID
	rev := e.Revision()
	out, err := e.Promote(cmd, rev)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := e.FindTask(b.ID)
	leaf := getTree(t, e, b.ID)[c.ID]
	original := getTree(t, e, a.ID)
	if root.ParentID != nil || root.RootID != b.ID || val(root.Date) != "2026-10-20" || leaf.RootID != b.ID || val(leaf.ParentID) != b.ID || len(original) != 1 || root.SeriesID != nil || root.RemindAt != nil {
		t.Fatal(root, leaf, original)
	}
	target, _ := e.FindTask(other.ID)
	if root.Order >= target.Order {
		t.Fatal("wrong insertion order")
	}
	after := e.Revision()
	repeat, err := e.Promote(cmd, rev)
	if err != nil || e.Revision() != after || len(out) != len(repeat) {
		t.Fatal("not idempotent", err)
	}
	if fmt.Sprint(root.Version) == cmd.ExpectedVersion {
		t.Fatal("root version not advanced")
	}
}
func TestPromoteFailureLeavesTreeIntact(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "parent", nil)
	b := child(t, e, a, "child")
	c := cmd(t, e, b)
	c.BeforeID = "missing"
	rev := e.Revision()
	if _, err := e.Promote(c, rev); err == nil {
		t.Fatal("accepted missing target")
	}
	if e.Revision() != rev || len(getTree(t, e, a.ID)) != 2 {
		t.Fatal("partial update")
	}
	c.BeforeID = ""
	if _, err := e.Promote(c, "old"); err == nil {
		t.Fatal("accepted stale list")
	}
	if err := e.SaveSettings(map[string]string{"expandDepth": "6"}); err == nil {
		t.Fatal("accepted invalid depth")
	}
	settings, _ := e.Settings()
	if settings["expandDepth"] != "2" || settings["autoPostpone"] != "true" {
		t.Fatal(settings)
	}
}
