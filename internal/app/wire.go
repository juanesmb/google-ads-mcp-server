package app

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpServerInstructions string = "This is a Google Ads MCP server."

func initServer(configs Configs) *mcp.Server {
	implementation := initImplementation()
	options := getMCPOptions()

	server := mcp.NewServer(implementation, options)

	return server
}

func initImplementation() *mcp.Implementation {
	return &mcp.Implementation{
		Name:    "Google Ads MCP",
		Version: "v1.0.0",
		Title:   "Google Ads MCP server.",
	}
}

func getMCPOptions() *mcp.ServerOptions {
	return &mcp.ServerOptions{
		Instructions: mcpServerInstructions,
	}
}
