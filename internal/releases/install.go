package releases

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"debug/macho"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type InstallPlan struct {
	Target   string `json:"target"`
	Stage    string `json:"stage"`
	Payload  string `json:"payload"`
	PID      int    `json:"pid"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Result   string `json:"result"`
}
type InstallResult struct {
	Version string `json:"version"`
	Error   string `json:"error"`
	Stage   string `json:"stage"`
}

func InstallTarget(executable, platform string) (string, error) {
	executable, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	if platform == "windows" && strings.EqualFold(filepath.Base(executable), "simpletodo.exe") {
		return executable, nil
	}
	if platform == "darwin" && filepath.Base(executable) == "simpletodo" && filepath.Base(filepath.Dir(executable)) == "MacOS" && filepath.Base(filepath.Dir(filepath.Dir(executable))) == "Contents" {
		app := filepath.Dir(filepath.Dir(filepath.Dir(executable)))
		if strings.HasSuffix(app, ".app") {
			if strings.Contains(app, "/AppTranslocation/") {
				return "", fmt.Errorf("请先将应用移到“应用程序”文件夹再升级")
			}
			return app, nil
		}
	}
	return "", fmt.Errorf("请使用正式安装的 SimpleTodo 应用执行自动升级")
}

// PrepareInstall only accepts the expected release layout and verifies the archive before extracting.
func PrepareInstall(archive, target, platform, version, digest, result string) (plan InstallPlan, err error) {
	if !strings.HasPrefix(digest, "sha256:") {
		return plan, fmt.Errorf("发布包缺少 SHA-256 校验信息，请稍后重试或手动下载")
	}
	sum, err := hex.DecodeString(strings.TrimPrefix(digest, "sha256:"))
	if err != nil || len(sum) != 32 {
		return plan, fmt.Errorf("无效的发布包校验信息")
	}
	f, err := os.Open(archive)
	if err != nil {
		return plan, err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	f.Close()
	if err != nil || !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), hex.EncodeToString(sum)) {
		return plan, fmt.Errorf("发布包校验失败，原应用未更改")
	}
	stage, err := os.MkdirTemp(filepath.Dir(target), ".simpletodo-stage-")
	if err != nil {
		return plan, fmt.Errorf("应用所在目录不可写，请将应用移到有写入权限的位置再升级")
	}
	defer func() {
		if err != nil {
			os.RemoveAll(stage)
		}
	}()
	root := "simpletodo.exe"
	if platform == "darwin" {
		root = "simpletodo.app"
	} else if platform != "windows" {
		return plan, fmt.Errorf("当前平台不支持自动安装")
	}
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return plan, fmt.Errorf("发布包不是有效的 ZIP 文件")
	}
	defer zr.Close()
	var total uint64
	seen := map[string]bool{}
	for _, file := range zr.File {
		name := file.Name
		if strings.HasPrefix(name, "__MACOSX/") {
			continue
		}
		clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
		if strings.Contains(name, "\\") || strings.Contains(name, ":") || strings.HasPrefix(name, "/") || clean == ".." || strings.HasPrefix(clean, "../") || (clean != root && !strings.HasPrefix(clean, root+"/")) {
			return plan, fmt.Errorf("发布包包含无效路径")
		}
		if file.Mode()&os.ModeSymlink != 0 || (!file.Mode().IsRegular() && !file.FileInfo().IsDir()) {
			return plan, fmt.Errorf("发布包包含不支持的文件类型")
		}
		if seen[clean] {
			return plan, fmt.Errorf("发布包包含重复文件")
		}
		seen[clean] = true
		total += file.UncompressedSize64
		if total > 1<<30 || len(seen) > 20000 {
			return plan, fmt.Errorf("发布包解压大小超出限制")
		}
		dest := filepath.Join(stage, "payload", filepath.FromSlash(clean))
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(dest, 0755); err != nil {
				return plan, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return plan, err
		}
		src, er := file.Open()
		if er != nil {
			return plan, er
		}
		mode := os.FileMode(0644)
		if file.Mode()&0111 != 0 {
			mode = 0755
		}
		dst, er := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
		if er != nil {
			src.Close()
			return plan, er
		}
		n, er := io.Copy(dst, io.LimitReader(src, int64(file.UncompressedSize64)+1))
		src.Close()
		closeErr := dst.Close()
		if er != nil || closeErr != nil || n != int64(file.UncompressedSize64) {
			return plan, fmt.Errorf("发布包解压失败")
		}
	}
	payload := filepath.Join(stage, "payload", root)
	if err = validatePayload(payload, platform, version); err != nil {
		return plan, err
	}
	plan = InstallPlan{Target: target, Stage: stage, Payload: payload, PID: os.Getpid(), Platform: platform, Version: version, Result: result}
	return plan, nil
}
func validatePayload(payload, platform, version string) error {
	if platform == "windows" {
		f, err := pe.Open(payload)
		if err != nil {
			return fmt.Errorf("发布包缺少有效的 Windows 程序")
		}
		defer f.Close()
		if f.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return fmt.Errorf("安装包平台不匹配")
		}
		return nil
	}
	data, err := os.ReadFile(filepath.Join(payload, "Contents", "Info.plist"))
	if err != nil {
		return fmt.Errorf("发布包缺少应用信息")
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	values := map[string]string{}
	key := ""
	for {
		tok, e := dec.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return fmt.Errorf("无效应用信息")
		}
		if el, ok := tok.(xml.StartElement); ok && (el.Name.Local == "key" || el.Name.Local == "string") {
			var text string
			if dec.DecodeElement(&text, &el) != nil {
				return fmt.Errorf("无效应用信息")
			}
			if el.Name.Local == "key" {
				key = text
			} else {
				values[key] = text
			}
		}
	}
	if values["CFBundleIdentifier"] != "com.junyi.simpletodo" || values["CFBundleExecutable"] != "simpletodo" || values["CFBundleShortVersionString"] != version {
		return fmt.Errorf("安装包应用标识或版本不匹配")
	}
	binary := filepath.Join(payload, "Contents", "MacOS", "simpletodo")
	m, err := macho.Open(binary)
	if err != nil {
		return fmt.Errorf("发布包缺少有效的 macOS 程序")
	}
	defer m.Close()
	if m.Cpu != macho.CpuArm64 {
		return fmt.Errorf("安装包平台不匹配")
	}
	if err = os.Chmod(binary, 0755); err != nil {
		return err
	}
	if err = exec.Command("/usr/bin/codesign", "--verify", "--deep", "--strict", payload).Run(); err != nil {
		return fmt.Errorf("应用签名校验失败，原应用未更改")
	}
	return nil
}
func (p InstallPlan) validate() error {
	if (p.Platform != "darwin" && p.Platform != "windows") || (p.Platform == "darwin" && !strings.HasSuffix(p.Target, ".app")) || (p.Platform == "windows" && !strings.EqualFold(filepath.Base(p.Target), "simpletodo.exe")) {
		return fmt.Errorf("无效安装目标")
	}
	stage, err := filepath.Abs(p.Stage)
	if err != nil || stage != p.Stage || !strings.HasPrefix(filepath.Base(stage), ".simpletodo-stage-") || filepath.Dir(stage) != filepath.Dir(p.Target) || p.Payload != filepath.Join(stage, "payload", map[string]string{"darwin": "simpletodo.app", "windows": "simpletodo.exe"}[p.Platform]) || p.PID <= 0 || !filepath.IsAbs(p.Target) || !filepath.IsAbs(p.Result) {
		return fmt.Errorf("无效升级计划")
	}
	return nil
}
func LaunchInstaller(p InstallPlan, executable string) error {
	if err := p.validate(); err != nil {
		return err
	}
	helper := filepath.Join(p.Stage, "update-helper")
	if p.Platform == "windows" {
		helper += ".exe"
	}
	in, err := os.Open(executable)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(helper, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	closeErr := out.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	data, _ := json.Marshal(p)
	path := filepath.Join(p.Stage, "plan.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(p.Stage, "helper.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	cmd := exec.Command(helper, "--apply-update", path)
	cmd.Stdout = log
	cmd.Stderr = log
	detach(cmd)
	if err = cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// ApplyInstall can be exercised against disposable fixtures without launching an app.
func ApplyInstall(ctx context.Context, p InstallPlan, alive func(int) bool, launch func(string, string) error) error {
	if err := p.validate(); err != nil {
		return err
	}
	for alive(p.PID) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待原应用退出超时，原应用未更改")
		case <-time.After(200 * time.Millisecond):
		}
	}
	backup := filepath.Join(p.Stage, "previous")
	if err := os.Rename(p.Target, backup); err != nil {
		if restart := launch(p.Target, p.Platform); restart != nil {
			return fmt.Errorf("无法替换原应用，原文件保留但无法自动重启：%w", restart)
		}
		return fmt.Errorf("无法替换原应用，已重新启动旧版：%w", err)
	}
	if err := os.Rename(p.Payload, p.Target); err != nil {
		if rollback := os.Rename(backup, p.Target); rollback != nil {
			return fmt.Errorf("安装失败，原程序保留于 %s", backup)
		}
		if restart := launch(p.Target, p.Platform); restart != nil {
			return fmt.Errorf("安装失败，原文件已恢复但无法自动重启：%w", restart)
		}
		return fmt.Errorf("安装失败，已恢复并重新启动原应用：%w", err)
	}
	if err := launch(p.Target, p.Platform); err != nil {
		os.Rename(p.Target, p.Payload)
		if rollback := os.Rename(backup, p.Target); rollback != nil {
			return fmt.Errorf("启动失败，原程序保留于 %s", backup)
		}
		_ = launch(p.Target, p.Platform)
		return fmt.Errorf("新版启动失败，已恢复原应用：%w", err)
	}
	return nil
}
func RunInstaller(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var p InstallPlan
	if err = json.Unmarshal(data, &p); err != nil {
		return err
	}
	if err = p.validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err = ApplyInstall(ctx, p, processAlive, func(target, platform string) error {
		var cmd *exec.Cmd
		if platform == "darwin" {
			return exec.Command("/usr/bin/open", "-n", target).Run()
		} else {
			cmd = exec.Command(target)
		}
		detach(cmd)
		if er := cmd.Start(); er != nil {
			return er
		}
		return cmd.Process.Release()
	})
	result := InstallResult{Version: p.Version, Stage: p.Stage}
	if err != nil {
		result.Error = err.Error()
	}
	b, _ := json.Marshal(result)
	_ = os.MkdirAll(filepath.Dir(p.Result), 0700)
	_ = os.WriteFile(p.Result, b, 0600)
	// The running helper is removed by a later app launch on Windows.
	if p.Platform != "windows" && err == nil {
		_ = os.RemoveAll(p.Stage)
	}
	return err
}
