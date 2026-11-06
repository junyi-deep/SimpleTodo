package todo

import (
	"fmt"
	"testing"
	"time"
)

func TestPeriodCompletionStopsFutureCopiesAndCanBeUndone(t *testing.T) {
	for _, batch := range []bool{false, true} {
		name := "single"
		if batch {
			name = "batch"
		}
		t.Run(name, func(t *testing.T) {
			e := testEngine(t)
			if err := e.UpgradeTaskPeriods(); err != nil {
				t.Fatal(err)
			}
			now := e.localNow()
			day, tomorrow := today(now), today(now.AddDate(0, 0, 1))
			end := now.AddDate(0, 0, 2).UnixMilli()
			reminder := now.Add(2 * time.Hour)
			out, err := e.Create(Command{CommandID: ID(), Title: "period", Date: &day, Span: &TaskSpan{Start: now.UnixMilli(), End: &end}, RemindAt: ptr(reminder.UnixMilli()), ReminderTime: ptr(reminder.Format("15:04"))})
			if err != nil {
				t.Fatal(err)
			}
			root := rootOf(out)
			if err = e.EnsureThrough(tomorrow); err != nil {
				t.Fatal(err)
			}
			future, err := e.Query(Query{View: "day", Date: tomorrow})
			if err != nil || len(future.Tasks) != 1 {
				t.Fatal(future, err)
			}
			e.now = func() time.Time { return now.Add(time.Hour) }
			if batch {
				p, err := e.PreviewBatch([]string{root.ID}, "complete", e.Revision())
				if err != nil {
					t.Fatal(err)
				}
				if err = e.CommitBatch(p.Token); err != nil {
					t.Fatal(err)
				}
			} else {
				c := cmd(t, e, root)
				c.Completed = true
				if _, err = e.Complete(c); err != nil {
					t.Fatal(err)
				}
			}
			finished, err := e.FindTask(root.ID)
			if err != nil || finished.Status != "completed" || val(finished.EndAt) != e.now().UnixMilli() {
				t.Fatal(finished, err)
			}
			if err = e.EnsureThrough(today(now.AddDate(0, 0, 7))); err != nil {
				t.Fatal(err)
			}
			q, err := e.Query(Query{View: "day", Date: tomorrow})
			if err != nil || len(q.Tasks) != 0 {
				t.Fatal(q, err)
			}
			var reminderState string
			err = e.db.QueryRow("SELECT state FROM notification_jobs WHERE source_id=?", future.Tasks[0].ID).Scan(&reminderState)
			if err != nil || reminderState != "cancelled" {
				t.Fatal(reminderState, err)
			}
			report, err := e.TaskReport(Query{View: "gantt"})
			if err != nil || len(report) != 1 || report[0].ID != root.ID {
				t.Fatal(report, err)
			}
			copyDay := today(now.AddDate(0, 0, 8))
			if err = e.Duplicate(cmd(t, e, finished), &copyDay); err != nil {
				t.Fatal(err)
			}
			q, err = e.Query(Query{View: "day", Date: copyDay})
			if err != nil || len(q.Tasks) != 1 || q.Tasks[0].Status != "pending" {
				t.Fatal(q, err)
			}
			if _, err = e.Complete(cmd(t, e, finished)); err != nil {
				t.Fatal(err)
			}
			reopened, err := e.FindTask(root.ID)
			if err != nil || reopened.Status != "pending" || val(reopened.EndAt) != end {
				t.Fatal(reopened, err)
			}
			q, err = e.Query(Query{View: "day", Date: tomorrow})
			if err != nil || len(q.Tasks) != 1 || q.Tasks[0].Status != "pending" {
				t.Fatal(q, err)
			}
		})
	}
}

// A completed branch is history within an unfinished root, not a new pending task.
func TestPeriodRetainsCompletedChildrenOnLaterDays(t *testing.T) {
	for _, mode := range []string{"unbounded", "before-end", "auto-postpone"} {
		for _, batch := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/batch=%t", mode, batch), func(t *testing.T) {
				e := testEngine(t)
				if err := e.UpgradeTaskPeriods(); err != nil {
					t.Fatal(err)
				}
				now := e.localNow()
				day := today(now)
				var end *int64
				if mode != "unbounded" {
					days := 3
					if mode == "auto-postpone" {
						days = 0
					}
					end = ptr(now.AddDate(0, 0, days).UnixMilli())
				}
				out, err := e.Create(Command{CommandID: ID(), Title: "parent", Date: &day, Span: &TaskSpan{Start: now.UnixMilli(), End: end}})
				if err != nil {
					t.Fatal(err)
				}
				root := rootOf(out)
				branch := child(t, e, root, "finished branch")
				leaf := child(t, e, branch, "finished leaf")
				child(t, e, root, "pending sibling")
				tomorrow := today(now.AddDate(0, 0, 1))
				if err = e.EnsureThrough(tomorrow); err != nil {
					t.Fatal(err)
				}
				e.now = func() time.Time { return now.Add(time.Minute) }
				if batch {
					p, er := e.PreviewBatch([]string{branch.ID}, "complete", e.Revision())
					if er != nil {
						t.Fatal(er)
					}
					err = e.CommitBatch(p.Token)
				} else {
					c := cmd(t, e, branch)
					c.Completed = true
					_, err = e.Complete(c)
				}
				if err != nil {
					t.Fatal(err)
				}
				source := getTree(t, e, root.ID)
				if source[root.ID].Status != "pending" {
					t.Fatal("parent completed early")
				}
				for offset := 1; offset <= 2; offset++ {
					if mode == "auto-postpone" {
						at := now.AddDate(0, 0, offset)
						e.now = func() time.Time { return at }
						if err = e.Tick(); err != nil {
							t.Fatal(err)
						}
					}
					target := today(now.AddDate(0, 0, offset))
					q, er := e.Query(Query{View: "day", Date: target})
					if er != nil || len(q.Tasks) != 4 {
						t.Fatal(target, q.Tasks, er)
					}
					byTitle := map[string]Task{}
					for _, task := range q.Tasks {
						byTitle[task.Title] = task
					}
					for _, original := range []Task{source[branch.ID], source[leaf.ID]} {
						got := byTitle[original.Title]
						if got.Status != "completed" || val(got.CompletedAt) != val(original.CompletedAt) || val(got.EndAt) != val(original.EndAt) || got.CreatedAt != original.CreatedAt || got.UpdatedAt != original.UpdatedAt {
							t.Fatalf("completed metadata changed on %s: got %+v, source %+v", target, got, original)
						}
					}
					if val(byTitle["finished leaf"].ParentID) != byTitle["finished branch"].ID || val(byTitle["finished branch"].ParentID) != byTitle["parent"].ID {
						t.Fatal("nested branch detached", byTitle)
					}
					if byTitle["parent"].Status != "pending" || byTitle["pending sibling"].Status != "pending" {
						t.Fatal(byTitle)
					}
				}
				history := getTree(t, e, root.ID)
				if history[branch.ID].Status != "completed" || val(history[branch.ID].CompletedAt) != val(source[branch.ID].CompletedAt) {
					t.Fatal("history changed", history)
				}
			})
		}
	}
}

func TestPeriodUpgradeRebuildsOnlyUntouchedFutureSnapshots(t *testing.T) {
	e := testEngine(t)
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	now := e.localNow()
	day := today(now)
	out, err := e.Create(Command{CommandID: ID(), Title: "upgrade parent", Date: &day, Span: &TaskSpan{Start: now.UnixMilli()}})
	if err != nil {
		t.Fatal(err)
	}
	root := rootOf(out)
	done := child(t, e, root, "done")
	child(t, e, root, "pending")
	c := cmd(t, e, done)
	c.Completed = true
	if _, err := e.Complete(c); err != nil {
		t.Fatal(err)
	}
	tomorrow, later := today(now.AddDate(0, 0, 1)), today(now.AddDate(0, 0, 2))
	if err := e.EnsureThrough(later); err != nil {
		t.Fatal(err)
	}
	rules, err := e.Rules()
	if err != nil || len(rules) != 1 {
		t.Fatal(rules, err)
	}
	// Reproduce the old release's pre-generated copies, which omitted done children.
	tx, err := e.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.Exec("DELETE FROM tasks WHERE parent_id IS NULL AND series_id=? AND planned_date>?", val(root.SeriesID), day); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{tomorrow, later} {
		old := cloneTree(rules[0].Template, target, rules[0].SeriesID, rules[0].ID, now.UnixMilli(), false, e.location(), true, false)
		for _, node := range old {
			if target == later && node.ParentID == nil {
				node.Exception = true
				node.Title = "independently edited"
			}
			if err = putTask(tx, node); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err = tx.Exec("DELETE FROM app_meta WHERE key='period_completed_children_ready'"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || len(q.Tasks) != 2 {
		t.Fatal(q, err)
	}
	q, err = e.Query(Query{View: "day", Date: later})
	if err != nil {
		t.Fatal(err)
	}
	edited := rootOf(q.Tasks)
	e.now = func() time.Time { return now.AddDate(0, 0, 1) }
	if err = e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || len(q.Tasks) != 3 {
		t.Fatal(q, err)
	}
	for _, node := range q.Tasks {
		if node.Title == "done" && (node.Status != "completed" || node.CompletedAt == nil) {
			t.Fatal(node)
		}
	}
	restoredID := rootOf(q.Tasks).ID
	q, err = e.Query(Query{View: "day", Date: later})
	if err != nil || rootOf(q.Tasks).ID != edited.ID || rootOf(q.Tasks).Title != "independently edited" {
		t.Fatal("edited snapshot replaced", q, err)
	}
	if err = e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	q, err = e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || rootOf(q.Tasks).ID != restoredID {
		t.Fatal("upgrade was not idempotent", q, err)
	}
	if getTree(t, e, root.ID)[done.ID].Status != "completed" {
		t.Fatal("source history changed")
	}
}
