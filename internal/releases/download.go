package releases

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Download stages the archive next to its destination, leaving an existing file intact on failure.
func Download(ctx context.Context, client *http.Client, address, path string) error {
	if !strings.HasPrefix(address, "https://github.com/junyi-deep/SimpleTodo/releases/download/") {
		return fmt.Errorf("无效的发布包地址")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败，请检查网络与代理设置")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败（HTTP %d）", res.StatusCode)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".simpletodo-download-*")
	if err != nil {
		return fmt.Errorf("无法写入所选目录")
	}
	defer os.Remove(f.Name())
	const maxSize = 512 << 20
	n, err := io.Copy(f, io.LimitReader(res.Body, maxSize+1))
	closeErr := f.Close()
	if err != nil || closeErr != nil || n == 0 || n > maxSize || (res.ContentLength >= 0 && n != res.ContentLength) {
		return fmt.Errorf("下载未完成，请重试")
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("无法保存安装包，请关闭占用文件后重试")
	}
	return nil
}
