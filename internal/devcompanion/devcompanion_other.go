//go:build !windows

// Package devcompanion 的非 Windows 桩实现。
//
// 「伴生进程」这套逻辑是为 Windows 开发环境写的（start-dev.ps1 + tasklist/taskkill），
// 其他平台直接返回 0，保证跨平台编译通过。详见 devcompanion_windows.go。
package devcompanion

// FileName 与 Windows 实现保持一致，避免调用方按平台写死路径。
const FileName = ".dev-companions.json"

// Cleanup 在非 Windows 平台不做任何事。
func Cleanup(string) int { return 0 }
