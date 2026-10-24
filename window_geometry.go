package main

import "github.com/wailsapp/wails/v3/pkg/application"

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
