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
//   - 榜单数据源全部内置在程序中（internal/fanqie、qimao、qidian、feilu，
//     均为 Go 原生实现），打包产物不再需要 Python 桥脚本与 Python 运行时。
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

// String 必须容忍零值：flag 在打印 usage 的默认值时会用一个未初始化的
// 副本调 String()，此时 v 为 nil（直接解引用会让 -h 刷一屏 panic）。
func (s strValue) String() string {
	if s.v == nil {
		return ""
	}
	return *s.v
}
func (s strValue) Set(v string) error {
	*s.v = v
	return nil
}

type boolValue struct{ v *bool }

func (b boolValue) String() string {
	if b.v == nil {
		return "false"
	}
	return strconv.FormatBool(*b.v)
}
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
	strValue{&targetOS}.register("目标操作系统，逗号分隔（默认 windows）", "os", "OS")
	strValue{&targetArch}.register("目标架构，逗号分隔（默认 amd64）", "arch", "Arch")
	strValue{&verFlag}.register("版本号；留空则用 git commit，归档名据此生成", "version", "Version")
	boolValue{&skipFront}.register("跳过前端构建（web/dist 过期时会自动重建）", "skip-frontend", "SkipFrontend")
	boolValue{&includeMig}.register("额外编译 migrate 工具", "include-migrate", "IncludeMigrate")
	boolValue{&makeZip}.register("产出 zip 归档", "zip", "Zip")
	boolValue{&noZip}.register("明确不打归档（与 -zip 同给时以本项为准）", "no-zip", "NoZip")
}

func (s strValue) register(usage string, names ...string) {
	for _, n := range names {
		flag.Var(s, n, usage)
	}
}
func (b boolValue) register(usage string, names ...string) {
	for _, n := range names {
		flag.Var(b, n, usage)
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

// ─── 前置检查 ──────────────────────────────────────────────────────────

// probeOutputWritable 逐个检查目标产物能否被覆盖，**不改动任何文件**。
//
// 为什么用"试着以写方式打开 exe"而不是按进程名查：
//   - 按进程名查会误伤。开发时常见的形态是项目根目录有一个 xingmang.exe 在跑
//     （start-dev 起的后端），它用的是根目录数据，和 dist\<os>-<arch>\ 毫无关系，
//     打包完全不冲突。按名字一刀切会让「开着开发环境就不能打包」。
//   - 真正的冲突只有一种：**待覆盖的那份 exe 正在运行**。Windows 加载映像时
//     以 FILE_SHARE_READ 打开，所以再以写方式打开必然拿到 sharing violation。
//     拿它当探针最准，而且只打开不写入，探测本身不破坏任何东西。
func probeOutputWritable(distRoot, goos, goarch string) (string, error) {
	outDir := filepath.Join(distRoot, goos+"-"+goarch)
	exeName := "xingmang"
	if goos == "windows" {
		exeName = "xingmang.exe"
	}
	exe := filepath.Join(outDir, exeName)
	f, err := os.OpenFile(exe, os.O_WRONLY, 0)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return goos + "/" + goarch, err
	}
	f.Close()
	return "", nil
}

// preflight 在动手之前把「一定会失败」的条件查完。
//
// 这些失败都发生在前端构建（约 1 分钟）之后，白等；而且真正的原因藏在
// "Access is denied" 里，不像缺 go 那样一眼能看出该怎么退。
func preflight(root, distRoot string, skip bool) {
	if !skip {
		if _, err := os.Stat(filepath.Join(root, "web", "node_modules")); err != nil {
			fmt.Println("[FAIL]  web/node_modules 不存在，无法构建前端。")
			fmt.Println("        请先 cd web 执行 npm install；或加 -SkipFrontend 复用现有 web/dist。")
			os.Exit(1)
		}
	}

	for _, goos := range splitList(targetOS) {
		for _, goarch := range splitList(targetArch) {
			if _, err := probeOutputWritable(distRoot, goos, goarch); err != nil {
				fmt.Printf("[FAIL]  无法覆盖 %s/%s 的产物：%v\n", goos, goarch, err)
				fmt.Printf("        %s\\xingmang.exe 正在运行（或被杀软占用），Windows 不允许覆盖。\n",
					filepath.Join(distRoot, goos+"-"+goarch))
				fmt.Println("        请先退出那个程序：界面右上角头像 →「退出程序」；或任务管理器结束进程。")
				os.Exit(1)
			}
			// 目录里有跑出来的数据：产物目录被整个重建，这些数据属于临时产物，
			// 按约定直接清掉继续打包（删除由 buildTarget 的 RemoveAll 完成），
			// 这里只把被清掉的东西列出来，方便事后追。
			if found := runtimeDataIn(filepath.Join(distRoot, goos+"-"+goarch)); len(found) > 0 {
				warn("产物目录里有本地跑程序留下的运行时数据，将随目录一起清掉：")
				fmt.Printf("        %s\n", filepath.Join(distRoot, goos+"-"+goarch))
				for _, f := range found {
					fmt.Printf("          · %s\n", f)
				}
				fmt.Println("        （需要保留就先拷走；本次打包不中止。）")
			}
		}
	}

	// 图标资源：windows 构建时 Go 会自动链接包目录下的 rsrc_<os>_<arch>.syso。
	// 缺了不会导致构建失败，但产物会没有图标 —— 属于"打包成功、发出去难看"，所以提示一声。
	for _, arch := range splitList(targetArch) {
		if !strings.Contains(targetOS, "windows") {
			continue
		}
		syso := filepath.Join(root, "cmd", "xingmang", "rsrc_windows_"+arch+".syso")
		if _, err := os.Stat(syso); err != nil {
			warn(fmt.Sprintf("缺少图标资源 cmd/xingmang/rsrc_windows_%s.syso，产物将没有程序图标", arch))
			fmt.Println("        重新生成：go run ./tools/logogen web/public/brand/app.ico（见 docs/brand/logo-usage.md）")
		}
	}
}

// ─── 产物目录防护 ──────────────────────────────────────────────────────

// runtimeDataIn 检查产物目录里是否已经有「跑起来才会产生」的数据。
//
// 背景：产物目录会被整个 RemoveAll 再重建，而这个程序的数据落点就是 exe 所在目录。
// 也就是说，**任何一次在本地产物目录里双击跑过的 xingmang.exe，都会把作品和索引
// 写进 dist\<os>-<arch>\**。这些数据属于本地调试留下的临时产物，打包时直接清掉
// （RemoveAll 会一起删），这里只负责把它们列出来，不再中止打包。
//
// 归档侧仍单独排除 books\ / data\ / workflow_outputs\ / chapter-cache\ / logs\：
// 万一将来有人绕过这道提示，也不至于把稿件打进发布包。
//
// 更糟的是 zipDir 只排除 .db / .log，books\ 下的 .md 会被原样打进发布包 ——
// 等于把作者的全部稿件发给所有人。
//
// 所以这里两件事都做：有数据就中止（让人先把东西拷走），没数据也把 books\ 排除出归档。
func runtimeDataIn(outDir string) []string {
	var found []string
	for _, name := range []string{"xingmang.db", "books", "workflow_outputs", "data", "chapter-cache"} {
		p := filepath.Join(outDir, name)
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			// 空库（刚建表、没内容）不算，几百 KB 的骨架库是正常的
			if fi.Size() > 2<<20 {
				found = append(found, fmt.Sprintf("%s (%d 字节)", name, fi.Size()))
			}
			continue
		}
		n := 0
		filepath.Walk(p, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				n++
			}
			return nil
		})
		if n > 0 {
			found = append(found, fmt.Sprintf("%s\\ (%d 个文件)", name, n))
		}
	}
	return found
}


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

【功能模块】
  我的小说 · 提示词库 · 提示预览 · 变量抽卡 · AI 扫榜 · 星芒 Agent
  wiki 图谱 · 工作流 · 插件系统 · API 配置 · 个人中心
  子页面：创作（正文写作）· 拆书库 · 角色库 · 教程中心 · 功能速查表

【AI 扫榜 (榜单数据源)】
  四个数据源已内置在程序中（Go 原生实现，无需 Python/Node 环境），
  在「AI 扫榜」页面右上角切换，默认打开番茄小说：

    番茄小说    在读人数；名次变动由平台原生给出幅度
    七猫小说    频道 × 榜单（点击/字数/更新/收藏）× 题材
    起点中文网  性别 × 榜单 × 筛选（题材 / 月份 / 周期）；
                唯一支持整榜抓取与「掉榜」列表的数据源
    飞卢小说    题材 × 榜单（月票/点击/鲜花/收藏/打赏/字数/催更）

  - 名次变动幅度与「在榜天数」依赖跨天快照：首次打开时只有方向、
    或没有可对比的基准，攒够两次采样后会自动升级为 ▲N / ▼N 与「NEW」。
  - 右上角齿轮可把当前「数据源 · 分组 · 榜单 · 筛选」设为默认常用榜单。

【数据位置与备份】
  首次启动自动创建 xingmang.db (SQLite, WAL 模式) 与作品库目录。

  数据目录 = exe 所在目录。若该目录不可写（例如装在 Program Files），
  自动改存 %%AppData%%\xingmang；启动日志首行会写明本次实际使用的「数据目录」。

  数据目录下：
    xingmang.db           索引库：作品/章节/人物的条目、榜单快照、提示词、用量统计等
    books\                作品本体：一部作品一个文件夹，
                          正文 / 人物 / 世界观 / 大纲 / 灵感 / 备忘录 / 封面画廊 / 创作记忆
                          全部是 .md / .txt 明文，可直接用记事本或 VSCode 打开、编辑、拷贝
    data\editor-images    正文插图（正文里的图片指向这里，删掉后插图会失效）
    data\image-cache      外链图片缓存，可删，会重新抓取
    data\generated-images AI 生成图暂存区；封面会归位到 books\<书名>\gallery\
    chapter-cache\        章节抓取缓存（24 小时过期），可删，会重新抓取
    workflow_outputs\     工作流产物

  ⚠ 备份口径：xingmang.db 与 books\ 要一起拷。
    正文只存在 books\ 里，只备份 xingmang.db 会丢掉全部稿件。
    反过来也成立 —— 把 books\ 拷到另一台机器，用书架页「切换目录」指过去，
    点「接管」就能读回全部内容（不复制、不搬移任何文件）。

  作品库位置可改：我的小说（书架）→「切换目录」。

  退出时会自动合并 WAL，退出后 xingmang.db 是自包含的（索引只需要它一个）。
  程序若被强制结束（未走正常退出），最新数据仍在 xingmang.db-wal 里，
  下次启动会自动恢复；此时备份需连同 .db-wal 一起复制。

【退出程序】
  界面右上角用户头像 → 「退出程序」→ 确认。
  （也可以在控制台窗口按 Ctrl+C）
  - 服务会先处理完在途请求再停止，随后合并数据库并结束进程。
  - 仅允许在本机操作；若通过局域网地址访问界面，该按钮会被拒绝。
  - 关闭浏览器标签页不会退出后台服务，必须走上面的方式。

【升级】
  1. 先退出正在运行的程序（见上「退出程序」）
  2. 再解压覆盖本目录
  3. 重新双击 xingmang.exe

  - 旧进程没退出时解压会因 exe 被占用而失败；直接双击新程序也会检测到旧实例并
    中止启动，弹窗提示先退出 —— 这是为了避免「更新了、跑的还是旧版本」。
  - 解压覆盖只替换程序文件；books\ 与 xingmang.db 都在数据目录里，不受影响。
  - 查看当前版本：界面「个人中心」右上角的版本徽章，或鼠标悬停左侧栏底部「版本」。

【小贴士】
  - 单实例复用：服务已在运行时再次双击 %s，只会在浏览器里再开一个标签页，
    不会重复启动服务，也不会因端口被占用而报错。
  - 如果界面显示「版本 未知」，说明后台跑的是不支持版本上报的旧版本，
    请按「升级」章节退出旧进程后重新启动。

【注意】
  - 界面在系统默认浏览器中打开（不再依赖 Microsoft Edge / WebView2）
  - 数据库文件与会话日志位于程序所在目录
  - 发布包里**不含**任何作品/对话/图片等个人数据，首次启动是全新状态
`, version, buildTime, commit, goos, goarch,
		runCmd, runCmd, runCmd, runCmd, runCmd, runCmd)
	if err := os.WriteFile(filepath.Join(outDir, "README.txt"), []byte(readme), 0o644); err != nil {
		die("写入 README.txt 失败: " + err.Error())
	}
	ok("README.txt 已生成")
}

// zipDir 将 dir 目录内容递归压缩为 zip 文件。
//
// 排除运行时文件，避免脏包。除了 .db / .log，还有两类必须排除：
//   - books\ 与 workflow_outputs\ 是用户数据（正文、记忆、工作流产物）。
//     正常打包时产物目录是新建的，不会有这些；但如果有人在产物目录里跑过程序，
//     它们就会在里面 —— 打进去等于把作者稿件发出去。
//   - data\ 是插图 / 缓存暂存区，chapter-cache\ 是章节抓取缓存，同理。
func zipDir(dir, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	// 顶层运行时目录：整个跳过
	skipTop := map[string]bool{
		"books":            true,
		"workflow_outputs": true,
		"data":             true,
		"chapter-cache":    true,
		"logs":             true,
	}

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if skipTop[rel] {
				return filepath.SkipDir
			}
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

	// 前置检查放在最前面：缺前端依赖 / 产物 exe 被占用，都必须在
	// 前端构建之前中止，否则要白等一分钟才看到失败。
	distRoot := filepath.Join(root, "dist")
	preflight(root, distRoot, skipFront)

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
