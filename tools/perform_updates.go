package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"

	"github.com/blakerouse/sshai/ssh"
)

func init() {
	// register the tool in the registry
	Registry.Register(&PerformUpdates{})
}

// PerformUpdates is a tool that updates the packages on a remote machine.
//
// At the moment this is limited to Ubuntu.
type PerformUpdates struct{}

// Definition returns the mcp.Tool definition.
func (c *PerformUpdates) Definition() mcp.Tool {
	return mcp.NewTool("perform_updates",
		mcp.WithDescription("SSH into a remote machine and updates the operating system packages."),
		mcp.WithString("ssh_connection_string",
			mcp.Required(),
			mcp.Description("SSH connection string"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *PerformUpdates) Handler(aiClient openai.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshConnectionString, err := request.RequireString("ssh_connection_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshClient, err := ssh.NewClientFromString(sshConnectionString)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// connect over ssh
		err = sshClient.Connect()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer sshClient.Close()

		// too make this smarter we should now check what type of machine this is
		// is it a linux machine? running ubuntu, debian, centos, etc.
		// is it a windows machine? running powershell, etc.

		// for now we just assume Ubuntu

		// sudo is required to update and upgrade
		_, err = sshClient.Exec("sudo apt-get update")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// get the upgrade information from the host
		output, err := sshClient.Exec("sudo apt-get upgrade -y")
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to upgrade packages: %w", err).Error()), nil
		}

		// send the output to OpenAI to get a summary of what needs to be updated
		summary, err := summarizeAptUpgradeOutput(ctx, aiClient, output)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to summarize apt output: %w", err).Error()), nil
		}
		return mcp.NewToolResultText(summary), nil
	}
}

func summarizeAptUpgradeOutput(ctx context.Context, aiClient openai.Client, output []byte) (string, error) {
	// send the output to OpenAI to get a summary of what needs to be updated
	chat, err := aiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that summarizes the output of the 'apt-get upgrade -y' command."),
			openai.UserMessage(string(output)),
		},
	})
	if err != nil {
		return "", err
	}
	if len(chat.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI")
	}
	return chat.Choices[0].Message.Content, nil
}
