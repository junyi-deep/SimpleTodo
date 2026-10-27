package main

import (
	"changeme/internal/todo"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestRestoreEachViewsOwnPositionAndSize(t *testing.T) {
	area := application.Rect{X: -1920, Y: 25, Width: 1920, Height: 1055}
	saved := map[string]todo.ViewWindowState{}
	for i, view := range []string{"day", "inbox", "calendar", "focus"} {
		saved[view] = todo.ViewWindowState{View: view, X: -1800 + i*100, Y: 50 + i*20, Width: 800 + i*10, Height: 650 + i*10}
	}
	current := application.Rect{X: 200, Y: 100, Width: 1200, Height: 900}
	for _, view := range []string{"calendar", "focus", "day", "inbox", "day", "calendar"} {
		got := restoredViewBounds(view, saved, current, area)
		if want := viewBounds(saved[view]); got != want {
			t.Fatalf("switch %s: got %+v, want %+v", view, got, want)
		}
		current = got
	}
}

func TestViewDefaultsClampingAndMaximisedNormalBounds(t *testing.T) {
	current := application.Rect{X: 300, Y: 100, Width: 1200, Height: 900}
	saved := map[string]todo.ViewWindowState{}
	for _, view := range []string{"day", "inbox", "calendar", "focus"} {
		got := restoredViewBounds(view, saved, current, application.Rect{})
		width := 480
		if view == "calendar" {
			width = 1080
		}
		if got.X != 300 || got.Y != 100 || got.Width != width || got.Height != 720 {
			t.Fatalf("wrong first-use bounds for %s: %+v", view, got)
		}
	}
	// A disconnected display and an oversized calendar must not strand the window.
	saved["calendar"] = todo.ViewWindowState{View: "calendar", X: -3000, Y: 1800, Width: 2000, Height: 1400}
	area := application.Rect{X: 0, Y: 25, Width: 700, Height: 550}
	if got := restoredViewBounds("calendar", saved, current, area); got != area {
		t.Fatalf("disconnected display was not clamped: %+v", got)
	}
	normal := application.Rect{X: 120, Y: 90, Width: 530, Height: 810}
	saved["day"] = capturedViewWindow("day", saved, normal, false)
	maximised := capturedViewWindow("day", saved, application.Rect{Width: 1920, Height: 1080}, true)
	if viewBounds(maximised) != normal || !maximised.Maximised {
		t.Fatalf("maximise destroyed normal bounds: %+v", maximised)
	}
	saved["day"] = maximised
	if restoredViewBounds("day", saved, current, application.Rect{}) != normal {
		t.Fatal("maximised view did not preserve restore rectangle")
	}
}

func TestChecklistWindowDefaultsAndWindowsTitlebar(t *testing.T) {
	for _, platform := range []string{"windows", "darwin"} {
		opts := mainWindowOptions(platform)
		if opts.Width != 480 || opts.Height != 720 || opts.MinWidth != 400 || opts.MinHeight != 440 || opts.URL != "/?window=main" {
			t.Fatalf("unexpected checklist window defaults: %+v", opts)
		}
		if opts.MinimiseButtonState != application.ButtonHidden || opts.CloseButtonState != application.ButtonHidden || opts.MaximiseButtonState != application.ButtonHidden {
			t.Fatal("native buttons must be hidden")
		}
		if opts.BackgroundType != application.BackgroundTypeTransparent || opts.BackgroundColour.Alpha != 0 || opts.Mac.Backdrop != application.MacBackdropTransparent {
			t.Fatal("window must allow a CSS background alpha without fading text")
		}
		if opts.Frameless != (platform == "windows") {
			t.Fatalf("incorrect native titlebar on %s", platform)
		}
	}
}

func TestViewBoundsStayInsideDisplay(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		bounds, area, expect application.Rect
	}{
		{"right edge", application.Rect{X: 900, Y: 50, Width: 1080, Height: 720}, application.Rect{X: 0, Y: 25, Width: 1512, Height: 900}, application.Rect{X: 432, Y: 50, Width: 1080, Height: 720}},
		{"small display", application.Rect{X: 800, Y: 500, Width: 1080, Height: 720}, application.Rect{X: 0, Y: 25, Width: 900, Height: 600}, application.Rect{X: 0, Y: 25, Width: 900, Height: 600}},
		{"negative monitor", application.Rect{X: -200, Y: -20, Width: 1080, Height: 720}, application.Rect{X: -1440, Y: -900, Width: 1440, Height: 875}, application.Rect{X: -1080, Y: -745, Width: 1080, Height: 720}},
		{"restore list", application.Rect{X: 900, Y: 100, Width: 440, Height: 580}, application.Rect{X: 0, Y: 25, Width: 1512, Height: 900}, application.Rect{X: 900, Y: 100, Width: 440, Height: 580}},
		{"unknown display", application.Rect{X: 900, Y: 50, Width: 1080, Height: 720}, application.Rect{}, application.Rect{X: 900, Y: 50, Width: 1080, Height: 720}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := fitWindowBounds(tc.bounds, tc.area); got != tc.expect {
				t.Fatalf("got %+v, want %+v", got, tc.expect)
			}
		})
	}
}
