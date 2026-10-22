//go:build windows

// Package devcompanion 清理「开发环境伴生进程」——由 start-dev.ps1 启动的
// Vite 开发服务器与番茄桥。
//
// 为什么需要它：这两者与后端是**兄弟进程**（同一个 PowerShell 的子进程），
// 后端退出时操作系统**不会**自动带走它们。而用户的直觉是「关掉后端 = 关掉整个程序」，
// 关掉那个黑窗口后留下两个占端口的孤儿进程，下次启动就会撞端口。
//
// start-dev.ps1 自身有守护循环（任一服务退出就 Stop-All），但那份保障依赖脚本进程存活：
// 用户直接点关闭按钮 / 结束 powershell 时，脚本的 finally 不会执行，子进程就没人管了。
// 因此这里补一条**不依赖脚本存活**的路径：后端收尾时读脚本落盘的 PID 清单，主动结束它们。
//
// 只对开发环境生效：生产环境（双击 exe）根本没有这份清单文件，Cleanup 会直接返回。
package devcompanion

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// FileName 是 start-dev.ps1 落盘的清单文件名（位于 logs/ 下）。
const FileName = ".dev-companions.json"

// maxAge 是清单的有效期。
//
// 清单本应「用完即删」，但脚本被强杀时它会留下。若隔了很久才再启动后端，
// 里面的 PID 极可能已被系统复用给别的进程 —— 过期就直接丢弃，不去猜。
const maxAge = 12 * time.Hour

// Entry 描述一个待清理的伴生进程。
type Entry struct {
	Name string `json:"name"`
	PID  int    `json:"pid"`
	// Exe 是启动时记录下来的映像名（node.exe / python.exe），用于杀死前的身份校验。
	Exe string `json:"exe"`
	// Port 仅用于日志，让人一眼看出杀掉的是哪个服务。
	Port int `json:"port"`
}

// manifest 是清单文件的结构。
//
// Companions 用 RawMessage 接收：Windows PowerShell 5.1 的 ConvertTo-Json
// 对**单元素数组**会降级成对象（只起 Vite 不起番茄桥时就是这种），
// 固定声明成 []Entry 会解析失败。下面按数组优先、失败退回单对象。
type manifest struct {
	Created    string          `json:"created"`
	Companions json.RawMessage `json:"companions"`
}

func (m *manifest) entries() []Entry {
	raw := strings.TrimSpace(string(m.Companions))
	if raw == "" || raw == "null" {
		return nil
	}
	var list []Entry
	if err := json.Unmarshal(m.Companions, &list); err == nil {
		return list
	}
	var one Entry
	if err := json.Unmarshal(m.Companions, &one); err == nil && one.PID > 0 {
		return []Entry{one}
	}
	return nil
}

// Cleanup 读取 logDir 下的清单并结束其中的伴生进程，返回成功结束的个数。
//
// 设计约定：**永不返回错误，也永不 panic**。它跑在程序退出的收尾路径上，
// 任何异常都不该阻止后端自己退出。清单不存在（生产环境）时静默返回 0。
func Cleanup(logDir string) int {
	path := filepath.Join(logDir, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0 // 生产环境没有这份清单，属正常
	}
	// 清单是一次性的：无论下面清理成败都删掉，避免下次拿陈旧 PID 去误杀。
	defer func() { _ = os.Remove(path) }()

	// ⚠ 剥掉 UTF-8 BOM。PowerShell 5.1 的 `Set-Content -Encoding UTF8` 会写入 BOM
	// （没有 UTF8NoBOM 这个选项，那是 PS 6+ 才有的），而 encoding/json 遇到 BOM
	// 直接报 "invalid character 'ï'" → 整个清理逻辑静默失效。
	// 真实踩过：脚本侧看起来一切正常，清单内容也正确，就是永远不清理。
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})

	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		log.Printf("[companion] 清单解析失败，已丢弃: %v", err)
		return 0
	}

	if t, err := time.Parse(time.RFC3339, m.Created); err == nil {
		if age := time.Since(t); age > maxAge {
			log.Printf("[companion] 清单已过期（%s 前），跳过清理", age.Round(time.Minute))
			return 0
		}
	}

	killed := 0
	for _, e := range m.entries() {
		if e.PID <= 0 {
			continue
		}

		imageName, alive := imageOf(e.PID)
		if !alive {
			continue // 已经退出了，无需处理
		}
		// ⚠ 身份校验是这里唯一的安全阀：PID 会被系统复用，若清单里的进程早已退出，
		// 同一个 PID 可能已经属于别的程序。映像名对不上就绝不碰它。
		if e.Exe != "" && !strings.EqualFold(imageName, e.Exe) {
			log.Printf("[companion] PID %d 现在是 %s，与清单记录的 %s 不符，跳过",
				e.PID, imageName, e.Exe)
			continue
		}

		if err := killTree(e.PID); err != nil {
			log.Printf("[companion] 结束 %s (PID %d) 失败: %v", e.Name, e.PID, err)
			continue
		}
		log.Printf("[companion] 已结束伴生进程 %s (PID %d, %s, 端口 %d)", e.Name, e.PID, imageName, e.Port)
		killed++
	}
	return killed
}

// imageOf 返回指定 PID 的映像名（如 node.exe），以及该进程是否存活。
func imageOf(pid int) (string, bool) {
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return "", false
	}
	// 未命中时 tasklist 输出 "INFO: No tasks are running which match..."
	line := strings.TrimSpace(string(out))
	if !strings.HasPrefix(line, "\"") {
		return "", false
	}
	// 形如: "node.exe","22032","Console","1","230,840 K"
	fields := strings.Split(line, "\",\"")
	if len(fields) < 2 {
		return "", false
	}
	name := strings.Trim(fields[0], "\"")
	// 逐字段比对 PID，不用 strings.Contains —— 否则查 123 会被 1234 命中。
	if strings.Trim(fields[1], "\"") != strconv.Itoa(pid) {
		return "", false
	}
	return name, true
}

// killTree 结束进程及其整棵子树。
//
// node 会派生 esbuild 等子进程、番茄桥也可能有子进程，只杀父进程会留下孤儿占端口，
// 所以用 taskkill /T。
func killTree(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
