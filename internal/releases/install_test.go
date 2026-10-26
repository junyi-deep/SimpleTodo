package releases

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func zipFixture(t *testing.T, name string) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "update.zip")
	f, _ := os.Create(path)
	z := zip.NewWriter(f)
	w, _ := z.Create(name)
	w.Write([]byte("bad executable"))
	z.Close()
	f.Close()
	b, _ := os.ReadFile(path)
	return path, fmt.Sprintf("sha256:%x", sha256.Sum256(b))
}
func TestPrepareRejectsUntrustedArchives(t *testing.T) {
	for _, name := range []string{"../simpletodo.exe", "/simpletodo.exe", "other.exe", "simpletodo.exe"} {
		t.Run(name, func(t *testing.T) {
			path, sum := zipFixture(t, name)
			target := filepath.Join(t.TempDir(), "simpletodo.exe")
			os.WriteFile(target, []byte("original"), 0700)
			if _, err := PrepareInstall(path, target, "windows", "0.2.0", sum, filepath.Join(t.TempDir(), "result")); err == nil {
				t.Fatal("accepted invalid archive")
			}
			b, _ := os.ReadFile(target)
			if string(b) != "original" {
				t.Fatal("changed installed app")
			}
			stages, _ := filepath.Glob(filepath.Join(filepath.Dir(target), ".simpletodo-stage-*"))
			if len(stages) != 0 {
				t.Fatal("failed preparation leaked stage")
			}
		})
	}
	path, _ := zipFixture(t, "simpletodo.exe")
	if _, err := PrepareInstall(path, filepath.Join(t.TempDir(), "simpletodo.exe"), "windows", "0.2.0", "sha256:0000", filepath.Join(t.TempDir(), "result")); err == nil {
		t.Fatal("accepted bad hash")
	}
}
func planFixture(t *testing.T) InstallPlan {
	t.Helper()
	dir := t.TempDir()
	stage, _ := os.MkdirTemp(dir, ".simpletodo-stage-")
	payload := filepath.Join(stage, "payload", "simpletodo.exe")
	os.MkdirAll(filepath.Dir(payload), 0700)
	os.WriteFile(payload, []byte("new"), 0700)
	target := filepath.Join(dir, "simpletodo.exe")
	os.WriteFile(target, []byte("old"), 0700)
	return InstallPlan{Target: target, Stage: stage, Payload: payload, PID: 12345, Platform: "windows", Version: "0.2.0", Result: filepath.Join(t.TempDir(), "result")}
}
func TestInstallSwapRollbackAndWait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := planFixture(t)
		called := false
		err := ApplyInstall(context.Background(), p, func(int) bool { return false }, func(path, platform string) error {
			called = true
			b, _ := os.ReadFile(path)
			if string(b) != "new" {
				t.Fatal("launched wrong app")
			}
			return nil
		})
		if err != nil || !called {
			t.Fatal(err)
		}
	})
	t.Run("rollback", func(t *testing.T) {
		p := planFixture(t)
		err := ApplyInstall(context.Background(), p, func(int) bool { return false }, func(string, string) error { return fmt.Errorf("failed launch") })
		b, _ := os.ReadFile(p.Target)
		if err == nil || string(b) != "old" {
			t.Fatal("did not restore original", err)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		p := planFixture(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := ApplyInstall(ctx, p, func(int) bool { return true }, func(string, string) error { t.Fatal("launched before parent exit"); return nil })
		b, _ := os.ReadFile(p.Target)
		if err == nil || string(b) != "old" {
			t.Fatal(err)
		}
	})
	t.Run("original rename fails", func(t *testing.T) {
		p := planFixture(t)
		os.MkdirAll(filepath.Join(p.Stage, "previous"), 0700)
		os.WriteFile(filepath.Join(p.Stage, "previous", "occupied"), []byte("x"), 0600)
		restarted := false
		err := ApplyInstall(context.Background(), p, func(int) bool { return false }, func(path, platform string) error {
			restarted = true
			b, _ := os.ReadFile(path)
			if string(b) != "old" {
				t.Fatal("restarted wrong app")
			}
			return nil
		})
		if err == nil || !restarted {
			t.Fatal("original app was not restarted")
		}
	})
	t.Run("missing payload", func(t *testing.T) {
		p := planFixture(t)
		os.Remove(p.Payload)
		restarted := false
		err := ApplyInstall(context.Background(), p, func(int) bool { return false }, func(path, platform string) error {
			restarted = true
			b, _ := os.ReadFile(path)
			if string(b) != "old" {
				t.Fatal("restarted wrong payload")
			}
			return nil
		})
		b, _ := os.ReadFile(p.Target)
		if err == nil || string(b) != "old" || !restarted {
			t.Fatal("rename failure did not roll back and restart")
		}
	})
}
