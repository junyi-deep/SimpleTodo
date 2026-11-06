package todo

import (
	"testing"
	"time"
)

func TestDefaultCreationEndsOnPlannedDayAndBrowsingDoesNotPostpone(t *testing.T) {
	for _, offset := range []int{0, 2} {
		e := testEngine(t)
		if err := e.UpgradeTaskPeriods(); err != nil {
			t.Fatal(err)
		}
		now := e.localNow()
		day := today(now.AddDate(0, 0, offset))
		c := Command{CommandID: ID(), Title: "single day", Date: &day}
		rows, err := e.Create(c)
		if err != nil {
			t.Fatal(err)
		}
		root := rootOf(rows)
		at := time.UnixMilli(root.StartAt).In(e.location())
		expectedEnd := time.Date(at.Year(), at.Month(), at.Day()+1, 0, 0, 0, 0, e.location()).UnixMilli() - 1
		if today(at) != day || val(root.EndAt) != expectedEnd {
			t.Fatal(root)
		}
		for _, distance := range []int{1, 2, 10} {
			future := today(now.AddDate(0, 0, offset+distance))
			q, er := e.Query(Query{View: "day", Date: future})
			if er != nil || len(q.Tasks) != 0 {
				t.Fatal("browsing extended task", future, q, er)
			}
			current, er := e.FindTask(root.ID)
			if er != nil || val(current.EndAt) != expectedEnd || current.StartAt != root.StartAt {
				t.Fatal(current, er)
			}
		}
		retry, err := e.Create(c)
		if err != nil || rootOf(retry).ID != root.ID {
			t.Fatal("default span broke retry", retry, err)
		}
	}
}

func TestDefaultTaskPostponesOnlyEndToActualToday(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		e := testEngine(t)
		if err := e.UpgradeTaskPeriods(); err != nil {
			t.Fatal(err)
		}
		if !enabled {
			if err := e.SaveSettings(map[string]string{"autoPostpone": "false"}); err != nil {
				t.Fatal(err)
			}
		}
		now := e.localNow()
		day := today(now)
		rows, err := e.Create(Command{CommandID: ID(), Title: "postpone", Date: &day})
		if err != nil {
			t.Fatal(err)
		}
		root := rootOf(rows)
		branch := child(t, e, root, "done")
		child(t, e, root, "pending")
		c := cmd(t, e, branch)
		c.Completed = true
		if _, err = e.Complete(c); err != nil {
			t.Fatal(err)
		}
		originalChild := getTree(t, e, root.ID)[branch.ID]
		for _, elapsed := range []int{1, 3} {
			next := now.AddDate(0, 0, elapsed)
			e.now = func() time.Time { return next }
			if err = e.Tick(); err != nil {
				t.Fatal(err)
			}
			q, er := e.Query(Query{View: "day", Date: today(next)})
			if er != nil {
				t.Fatal(er)
			}
			if !enabled {
				if len(q.Tasks) != 0 {
					t.Fatal("disabled task postponed", q)
				}
				continue
			}
			if len(q.Tasks) != 3 {
				t.Fatal(q)
			}
			current, er := e.FindTask(root.ID)
			if er != nil || current.StartAt != root.StartAt || val(current.Date) != day || current.CreatedAt != root.CreatedAt || current.EndAt == nil || today(time.UnixMilli(*current.EndAt).In(e.location())) != today(next) {
				t.Fatal(current, er)
			}
			for _, task := range q.Tasks {
				if task.Title == branch.Title && (task.Status != "completed" || val(task.CompletedAt) != val(originalChild.CompletedAt) || val(task.EndAt) != val(originalChild.EndAt)) {
					t.Fatal("completed child changed", task)
				}
			}
			future, er := e.Query(Query{View: "day", Date: today(next.AddDate(0, 0, 1))})
			if er != nil || len(future.Tasks) != 0 {
				t.Fatal("postpone extended into future", future, er)
			}
		}
	}
}
