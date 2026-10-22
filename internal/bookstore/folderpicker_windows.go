//go:build windows

package bookstore

import (
	"fmt"
	"syscall"
	"unsafe"
)

type browseInfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

var (
	shell32Pick       = syscall.NewLazyDLL("shell32.dll")
	ole32Pick         = syscall.NewLazyDLL("ole32.dll")
	procBrowse        = shell32Pick.NewProc("SHBrowseForFolderW")
	procPathFromIDL   = shell32Pick.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree = ole32Pick.NewProc("CoTaskMemFree")
	procCoInitialize  = ole32Pick.NewProc("CoInitialize")
	user32Pick        = syscall.NewLazyDLL("user32.dll")
	procForeground    = user32Pick.NewProc("GetForegroundWindow")
	procCoUninit      = ole32Pick.NewProc("CoUninitialize")
)

const (
	bifReturnOnlyFSDirs = 0x0001
	bifNewDialogStyle   = 0x0040
)

// PickFolder 弹出系统文件夹选择框，返回用户选择的绝对路径。
//
// 浏览器拿不到真实路径（安全限制），所以只能在**本机**用系统 API 选。
func PickFolder(title string) (string, error) {
	// SHBrowseForFolder 需要 COM 已初始化；配对调用 CoUninitialize。
	hr, _, _ := procCoInitialize.Call(0)
	if int32(hr) >= 0 {
		defer procCoUninit.Call()
	}

	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, syscall.MAX_PATH)
	// 挂在当前前台窗口上，避免对话框弹到浏览器后面
	hwnd, _, _ := procForeground.Call()
	bi := browseInfo{
		hwndOwner:      hwnd,
		pszDisplayName: &buf[0],
		lpszTitle:      titlePtr,
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
	}
	pidl, _, _ := procBrowse.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", ErrPickCancelled
	}
	defer procCoTaskMemFree.Call(pidl)

	pathBuf := make([]uint16, syscall.MAX_PATH)
	ok, _, _ := procPathFromIDL.Call(pidl, uintptr(unsafe.Pointer(&pathBuf[0])))
	if ok == 0 {
		return "", fmt.Errorf("无法解析所选目录路径")
	}
	return syscall.UTF16ToString(pathBuf), nil
}
