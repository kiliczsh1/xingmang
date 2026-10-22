// Package appdir 解析「运行根目录」，即 xingmang.exe 所在目录。
//
// 背景：数据库、data/、workflow_outputs/ 等运行时数据历史上都按「进程工作目录（cwd）」
// 解析相对路径。一旦发生下面两种情况，数据就会落错地方：
//
//   - 用户把新版本解压到**另一个目录**再双击 → 新目录建空库，旧数据留在老目录，
//     用户看到的是「升级后数据没了」或「回到旧目录还是旧数据」；
//   - 用户**直接从压缩包双击 exe**（7-Zip/WinRAR 会先解压到 %TEMP% 再运行）
//     → 数据建在系统临时目录，程序一退出就被清理。
//
// 桌面便携程序的数据应当跟着 exe 走，这也是 README 对用户的承诺
// （见 tools/release/main.go：「数据库文件与会话日志位于程序所在目录」）。
package appdir

import (
	"os"
	"path/filepath"
	"strings"
)

// goRunMarker 是 `go run` 产出的临时可执行文件路径特征，形如
//
//	%TEMP%\go-build1234567\b001\exe\main.exe
//
// 这种场景必须回退到 cwd（源码根目录），否则开发时数据会落到系统临时目录。
const goRunMarker = "go-build"

// Root 返回运行根目录（绝对路径）。
//
// 优先取 exe 所在目录；无法判定时（`go run`、取不到 exe 路径）回退当前工作目录。
func Root() string {
	if exe, err := os.Executable(); err == nil {
		if dir := filepath.Dir(exe); !isTemporaryExecution(dir) {
			if abs, aerr := filepath.Abs(dir); aerr == nil {
				return abs
			}
			return dir
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

// isTemporaryExecution 判断 exe 是否位于「临时构建 / 临时解压」目录。
// 这类路径不能当作数据根，否则程序退出后数据会被一并清理。
func isTemporaryExecution(dir string) bool {
	if strings.Contains(strings.ToLower(dir), goRunMarker) {
		return true
	}
	if tmp := os.TempDir(); tmp != "" {
		if rel, err := filepath.Rel(tmp, dir); err == nil {
			// rel 不以 ".." 开头 => dir 位于 tmp 之内
			up := ".." + string(filepath.Separator)
			if rel != ".." && !strings.HasPrefix(rel, up) {
				return true
			}
		}
	}
	return false
}

// UserDataRoot 返回「程序目录不可写」时的兜底数据目录。
// Windows 落在 %AppData%\xingmang，macOS/Linux 落在各自的标准配置目录下。
func UserDataRoot() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		if home, herr := os.UserHomeDir(); herr == nil && home != "" {
			return filepath.Join(home, ".xingmang")
		}
		return ""
	}
	return filepath.Join(base, "xingmang")
}

// Prepare 决定并使用最终的运行根目录，返回实际生效的目录。
//
// 所有运行时数据（xingmang.db、data/、workflow_outputs/）都按相对 cwd 解析，
// 因此这里直接切换进程工作目录，一次性把它们统一到 exe 所在目录，无需逐处改路径。
//
// 顺序：exe 目录（可写）→ 用户数据目录（exe 目录不可写时，如装在 Program Files）
// → 保持当前目录（调用方自行告警）。
// fallback 为 true 表示走了用户数据目录这条兜底路径。
func Prepare() (dir string, fallback bool) {
	if root := Root(); Writable(root) {
		if err := os.Chdir(root); err == nil {
			return root, false
		}
	}

	if ud := UserDataRoot(); ud != "" {
		if err := os.MkdirAll(ud, 0o755); err == nil && Writable(ud) {
			if err := os.Chdir(ud); err == nil {
				return ud, true
			}
		}
	}

	cwd, _ := os.Getwd()
	return cwd, false
}

// Writable 判断目录是否可写（用创建临时文件实测，而非只看权限位）。
func Writable(dir string) bool {
	f, err := os.CreateTemp(dir, ".xingmang-write-test-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}
