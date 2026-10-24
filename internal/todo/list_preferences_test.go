package todo

import "testing"

func TestListFilterDefaultsAndPersistence(t *testing.T) {
	dir := t.TempDir()
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := e.Settings()
	if m["dayShowCompleted"] != "true" || m["inboxShowCompleted"] != "false" || m["dayShowRecurring"] != "true" || m["inboxShowRecurring"] != "true" {
		t.Fatal(m)
	}
	if err = e.SaveSettings(map[string]string{"dayShowCompleted": "false", "inboxShowCompleted": "true", "dayShowRecurring": "false"}); err != nil {
		t.Fatal(err)
	}
	if err = e.SaveSettings(map[string]string{"inboxShowRecurring": "invalid"}); err == nil {
		t.Fatal("accepted invalid switch")
	}
	e.Close()
	e, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	m, _ = e.Settings()
	if m["dayShowCompleted"] != "false" || m["inboxShowCompleted"] != "true" || m["dayShowRecurring"] != "false" || m["inboxShowRecurring"] != "true" {
		t.Fatal(m)
	}
}
func TestInboxOrderingPreservesDatesAndRejectsDuplicateIDs(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "scheduled", ptr("2026-10-23"))
	b := mustRoot(t, e, "inbox", nil)
	if err := e.ReorderRoots([]string{b.ID, a.ID}, e.Revision()); err != nil {
		t.Fatal(err)
	}
	x, _ := e.FindTask(a.ID)
	y, _ := e.FindTask(b.ID)
	if x.Date == nil || *x.Date != "2026-10-23" || y.Date != nil || y.Order >= x.Order {
		t.Fatal(x, y)
	}
	before := e.Revision()
	if err := e.ReorderRoots([]string{a.ID, a.ID}, before); err == nil {
		t.Fatal("accepted duplicate ID")
	}
	if e.Revision() != before {
		t.Fatal("invalid reorder changed revision")
	}
}
