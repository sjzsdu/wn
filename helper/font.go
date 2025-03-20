package helper

import (
	"fmt"
	"os"
)

func FindFont() (string, error) {
	fontPaths := []string{
		"/System/Library/Fonts/Songti.ttc", // 常用中文字体
		"/System/Library/Fonts/SimSun.ttf", // 常用中文字体
		"/System/Library/Fonts/SimHei.ttf", // 常用中文字体
		"/System/Library/Fonts/PingFang.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/System/Library/Fonts/STHeiti Medium.ttc",
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no suitable font found")
}
