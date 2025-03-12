package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: lang.T("Set config"),
	Long:  lang.T("Set global configuration"),
	Run:   runConfig,
}

var (
	configMap map[string]string
	flagKeys  = []string{"lang", "deepseek_apikey"}
	listFlag  bool
)

func init() {
	configMap = make(map[string]string)

	// 加载配置文件
	if err := loadConfig(); err == nil {
		// 设置环境变量
		for key, value := range configMap {
			os.Setenv(key, value)
		}
	}

	// 如果环境变量未设置，使用默认值
	if GetConfig("lang") == "" {
		os.Setenv("WN_LANG", "en")
	}

	rootCmd.AddCommand(configCmd)
	configCmd.Flags().String("lang", GetConfig("lang"), lang.T("Set language"))
	configCmd.Flags().String("deepseek_apikey", GetConfig("deepseek_apikey"), lang.T("Set DeepSeek API Key"))
	configCmd.Flags().BoolVar(&listFlag, "list", false, lang.T("List all configurations"))
}

// GetConfig 获取配置值，支持 lang 或 WN_LANG 格式的 key
func GetConfig(key string) string {
	envKey := key
	if !strings.HasPrefix(key, "WN_") {
		envKey = getEnvKey(key)
	}
	return os.Getenv(envKey)
}

func getEnvKey(flagKey string) string {
	return "WN_" + strings.ToUpper(flagKey)
}

func runConfig(cmd *cobra.Command, args []string) {
	if err := loadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	if listFlag {
		fmt.Println(lang.T("Current configurations:"))
		for _, key := range flagKeys {
			envKey := getEnvKey(key)
			value := GetConfig(key)
			if value != "" {
				fmt.Printf("%s=%s\n", envKey, value)
			}
		}
		return
	}

	configChanged := false
	for _, key := range flagKeys {
		if cmd.Flag(key).Changed {
			value, _ := cmd.Flags().GetString(key)
			envKey := getEnvKey(key)
			configMap[envKey] = value
			os.Setenv(envKey, value)
			configChanged = true
		}
	}

	if configChanged {
		if err := saveConfig(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
	}

	// 始终输出可导出的配置格式
	for _, key := range flagKeys {
		value := GetConfig(key)
		if value != "" {
			fmt.Printf("export %s=%s\n", getEnvKey(key), value)
		}
	}
}

func loadConfig() error {
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			configMap[parts[0]] = parts[1]
		}
	}
	return scanner.Err()
}

func saveConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".wn")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config")
	content := ""
	for key, value := range configMap {
		content += fmt.Sprintf("%s=%s\n", key, value)
	}

	return os.WriteFile(configFile, []byte(content), 0644)
}
