//go:build !windows

package bookstore

import "errors"

// moveToRecycleBin 在非 Windows 上不可用，调用方会回退到库内 _trash 目录。
func moveToRecycleBin(string) error {
	return errors.New("当前平台不支持系统回收站")
}
