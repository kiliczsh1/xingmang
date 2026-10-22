//go:build !windows

package bookstore

import "errors"

// PickFolder 仅在 Windows 提供；其他平台返回错误，由调用方提示改用路径输入。
func PickFolder(string) (string, error) {
	return "", errors.New("当前平台不支持系统文件夹选择")
}
