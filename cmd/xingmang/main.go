package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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

func main() {
	showVer := flag.Bool("version", false, "Print version info and exit")
	devMode := flag.Bool("dev", false, "Development mode (connect to Vite dev server)")
	port := flag.Int("port", 19825, "Server port")
	noWindow := flag.Bool("no-window", false, "Don't open browser/window")
	flag.Parse()

	if *showVer {
		fmt.Printf("xingmang %s (commit: %s, built: %s)\n", version, commit, buildTime)
		return
	}

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	log.Printf("xingmang %s starting (commit: %s, built: %s)", version, commit, buildTime)

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *devMode {
		log.Println("开发模式：请通过 Vite 开发服务器访问")
		go server.Start(addr, true, webFS)
		<-sigChan
		log.Println("正在关闭服务...")
		return
	}

	// 单实例复用：探活已有后端，命中则只打开界面、不重复启动服务。
	// 解决"后端还活着、界面已关闭"导致找不到入口的问题——
	// 再次双击 exe 即可重新打开界面，不会因端口占用而崩溃。
	if !*noWindow && probeHealth(url) {
		log.Printf("检测到已有实例在运行 (%s)，重新打开界面...", url)
		openBrowser(url)
		return
	}

	// Production mode: start server in goroutine
	go server.Start(addr, false, webFS)

	// Wait for server to be ready
	for i := 0; i < 100; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", *port))
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if *noWindow {
		log.Printf("服务已启动: %s (按 Ctrl+C 退出)", url)
		<-sigChan
		log.Println("正在关闭服务...")
		return
	}

	// 在系统默认浏览器中打开界面（普通标签页）
	openBrowser(url)
	log.Printf("服务已启动: %s (按 Ctrl+C 退出)", url)
	<-sigChan
	log.Println("正在关闭服务...")
}

// probeHealth 探测给定 URL 的后端是否已在运行（GET /api/health）。
// 超时短（1s），只用于判断"是否已有实例"，不阻塞正常冷启动。
func probeHealth(url string) bool {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(url + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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
