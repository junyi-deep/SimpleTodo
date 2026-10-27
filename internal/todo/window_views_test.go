package todo

import "testing"

func TestViewWindowsPersistIndependentlyAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	e, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	configured := WindowState{Key: "main", Mode: "top", Opacity: .75, X: 40, Y: 50, Width: 480, Height: 720}
	if err = e.SaveWindow(configured); err != nil {
		t.Fatal(err)
	}
	want := map[string]ViewWindowState{}
	for i, view := range []string{"day", "inbox", "calendar", "focus"} {
		state := ViewWindowState{View: view, X: -1000 + i*100, Y: 20 + i*10, Width: 500 + i*100, Height: 600 + i*50, Maximised: view == "focus"}
		want[view] = state
		if err = e.SaveViewWindow(state); err != nil {
			t.Fatal(err)
		}
	}
	// Geometry events for the physical window must not overwrite any view.
	if err = e.SaveWindowBounds(WindowState{Key: "main", X: 30, Y: 40, Width: 900, Height: 800}); err != nil {
		t.Fatal(err)
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	e, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	got, err := e.ViewWindows()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d views, want %d", len(got), len(want))
	}
	for view, state := range want {
		if got[view] != state {
			t.Fatalf("%s: got %+v, want %+v", view, got[view], state)
		}
	}
	windows, err := e.Windows()
	if err != nil || len(windows) != 1 || windows[0].Key != "main" || windows[0].Mode != "top" || windows[0].Opacity != .75 {
		t.Fatalf("view records leaked into global display configuration: %+v, %v", windows, err)
	}
	for _, invalid := range []ViewWindowState{{View: "unknown", Width: 480, Height: 720}, {View: "day", Width: 0, Height: 720}, {View: "focus", Width: 480, Height: -1}} {
		if err = e.SaveViewWindow(invalid); err == nil {
			t.Fatalf("accepted invalid state: %+v", invalid)
		}
	}
}
