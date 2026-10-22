//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Windows MessageBox 常量（winuser.h）
const (
	mbOK            = 0x00000000
	mbIconWarning   = 0x00000030
	mbSetForeground = 0x00010000
	mbTopMost       = 0x00040000
)

var (
	user32DLL      = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32DLL.NewProc("MessageBoxW")
)

// messageBox 弹出阻塞式消息框，返回是否成功弹出。
//
// 用户双击 exe 时可能没有可见的控制台（快捷方式 / --no-window / 窗口一闪而过），
// 只写日志等于没说。这里直接调 user32，避免引入第三方依赖。
// 调用失败返回 false，由调用方回退到日志输出。
func messageBox(title, text string) bool {
	t, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return false
	}
	c, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return false
	}
	ret, _, _ := procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(t)),
		uintptr(mbOK|mbIconWarning|mbSetForeground|mbTopMost),
	)
	return ret != 0
}
