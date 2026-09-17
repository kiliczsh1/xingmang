// 星芒桌面版发布打包工具（Go 版，替代 build.ps1）。
//
// 用法:
//   go run ./tools/release                        # 默认: windows/amd64, 含前端
//   go run ./tools/release -skip-frontend         # 跳过前端构建；web/ 过期会自动重建
//   go run ./tools/release -os linux -arch arm64  # 交叉编译
//   go run ./tools/release -include-migrate       # 额外编译 migrate 工具
//   go run ./tools/release -zip                   # 同时产出 zip 归档
//   go run ./tools/release -no-zip                # 明确不打归档（-zip 与 -no-zip 同时给时以 -no-zip 为准）
//   go run ./tools/release -os windows -arch amd64,arm64
//   go run ./tools/release -version 0.736 -zip    # 指定版本号，归档名 V-0.736.zip
//
// 参数同时兼容 PowerShell 大写形式（-OS / -Arch / -SkipFrontend ...），
// 以便 build.bat / build.ps1 原样转调。
//
// 产物:
//   dist/<os>-<arch>/
//     xingmang.exe (或 xingmang)
//     migrate.exe  (可选, -include-migrate)
//     README.txt
//   dist/<os>-<arch>.zip                       (可选, -zip; 未指定版本号)
//   dist/V-<version>.zip                       (可选, -zip 且指定 -version, 单目标)
//   dist/V-<version>-<os>-<arch>.zip           (可选, -zip 且指定 -version, 多目标)
//
// 说明:
//   - 番茄榜单数据源已内置在程序中（internal/fanqie，Go 原生实现），
//     打包产物不再需要 Python 桥脚本与 Python 运行时。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ─── flag（兼容双大小写写法） ──────────────────────────────────────────

type strValue struct{ v *string }

func (s strValue) String() string { return *s.v }
func (s strValue) Set(v string) error {
	*s.v = v
	return nil
}

type boolValue struct{ v *bool }

func (b boolValue) String() string { return strconv.FormatBool(*b.v) }
func (b boolValue) Set(v string) error {
	val, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}
	*b.v = val
	return nil
}
func (b boolValue) IsBoolFlag() bool { return true }

var (
	targetOS   string
	targetArch string
	skipFront  bool
	includeMig bool
	makeZip    bool
	noZip      bool
	verFlag    string
)

func init() {
	strValue{&targetOS}.register("os", "OS")
	strValue{&targetArch}.register("arch", "Arch")
	strValue{&verFlag}.register("version", "Version")
	boolValue{&skipFront}.register("skip-frontend", "SkipFrontend")
	boolValue{&includeMig}.register("include-migrate", "IncludeMigrate")
	boolValue{&makeZip}.register("zip", "Zip")
	boolValue{&noZip}.register("no-zip", "NoZip")
}

func (s strValue) register(names ...string) {
	for _, n := range names {
		flag.Var(s, n, "")
	}
}
func (b boolValue) register(names ...string) {
	for _, n := range names {
		flag.Var(b, n, "")
	}
}

// ─── 输出辅助 ──────────────────────────────────────────────────────────

func step(msg string) { fmt.Printf("[BUILD] %s\n", msg) }
func ok(msg string)   { fmt.Printf("[OK]    %s\n", msg) }
func warn(msg string) { fmt.Printf("[WARN]  %s\n", msg) }
func die(msg string)  { fmt.Printf("[FAIL]  %s\n", msg); os.Exit(1) }

// ─── 工具函数 ──────────────────────────────────────────────────────────

// findRoot 从当前目录向上查找项目根（含 go.mod 的目录）。
func findRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		die("无法获取当前目录: " + err.Error())
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			die("未找到项目根目录 (go.mod)，请在项目目录内运行")
		}
		dir = parent
	}
}

// splitList 拆分逗号分隔的目标列表（如 -os windows -arch amd64,arm64），丢弃空白项。
func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// run 在指定目录执行命令，输出透传。
func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runOut 在指定目录执行命令并返回输出（去除首尾空白）。
func runOut(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// newestFileModTime 返回一组路径中最新的文件修改时间（目录则递归其下文件）。
func newestFileModTime(paths ...string) time.Time {
	var t time.Time
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			if fi.ModTime().After(t) {
				t = fi.ModTime()
			}
			continue
		}
		filepath.Walk(p, func(_ string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && fi.ModTime().After(t) {
				t = fi.ModTime()
			}
			return nil
		})
	}
	return t
}

// webSourceMaxTime 返回 web/ 下前端源码的最新修改时间。
// 用于判断 web/dist 是否过期：只要 src/、public/、vite 配置、package.json(.lock)
// 等任一文件比产物新，就认为需要重建。dist/ 与 node_modules/ 属于产物/依赖，跳过。
func webSourceMaxTime(webDir string) time.Time {
	var t time.Time
	filepath.Walk(webDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			if path != webDir {
				switch filepath.Base(path) {
				case "dist", "node_modules", ".git":
					return filepath.SkipDir
				}
			}
			return nil
		}
		if fi.ModTime().After(t) {
			t = fi.ModTime()
		}
		return nil
	})
	return t
}

// ─── 前端构建 ──────────────────────────────────────────────────────────

func buildFrontend(root string, skip bool) {
	webDir := filepath.Join(root, "web")
	if _, err := os.Stat(filepath.Join(webDir, "package.json")); err != nil {
		die("未找到 web/package.json")
	}
	if !skip {
		step("构建前端 (npm run build)...")
		if err := run(webDir, "npm", "run", "build"); err != nil {
			warn("npm run build 失败，回退到 vite build (跳过类型检查)...")
			if err := run(webDir, "npx", "vite", "build"); err != nil {
				die("vite build 失败: " + err.Error())
			}
		}
	} else {
		distDir := filepath.Join(webDir, "dist")
		if _, err := os.Stat(filepath.Join(distDir, "index.html")); err != nil {
			die("web/dist 不存在，无法跳过前端构建")
		}
		// 前端源码树（src/public/配置/依赖清单…）比产物新，说明 dist 已过期
		if webSourceMaxTime(webDir).After(newestFileModTime(distDir)) {
			warn("检测到 web/dist 已过期，-skip-frontend 被忽略，自动重建...")
			buildFrontend(root, false)
			return
		}
		ok("跳过前端构建 (web/dist 已是最新)")
	}
	if _, err := os.Stat(filepath.Join(webDir, "dist", "index.html")); err != nil {
		die("前端产物缺失: web/dist/index.html")
	}
	ok("前端构建完成")
}

// syncFrontend 同步前端产物到 embed 目录（供 go:embed 使用）。
func syncFrontend(root string) {
	step("同步前端产物到 cmd/xingmang/web/dist...")
	src := filepath.Join(root, "web", "dist")
	dst := filepath.Join(root, "cmd", "xingmang", "web", "dist")
	if err := os.RemoveAll(dst); err != nil {
		die("清理 embed 目录失败: " + err.Error())
	}
	if err := copyDir(src, dst); err != nil {
		die("同步前端产物失败: " + err.Error())
	}
	ok("前端产物已嵌入")
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// ─── 版本号 ────────────────────────────────────────────────────────────

func gitVersion(root string) (version, commit, dirty string) {
	commit = "unknown"
	if out, err := runOut(root, "git", "log", "-1", "--format=%h"); err == nil && out != "" {
		commit = out
	}
	if out, err := runOut(root, "git", "status", "--porcelain"); err == nil && out != "" {
		dirty = "-dirty"
	}
	if commit != "unknown" {
		version = commit + dirty
	} else {
		version = "dev"
	}
	return version, commit, dirty
}

// ─── 编译 ──────────────────────────────────────────────────────────────

// buildGo 用指定 GOOS/GOARCH 编译 main 包到 outExe。
func buildGo(root, outExe, goos, goarch, ldflags, pkg string) error {
	cmd := exec.Command("go", "build", "-ldflags="+ldflags, "-o", outExe, pkg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// archiveName 生成归档文件名：
//   - 未指定版本号：<os>-<arch>.zip（覆盖式「最新包」）
//   - 指定版本号：  V-<version>.zip（单目标）/ V-<version>-<os>-<arch>.zip（多目标）
//
// 版本号里自带的 V/v 前缀会先归一化，避免出现 V-V0.736.zip。
func archiveName(version, goos, goarch string, multiTarget bool) string {
	if version == "" {
		return goos + "-" + goarch + ".zip"
	}
	v := strings.TrimPrefix(version, "V")
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimLeft(v, "-")
	if multiTarget {
		return "V-" + v + "-" + goos + "-" + goarch + ".zip"
	}
	return "V-" + v + ".zip"
}

func buildTarget(root, outDir, goos, goarch, version, commit, buildTime string, includeMig, multiTarget bool) {
	if err := os.RemoveAll(outDir); err != nil {
		// 最常见原因：exe 正在运行，其目录下的 xingmang.db / server.log 被占用
		die("清理产物目录失败: " + err.Error() + "\n  请先退出正在运行的 xingmang.exe（任务管理器结束进程）后重新打包")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		die("创建产物目录失败: " + err.Error())
	}

	exeName := "xingmang"
	if goos == "windows" {
		exeName = "xingmang.exe"
	}
	outExe := filepath.Join(outDir, exeName)
	ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.buildTime=%s",
		version, commit, buildTime)

	step(fmt.Sprintf("编译主程序: %s/%s -> %s", goos, goarch, outExe))
	if err := buildGo(root, outExe, goos, goarch, ldflags, "./cmd/xingmang/"); err != nil {
		die("go build 失败 (" + goos + "/" + goarch + "): " + err.Error())
	}
	if size, err := os.Stat(outExe); err == nil {
		ok(fmt.Sprintf("主程序: %s (%d 字节)", exeName, size.Size()))
	}

	if includeMig {
		migName := "migrate"
		if goos == "windows" {
			migName = "migrate.exe"
		}
		outMig := filepath.Join(outDir, migName)
		step("编译迁移工具: " + goos + "/" + goarch + " -> " + migName)
		if err := buildGo(root, outMig, goos, goarch, "-s -w", "./cmd/migrate/"); err != nil {
			die("migrate go build 失败: " + err.Error())
		}
		ok("迁移工具已生成")
	}

	writeReadme(outDir, version, commit, buildTime, goos, goarch)

	if makeZip && !noZip {
		zipPath := filepath.Join(root, "dist", archiveName(verFlag, goos, goarch, multiTarget))
		if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
			die("清理旧归档失败: " + err.Error())
		}
		step("压缩归档: " + filepath.Base(zipPath))
		if err := zipDir(outDir, zipPath); err != nil {
			die("zip 失败: " + err.Error())
		}
		if size, err := os.Stat(zipPath); err == nil {
			ok(fmt.Sprintf("归档: %s (%d 字节)", zipPath, size.Size()))
		}
	}
}

func writeReadme(outDir, version, commit, buildTime, goos, goarch string) {
	runCmd := "./xingmang"
	if goos == "windows" {
		runCmd = "xingmang.exe"
	}
	readme := fmt.Sprintf(`星芒桌面版  %s
构建时间:   %s
Git Commit: %s
目标平台:   %s/%s

【运行方式】
  %s                  # 默认启动 (端口 19825, 自动在系统默认浏览器打开界面)
  %s --no-window      # 仅启动服务，不打开界面
  %s --port 8080      # 指定端口
  %s --dev            # 开发模式 (连接 Vite)
  %s --version        # 查看版本信息

【番茄榜单功能 (AI 扫榜)】
  榜单数据源已内置在程序中（Go 原生实现，无需 Python/Node 环境）。
  首次打开「AI 扫榜」即可获取真实榜单数据，开箱即用。

【数据库】
  首次启动自动创建 xingmang.db (SQLite, WAL 模式)

【小贴士】
  - 单实例复用：服务已在运行时再次双击 %s，只会在浏览器里再开一个标签页，
    不会重复启动服务，也不会因端口被占用而报错。
  - 关闭标签页不会退出后台服务；要彻底退出请在「任务管理器」结束 xingmang.exe。

【注意】
  - 界面在系统默认浏览器中打开（不再依赖 Microsoft Edge / WebView2）
  - 数据库文件与会话日志位于程序所在目录
`, version, buildTime, commit, goos, goarch,
		runCmd, runCmd, runCmd, runCmd, runCmd, runCmd)
	if err := os.WriteFile(filepath.Join(outDir, "README.txt"), []byte(readme), 0o644); err != nil {
		die("写入 README.txt 失败: " + err.Error())
	}
	ok("README.txt 已生成")
}

// zipDir 将 dir 目录内容递归压缩为 zip 文件。
// 排除运行时文件（数据库/日志），避免脏包。
func zipDir(dir, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		// 跳过运行时生成的文件（数据库 / 日志），它们不属于发布包
		base := filepath.Base(rel)
		if strings.HasSuffix(base, ".db") || strings.HasSuffix(base, ".db-wal") ||
			strings.HasSuffix(base, ".db-shm") || strings.HasSuffix(base, ".log") {
			return nil
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		f.Close()
		return err
	})
}

// ─── 主流程 ───────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	if targetOS == "" {
		targetOS = "windows"
	}
	if targetArch == "" {
		targetArch = "amd64"
	}

	root := findRoot()
	if _, err := exec.LookPath("go"); err != nil {
		die("未找到 go 命令，请先安装 Go。")
	}
	if out, err := runOut(root, "go", "version"); err == nil {
		ok(strings.TrimPrefix(out, "go version "))
	}

	// 版本号（-version 指定优先，否则用 git commit；commit 始终注入，便于追溯）
	autoVersion, commit, _ := gitVersion(root)
	version := verFlag
	if version == "" {
		version = autoVersion
	}
	if commit == "" {
		commit = "unknown"
	}
	buildTime := time.Now().Format("2006-01-02T15:04:05-07:00")

	fmt.Printf("\n============================================\n")
	fmt.Printf("  星芒桌面版 - 发布打包 (Go 打包工具)\n")
	fmt.Printf("============================================\n")
	fmt.Printf("  Version:   %s\n", version)
	fmt.Printf("  Commit:    %s\n", commit)
	fmt.Printf("  BuildTime: %s\n", buildTime)
	fmt.Printf("  Target:    %s / %s\n", targetOS, targetArch)
	fmt.Printf("============================================\n\n")

	// 前端构建 + 同步 embed
	buildFrontend(root, skipFront)
	syncFrontend(root)

	distRoot := filepath.Join(root, "dist")
	if err := os.MkdirAll(distRoot, 0o755); err != nil {
		die("创建 dist 目录失败: " + err.Error())
	}

	oses := splitList(targetOS)
	arches := splitList(targetArch)
	multiTarget := len(oses)*len(arches) > 1

	for _, goos := range oses {
		for _, goarch := range arches {
			buildTarget(root, filepath.Join(distRoot, goos+"-"+goarch),
				goos, goarch, version, commit, buildTime, includeMig, multiTarget)
		}
	}

	// 总结
	fmt.Printf("\n============================================\n")
	fmt.Printf("  构建成功\n")
	fmt.Printf("============================================\n")
	entries, err := os.ReadDir(distRoot)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				exe := filepath.Join(distRoot, e.Name(), "xingmang.exe")
				if _, err := os.Stat(exe); err != nil {
					exe = filepath.Join(distRoot, e.Name(), "xingmang")
				}
				size := "?"
				if fi, err := os.Stat(exe); err == nil {
					size = strconv.FormatInt(fi.Size(), 10)
				}
				fmt.Printf("  dist\\%s\\  (%s 字节)\n", e.Name(), size)
			} else {
				if info, err := e.Info(); err == nil {
					fmt.Printf("  dist\\%s  (%d 字节)\n", e.Name(), info.Size())
				} else {
					fmt.Printf("  dist\\%s\n", e.Name())
				}
			}
		}
	}
	fmt.Printf("============================================\n\n")
}
