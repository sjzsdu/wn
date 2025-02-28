package helper

import (
	"fmt"
	"os"
)

func FindFont() (string, error) {
	fontPaths := []string{
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
