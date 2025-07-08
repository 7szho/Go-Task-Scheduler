package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// Exists checks whether the file/folder in the given path exists
// Exists 检查给定的路径(文件或文件夹)是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) // os.Stat 获取文件信息
	if err == nil {
		// 没有错误, 路径存在
		return true
	}
	if os.IsNotExist(err) {
		// 文件不存在
		return false
	}
	// 其他错误, 路径存在, 但无法访问
	return true
}

// Ext returns the file name extension used by path.
// it is empty if there is no dot.
// Ext 函数返回路径中文件名的扩展名
func Ext(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}
