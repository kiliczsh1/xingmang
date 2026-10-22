//go:build windows

package bookstore

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	foDelete          = 3
	fofSilent         = 0x0004
	fofNoConfirmation = 0x0010
	fofAllowUndo      = 0x0040
	fofNoErrorUI      = 0x0400
)

type shFileOpStruct struct {
	hwnd                  uintptr
	wFunc                 uint32
	pFrom                 *uint16
	pTo                   *uint16
	fFlags                uint16
	fAnyOperationsAborted int32
	hNameMappings         uintptr
	lpszProgressTitle     *uint16
}

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperati = shell32.NewProc("SHFileOperationW")
)

// moveToRecycleBin 把文件/目录送进 Windows 回收站（可撤销）。
//
// 用 SHFileOperationW 而不是 SHFileOperation（ANSI）：中文路径在 ANSI 版本下会乱码。
// pFrom 必须是以 \0\0 结尾的双空字符串。
func moveToRecycleBin(path string) error {
	from, err := syscall.UTF16FromString(path)
	if err != nil {
		return err
	}
	// UTF16FromString 已带一个 \0，再补一个构成双终止
	from = append(from, 0)

	op := shFileOpStruct{
		wFunc:  foDelete,
		pFrom:  &from[0],
		fFlags: fofAllowUndo | fofNoConfirmation | fofSilent | fofNoErrorUI,
	}
	ret, _, _ := procSHFileOperati.Call(uintptr(unsafe.Pointer(&op)))
	if ret != 0 {
		return errors.New("系统回收站操作失败（SHFileOperation 返回非 0）")
	}
	if op.fAnyOperationsAborted != 0 {
		return errors.New("系统回收站操作被取消")
	}
	return nil
}
