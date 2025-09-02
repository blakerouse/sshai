package tools

import (
	"github.com/blakerouse/sshai/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"
)

// Tool defines the interface that provides both the definition and the handler for a tool.
type Tool interface {
	Definition() mcp.Tool
	Handler(*storage.Engine, openai.Client) server.ToolHandlerFunc
}
