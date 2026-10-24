package todo

import "testing"

func TestCreateSiblingInsertionIsAtomicAndIdempotent(t *testing.T) {
	e := testEngine(t)
	a := mustRoot(t, e, "a", nil)
	b := mustRoot(t, e, "b", nil)
	if err := e.ReorderRoots([]string{a.ID, b.ID}, e.Revision()); err != nil {
		t.Fatal(err)
	}
	c := Command{CommandID: ID(), Title: "between", AfterID: a.ID}
	out, err := e.Create(c)
	if err != nil {
		t.Fatal(err)
	}
	middle := rootOf(out)
	repeated, err := e.Create(c)
	if err != nil || rootOf(repeated).ID != middle.ID {
		t.Fatalf("retry: %v", err)
	}
	q, _ := e.Query(Query{View: "inbox"})
	orders := map[string]int{}
	for _, n := range q.Tasks {
		orders[n.ID] = n.Order
	}
	if !(orders[a.ID] < orders[middle.ID] && orders[middle.ID] < orders[b.ID]) {
		t.Fatal(orders)
	}
	before := e.Revision()
	if _, err = e.Create(Command{CommandID: ID(), Title: "rollback", AfterID: "missing"}); err == nil {
		t.Fatal("accepted missing sibling")
	}
	if e.Revision() != before {
		t.Fatal("invalid insertion changed revision")
	}
	x := child(t, e, a, "x")
	y := child(t, e, a, "y")
	c = cmd(t, e, x)
	c.ParentID = ptr(a.ID)
	c.Title = "z"
	c.AfterID = x.ID
	out, err = e.Create(c)
	if err != nil {
		t.Fatal(err)
	}
	var z Task
	for _, n := range out {
		if n.Title == "z" {
			z = n
		}
	}
	tree := getTree(t, e, a.ID)
	if !(tree[x.ID].Order < z.Order && z.Order < tree[y.ID].Order) {
		t.Fatal(tree)
	}
}

func TestCreateWeeklyRepeatForeverAndBoundedRange(t *testing.T) {
	for _, bounded := range []bool{false, true} {
		t.Run(map[bool]string{false: "forever", true: "bounded"}[bounded], func(t *testing.T) {
			e := testEngine(t)
			r := &Rule{Frequency: "weekly", Interval: 1, Weekdays: []int{2, 4}}
			if bounded {
				r.Until = ptr("2026-10-29")
			}
			out, err := e.Create(Command{CommandID: ID(), Title: "repeat", Date: ptr("2026-10-20"), RepeatRule: r})
			if err != nil {
				t.Fatal(err)
			}
			first := rootOf(out)
			q, err := e.Query(Query{View: "month", Date: "2026-10-19", End: "2026-11-05"})
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]bool{"2026-10-20": true, "2026-10-22": true, "2026-10-27": true, "2026-10-29": true}
			if !bounded {
				want["2026-11-03"] = true
				want["2026-11-05"] = true
			}
			if len(q.Tasks) != len(want) {
				t.Fatalf("%d != %d", len(q.Tasks), len(want))
			}
			for _, n := range q.Tasks {
				if !want[val(n.Date)] || val(n.SeriesID) != val(first.SeriesID) {
					t.Fatal(n)
				}
			}
			rules, _ := e.Rules()
			if len(rules) != 1 || (rules[0].Until != nil) != bounded {
				t.Fatal(rules)
			}
		})
	}
}
