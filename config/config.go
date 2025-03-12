package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var configMap map[string]string

func init() {
	configMap = make(map[string]string)
	if err := LoadConfig(); err == nil {
		for key, value := range configMap {
			os.Setenv(key, value)
		}
	}
}

func GetConfig(key string) string {
	envKey := key
	if !strings.HasPrefix(key, "WN_") {
		envKey = getEnvKey(key)
	}
	return os.Getenv(envKey)
}

func GetEnvKey(flagKey string) string {
	return "WN_" + strings.ToUpper(flagKey)
}

func LoadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(homeDir, ".wn", "config")
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	// 清空现有配置
	configMap = make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			configMap[parts[0]] = parts[1]
			os.Setenv(parts[0], parts[1])
		}
	}
	return scanner.Err()
}

func SaveConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".wn")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config")
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 确保写入所有配置项
	for key, value := range configMap {
		if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
			return err
		}
	}
	return file.Sync() // 确保数据写入磁盘
}

func getEnvKey(flagKey string) string {
	return "WN_" + strings.ToUpper(flagKey)
}

// SetConfig 设置配置值并更新环境变量
func SetConfig(key, value string) {
	envKey := key
	if !strings.HasPrefix(key, "WN_") {
		envKey = getEnvKey(key)
	}
	configMap[envKey] = value
	os.Setenv(envKey, value)
}

// ClearConfig 清除指定配置
func ClearConfig(key string) {
	envKey := key
	if !strings.HasPrefix(key, "WN_") {
		envKey = getEnvKey(key)
	}
	delete(configMap, envKey)
	os.Unsetenv(envKey)
}

// ClearAllConfig 清除所有配置
func ClearAllConfig() {
	for key := range configMap {
		os.Unsetenv(key)
	}
	configMap = make(map[string]string)
}
