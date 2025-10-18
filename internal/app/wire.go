package app

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func initServer(configs Configs) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Google Ads MCP",
		Version: "v1.0.0",
		Title:   "Google Ads MCP server. Use 'system_guidelines' prompt first.",
	}, nil)

	return server
}
