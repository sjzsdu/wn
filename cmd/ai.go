package cmd

import (
	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: lang.T("Chat with AI"),
	Long:  lang.T("Start an interactive chat session with AI using configured LLM provider"),
	Run:   runAI,
}

var (
	listProviders bool
	listModels    bool
)

func init() {
	rootCmd.AddCommand(aiCmd)

	aiCmd.Flags().BoolVar(&listProviders, "providers", false, lang.T("List available LLM providers"))
	aiCmd.Flags().BoolVar(&listModels, "models", false, lang.T("List available models for current provider"))
}

func runAI(cmd *cobra.Command, args []string) {
	if listProviders {
		providers := aigc.GetAvailableProviders()
		for _, p := range providers {
			cmd.Printf("- %s\n", p)
		}
		return
	}

	if listModels {
		models, err := aigc.GetAvailableModels(llmName)
		if err != nil {
			cmd.PrintErrf(lang.T("Error getting models")+": %v\n", err)
			return
		}
		cmd.Printf(lang.T("Available models for provider")+" (%s):\n", llmName)
		for _, m := range models {
			cmd.Printf("- %s\n", m)
		}
		return
	}
}
