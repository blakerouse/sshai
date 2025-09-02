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
	Registry.Register(&CheckForUpdates{})
}

// CheckForUpdates is a tool that checks for updates on a remote machine.
//
// At the moment this is limited to Ubuntu.
type CheckForUpdates struct{}

// Definition returns the mcp.Tool definition.
func (c *CheckForUpdates) Definition() mcp.Tool {
	return mcp.NewTool("check_for_updates",
		mcp.WithDescription("SSH into a remote machines and checks if the operating system is up to date. If it has any updates, it will return the list of packages that can be upgraded."),
		mcp.WithArray("name_of_hosts",
			mcp.Required(),
			mcp.Description("Name of the hosts"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *CheckForUpdates) Handler(storageEngine *storage.Engine, aiClient openai.Client) server.ToolHandlerFunc {
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
			// ensure that the repository is up to date
			// this is allowed to fail, because the user might not have permission or sudo
			// to perform this action (in that case we just use the repository as it is)
			_ = updateRepo(ctx, sshClient)

			// get the upgrade information from the host
			output, err := sshClient.Exec("apt list --upgradable")
			if err != nil {
				return "", fmt.Errorf("failed to get upgrade information: %w", err)
			}

			return string(output), nil
		})

		return mcp.NewToolResultStructuredOnly(result), nil
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
