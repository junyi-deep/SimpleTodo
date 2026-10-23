package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestChecklistWindowDefaultsAndWindowsTitlebar(t *testing.T) {
	for _, platform := range []string{"windows", "darwin"} {
		opts := mainWindowOptions(platform)
		if opts.Width != 480 || opts.Height != 720 || opts.MinWidth != 400 || opts.MinHeight != 440 || opts.URL != "/?window=main" {
			t.Fatalf("unexpected checklist window defaults: %+v", opts)
		}
		if opts.MinimiseButtonState != application.ButtonHidden || opts.CloseButtonState != application.ButtonHidden || opts.MaximiseButtonState != application.ButtonHidden {
			t.Fatal("native buttons must be hidden")
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
