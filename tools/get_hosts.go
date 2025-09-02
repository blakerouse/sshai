package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"

	"github.com/blakerouse/sshai/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&GetHosts{})
}

// GetHosts is a tool that retrieves the list of hosts from the SSH configuration.
type GetHosts struct{}

// Definition returns the mcp.Tool definition.
func (c *GetHosts) Definition() mcp.Tool {
	return mcp.NewTool("get_hosts",
		mcp.WithDescription("Retrieves the list of hosts from the SSH configuration."),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *GetHosts) Handler(storageEngine *storage.Engine, aiClient openai.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hosts, err := storageEngine.List()
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to list hosts: %w", err).Error()), nil
		}
		list := make([]string, 0, len(hosts))
		for _, host := range hosts {
			list = append(list, host.Name)
		}
		return mcp.NewToolResultStructured(hosts, strings.Join(list, ", ")), nil
	}
}
