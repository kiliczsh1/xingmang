//go:build windows

package devcompanion

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// spawnSleeper 起一个能稳定挂住一段时间的真实进程，返回 PID 与映像名。
//
// 用 ping 而不是自己造进程：它是系统自带、映像名确定、能挂住且无副作用。
func spawnSleeper(t *testing.T) (int, string) {
	t.Helper()
	cmd := exec.Command("ping", "-n", "60", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("启动测试进程失败: %v", err)
	}
	pid := cmd.Process.Pid
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})

	name, alive := imageOf(pid)
	if !alive {
		t.Fatalf("测试进程 PID %d 未存活，imageOf 判定有误", pid)
	}
	return pid, name
}

func alive(pid int) bool {
	_, ok := imageOf(pid)
	return ok
}

func writeManifestFile(t *testing.T, dir string, created time.Time, entries []Entry) string {
	t.Helper()
	payload := map[string]any{
		"created":    created.UTC().Format(time.RFC3339),
		"companions": entries,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("序列化清单失败: %v", err)
	}
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("写清单失败: %v", err)
	}
	return path
}

// 清单不存在（生产环境的常态）→ 静默返回 0，不报错。
func TestCleanup_NoManifest(t *testing.T) {
	if n := Cleanup(t.TempDir()); n != 0 {
		t.Fatalf("无清单时应返回 0，实得 %d", n)
	}
}

// PID 已不存在 → 跳过，返回 0。
func TestCleanup_DeadPid(t *testing.T) {
	dir := t.TempDir()
	writeManifestFile(t, dir, time.Now(), []Entry{{Name: "vite", PID: 999999, Exe: "node.exe", Port: 5173}})
	if n := Cleanup(dir); n != 0 {
		t.Fatalf("PID 不存在时应返回 0，实得 %d", n)
	}
}

// 正常路径：存活的伴生进程被结束，清单被删除。
func TestCleanup_KillsLiveProcess(t *testing.T) {
	dir := t.TempDir()
	pid, name := spawnSleeper(t)
	path := writeManifestFile(t, dir, time.Now(), []Entry{{Name: "vite", PID: pid, Exe: name, Port: 5173}})

	if n := Cleanup(dir); n != 1 {
		t.Fatalf("应结束 1 个进程，实得 %d", n)
	}
	// taskkill /F 是同步的，但进程句柄回收有一瞬延迟，轮询确认。
	dead := false
	for i := 0; i < 20; i++ {
		if !alive(pid) {
			dead = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !dead {
		t.Fatalf("PID %d 应已被结束，但仍在运行", pid)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("清单应已被删除，stat err=%v", err)
	}
}

// ⚠ 最关键的一条：映像名对不上时必须**跳过**，绝不能误杀。
// 场景：清单里的进程早已退出，PID 被系统复用给了别的程序。
func TestCleanup_RefusesMismatchedImage(t *testing.T) {
	dir := t.TempDir()
	pid, realName := spawnSleeper(t)
	// 故意写一个与真实映像名不同的值
	bogus := "definitely-not-" + realName
	writeManifestFile(t, dir, time.Now(), []Entry{{Name: "vite", PID: pid, Exe: bogus, Port: 5173}})

	if n := Cleanup(dir); n != 0 {
		t.Fatalf("映像名不符时应跳过（返回 0），实得 %d", n)
	}
	if !alive(pid) {
		t.Fatalf("映像名不符的进程被误杀：PID %d", pid)
	}
}

// 过期清单直接丢弃，不碰任何进程。
func TestCleanup_ExpiredManifest(t *testing.T) {
	dir := t.TempDir()
	pid, name := spawnSleeper(t)
	path := writeManifestFile(t, dir, time.Now().Add(-(maxAge + time.Hour)),
		[]Entry{{Name: "vite", PID: pid, Exe: name, Port: 5173}})

	if n := Cleanup(dir); n != 0 {
		t.Fatalf("过期清单应返回 0，实得 %d", n)
	}
	if !alive(pid) {
		t.Fatalf("过期清单不应结束任何进程，PID %d 被杀了", pid)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("过期清单也应被删除，stat err=%v", err)
	}
}

// PowerShell 5.1 的 ConvertTo-Json 会把**单元素数组**降级成对象，
// 只起 Vite（-SkipBridge）时就是这种形状，必须能解析。
func TestCleanup_SingleObjectCompanions(t *testing.T) {
	dir := t.TempDir()
	pid, name := spawnSleeper(t)
	payload := map[string]any{
		"created": time.Now().UTC().Format(time.RFC3339),
		// 注意：这里刻意是对象而不是数组
		"companions": Entry{Name: "vite", PID: pid, Exe: name, Port: 5173},
	}
	b, _ := json.Marshal(payload)
	if err := os.WriteFile(filepath.Join(dir, FileName), b, 0o644); err != nil {
		t.Fatalf("写清单失败: %v", err)
	}

	if n := Cleanup(dir); n != 1 {
		t.Fatalf("单对象形状应也能清理 1 个，实得 %d", n)
	}
}

// ⚠ 回归测试：清单带 UTF-8 BOM 时仍须正常解析。
//
// PowerShell 5.1 的 `Set-Content -Encoding UTF8` 默认写 BOM，而 encoding/json
// 会因此直接报错。脚本侧已改为写无 BOM，但这里必须有兜底 —— 否则任何一次
// 脚本改回 Set-Content，整个清理功能就会静默失效（不报错、不清理，最难排查）。
func TestCleanup_StripsUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	pid, name := spawnSleeper(t)
	payload := map[string]any{
		"created":    time.Now().UTC().Format(time.RFC3339),
		"companions": []Entry{{Name: "vite", PID: pid, Exe: name, Port: 5173}},
	}
	b, _ := json.Marshal(payload)
	// 手动前置 BOM
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, b...)
	if err := os.WriteFile(filepath.Join(dir, FileName), withBOM, 0o644); err != nil {
		t.Fatalf("写带 BOM 清单失败: %v", err)
	}

	if n := Cleanup(dir); n != 1 {
		t.Fatalf("带 BOM 的清单应仍能清理 1 个进程，实得 %d（BOM 未被剥离）", n)
	}
}

// 损坏的清单不应让退出路径崩溃，且应被丢弃。
func TestCleanup_MalformedManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte("{ not json at all"), 0o644); err != nil {
		t.Fatalf("写损坏清单失败: %v", err)
	}
	if n := Cleanup(dir); n != 0 {
		t.Fatalf("损坏清单应返回 0，实得 %d", n)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("损坏清单应被丢弃，stat err=%v", err)
	}
}

// created 缺失（手动构造的清单）时应照常执行清理，不能被时间检查挡掉。
func TestCleanup_MissingCreatedStillWorks(t *testing.T) {
	dir := t.TempDir()
	pid, name := spawnSleeper(t)
	payload := map[string]any{"companions": []Entry{{Name: "vite", PID: pid, Exe: name, Port: 5173}}}
	b, _ := json.Marshal(payload)
	if err := os.WriteFile(filepath.Join(dir, FileName), b, 0o644); err != nil {
		t.Fatalf("写清单失败: %v", err)
	}
	if n := Cleanup(dir); n != 1 {
		t.Fatalf("缺 created 时应仍清理，实得 %d", n)
	}
}
