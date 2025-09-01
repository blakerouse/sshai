package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/spf13/cobra"

	"github.com/blakerouse/sshai/tools"
)

var rootCmd = &cobra.Command{
	Use:   "sshai",
	Short: "SSHAI is a tool that allows you to manage remote machines over SSH using AI.",
	Run: func(cmd *cobra.Command, args []string) {
		// Start the server
		err := run(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("openai", "", "OpenAI API key")
}

func main() {
	err := rootCmd.Execute()
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	apiKey := cmd.Flag("openai").Value.String()
	if apiKey == "" {
		return errors.New("OpenAI API key is required")
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	s := server.NewMCPServer(
		"SSH",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	for _, tool := range tools.Registry.Tools() {
		s.AddTool(tool.Definition(), tool.Handler(client))
	}

	// start the stdio server
	stdio := server.NewStdioServer(s)
	return stdio.Listen(ctx, os.Stdin, os.Stdout)
}
