package todo

import (
	"database/sql"
	"os"
	"path/filepath"
)

type BackupInfo struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	Roots     int    `json:"roots"`
	Nodes     int    `json:"nodes"`
}

func InspectBackup(path string) (BackupInfo, error) {
	var info BackupInfo
	if err := checkBackup(path); err != nil {
		return info, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return info, err
	}
	info = BackupInfo{Path: path, Name: filepath.Base(path), CreatedAt: stat.ModTime().UnixMilli()}
	db, err := sql.Open("sqlite", sqliteURI(path)+"?mode=ro")
	if err != nil {
		return info, err
	}
	defer db.Close()
	err = db.QueryRow("SELECT COALESCE(SUM(parent_id IS NULL AND status<>'skipped'),0),count(*) FROM tasks").Scan(&info.Roots, &info.Nodes)
	return info, err
}
func (e *Engine) QuitStatus() (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var count int
	err := e.db.QueryRow("SELECT (SELECT count(*) FROM focus_sessions WHERE state IN ('running','paused'))+(SELECT count(*) FROM notification_jobs WHERE state='pending' AND fire_at BETWEEN ? AND ?)", e.now().UnixMilli(), e.now().UnixMilli()+3600000).Scan(&count)
	return count > 0, err
}
