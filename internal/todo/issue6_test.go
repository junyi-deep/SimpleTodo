package todo

import (
	"changeme/internal/localcalendar"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestBatchChildBranchesAreDeduplicatedAndDoNotChangeSiblings(t *testing.T) {
	e := testEngine(t)
	r := mustRoot(t, e, "root", nil)
	a := child(t, e, r, "a")
	leaf := child(t, e, a, "leaf")
	b := child(t, e, r, "b")
	p, err := e.PreviewBatch([]string{a.ID, leaf.ID, a.ID}, "complete", e.Revision())
	if err != nil || p.Roots != 1 || p.Nodes != 2 {
		t.Fatal(p, err)
	}
	if err = e.CommitBatch(p.Token); err != nil {
		t.Fatal(err)
	}
	tree := getTree(t, e, r.ID)
	if tree[a.ID].Status != "completed" || tree[leaf.ID].Status != "completed" || tree[b.ID].Status != "pending" || tree[r.ID].Status != "pending" {
		t.Fatal(tree)
	}
	p, err = e.PreviewBatch([]string{leaf.ID}, "pending", e.Revision())
	if err != nil {
		t.Fatal(err)
	}
	if err = e.CommitBatch(p.Token); err != nil {
		t.Fatal(err)
	}
	tree = getTree(t, e, r.ID)
	if tree[a.ID].Status != "pending" || tree[leaf.ID].Status != "pending" {
		t.Fatal(tree)
	}
	p, err = e.PreviewBatch([]string{a.ID, leaf.ID}, "delete", e.Revision())
	if err != nil {
		t.Fatal(err)
	}
	if err = e.CommitBatch(p.Token); err != nil {
		t.Fatal(err)
	}
	tree = getTree(t, e, r.ID)
	if len(tree) != 2 || tree[b.ID].ID == "" {
		t.Fatal(tree)
	}
	p, err = e.PreviewBatch([]string{b.ID}, "delete", e.Revision())
	if err != nil {
		t.Fatal(err)
	}
	child(t, e, r, "concurrent")
	if e.CommitBatch(p.Token) == nil {
		t.Fatal("stale batch committed")
	}
	if getTree(t, e, r.ID)[b.ID].ID == "" {
		t.Fatal("stale batch deleted data")
	}
}

func TestExpansionAndCalendarCacheSurviveRestartWithoutTaskWrites(t *testing.T) {
	dir := t.TempDir()
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := mustRoot(t, e, "root", nil)
	a := child(t, e, r, "a")
	rev := e.Revision()
	if err = e.SaveTaskExpansion(map[string]bool{r.ID: false, a.ID: true}); err != nil {
		t.Fatal(err)
	}
	events := []localcalendar.Event{{ID: "event-1", CalendarID: "outlook-local", Title: "meeting", Start: 1000, End: 3000}}
	if err = e.SaveLocalCalendar([]string{"outlook-local"}, events); err != nil {
		t.Fatal(err)
	}
	if e.Revision() != rev {
		t.Fatal("preferences/cache changed task revision")
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	e, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	states, err := e.TaskExpansion()
	if err != nil || states[r.ID] || !states[a.ID] {
		t.Fatal(states, err)
	}
	state, err := e.LocalCalendarState(1500, 2000)
	if err != nil || len(state.Events) != 1 || len(state.Sources) != 1 {
		t.Fatal(state, err)
	}
	if err = e.SaveLocalCalendar([]string{"different"}, events); err == nil {
		t.Fatal("unselected source accepted")
	}
	state, _ = e.LocalCalendarState(1500, 2000)
	if len(state.Events) != 1 {
		t.Fatal("failed refresh lost cache")
	}
	if len(getTree(t, e, r.ID)) != 2 {
		t.Fatal("calendar data entered task tree")
	}
	if err = e.SaveLocalCalendar(nil, nil); err != nil {
		t.Fatal(err)
	}
	state, _ = e.LocalCalendarState(0, 9999)
	if len(state.Events) != 0 || len(state.Sources) != 0 {
		t.Fatal(state)
	}
}

func TestSchemaOneMigratesWithProtectiveBackupAndPreservesTasks(t *testing.T) {
	dir := t.TempDir()
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := mustRoot(t, e, "legacy", nil)
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", sqliteURI(filepath.Join(dir, "simpletodo.db")))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DROP TABLE task_expansion;DROP TABLE local_calendar_events;DROP TABLE local_calendar_config;DELETE FROM schema_migrations WHERE version=2")
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	e, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if getTree(t, e, r.ID)[r.ID].Title != "legacy" {
		t.Fatal("migration lost task")
	}
	backups, err := filepath.Glob(filepath.Join(dir, "backups", "pre-schema-v2-*.db"))
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	db, err = sql.Open("sqlite", sqliteURI(backups[0])+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var v int
	if err = db.QueryRow("SELECT max(version) FROM schema_migrations").Scan(&v); err != nil || v != 1 {
		t.Fatal(v, err)
	}
	if err = e.SaveTaskExpansion(map[string]bool{r.ID: true}); err != nil {
		t.Fatal(err)
	}
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	if err = e.Restore(b.Name); err != nil {
		t.Fatal(err)
	}
	states, err := e.TaskExpansion()
	if err != nil || !states[r.ID] {
		t.Fatal(states, err)
	}
}

func TestLocalCalendarAcceptsInstantAppointments(t *testing.T) {
	e := testEngine(t)
	events := []localcalendar.Event{{ID: "instant", CalendarID: "local", Title: "appointment", Start: 1000, End: 1000}}
	if err := e.SaveLocalCalendar([]string{"local"}, events); err != nil {
		t.Fatal(err)
	}
	state, err := e.LocalCalendarState(1000, 2000)
	if err != nil || len(state.Events) != 1 || state.Events[0].End != 1001 || events[0].End != 1000 {
		t.Fatal(state, err)
	}
}
