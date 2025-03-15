package cmd

import (
	"fmt"

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
	flagKeys = []string{"lang", "default_provider", "default_agent", "deepseek_apikey", "deepseek_model", "openai_apikey", "openai_model"}
	listFlag bool
)

func init() {
	if config.GetConfig("lang") == "" {
		config.SetConfig("lang", "en")
	}

	rootCmd.AddCommand(configCmd)
	configCmd.Flags().String("lang", config.GetConfig("lang"), lang.T("Set language"))
	configCmd.Flags().String("default_provider", config.GetConfig("default_provider"), lang.T("Set default LLM provider"))
	configCmd.Flags().String("default_agent", config.GetConfig("default_agent"), lang.T("Set default agent"))
	configCmd.Flags().String("deepseek_apikey", config.GetConfig("deepseek_apikey"), lang.T("Set DeepSeek API Key"))
	configCmd.Flags().String("deepseek_model", config.GetConfig("deepseek_model"), lang.T("Set DeepSeek default model"))
	configCmd.Flags().String("openai_apikey", config.GetConfig("openai_apikey"), lang.T("Set Openai API Key"))
	configCmd.Flags().String("openai_model", config.GetConfig("openai_model"), lang.T("Set Openai default model"))
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
	// 处理 flagKeys 中的标准配置项
	for _, key := range flagKeys {
		flag := cmd.Flag(key)
		if flag != nil && flag.Changed {
			value, _ := cmd.Flags().GetString(key)
			config.SetConfig(key, value)
			configChanged = true
		}
	}

	// 特殊处理 default_provider 标志，将其映射到 llm 配置项
	defaultProviderFlag := cmd.Flag("default_provider")
	if defaultProviderFlag != nil && defaultProviderFlag.Changed {
		value, err := cmd.Flags().GetString("default_provider")
		if err == nil {
			envKey := config.GetEnvKey("default_provider")
			if envKey != "" {
				config.SetConfig(envKey, value)
				configChanged = true
			}
		}
	}

	if configChanged {
		if err := config.SaveConfig(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
	}
}
