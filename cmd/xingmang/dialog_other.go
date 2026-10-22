//go:build !windows

package main

// messageBox 在非 Windows 平台不可用，恒返回 false，
// 调用方据此回退到控制台日志输出。
func messageBox(_, _ string) bool { return false }
