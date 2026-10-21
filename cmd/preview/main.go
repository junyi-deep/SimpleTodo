// Development-only browser harness. The desktop app does not start an HTTP API.
package main

import (
	"changeme/internal/todo"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"
)

func main() {
	dir := flag.String("data-dir", "", "isolated test data directory")
	flag.Parse()
	if *dir == "" {
		tmp, err := os.MkdirTemp("", "simpletodo-preview-")
		if err != nil {
			log.Fatal(err)
		}
		*dir = tmp
		defer os.RemoveAll(tmp)
	}
	e, err := todo.Open(*dir)
	if err != nil {
		log.Fatal(err)
	}
	defer e.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go e.Run(ctx)
	methods := map[string]any{"Query": e.Query, "SetTimezone": e.SetTimezone, "SeriesImpact": e.SeriesImpact, "FindTask": e.FindTask, "ReadTree": e.ReadTree, "Duplicate": e.Duplicate, "Create": e.Create, "Update": func(c todo.Command) ([]todo.Task, error) {
		out, err := e.Update(c)
		if err == nil && c.RepeatRule != nil {
			if er := e.EnsureThrough(time.Now().In(mustZone(e.Timezone())).AddDate(0, 0, 7).Format("2006-01-02")); er != nil {
				log.Print(er)
			}
		}
		return out, err
	}, "Complete": e.Complete, "Delete": e.Delete, "Move": e.Move, "Reparent": e.Reparent, "Reorder": e.ReorderRoots, "Repeat": func(r todo.RepeatRequest) error {
		if err := e.Repeat(r); err != nil {
			return err
		}
		return e.EnsureThrough(time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
	}, "Rules": e.Rules, "StopFollowing": e.StopFollowing, "Settings": e.SaveSettings, "Focus": e.FocusAction, "SaveCountdown": e.SaveCountdown, "ToggleCountdown": e.ToggleCountdown, "DeleteCountdown": e.DeleteCountdown, "DismissNotices": e.DismissNotices, "Holidays": e.Holidays, "PreviewImport": e.PreviewImport, "CommitImport": e.CommitImport, "PreviewBatch": e.PreviewBatch, "CommitBatch": e.CommitBatch, "PreviewPaste": e.PreviewPaste, "CopyTasks": e.CopyTasks, "Export": e.Export, "Backups": e.Backups, "Backup": e.CreateBackup, "Restore": e.Restore, "Permission": func() (bool, error) {
		return false, fmt.Errorf("浏览器预览不申请系统通知权限，请运行桌面应用")
	}, "Boot": func() (map[string]any, error) {
		settings, err := e.Settings()
		if err != nil {
			return nil, err
		}
		f, err := e.ActiveFocus()
		if err != nil {
			return nil, err
		}
		cs, err := e.Countdowns()
		if err != nil {
			return nil, err
		}
		ns, err := e.Notices()
		if err != nil {
			return nil, err
		}
		ws, err := e.Windows()
		return map[string]any{"settings": settings, "version": "0.1.1", "dataDir": e.Dir, "today": time.Now().Format("2006-01-02"), "focus": f, "timezone": e.Timezone(), "countdowns": cs, "windows": ws, "notices": ns, "warnings": []string{}, "autostart": false, "notificationAllowed": false, "capabilities": map[string]any{"desktopMode": false, "edgeHide": false, "platform": "preview"}}, err
	}}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "http://127.0.0.1:9245" && origin != "http://localhost:9245" {
			http.Error(w, "invalid origin", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		fn, ok := methods[strings.TrimPrefix(r.URL.Path, "/api/")]
		if !ok {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "此功能需要桌面环境"})
			return
		}
		v := reflect.ValueOf(fn)
		var args []json.RawMessage
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 30*1024*1024)).Decode(&args); err != nil || len(args) != v.Type().NumIn() {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "参数无效"})
			return
		}
		in := []reflect.Value{}
		for i, a := range args {
			p := reflect.New(v.Type().In(i))
			if err := json.Unmarshal(a, p.Interface()); err != nil {
				w.WriteHeader(400)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			in = append(in, p.Elem())
		}
		out := v.Call(in)
		var result any
		for _, x := range out {
			if x.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if !x.IsNil() {
					w.WriteHeader(400)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": x.Interface().(error).Error()})
					return
				}
			} else {
				result = x.Interface()
			}
		}
		_ = json.NewEncoder(w).Encode(result)
	})
	server := &http.Server{Addr: "127.0.0.1:9246", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { <-ctx.Done(); _ = server.Close() }()
	log.Println("Isolated preview API: http://127.0.0.1:9246")
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func mustZone(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}
