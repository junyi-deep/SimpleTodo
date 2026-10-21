package releases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestPlatformSelectionAndErrors(t *testing.T) {
	body := `{"tag_name":"v0.2.0","html_url":"https://github.com/junyi-deep/SimpleTodo/releases/tag/v0.2.0","assets":[{"name":"SimpleTodo-macos-arm64.zip","browser_download_url":"https://github.com/junyi-deep/SimpleTodo/releases/download/v0.2.0/SimpleTodo-macos-arm64.zip"},{"name":"SimpleTodo-windows-amd64.zip","browser_download_url":"https://github.com/junyi-deep/SimpleTodo/releases/download/v0.2.0/SimpleTodo-windows-amd64.zip"}]}`
	status := 200
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing user agent")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	for _, platform := range []string{"darwin", "windows"} {
		arch := "arm64"
		if platform == "windows" {
			arch = "amd64"
		}
		info, err := Latest(context.Background(), srv.Client(), srv.URL, "0.1.1", platform, arch)
		if err != nil || !info.Available || info.Version != "0.2.0" {
			t.Fatal(info, err)
		}
	}
	if _, err := Latest(context.Background(), srv.Client(), srv.URL, "0.1.1", "darwin", "amd64"); err == nil {
		t.Fatal("unsupported arch accepted")
	}
	status = 403
	if _, err := Latest(context.Background(), srv.Client(), srv.URL, "0.1.1", "darwin", "arm64"); err == nil {
		t.Fatal("HTTP failure ignored")
	}
	if newer("v0.1.1", "0.1.2") || newer("v0.1.1", "0.1.1") || !newer("v0.10.0", "0.2.0") {
		t.Fatal("version comparison")
	}
}
