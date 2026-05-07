//go:build !windows

package handlers

import "syscall"

// getDiskFree 返回指定路径所在磁盘的可用字节数（Unix 实现）
func getDiskFree(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
