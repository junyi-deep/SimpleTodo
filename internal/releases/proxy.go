package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-ieproxy"
)

var systemProxyMu sync.Mutex

// ProxyConfig is stored separately from task backups. Password is never returned to the UI.
type ProxyConfig struct {
	Mode         string `json:"mode"`
	Address      string `json:"address"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	HasPassword  bool   `json:"hasPassword,omitempty"`
	KeepPassword bool   `json:"keepPassword,omitempty"`
}

func (c ProxyConfig) Public() ProxyConfig {
	c.HasPassword = c.Password != ""
	c.Password = ""
	c.KeepPassword = c.HasPassword
	return c
}
func (c ProxyConfig) URL() (*url.URL, error) {
	raw := strings.TrimSpace(c.Address)
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("请输入 HTTP 代理地址，如 http://127.0.0.1:7890；用户名和密码请单独填写")
	}
	if u.Port() != "" {
		n, err := strconv.Atoi(u.Port())
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("代理端口必须为 1—65535")
		}
	}
	if c.Username != "" || c.Password != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u, nil
}
func (c ProxyConfig) Validate() error {
	if c.Mode != "system" && c.Mode != "none" && c.Mode != "custom" {
		return fmt.Errorf("请选择系统代理、无代理或自定义代理")
	}
	if len(c.Address) > 2048 || len(c.Username) > 512 || len(c.Password) > 512 {
		return fmt.Errorf("代理设置内容过长")
	}
	if c.Mode == "custom" {
		_, err := c.URL()
		return err
	}
	return nil
}
func ReadProxy(path string) (ProxyConfig, error) {
	c := ProxyConfig{Mode: "system"}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, fmt.Errorf("无法读取更新代理设置")
	}
	if json.Unmarshal(b, &c) != nil {
		return c, fmt.Errorf("更新代理设置损坏，请重新保存")
	}
	return c, c.Validate()
}
func SaveProxy(path string, c ProxyConfig) error {
	if err := c.Validate(); err != nil {
		return err
	}
	c.HasPassword = false
	c.KeepPassword = false
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".update-proxy-*")
	if err != nil {
		return fmt.Errorf("无法保存更新代理设置")
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func NewClient(c ProxyConfig, timeout time.Duration) (*http.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	switch c.Mode {
	case "none":
		transport.Proxy = nil
	case "custom":
		u, err := c.URL()
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(u)
	default:
		systemProxyMu.Lock()
		ieproxy.ReloadConf()
		system := ieproxy.GetProxyFunc()
		systemProxyMu.Unlock()
		transport.Proxy = func(r *http.Request) (*url.URL, error) {
			u, err := system(r)
			if err != nil {
				return nil, fmt.Errorf("无法读取系统代理")
			}
			if u != nil && u.Scheme == "" {
				u.Scheme = "http"
			}
			if u != nil && u.Scheme != "http" {
				return nil, fmt.Errorf("当前仅支持 HTTP 代理，请调整系统代理或选择无代理")
			}
			return u, nil
		}
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}
