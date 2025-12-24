package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kbesada/flux-code-cli/internal/app"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "flux",
	Short: "AI-powered coding assistant",
	Long: `Flux is a terminal-based AI coding assistant.
It provides an interactive TUI for AI-powered code assistance
using open-source AI models with your own API keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := app.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
}
