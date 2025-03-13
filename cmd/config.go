package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/wn/config"
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
	flagKeys = []string{"lang", "deepseek_apikey", "llm"}
	listFlag bool
)

func init() {
	if config.GetConfig("lang") == "" {
		os.Setenv("WN_LANG", "en")
	}

	rootCmd.AddCommand(configCmd)
	configCmd.Flags().String("lang", config.GetConfig("lang"), lang.T("Set language"))
	configCmd.Flags().String("default_provider", config.GetConfig("llm"), lang.T("Set default LLM provider"))
	configCmd.Flags().String("deepseek_apikey", config.GetConfig("deepseek_apikey"), lang.T("Set DeepSeek API Key"))
	configCmd.Flags().String("openai_apikey", config.GetConfig("openai_apikey"), lang.T("Set Openai API Key"))
	configCmd.Flags().BoolVar(&listFlag, "list", false, lang.T("List all configurations"))
}

func runConfig(cmd *cobra.Command, args []string) {
	if err := config.LoadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	if listFlag {
		fmt.Println(lang.T("Current configurations:"))
		for _, key := range flagKeys {
			value := config.GetConfig(key)
			if value != "" {
				fmt.Printf("%s=%s\n", config.GetEnvKey(key), value)
			}
		}
		return
	}

	configChanged := false
	for _, key := range flagKeys {
		if cmd.Flag(key).Changed {
			value, _ := cmd.Flags().GetString(key)
			os.Setenv(config.GetEnvKey(key), value)
			configChanged = true
		}
	}

	if configChanged {
		if err := config.SaveConfig(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
	}

	// 始终输出可导出的配置格式
	for _, key := range flagKeys {
		value := config.GetConfig(key)
		if value != "" {
			fmt.Printf("export %s=%s\n", config.GetEnvKey(key), value)
		}
	}
}
