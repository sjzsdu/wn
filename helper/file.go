package helper

import (
	"os"
	"path/filepath"
)

// FileInfo 包含文件的基本信息
type FileInfo struct {
	Path string
	Info os.FileInfo
}

// WalkFunc 是一个回调函数类型，用于处理每个文件
type WalkFunc func(fileInfo FileInfo) error

// WalkDir 递归遍历目录，对每个文件调用回调函数
func WalkDir(root string, callback WalkFunc) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo := FileInfo{
			Path: path,
			Info: info,
		}

		return callback(fileInfo)
	})
}

// FilterFiles 使用 WalkDir 来过滤文件
func FilterFiles(root string, extensions []string, excludes []string) ([]string, error) {
	var files []string

	err := WalkDir(root, func(fileInfo FileInfo) error {
		if fileInfo.Info.IsDir() {
			return nil
		}

		ext := filepath.Ext(fileInfo.Path)
		if len(ext) > 0 {
			ext = ext[1:] // 移除开头的点
		}

		if contains(extensions, ext) && !isExcluded(fileInfo.Path, excludes) {
			files = append(files, fileInfo.Path)
		}

		return nil
	})

	return files, err
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isExcluded(path string, excludes []string) bool {
	for _, pattern := range excludes {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}
