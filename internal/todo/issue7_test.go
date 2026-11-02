package todo

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

func TestTagColorsMigrateFromSchema3(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", sqliteURI(filepath.Join(dir, "simpletodo.db")))
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"001_init.sql", "002_interaction_calendar.sql", "003_workspace.sql"} {
		data, err := migrations.ReadFile("migrations/" + name)
		if err == nil {
			_, err = db.Exec(string(data))
		}
		if err == nil {
			_, err = db.Exec("INSERT INTO schema_migrations VALUES(?,?,0)", i+1, fmt.Sprintf("%x", sha256.Sum256(data)))
		}
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
	}
	if _, err = db.Exec("INSERT INTO tags VALUES('custom','旧标签',2)"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := e.Tags()
	if err != nil || len(tags) != 3 || tags[0].Color != "#b93838" || tags[1].Color != "#b45309" || tags[2].Color != defaultTagColor {
		t.Fatal(tags, err)
	}
	backups, err := e.Backups()
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	backup := filepath.Join(dir, "backups", backups[0].Name)
	if err = checkBackup(backup); err != nil {
		t.Fatal(err)
	}
	e.Close()
	e, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	tags, err = e.Tags()
	if err != nil || tags[2].Name != "旧标签" || tags[2].Color != defaultTagColor {
		t.Fatal(tags, err)
	}
	if err = e.Restore(backups[0].Name); err != nil {
		t.Fatal("restore schema 3 backup", err)
	}
	tags, err = e.Tags()
	if err != nil || len(tags) != 3 || tags[2].Color != defaultTagColor {
		t.Fatal(tags, err)
	}
}

func TestTagColorsPersistAcrossExportAndBackup(t *testing.T) {
	e := testEngine(t)
	if err := e.SaveTags([]Tag{{ID: "project", Name: "项目", Color: "#3563C1"}}); err != nil {
		t.Fatal(err)
	}
	// An older client omitting color must preserve the existing choice.
	if err := e.SaveTags([]Tag{{ID: "project", Name: "项目"}}); err != nil {
		t.Fatal(err)
	}
	tags, _ := e.Tags()
	if tags[0].Color != "#3563c1" {
		t.Fatal(tags)
	}
	b, err := e.CreateBackup()
	if err != nil {
		t.Fatal(err)
	}
	if err = e.SaveTags([]Tag{{ID: "project", Name: "项目", Color: "#b93838"}}); err != nil {
		t.Fatal(err)
	}
	if err = e.Restore(b.Name); err != nil {
		t.Fatal(err)
	}
	tags, _ = e.Tags()
	if tags[0].Color != "#3563c1" {
		t.Fatal(tags)
	}
	content, err := e.Export("json", Query{View: "all"}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	target := testEngine(t)
	p, err := target.PreviewImport(content, "json", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = target.CommitImport(p.Token); err != nil {
		t.Fatal(err)
	}
	tags, _ = target.Tags()
	if len(tags) != 1 || tags[0].Color != "#3563c1" {
		t.Fatal(tags)
	}
	if err = target.SaveTags([]Tag{{ID: "project", Name: "项目", Color: "red"}}); err == nil {
		t.Fatal("invalid color accepted")
	}
	tags, _ = target.Tags()
	if tags[0].Color != "#3563c1" {
		t.Fatal("invalid write changed tag", tags)
	}
	_, err = target.PreviewImport(`{"formatVersion":1,"tasks":[],"tags":[{"id":"bad","name":"bad","color":"url(foo)"}]}`, "json", nil, false)
	if err == nil {
		t.Fatal("invalid imported color accepted")
	}
}

func TestCalendarCompletedPreferenceIsIndependentAndPersists(t *testing.T) {
	e := testEngine(t)
	settings, _ := e.Settings()
	if settings["calendarShowCompleted"] != "false" || settings["dayShowCompleted"] != "true" {
		t.Fatal(settings)
	}
	if err := e.SaveSettings(map[string]string{"calendarShowCompleted": "true"}); err != nil {
		t.Fatal(err)
	}
	settings, _ = e.Settings()
	if settings["calendarShowCompleted"] != "true" || settings["inboxShowCompleted"] != "false" {
		t.Fatal(settings)
	}
	if err := e.SaveSettings(map[string]string{"calendarShowCompleted": "invalid"}); err == nil {
		t.Fatal("invalid preference accepted")
	}
}
