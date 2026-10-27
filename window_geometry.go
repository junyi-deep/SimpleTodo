package main

import (
	"changeme/internal/todo"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func viewMinimum(view string) (int, int) {
	if view == "calendar" {
		return 760, 500
	}
	return 400, 440
}

func viewBounds(state todo.ViewWindowState) application.Rect {
	return application.Rect{X: state.X, Y: state.Y, Width: state.Width, Height: state.Height}
}

func capturedViewWindow(view string, saved map[string]todo.ViewWindowState, current application.Rect, maximised bool) todo.ViewWindowState {
	bounds := current
	if maximised {
		// A maximise event reports the display size, not the last normal rectangle.
		bounds = restoredViewBounds(view, saved, current, application.Rect{})
	}
	return todo.ViewWindowState{View: view, X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: bounds.Height, Maximised: maximised}
}

// An unseen view starts at the current position, with its own suitable size.
func restoredViewBounds(view string, saved map[string]todo.ViewWindowState, current, area application.Rect) application.Rect {
	bounds := current
	bounds.Width, bounds.Height = 480, 720
	if view == "calendar" {
		bounds.Width = 1080
	}
	if state, ok := saved[view]; ok {
		bounds = viewBounds(state)
	}
	minW, minH := viewMinimum(view)
	bounds.Width, bounds.Height = max(minW, bounds.Width), max(minH, bounds.Height)
	return fitWindowBounds(bounds, area)
}

func mainWindowOptions(platform string) application.WebviewWindowOptions {
	return application.WebviewWindowOptions{Name: "main", Title: "SimpleTodo", Width: 480, Height: 720, MinWidth: 400, MinHeight: 440, URL: "/?window=main", MinimiseButtonState: application.ButtonHidden, MaximiseButtonState: application.ButtonHidden, CloseButtonState: application.ButtonHidden, Frameless: platform == "windows", BackgroundType: application.BackgroundTypeTransparent, BackgroundColour: application.NewRGBA(0, 0, 0, 0), Mac: application.MacWindow{Backdrop: application.MacBackdropTransparent, TitleBar: application.MacTitleBarHiddenInset, InvisibleTitleBarHeight: 36}}
}

// Keep enlarged views entirely inside the active display, including displays
// positioned to the left or above the primary display.
func fitWindowBounds(bounds, area application.Rect) application.Rect {
	if area.Width <= 0 || area.Height <= 0 {
		return bounds
	}
	bounds.Width = min(bounds.Width, area.Width)
	bounds.Height = min(bounds.Height, area.Height)
	bounds.X = max(area.X, min(bounds.X, area.X+area.Width-bounds.Width))
	bounds.Y = max(area.Y, min(bounds.Y, area.Y+area.Height-bounds.Height))
	return bounds
}
