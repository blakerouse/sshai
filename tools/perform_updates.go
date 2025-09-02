package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"

	"github.com/blakerouse/sshai/ssh"
	"github.com/blakerouse/sshai/storage"
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
		mcp.WithArray("name_of_hosts",
			mcp.Required(),
			mcp.Description("Name of the hosts"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *PerformUpdates) Handler(storageEngine *storage.Engine, aiClient openai.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshNameOfHosts, err := request.RequireStringSlice("name_of_hosts")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(sshNameOfHosts) == 0 {
			return mcp.NewToolResultError("no hosts provided"), nil
		}

		found, err := getHostsFromStorage(storageEngine, sshNameOfHosts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(found) == 0 {
			return mcp.NewToolResultError("no matching hosts found"), nil
		}

		// too make this smarter we should now check what type of machine this is
		// is it a linux machine? running ubuntu, debian, centos, etc.
		// is it a windows machine? running powershell, etc.

		// for now we just assume Ubuntu

		result := performTasksOnHosts(found, func(sshClient *ssh.Client) (string, error) {
			// sudo is required to update and upgrade
			_, err = sshClient.Exec("sudo apt-get update")
			if err != nil {
				return "", fmt.Errorf("failed to update packages: %w", err)
			}

			// get the upgrade information from the host
			output, err := sshClient.Exec("sudo apt-get upgrade -y")
			if err != nil {
				return "", fmt.Errorf("failed to upgrade packages: %w", err)
			}

			return string(output), nil
		})

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}
