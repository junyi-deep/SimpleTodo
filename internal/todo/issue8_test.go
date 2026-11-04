package todo

import (
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

func TestPeriodCompletionKeepsOnlyUnfinishedChildrenOnLaterDays(t *testing.T) {
	e := testEngine(t)
	if err := e.UpgradeTaskPeriods(); err != nil {
		t.Fatal(err)
	}
	now := e.localNow()
	day, tomorrow := today(now), today(now.AddDate(0, 0, 1))
	out, err := e.Create(Command{CommandID: ID(), Title: "parent", Date: &day, Span: &TaskSpan{Start: now.UnixMilli()}})
	if err != nil {
		t.Fatal(err)
	}
	root := rootOf(out)
	finished := child(t, e, root, "finished")
	child(t, e, root, "pending")
	if err = e.EnsureThrough(tomorrow); err != nil {
		t.Fatal(err)
	}
	c := cmd(t, e, finished)
	c.Completed = true
	if _, err = e.Complete(c); err != nil {
		t.Fatal(err)
	}
	q, err := e.Query(Query{View: "day", Date: tomorrow})
	if err != nil || len(q.Tasks) != 2 {
		t.Fatal(q, err)
	}
	for _, task := range q.Tasks {
		if task.Title == "finished" || task.Status != "pending" {
			t.Fatal(task)
		}
	}
}
