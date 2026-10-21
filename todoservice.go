package main

import (
	"changeme/internal/desktop"
	"changeme/internal/todo"
	"context"
	"encoding/json"
	"fmt"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type TodoService struct {
	forceQuit   atomic.Bool
	maintenance atomic.Bool
	confirming  atomic.Int32
	viewMode    string
	listBounds  application.Rect
	edgeEditing bool
	edgeTicks   int
	shortcutMu  sync.Mutex
	engine      atomic.Pointer[todo.Engine]
	startupErr  atomic.Pointer[string]
	dataPath    string
	ctx         context.Context
	app         *application.App
	ns          *notifications.NotificationService
	mu          sync.Mutex
	windows     map[string]*application.WebviewWindow
	shortcuts   map[string]string
	bootstrap   string
	warnings    []string
	edgeHidden  bool
	edgeBounds  application.Rect
	edgeSince   time.Time
	edgeBusy    bool
	modalBusy   bool
	menuBusy    bool
}

func (s *TodoService) Boot() (map[string]any, error) {
	if s.core() == nil {
		msg := "数据库尚未打开"
		if p := s.startupErr.Load(); p != nil {
			msg = *p
		}
		return map[string]any{"startupError": msg, "settings": map[string]string{"theme": "system", "onboarding": "true"}, "version": "0.1.1", "dataDir": s.dataPath, "today": time.Now().Format("2006-01-02"), "windows": []todo.WindowState{}, "notices": []todo.Notice{}, "countdowns": []todo.Countdown{}, "warnings": s.warnings, "capabilities": map[string]any{"platform": runtime.GOOS}}, nil
	}

	settings, err := s.core().Settings()
	if err != nil {
		return nil, err
	}
	f, err := s.core().ActiveFocus()
	if err != nil {
		return nil, err
	}
	cs, err := s.core().Countdowns()
	if err != nil {
		return nil, err
	}
	ws, err := s.core().Windows()
	if err != nil {
		return nil, err
	}
	ns, err := s.core().Notices()
	if err != nil {
		return nil, err
	}
	allowed, _ := s.ns.CheckNotificationAuthorization()
	auto, _ := s.app.Autostart.IsEnabled()
	return map[string]any{"settings": settings, "version": "0.1.1", "dataDir": s.core().DataDir(), "today": time.Now().In(mustLocation(s.core().Timezone())).Format("2006-01-02"), "timezone": s.core().Timezone(), "focus": f, "countdowns": cs, "windows": ws, "notices": ns, "warnings": s.warnings, "autostart": auto, "notificationAllowed": allowed, "capabilities": map[string]any{"desktopMode": desktop.Supported(), "edgeHide": desktop.Supported(), "platform": runtime.GOOS}}, nil
}
func (s *TodoService) Query(q todo.Query) (todo.Snapshot, error) {
	if s.core() == nil {
		return todo.Snapshot{}, fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	return s.core().Query(q)
}
func (s *TodoService) Create(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	return s.core().Create(c)
}
func (s *TodoService) Update(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	out, err := s.core().Update(c)
	if err == nil && c.RepeatRule != nil {
		if er := s.core().EnsureThrough(time.Now().AddDate(0, 0, 7).Format("2006-01-02")); er != nil {
			s.app.Event.Emit("todo:error", er.Error())
		}
	}
	return out, err
}
func (s *TodoService) Complete(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	return s.core().Complete(c)
}
func (s *TodoService) Delete(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	return s.core().Delete(c)
}
func (s *TodoService) Move(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	return s.core().Move(c)
}
func (s *TodoService) Reparent(c todo.Command) ([]todo.Task, error) {
	if s.maintenance.Load() {
		return nil, fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	return s.core().Reparent(c)
}
func (s *TodoService) Reorder(ids []string, revision string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().ReorderRoots(ids, revision)
}
func (s *TodoService) Repeat(r todo.RepeatRequest) error {
	if err := s.ready(); err != nil {
		return err
	}
	if err := s.core().Repeat(r); err != nil {
		return err
	}
	return s.core().EnsureThrough(time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
}
func (s *TodoService) Rules() ([]todo.Rule, error) { return s.core().Rules() }
func (s *TodoService) StopFollowing(c todo.Command) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().StopFollowing(c)
}
func (s *TodoService) Settings(m map[string]string) error {
	if err := s.ready(); err != nil {
		return err
	}
	err := s.core().SaveSettings(m)
	if err == nil && m["edgeHide"] == "false" {
		s.mu.Lock()
		hidden := s.edgeHidden
		b := s.edgeBounds
		s.edgeHidden = false
		s.mu.Unlock()
		if hidden {
			s.windows["mini"].SetBounds(b)
		}
	}
	return err
}
func (s *TodoService) Focus(action string, minutes int, id *string) (*todo.Focus, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().FocusAction(action, minutes, id)
}
func (s *TodoService) SaveCountdown(c todo.Countdown) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().SaveCountdown(c)
}
func (s *TodoService) DeleteCountdown(id string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().DeleteCountdown(id)
}
func (s *TodoService) Permission() (bool, error)         { return s.ns.RequestNotificationAuthorization() }
func (s *TodoService) DismissNotices() error             { return s.core().DismissNotices() }
func (s *TodoService) Holidays() ([]todo.Holiday, error) { return s.core().Holidays() }
func (s *TodoService) PreviewImport(content, format string, date *string, replace bool) (todo.Preview, error) {
	return s.core().PreviewImport(content, format, date, replace)
}
func (s *TodoService) CommitImport(token string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().CommitImport(token)
}
func (s *TodoService) PreviewBatch(ids []string, action, revision string) (todo.Preview, error) {
	return s.core().PreviewBatch(ids, action, revision)
}
func (s *TodoService) CommitBatch(token string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().CommitBatch(token)
}
func (s *TodoService) Export(format string, q todo.Query, inbox bool, ids []string) (string, error) {
	return s.core().Export(format, q, inbox, ids)
}
func (s *TodoService) SaveExport(format string, q todo.Query, inbox bool, ids []string) (string, error) {
	text, err := s.core().Export(format, q, inbox, ids)
	if err != nil {
		return "", err
	}
	ext := format
	if ext == "markdown" {
		ext = "md"
	}
	if ext == "text" {
		ext = "txt"
	}
	path, err := s.app.Dialog.SaveFile().SetFilename("SimpleTodo-" + time.Now().Format("20060102") + "." + ext).PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	return path, os.WriteFile(path, []byte(text), 0600)
}
func (s *TodoService) PickImport() (map[string]string, error) {
	path, err := s.app.Dialog.OpenFile().SetTitle("选择待办文件").AddFilter("待办文件", "*.json;*.md;*.txt").PromptForSingleSelection()
	if err != nil || path == "" {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 20*1024*1024 {
		return nil, fmt.Errorf("文件超过 20 MiB")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	format := "text"
	if strings.HasSuffix(strings.ToLower(path), ".json") {
		format = "json"
	} else if strings.HasSuffix(strings.ToLower(path), ".md") {
		format = "markdown"
	}
	return map[string]string{"content": string(b), "format": format}, nil
}
func (s *TodoService) ClipboardRead() (string, error) {
	text, ok := s.app.Clipboard.Text()
	if !ok {
		return "", fmt.Errorf("无法读取剪贴板文本，请复制文本后重试")
	}
	return text, nil
}
func (s *TodoService) ClipboardWrite(text string) error {
	if !s.app.Clipboard.SetText(text) {
		return fmt.Errorf("无法写入剪贴板")
	}
	return nil
}
func (s *TodoService) ImportHolidays() error {
	if err := s.ready(); err != nil {
		return err
	}
	path, err := s.app.Dialog.OpenFile().SetTitle("导入节假日 JSON").AddFilter("JSON", "*.json").PromptForSingleSelection()
	if err != nil || path == "" {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > 2*1024*1024 {
		return fmt.Errorf("文件超过 2 MiB")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return s.core().SaveHolidayData(string(b))
}
func (s *TodoService) Backup() (todo.Backup, error)    { return s.core().CreateBackup() }
func (s *TodoService) Backups() ([]todo.Backup, error) { return s.core().Backups() }
func (s *TodoService) Restore(name string) error {
	if name == "" {
		path, err := s.app.Dialog.OpenFile().SetTitle("恢复 SimpleTodo 数据库备份").AddFilter("SQLite 备份", "*.db").PromptForSingleSelection()
		if err != nil || path == "" {
			return err
		}
		if s.core() == nil {
			if err = todo.RecoverDatabase(s.dataPath, path); err != nil {
				return err
			}
			e, err := todo.Open(s.dataPath)
			if err != nil {
				return err
			}
			s.installEngine(e)
			s.app.Event.Emit("todo:changed")
			return nil
		}
		return s.core().RestoreFile(path)
	}
	return s.core().Restore(name)
}
func (s *TodoService) Migrate() error {
	dir, err := s.app.Dialog.OpenFile().SetTitle("迁移数据到本地目录").CanChooseFiles(false).CanChooseDirectories(true).PromptForSingleSelection()
	if err != nil || dir == "" {
		return err
	}
	if !s.maintenance.CompareAndSwap(false, true) {
		return fmt.Errorf("MAINTENANCE: 正在维护数据")
	}
	s.app.Event.Emit("todo:maintenance", true)
	defer func() { s.maintenance.Store(false); s.app.Event.Emit("todo:maintenance", false) }()
	return s.core().MigrateWithCommit(dir, func(next string) error {
		b, _ := json.Marshal(map[string]string{"dataDir": next})
		tmp := s.bootstrap + ".tmp"
		if err := os.WriteFile(tmp, b, 0600); err != nil {
			return err
		}
		return os.Rename(tmp, s.bootstrap)
	})
}
func (s *TodoService) OpenDataDir() error {
	dir := s.dataPath
	if s.core() != nil {
		dir = s.core().DataDir()
	}
	return s.app.Browser.OpenFile(dir)
}
func (s *TodoService) CheckUpdate() error {
	m, err := s.core().Settings()
	if err != nil {
		return err
	}
	url := m["releaseURL"]
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("尚未配置有效的 HTTPS 项目发布页")
	}
	return s.app.Browser.OpenURL(url)
}
func (s *TodoService) Autostart(enabled bool) error {
	if enabled {
		return s.app.Autostart.EnableWithOptions(application.AutostartOptions{Arguments: []string{"--background"}})
	}
	return s.app.Autostart.Disable()
}
func (s *TodoService) Shortcut(action, key string) error {
	s.shortcutMu.Lock()
	defer s.shortcutMu.Unlock()
	if !containsAction(action) {
		return fmt.Errorf("未知快捷键操作")
	}
	old := s.shortcuts[action]
	if old == key {
		return nil
	}
	if key != "" {
		err := s.app.GlobalShortcut.Register(key, func() {
			switch action {
			case "main":
				s.ToggleWindow("main")
			case "mini":
				s.ToggleWindow("mini")
			case "quick":
				s.ShowWindow("quick")
			}
		})
		if err != nil {
			return err
		}
	}
	if err := s.core().SaveSettings(map[string]string{"shortcut" + strings.Title(action): key}); err != nil {
		if key != "" {
			_ = s.app.GlobalShortcut.Unregister(key)
		}
		return err
	}
	if old != "" {
		_ = s.app.GlobalShortcut.Unregister(old)
	}
	s.shortcuts[action] = key
	return nil
}
func containsAction(v string) bool { return v == "main" || v == "mini" || v == "quick" }
func (s *TodoService) ShowWindow(key string) error {
	w := s.windows["main"]
	if w == nil {
		return fmt.Errorf("窗口尚未就绪")
	}
	w.Show()
	w.Focus()
	switch key {
	case "overview":
		s.app.Event.Emit("todo:navigate", "calendar")
		return s.SetView("calendar")
	case "quick":
		if err := s.SetView("tasks"); err != nil {
			return err
		}
		s.app.Event.Emit("todo:navigate", "quick")
	}
	return nil
}

// Every view uses the same native window. Only its dimensions change.
func (s *TodoService) SetView(view string) error {
	if view != "tasks" && view != "calendar" && view != "focus" {
		return fmt.Errorf("未知视图")
	}
	w := s.windows["main"]
	if w == nil {
		return fmt.Errorf("窗口尚未就绪")
	}
	s.mu.Lock()
	old := s.viewMode
	s.viewMode = view
	if view == "calendar" && old != "calendar" {
		s.listBounds = w.Bounds()
	}
	restore := s.listBounds
	s.mu.Unlock()
	application.InvokeSync(func() {
		var area application.Rect
		// ScreenManager exposes DIP work areas; GetScreen on macOS returns
		// unnormalised native values on Retina displays in this Wails version.
		if screen := s.app.Screen.ScreenNearestDipRect(w.Bounds()); screen != nil {
			area = screen.WorkArea
		}
		minW, minH := 440, 380
		if view == "calendar" {
			minW, minH = 760, 500
		}
		if area.Width > 0 && area.Height > 0 {
			minW, minH = min(minW, area.Width), min(minH, area.Height)
		}
		w.SetMinSize(minW, minH)
		if view == "calendar" {
			if !w.IsMaximised() {
				bounds := w.Bounds()
				bounds.Width, bounds.Height = 1080, 720
				w.SetBounds(fitWindowBounds(bounds, area))
			}
		} else {
			if old == "calendar" && !w.IsMaximised() {
				if restore.Width < 440 {
					restore = w.Bounds()
					restore.Width, restore.Height = 600, 580
				}
				w.SetBounds(fitWindowBounds(restore, area))
			}
		}
	})
	return nil
}
func (s *TodoService) WindowAction(action string) error {
	w := s.windows["main"]
	if w == nil {
		return fmt.Errorf("窗口尚未就绪")
	}
	switch action {
	case "minimize":
		w.Minimise()
	case "maximize":
		w.ToggleMaximise()
	default:
		return fmt.Errorf("未知窗口操作")
	}
	return nil
}
func (s *TodoService) ToggleWindow(key string) {
	s.mu.Lock()
	w := s.windows[key]
	s.mu.Unlock()
	if w != nil {
		if w.IsVisible() {
			w.Hide()
		} else {
			_ = s.ShowWindow(key)
		}
	}
}
func (s *TodoService) HideWindow(key string) error {
	if w := s.windows[key]; w != nil {
		w.Hide()
		return nil
	}
	return fmt.Errorf("未知窗口")
}

// Confirm uses a native panel: WKWebView's browser confirm() is unavailable.
// No affirmative default: Escape cancels without mutating application data.
func (s *TodoService) Confirm(key, message string) (bool, error) {
	w := s.windows[key]
	if w == nil || strings.TrimSpace(message) == "" || len(message) > 20000 {
		return false, fmt.Errorf("无效确认请求")
	}
	answer := make(chan bool, 1)
	s.confirming.Add(1)
	defer s.confirming.Add(-1)
	d := s.app.Dialog.Question().SetTitle("SimpleTodo").SetMessage(message).AttachToWindow(w)
	// Wails maps a default button to Return before checking its cancel flag on macOS.
	// Keep cancellation separate so Escape remains available and Return cannot approve.
	d.AddButton("取消").SetAsCancel().OnClick(func() { answer <- false })
	d.AddButton("确认").OnClick(func() { answer <- true })
	d.Show()
	select {
	case ok := <-answer:
		return ok, nil
	case <-s.ctx.Done():
		return false, nil
	}
}
func (s *TodoService) WindowConfig(key, mode string, opacity float64, collapsed bool) error {
	w := s.windows[key]
	if w == nil {
		return fmt.Errorf("未知窗口")
	}
	if opacity < 0.4 || opacity > 1 {
		return fmt.Errorf("背景不透明度为 40%%—100%%")
	}
	if mode != "normal" && mode != "top" && mode != "desktop" {
		return fmt.Errorf("无效窗口模式")
	}
	if err := application.InvokeSyncWithError(func() error { return desktop.SetLevel(w.NativeWindow(), mode) }); err != nil {
		return err
	}
	bounds := w.Bounds()
	if key == "mini" {
		if collapsed {
			bounds.Height = 62
		} else if bounds.Height < 160 {
			bounds.Height = 420
		}
		w.SetBounds(bounds)
	}
	state := todo.WindowState{Key: key, X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: bounds.Height, Mode: mode, Opacity: opacity, Collapsed: collapsed}
	if err := s.core().SaveWindow(state); err != nil {
		return err
	}
	s.app.Event.Emit("todo:changed")
	return nil
}
func (s *TodoService) ResetWindows() error {
	for key, w := range s.windows {
		w.Center()
		if key == "mini" {
			s.mu.Lock()
			s.edgeHidden = false
			s.mu.Unlock()
			w.SetSize(300, 420)
			states, _ := s.core().Windows()
			for _, state := range states {
				if state.Key == "mini" {
					b := w.Bounds()
					state.X = b.X
					state.Y = b.Y
					state.Width = 300
					state.Height = 420
					state.Collapsed = false
					_ = s.core().SaveWindow(state)
				}
			}
		}
		s.savePosition(key, w)
	}
	s.app.Event.Emit("todo:changed")
	return nil
}
func (s *TodoService) SetEdgeBusy(busy bool) { s.mu.Lock(); s.edgeBusy = busy; s.mu.Unlock() }
func (s *TodoService) edgeTick() {
	if s.core() == nil {
		return
	}
	s.ensureVisibleWindows()
	settings, err := s.core().Settings()
	if err != nil || settings["edgeHide"] != "true" || !desktop.Supported() {
		return
	}
	application.InvokeSync(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		w := s.windows["mini"]
		if w == nil || !w.IsVisible() || s.modalBusy || s.menuBusy || s.edgeEditing || s.confirming.Load() > 0 {
			return
		}
		if s.edgeBusy && desktop.MouseDown() {
			return
		}
		s.edgeBusy = false
		bounds := w.Bounds()
		screen, err := w.GetScreen()
		if err != nil || screen == nil {
			return
		}
		area := screen.WorkArea
		x, y := desktop.Pointer()
		if runtime.GOOS == "windows" {
			point := s.app.Screen.PhysicalToDipPoint(application.Point{X: x, Y: y})
			x, y = point.X, point.Y
		}
		inside := x >= bounds.X && x < bounds.X+bounds.Width && y >= bounds.Y && y < bounds.Y+bounds.Height
		if s.edgeHidden {
			if inside {
				if s.edgeSince.IsZero() {
					s.edgeSince = time.Now()
				}
				if time.Since(s.edgeSince) > 200*time.Millisecond {
					w.SetBounds(s.edgeBounds)
					s.edgeHidden = false
					s.edgeSince = time.Time{}
				}
			} else {
				s.edgeSince = time.Time{}
			}
			return
		}
		if inside {
			s.edgeSince = time.Time{}
			return
		}
		top := bounds.Y <= area.Y+12
		bottom := bounds.Y+bounds.Height >= area.Y+area.Height-12
		left := bounds.X <= area.X+12
		right := bounds.X+bounds.Width >= area.X+area.Width-12
		if !left && !right && !top && !bottom {
			return
		}
		if s.edgeSince.IsZero() {
			s.edgeSince = time.Now()
		}
		if time.Since(s.edgeSince) < 800*time.Millisecond {
			return
		}
		s.edgeBounds = bounds
		if left {
			bounds.X = area.X - bounds.Width + 6
		} else if right {
			bounds.X = area.X + area.Width - 6
		} else if top {
			bounds.Y = area.Y - bounds.Height + 6
		} else {
			bounds.Y = area.Y + area.Height - 6
		}
		w.SetBounds(bounds)
		s.edgeHidden = true
		s.edgeSince = time.Time{}
	})
}
func (s *TodoService) savePosition(key string, w *application.WebviewWindow) {
	if s.core() == nil {
		return
	}
	s.mu.Lock()
	if key == "mini" && s.edgeHidden {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	b := w.Bounds()
	_ = s.core().SaveWindowBounds(todo.WindowState{Key: key, X: b.X, Y: b.Y, Width: b.Width, Height: b.Height})
}
func (s *TodoService) setupWindows(background bool) {
	states := []todo.WindowState{}
	if s.core() != nil {
		states, _ = s.core().Windows()
	}
	saved := map[string]todo.WindowState{}
	for _, w := range states {
		saved[w.Key] = w
	}
	for _, key := range []string{"main"} {
		width, height, minW, minH := 600, 580, 440, 380
		title := "SimpleTodo"
		options := application.WebviewWindowOptions{Name: key, Title: title, Width: width, Height: height, MinWidth: minW, MinHeight: minH, Hidden: key != "main" || background, AlwaysOnTop: false, URL: "/?window=" + key, BackgroundColour: application.NewRGBA(255, 255, 255, 255), Mac: application.MacWindow{TitleBar: application.MacTitleBarHiddenInset, InvisibleTitleBarHeight: 36}}
		if state, ok := saved[key]; ok {
			options.Width = max(440, state.Width)
			options.Height = max(380, state.Height)
			options.X = state.X
			options.Y = state.Y
			visible := false
			for _, screen := range s.app.Screen.GetAll() {
				a := screen.WorkArea
				if state.X+state.Width > a.X && state.X < a.X+a.Width && state.Y+state.Height > a.Y && state.Y < a.Y+a.Height {
					visible = true
				}
			}
			if visible {
				options.InitialPosition = application.WindowXY
			}
			options.AlwaysOnTop = state.Mode == "top"
		}
		w := s.app.Window.NewWithOptions(options)
		s.windows[key] = w
		windowKey := key
		w.RegisterHook(events.Common.WindowClosing, func(ev *application.WindowEvent) { ev.Cancel(); w.Hide(); go s.savePosition(windowKey, w) })
		w.OnWindowEvent(events.Common.WindowDidMove, func(ev *application.WindowEvent) { go s.savePosition(windowKey, w) })
		w.OnWindowEvent(events.Common.WindowDidResize, func(ev *application.WindowEvent) { go s.savePosition(windowKey, w) })
		if state, ok := saved[key]; ok && state.Mode == "desktop" {
			w.OnWindowEvent(events.Common.WindowRuntimeReady, func(ev *application.WindowEvent) {
				application.InvokeAsync(func() {
					if err := desktop.SetLevel(w.NativeWindow(), "desktop"); err != nil {
						s.app.Event.Emit("todo:error", err.Error())
					}
				})
			})
		}
	}
}
func (s *TodoService) Quit()        { s.forceQuit.Store(true); s.app.Quit() }
func (s *TodoService) requestQuit() { _ = s.ShowWindow("main"); s.app.Event.Emit("todo:quit-request") }

func (s *TodoService) CopyTasks(q todo.Query, ids []string) (string, error) {
	return s.core().CopyTasks(q, ids)
}
func (s *TodoService) PreviewPaste(content, target string) (todo.Preview, error) {
	return s.core().PreviewPaste(content, target)
}
func (s *TodoService) ToggleCountdown(id string, enabled bool) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().ToggleCountdown(id, enabled)
}
func (s *TodoService) FindTask(id string) (todo.Task, error) { return s.core().FindTask(id) }
func (s *TodoService) Duplicate(c todo.Command, date *string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().Duplicate(c, date)
}

func (s *TodoService) SetEdgeModalBusy(busy bool) { s.mu.Lock(); s.modalBusy = busy; s.mu.Unlock() }
func (s *TodoService) SetEdgeMenuBusy(busy bool)  { s.mu.Lock(); s.menuBusy = busy; s.mu.Unlock() }
func (s *TodoService) OpenNotificationSettings() error {
	if runtime.GOOS == "darwin" {
		return s.app.Browser.OpenURL("x-apple.systempreferences:com.apple.preference.notifications")
	}
	if runtime.GOOS == "windows" {
		return s.app.Browser.OpenURL("ms-settings:notifications")
	}
	return fmt.Errorf("当前平台请手动打开系统通知设置")
}

func (s *TodoService) core() *todo.Engine { return s.engine.Load() }
func (s *TodoService) installEngine(e *todo.Engine) {
	e.Changed = func() { s.app.Event.Emit("todo:changed") }
	e.Report = func(err error) { s.app.Event.Emit("todo:error", err.Error()) }
	e.Notify = func(n todo.Notice) error {
		return s.ns.SendNotification(notifications.NotificationOptions{ID: n.ID, Title: "SimpleTodo", Body: n.Title, Sound: &notifications.NotificationSound{Silent: n.Silent}, Data: map[string]interface{}{"sourceId": n.SourceID, "sourceType": n.SourceType}})
	}
	s.engine.Store(e)
	s.startupErr.Store(nil)
	go e.Run(s.ctx)
}

func (s *TodoService) SeriesImpact(c todo.Command) (todo.SeriesImpact, error) {
	return s.core().SeriesImpact(c)
}
func (s *TodoService) SetTimezone(name string) (int, error) { return s.core().SetTimezone(name) }

func mustLocation(name string) *time.Location {
	l, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return l
}

func (s *TodoService) SetEdgeEditing(v bool) { s.mu.Lock(); s.edgeEditing = v; s.mu.Unlock() }
func (s *TodoService) ensureVisibleWindows() {
	application.InvokeSync(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.edgeTicks++
		if s.edgeTicks%10 != 0 {
			return
		}
		screens := s.app.Screen.GetAll()
		if len(screens) == 0 {
			return
		}
		for key, w := range s.windows {
			b := w.Bounds()
			if key == "mini" && s.edgeHidden {
				b = s.edgeBounds
			}
			visible := false
			for _, screen := range screens {
				a := screen.WorkArea
				if b.X+b.Width > a.X+12 && b.X < a.X+a.Width-12 && b.Y+b.Height > a.Y+12 && b.Y < a.Y+a.Height-12 {
					visible = true
					break
				}
			}
			if !visible {
				if key == "mini" {
					s.edgeHidden = false
				}
				w.Center()
			}
		}
	})
}
func (s *TodoService) QuitStatus() (bool, error) {
	if s.core() == nil {
		return false, nil
	}
	return s.core().QuitStatus()
}
func (s *TodoService) PrepareRestore(name string) (*todo.BackupInfo, error) {
	var path string
	if name == "" {
		var err error
		path, err = s.app.Dialog.OpenFile().SetTitle("选择完整 SQLite 备份").AddFilter("SQLite 备份", "*.db").PromptForSingleSelection()
		if err != nil || path == "" {
			return nil, err
		}
	} else {
		if filepath.Base(name) != name {
			return nil, fmt.Errorf("无效备份名称")
		}
		path = filepath.Join(s.core().DataDir(), "backups", name)
	}
	info, err := todo.InspectBackup(path)
	return &info, err
}
func (s *TodoService) RestorePrepared(path string) error {
	if !s.maintenance.CompareAndSwap(false, true) {
		return fmt.Errorf("MAINTENANCE: 正在维护数据")
	}
	s.app.Event.Emit("todo:maintenance", true)
	defer func() { s.maintenance.Store(false); s.app.Event.Emit("todo:maintenance", false) }()
	if s.core() == nil {
		if err := todo.RecoverDatabase(s.dataPath, path); err != nil {
			return err
		}
		e, err := todo.Open(s.dataPath)
		if err != nil {
			return err
		}
		s.installEngine(e)
		return nil
	}
	return s.core().RestoreFile(path)
}

func (s *TodoService) ready() error {
	if s.maintenance.Load() {
		return fmt.Errorf("MAINTENANCE: 正在维护数据，请稍后重试")
	}
	if s.core() == nil {
		return fmt.Errorf("RECOVERY_REQUIRED: 请先恢复数据库")
	}
	return nil
}

func (s *TodoService) ReadTree(id string) ([]todo.Task, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().ReadTree(id)
}
