package main

import (
	"changeme/internal/todo"
	"context"
	"embed"
	"encoding/json"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

//go:embed build/trayicon.png
var templateTrayIcon []byte

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ns := notifications.New()
	s := &TodoService{ns: ns, windows: map[string]*application.WebviewWindow{}, shortcuts: map[string]string{}, warnings: []string{}}
	app := application.New(application.Options{Name: "SimpleTodo", Description: "本地优先的个人待办与专注工具", OnShutdown: cancel, ShouldQuit: func() bool {
		if s.forceQuit.Load() {
			return true
		}
		s.requestQuit()
		return false
	}, Services: []application.Service{application.NewService(s), application.NewService(ns)}, Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)}, Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: false}, SingleInstance: &application.SingleInstanceOptions{UniqueID: "com.junyi.simpletodo", OnSecondInstanceLaunch: func(data application.SecondInstanceData) { _ = s.ShowWindow("main") }}})
	s.app = app
	s.ctx = ctx
	base, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" && os.Getenv("LOCALAPPDATA") != "" {
		base = os.Getenv("LOCALAPPDATA")
	}
	base = filepath.Join(base, "SimpleTodo")
	if err = os.MkdirAll(base, 0700); err != nil {
		log.Fatal(err)
	}
	s.bootstrap = filepath.Join(base, "bootstrap.json")
	dir := base
	if custom := os.Getenv("SIMPLETODO_DATA_DIR"); custom != "" {
		dir = custom
	} else if b, er := os.ReadFile(s.bootstrap); er == nil {
		var c map[string]string
		if json.Unmarshal(b, &c) == nil && c["dataDir"] != "" {
			dir = c["dataDir"]
		}
	}
	s.dataPath = dir
	engine, err := todo.Open(dir)
	if err != nil {
		msg := "无法打开原数据库：" + err.Error()
		s.startupErr.Store(&msg)
	} else {
		if err = engine.DailyBackup(); err != nil {
			s.warnings = append(s.warnings, "每日备份失败："+err.Error())
		}
		s.installEngine(engine)
	}
	defer func() {
		if e := s.core(); e != nil {
			_ = e.Close()
		}
	}()

	// Notification callbacks never mutate task state implicitly.
	ns.OnNotificationResponse(func(result notifications.NotificationResult) {
		_ = s.ShowWindow("main")
		app.Event.Emit("todo:notification", result.Response.UserInfo)
	})
	background := false
	for _, arg := range os.Args[1:] {
		if arg == "--background" {
			background = true
		}
	}
	s.setupWindows(background)
	tray := app.SystemTray.New()
	tray.SetIcon(trayIcon)
	tray.SetTooltip("SimpleTodo · 安排好每一天")
	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(templateTrayIcon)
	}
	menu := application.NewMenu()
	for _, item := range []struct{ Label, Window string }{{"打开 SimpleTodo", "main"}, {"快速记录", "quick"}, {"日历视图", "overview"}} {
		key := item.Window
		menu.Add(item.Label).OnClick(func(ctx *application.Context) { _ = s.ShowWindow(key) })
	}
	menu.Add("专注时钟").OnClick(func(ctx *application.Context) { _ = s.ShowWindow("main"); app.Event.Emit("todo:navigate", "focus") })
	menu.AddSeparator()
	menu.Add("退出 SimpleTodo").OnClick(func(ctx *application.Context) { s.requestQuit() })
	tray.SetMenu(menu)
	tray.OnClick(func() { s.ToggleWindow("main") })
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.edgeTick()
			}
		}
	}()
	// Register shortcuts after startup so registration failures are observable.
	go func() {
		time.Sleep(time.Second)
		if s.core() == nil {
			return
		}
		m, er := s.core().Settings()
		if er != nil {
			return
		}
		if m["autostartConfigured"] != "true" {
			if er = s.Autostart(true); er != nil {
				app.Event.Emit("todo:error", "登录启动未能开启："+er.Error())
			}
		}
		for _, action := range []string{"main", "quick"} {
			if key := m["shortcut"+map[string]string{"main": "Main", "mini": "Mini", "quick": "Quick"}[action]]; key != "" {
				if er = s.Shortcut(action, key); er != nil {
					app.Event.Emit("todo:error", er.Error())
				}
			}
		}
	}()
	if err = app.Run(); err != nil {
		log.Print(err)
	}
	cancel()
}
