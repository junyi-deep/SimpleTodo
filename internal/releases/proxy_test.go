package releases

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestProxyValidationAndPrivatePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proxy.json")
	c, err := ReadProxy(path)
	if err != nil || c.Mode != "system" {
		t.Fatalf("default: %+v %v", c, err)
	}
	for _, address := range []string{"socks5://localhost:80", "https://localhost:80", "http://user:secret@localhost:80", "http://localhost:99999", "http://localhost:80/path", "http://", "localhost:abc"} {
		if (ProxyConfig{Mode: "custom", Address: address}).Validate() == nil {
			t.Errorf("accepted %q", address)
		}
	}
	c = ProxyConfig{Mode: "custom", Address: "127.0.0.1:7890", Username: "user@name", Password: "secret:#"}
	if err = SaveProxy(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := ReadProxy(path)
	if err != nil || got.Password != c.Password {
		t.Fatalf("persistence: %v", err)
	}
	public := got.Public()
	if public.Password != "" || !public.HasPassword || !public.KeepPassword {
		t.Fatal("password exposed or retention lost")
	}
	u, err := got.URL()
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != c.Password || u.User.Username() != c.Username {
		t.Fatal("credentials escaped incorrectly")
	}
	if err = SaveProxy(path, ProxyConfig{Mode: "invalid"}); err == nil {
		t.Fatal("invalid mode accepted")
	}
	got, _ = ReadProxy(path)
	if got.Password != c.Password {
		t.Fatal("invalid save changed file")
	}
}
func TestCustomProxyAuthenticationAndDirectBypass(t *testing.T) {
	var auth string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Proxy-Authorization")
		io.WriteString(w, "proxied")
	}))
	defer proxy.Close()
	client, err := NewClient(ProxyConfig{Mode: "custom", Address: proxy.URL, Username: "name", Password: "password"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	response, err := client.Get("http://does-not-exist.invalid/update")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if auth != "Basic "+base64.StdEncoding.EncodeToString([]byte("name:password")) {
		t.Fatal("missing proxy authentication")
	}
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	direct, err := NewClient(ProxyConfig{Mode: "none"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if direct.Transport.(*http.Transport).Proxy != nil {
		t.Fatal("direct mode inherited proxy")
	}
	anonymous, err := NewClient(ProxyConfig{Mode: "custom", Address: proxy.URL}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer anonymous.CloseIdleConnections()
	response, err = anonymous.Get("http://does-not-exist.invalid/update")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if auth != "" {
		t.Fatal("optional credentials sent")
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestDownloadAtomicAndBounded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.zip")
	os.WriteFile(path, []byte("old"), 0600)
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("PK\x03\x04archive")), ContentLength: 100}, nil
	})}
	url := "https://github.com/junyi-deep/SimpleTodo/releases/download/v1/SimpleTodo.zip"
	if err := Download(context.Background(), client, url, path); err == nil {
		t.Fatal("truncated archive accepted")
	}
	old, _ := os.ReadFile(path)
	if string(old) != "old" {
		t.Fatal("existing file damaged")
	}
	client.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("PK\x03\x04archive")), ContentLength: 11}, nil
	})
	if err := Download(context.Background(), client, url, path); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(path)
	if string(saved) != "PK\x03\x04archive" {
		t.Fatal("wrong download")
	}
	files, _ := os.ReadDir(filepath.Dir(path))
	if len(files) != 1 {
		t.Fatal("temporary download leaked")
	}
	if err := Download(context.Background(), client, "https://example.com/app.zip", path); err == nil {
		t.Fatal("foreign release accepted")
	}
}

func TestHTTPProxyTunnelForReleaseAndDownload(t *testing.T) {
	var connections atomic.Int32
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/latest") {
			io.WriteString(w, `{"tag_name":"v9.0.0","assets":[{"name":"SimpleTodo-macos-arm64.zip","browser_download_url":"https://github.com/junyi-deep/SimpleTodo/releases/download/v9.0.0/SimpleTodo-macos-arm64.zip"}]}`)
		} else {
			io.WriteString(w, "PK\x03\x04archive")
		}
	}))
	defer target.Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "CONNECT" || r.Header.Get("Proxy-Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("name:password")) {
			w.WriteHeader(407)
			return
		}
		upstream, err := net.Dial("tcp", target.Listener.Addr().String())
		if err != nil {
			t.Error(err)
			return
		}
		downstream, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			upstream.Close()
			t.Error(err)
			return
		}
		connections.Add(1)
		io.WriteString(downstream, "HTTP/1.1 200 Connection Established\r\n\r\n")
		go func() { defer upstream.Close(); defer downstream.Close(); io.Copy(upstream, downstream) }()
		io.Copy(downstream, upstream)
	}))
	defer proxy.Close()
	client, err := NewClient(ProxyConfig{Mode: "custom", Address: proxy.URL, Username: "name", Password: "password"}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	// Trust only this test's locally generated TLS endpoint; production retains default TLS verification.
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	info, err := Latest(context.Background(), client, Endpoint, "0.1.3", "darwin", "arm64")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available {
		t.Fatal("release not found")
	}
	if err = Download(context.Background(), client, info.DownloadURL, filepath.Join(t.TempDir(), "app.zip")); err != nil {
		t.Fatal(err)
	}
	if connections.Load() < 2 {
		t.Fatal("release and archive must both use the HTTP proxy tunnel")
	}
}
