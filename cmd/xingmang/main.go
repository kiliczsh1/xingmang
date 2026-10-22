package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/appdir"
	"github.com/xingmang/xingmang-desktop/internal/devcompanion"
	"github.com/xingmang/xingmang-desktop/internal/server"
)

//go:embed all:web/dist
var webFS embed.FS

// 通过 -ldflags -X 在构建时注入
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// exitReused 表示「本进程没有启动服务就退出了」：19825 上已经有一个本程序的后端
// 在跑（同版本复用 / 版本不一致拒绝启动）。
//
// ⚠ **只有发布版会返回它**。开发构建（isDevBuild）整段跳过版本握手，走不到这里；
// 开发环境端口冲突由 start-dev.ps1 的 releasing ports 清理，或由 bind 失败直接报错。
//
// ⚠ 必须是**非 0**。start-dev.ps1 的守护循环把「backend 退出码 0」当成用户点了
// 「退出程序」，会紧接着 Stop-All 把 Vite / 番茄桥按端口一起杀掉。用 0 退出等于让
// 脚本把「启动被拦下」误判成「正常收尾」，结果是一整套还活着的开发环境被连坐清掉。
// 脚本侧认这个值（见 start-dev.ps1 的 exit 3 分支）：不 Stop-All，只清掉自己多起的进程。
const exitReused = 3

// exitPortBusy 表示「端口被占，服务没能起来」。
//
// ⚠ 必须和 exitReused 分开，两者在脚本侧的处置完全相反：
//   - exitReused=3 → 「启动被拦下」，脚本**只清自己多起的进程**，不碰别人的服务；
//   - exitPortBusy=4 → 端口被占死了，服务根本没起来，脚本应该**整体收掉并报错**。
//
// 合成一个码会让脚本把「bind 失败」误判成「用户主动退出程序」→ 静默 exit 0，
// 留下一个只有半套服务的假环境（真实踩过）。
const exitPortBusy = 4

// isDevBuild 判断当前这份二进制是不是开发构建。
//
// 判据只有 version == "dev"：
//   - start-dev.ps1 用 `go build -ldflags="-s -w"` 编译，不注入 -X main.version，
//     version 保持默认值 "dev"；
//   - 发布包由 tools/release 注入 `-X main.version=<x.y.z>`，永远不会是 "dev"。
//
// ⚠ 不能用 -dev 标志判断：那只是「连接 Vite 开发服务器」的开关，start-dev.ps1 也不传它，
// 而且它跟「这份二进制是不是发布版」没有关系。
func isDevBuild() bool { return version == "dev" }

func main() {
	// 用 run() 的返回值决定退出码，而不是在中间直接 os.Exit：
	// run 内部有 defer（shutdown / 收尾），直接 os.Exit 会把它们全部跳过。
	os.Exit(run())
}

func run() int {
	showVer := flag.Bool("version", false, "Print version info and exit")
	devMode := flag.Bool("dev", false, "Development mode (connect to Vite dev server)")
	port := flag.Int("port", 19825, "Server port")
	noWindow := flag.Bool("no-window", false, "Don't open browser/window")
	flag.Parse()

	if *showVer {
		fmt.Printf("xingmang %s (commit: %s, built: %s)\n", version, commit, buildTime)
		return 0
	}

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	log.Printf("xingmang %s starting (commit: %s, built: %s)", version, commit, buildTime)

	// 数据根固定到 exe 所在目录（便携约定）。
	//
	// 所有运行时数据（xingmang.db / data/ / workflow_outputs/）都按相对「进程工作目录」
	// 解析。历史实现依赖 cwd，导致：解压到新目录 → 新目录建空库、旧数据留在老目录；
	// 从压缩包直接双击（解压到 %TEMP% 再运行）→ 数据落在临时目录，退出即被清理。
	// 这里切换一次 cwd 即可把四处落点统一，无需逐个改路径。
	root, fallback := appdir.Prepare()
	if fallback {
		log.Printf("程序所在目录不可写，数据改存用户目录: %s", root)
	}
	log.Printf("数据目录: %s", root)

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	build := server.BuildInfo{Version: version, Commit: commit, BuildTime: buildTime}

	// 单实例复用 + 版本握手 —— **只在发布版生效**（见 isDevBuild）。
	//
	// 放在启动服务之前：除了「避免重复启动」，更重要的是防止**两个进程同时打开
	// 同一个 SQLite 库**（WAL 下并发写会互相干扰）。
	//
	// ⚠ 开发版整段跳过。dev 的 buildInfo 恒为 dev/unknown/unknown，sameBuild 永远为真，
	// 任何残留的旧 dev 后端都会被判成「同版本实例」→ exitReused=3；而 start-dev.ps1
	// 见 3 就判定「本次启动被拦下」，把 vite / 番茄桥一起停掉。
	// 结果是一次热更新重启就把整个开发环境关掉 —— 这是要避免的。
	// 开发环境单机单人，端口冲突交给 start-dev.ps1 的 releasing ports 处理。
	//
	// 跳过握手不会让两个进程同时打开 SQLite 库：**bind 本身就是互斥的**，
	// 第二个进程在下面试占端口时就会失败退出，根本走不到 database.Init。
	if !isDevBuild() {
		if info := probeHealth(url); info != nil {
			if sameBuild(info) {
				if *noWindow {
					log.Printf("检测到同版本实例已在运行 (%s, %s)，无需重复启动", url, versionText(version, commit))
					return exitReused
				}
				log.Printf("检测到已有实例在运行 (%s, %s)，重新打开界面...", url, versionText(version, commit))
				openBrowser(url)
				return exitReused
			}

			// 版本不一致 → **不复用**。
			// 复用意味着用户看到的是旧界面 + 旧数据（前端产物是 //go:embed 进 exe 的），
			// 会直接得出「我明明更新了，怎么还是老的」——这是用户反馈里最主要的坑。
			log.Printf("检测到版本不一致的实例：%s 运行 %s，当前程序 %s",
				url, versionText(info.Version, info.Commit), versionText(version, commit))
			if !*noWindow {
				warnVersionConflict(info)
			}
			return exitReused
		}
	}

	// 试占一次端口，把「端口被占」变成一条说得清的错误。
	//
	// 交给 http.Server.ListenAndServe 的话，失败只会在服务收尾时打一行「服务异常终止」；
	// 而 start-dev.ps1 的就绪探测是按「端口有没有应答」判断的 —— 旧进程还在应答就会被
	// 误判成「新后端起来了」，于是打开浏览器、进守护循环，才发现退出的其实是新进程。
	//
	// ⚠ 探到就**立刻 Close**，不要把这个 listener 用到服务上：
	// 这里的 bind 发生在数据库初始化之前，而端口一旦 listening，探测方（另一个实例的
	// probeHealth）就会收到连接却等不到 health 应答 —— 建库那几百毫秒里第二个实例会
	// 误判成「端口被别的程序占用」，于是本该提示「已有实例」变成了 bind 失败。
	// 真正的监听仍由 server.Start 内部在初始化完之后建立。
	if probeLn, err := net.Listen("tcp", addr); err != nil {
		log.Printf("启动失败：无法监听 %s（%v）", addr, err)
		if info := probeHealth(url); info != nil {
			log.Printf("该端口上已有一个本程序后端在应答：%s（先退出它）",
				versionText(info.Version, info.Commit))
		} else {
			log.Printf("该端口被其他程序占用，请先结束占用者再重试")
		}
		return exitPortBusy
	} else {
		_ = probeLn.Close()
	}

	// 服务生命周期统一由 ctx 控制，两条退出路径共用同一套收尾逻辑：
	//   - SIGINT / SIGTERM（Ctrl+C、窗口关闭）
	//   - 界面上的「退出程序」→ POST /api/shutdown → 内部调用 shutdown()
	runCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	serverDone := make(chan struct{})
	go func() {
		server.Start(runCtx, addr, *devMode, webFS, build, shutdown)
		close(serverDone)
	}()

	if *devMode {
		log.Println("开发模式：请通过 Vite 开发服务器访问")
	} else {
		// Wait for server to be ready
		for i := 0; i < 100; i++ {
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", *port))
			if err == nil {
				conn.Close()
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !*noWindow {
			// 在系统默认浏览器中打开界面（普通标签页）
			openBrowser(url)
		}
		log.Printf("服务已启动: %s (界面内「退出程序」/ Ctrl+C 退出)", url)
	}

	select {
	case <-sigChan:
		log.Println("收到中断信号，正在关闭服务...")
		shutdown()
	case <-runCtx.Done():
		// 已由 /api/shutdown 触发，server.Start 正在收尾
	}

	// 等 server.Start 真正收尾（HTTP 优雅关闭 + WAL checkpoint + 关闭数据库）。
	// 直接 return 会跳过这些步骤，导致 -wal 里残留未合并的数据。
	<-serverDone

	// 最后一步：把开发环境的伴生进程（Vite / 番茄桥）一起带走。
	//
	// 用户口径：**关闭后端就等于关闭整个程序**。这两个进程由 start-dev.ps1 启动，
	// 与后端是兄弟关系，OS 不会自动回收 —— 不清理就会出现「界面关了、后台还占着端口」。
	// 放在服务完全停止之后是刻意的：此时前端已收到退出响应，再关它不会打断交互。
	//
	// 清单只在开发环境存在；生产环境（双击 exe）这里是空操作。
	if n := devcompanion.Cleanup(filepath.Join(root, "logs")); n > 0 {
		log.Printf("已同步结束 %d 个伴生进程", n)
	}

	return 0
}

// healthInfo 是 /api/health 的应答（只取握手需要的字段）。
type healthInfo struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// sameBuild 判断探测到的实例是否就是「当前这个构建」。
//
// 三个字段全等才算同一个 exe，只比 version 是不够的：用户本地重编译时若不显式传
// -Version，版本号来自 git commit，工作区有未提交改动时两次编译会得到同一个
// 「commit+dirty」——版本号相同但二进制不同，仍会误判为可复用。
// 加 buildTime（精确到编译时刻）就能把「重新编译过」这件事稳定识别出来。
//
// 老版本实例不上报这些字段，一律判为不一致，正是要提示用户先退出的场景。
func sameBuild(info *healthInfo) bool {
	return info.Version == version &&
		info.Commit == commit &&
		info.BuildTime == buildTime
}

// probeHealth 探测给定 URL 上是否已有本程序的后端在运行。
// 返回 nil 表示没有（可以正常冷启动）。超时短（1s），不阻塞冷启动。
//
// ⚠ 必须连版本一起取回来：只判断「端口有人应答」是不够的。旧版本进程仍在后台时，
// 双击新 exe 会直接复用旧实例，用户看到旧界面 + 旧数据，误以为更新没生效。
func probeHealth(url string) *healthInfo {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(url + "/api/health")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var info healthInfo
	// 老版本（health 里没有 version 字段）同样能解码成功，只是 Version 为空串。
	// 这正属于「版本不一致」，也正是要提示用户先退出的场景，不能当失败丢弃。
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil
	}
	return &info
}

// versionText 把版本信息渲染成给人看的字符串。
// 未上报版本号的实例会被明确标注，因为对用户来说「不知道对面是哪个版本」本身就是信息。
func versionText(v, c string) string {
	switch {
	case v == "":
		return "未知（该实例未上报版本号，多半是旧版本）"
	case v == "dev":
		if c != "" && c != "unknown" {
			return "dev(" + truncate(c, 7) + ")"
		}
		return "dev"
	default:
		return "v" + v
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// warnVersionConflict 提示用户「另一个版本的程序还在后台运行」，并中止本次启动。
//
// 刻意不自动结束旧进程：它可能正在写数据，强行 kill 有丢数据风险。
// 代价是用户要手动去任务管理器结束一次——比悄悄用旧版本或丢数据都划算。
func warnVersionConflict(old *healthInfo) {
	msg := fmt.Sprintf(
		"检测到另一个版本的程序正在后台运行，本次启动已取消。\n\n"+
			"后台运行的版本：%s\n"+
			"当前程序版本：  %s\n\n"+
			"请先退出旧版本，再重新启动本程序：\n"+
			"  1. 按 Ctrl+Shift+Esc 打开任务管理器\n"+
			"  2. 找到 xingmang.exe，右键「结束任务」\n"+
			"  3. 重新双击本程序\n\n"+
			"注意：关闭浏览器页面不会退出后台服务，必须结束进程。",
		versionText(old.Version, old.Commit), versionText(version, commit))

	if !messageBox("星芒桌面版 - 检测到版本冲突", msg) {
		// 弹窗不可用（无窗口模式 / 非 Windows）时退化到日志
		log.Println(msg)
	}
}

// openBrowser 在系统默认浏览器中打开界面。
// 走的是普通标签页：浏览器已打开时复用其窗口、只多一个标签页，
// 因此不依赖 Microsoft Edge / WebView2，装了任意浏览器即可。
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("打开浏览器失败，请手动访问 %s (err: %v)", url, err)
	}
}
