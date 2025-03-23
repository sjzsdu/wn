package helper

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetTargetPath(cmdPath string, gitURL string) (string, error) {
	var targetPath string

	if gitURL != "" {
		// 创建临时目录
		tempDir, err := CloneProject(gitURL)
		if err != nil {
			return "", fmt.Errorf("error cloning repository: %v", err)
		}
		targetPath = tempDir
	} else {
		if cmdPath == "" {
			// 获取当前工作目录
			currentDir, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("error getting current directory: %v", err)
			}
			targetPath = currentDir
		} else {
			// 处理相对路径
			if !filepath.IsAbs(cmdPath) {
				currentDir, err := os.Getwd()
				if err != nil {
					return "", fmt.Errorf("error getting current directory: %v", err)
				}
				targetPath = filepath.Join(currentDir, cmdPath)
			} else {
				targetPath = cmdPath
			}
		}
	}

	// 清理路径（移除 .. 和 . 等）
	return filepath.Clean(targetPath), nil
}
