package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go/v2"

	"github.com/blakerouse/sshai/ssh"
	"github.com/blakerouse/sshai/storage"
	"github.com/blakerouse/sshai/utils"
)

func init() {
	// register the tool in the registry
	Registry.Register(&AddHost{})
}

// AddHost is a tool that adds a new host to the SSH configuration.
type AddHost struct{}

// Definition returns the mcp.Tool definition.
func (c *AddHost) Definition() mcp.Tool {
	return mcp.NewTool("add_host",
		mcp.WithDescription("Adds a new host to the SSH configuration."),
		mcp.WithString("ssh_connection_string",
			mcp.Required(),
			mcp.Description("SSH connection string"),
		),
		mcp.WithString("name_of_host",
			mcp.Description("Name of the host"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *AddHost) Handler(storageEngine *storage.Engine, aiClient openai.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshConnectionString, err := request.RequireString("ssh_connection_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshNameOfHost := request.GetString("name_of_host", "")

		clientInfo, err := ssh.NewClientInfo(sshNameOfHost, sshConnectionString)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshClient := ssh.NewClient(clientInfo)

		// connect over ssh
		err = sshClient.Connect()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer sshClient.Close()

		// from this point forward it is very much assuming linux
		// this really should be improved to do more checks to see if this macOS or Windows

		// gather the /etc/os-release information to determine the OS type
		osRelease, err := sshClient.Exec("cat /etc/os-release")
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to get output of /etc/os-release: %w", err).Error()), nil
		}

		// send the output to OpenAI to get a summary of what needs to be updated
		osInfo, err := getOSInfo(ctx, aiClient, osRelease)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to summarize OS information: %w", err).Error()), nil
		}

		// set the OS info and store it for usage later
		clientInfo.OS = *osInfo
		err = storageEngine.Set(*clientInfo)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to add host to storage: %w", err).Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("successfully added %s", clientInfo.Name)), nil
	}
}

func getOSInfo(ctx context.Context, aiClient openai.Client, output []byte) (*ssh.OSInfo, error) {
	// setup the schema to ensure the output is structured
	schema := utils.GenerateSchema[ssh.OSInfo]()
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "os_info",
		Description: openai.String("OS information"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	// perform the chat to compute the output into the desired format
	chat, err := aiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that summarizes the output of the 'cat /etc/os-release' command."),
			openai.UserMessage(string(output)),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		// model for structured output
		Model: openai.ChatModelGPT4o2024_08_06,
	})
	if err != nil {
		return nil, err
	}
	if len(chat.Choices) == 0 {
		return nil, errors.New("no choices returned from OpenAI")
	}

	// extract into a well-typed struct
	var osInfo ssh.OSInfo
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &osInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OS information: %w", err)
	}

	return &osInfo, nil
}
