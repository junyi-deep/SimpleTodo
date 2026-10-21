package main

import "github.com/wailsapp/wails/v3/pkg/application"

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
