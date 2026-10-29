package todo

import "testing"

func TestCompletedParentRemainsEditableAndNewChildReopensAncestors(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "finished project", nil)
	p := child(t, e, r, "finished branch")
	old := child(t, e, p, "finished leaf")
	c := cmd(t, e, r)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	edit := editable(t, e, getTree(t, e, r.ID)[r.ID])
	edit.Title = "editable finished project"
	if _, err := e.Update(edit); err != nil {
		t.Fatal(err)
	}
	before := getTree(t, e, r.ID)
	if before[r.ID].Status != "completed" || before[r.ID].CompletedAt == nil || before[r.ID].Title != edit.Title {
		t.Fatal("editing changed completion", before)
	}
	next := child(t, e, p, "new pending step")
	after := getTree(t, e, r.ID)
	for _, id := range []string{r.ID, p.ID, next.ID} {
		if after[id].Status != "pending" || after[id].CompletedAt != nil {
			t.Fatalf("ancestor not reopened: %+v", after[id])
		}
	}
	if after[old.ID].Status != "completed" || after[old.ID].CompletedAt == nil {
		t.Fatal("old completed child was reopened")
	}
}

func TestCompletedChildCanBePromotedWithItsSubtree(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "parent", ptr("2026-10-20"))
	p := child(t, e, r, "completed branch")
	leaf := child(t, e, p, "completed leaf")
	c := cmd(t, e, r)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Promote(cmd(t, e, p), e.Revision()); err != nil {
		t.Fatal(err)
	}
	tree := getTree(t, e, p.ID)
	if tree[p.ID].ParentID != nil || tree[p.ID].Status != "completed" || tree[p.ID].CompletedAt == nil || val(tree[p.ID].Date) != "2026-10-20" || tree[leaf.ID].RootID != p.ID || tree[leaf.ID].Status != "completed" {
		t.Fatal("promotion lost completion or subtree", tree)
	}
}
