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
	Registry.Register(&CheckForUpdates{})
}

// CheckForUpdates is a tool that checks for updates on a remote machine.
//
// At the moment this is limited to Ubuntu.
type CheckForUpdates struct{}

// Definition returns the mcp.Tool definition.
func (c *CheckForUpdates) Definition() mcp.Tool {
	return mcp.NewTool("check_for_updates",
		mcp.WithDescription("SSH into a remote machine and checks if the operating system is up to date. If it has any updates, it will return the list of packages that can be upgraded."),
		mcp.WithString("ssh_connection_string",
			mcp.Required(),
			mcp.Description("SSH connection string"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *CheckForUpdates) Handler(aiClient openai.Client) server.ToolHandlerFunc {
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

		// ensure that the repository is up to date
		// this is allowed to fail, because the user might not have permission or sudo
		// to perform this action (in that case we just use the repository as it is)
		_ = updateRepo(ctx, sshClient)

		// get the upgrade information from the host
		output, err := sshClient.Exec("apt list --upgradable")
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to get upgrade information: %w", err).Error()), nil
		}

		// send the output to OpenAI to get a summary of what needs to be updated
		summary, err := summarizeAptOutput(ctx, aiClient, output)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to summarize apt output: %w", err).Error()), nil
		}
		return mcp.NewToolResultText(summary), nil
	}
}

func updateRepo(ctx context.Context, client *ssh.Client) error {
	_, err := client.Exec("apt-get update")
	if err == nil {
		// successful at updating
		return nil
	}
	// check for context cancellation before continuing
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// try it with sudo
	_, err = client.Exec("sudo apt-get update")
	if err == nil {
		return nil
	}
	// failed both times
	return err
}

func summarizeAptOutput(ctx context.Context, aiClient openai.Client, output []byte) (string, error) {
	// send the output to OpenAI to get a summary of what needs to be updated
	chat, err := aiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that summarizes the output of the 'apt list --upgradable' command."),
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
