package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const Endpoint = "https://api.github.com/repos/junyi-deep/SimpleTodo/releases/latest"

type Info struct {
	Version     string `json:"version"`
	Available   bool   `json:"available"`
	DownloadURL string `json:"downloadURL"`
	ReleaseURL  string `json:"releaseURL"`
}

func Latest(ctx context.Context, client *http.Client, endpoint, current, platform, arch string) (Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Info{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "SimpleTodo/"+current)
	res, err := client.Do(req)
	if err != nil {
		return Info{}, fmt.Errorf("无法连接 GitHub，请稍后重试：%w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Info{}, fmt.Errorf("GitHub 更新检查失败（HTTP %d），请稍后重试", res.StatusCode)
	}
	var release struct {
		Tag        string `json:"tag_name"`
		URL        string `json:"html_url"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&release); err != nil {
		return Info{}, fmt.Errorf("GitHub 返回了无效版本信息")
	}
	if release.Draft || release.Prerelease {
		return Info{}, fmt.Errorf("暂未找到稳定版本")
	}
	name := ""
	if platform == "darwin" && arch == "arm64" {
		name = "SimpleTodo-macos-arm64.zip"
	}
	if platform == "windows" && arch == "amd64" {
		name = "SimpleTodo-windows-amd64.zip"
	}
	if name == "" {
		return Info{}, fmt.Errorf("当前平台没有对应的发布包")
	}
	out := Info{Version: strings.TrimPrefix(release.Tag, "v"), Available: newer(release.Tag, current), ReleaseURL: release.URL}
	for _, a := range release.Assets {
		if a.Name == name && strings.HasPrefix(a.URL, "https://github.com/junyi-deep/SimpleTodo/releases/download/") {
			out.DownloadURL = a.URL
		}
	}
	if out.DownloadURL == "" {
		return Info{}, fmt.Errorf("最新版本尚无当前平台的下载包")
	}
	return out, nil
}
func newer(a, b string) bool {
	parse := func(s string) [3]int {
		var out [3]int
		for i, v := range strings.Split(strings.TrimPrefix(s, "v"), ".") {
			if i >= 3 {
				break
			}
			out[i], _ = strconv.Atoi(strings.SplitN(v, "-", 2)[0])
		}
		return out
	}
	x, y := parse(a), parse(b)
	for i := range x {
		if x[i] != y[i] {
			return x[i] > y[i]
		}
	}
	return false
}
