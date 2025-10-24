package app

import (
	repo "google-ads-mcp/internal/infrastructure/api/listadaccounts"
	"google-ads-mcp/internal/infrastructure/http"
	"google-ads-mcp/internal/infrastructure/log/local"
	"google-ads-mcp/internal/tools/listadaccounts"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpServerInstructions string = "This is a Google Ads MCP server."

func initServer(configs Configs) *mcp.Server {
	implementation := initImplementation()
	options := getMCPOptions()

	server := mcp.NewServer(implementation, options)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_ad_accounts",
		Description: "List Google Ads accounts",
	}, initListAdAccountsTool(configs).ListAdAccounts)

	return server
}

func initListAdAccountsTool(configs Configs) *listadaccounts.Tool {
	httpClient := http.NewClient(nil)
	logger := local.NewLogger()
	service := repo.NewService(httpClient, logger, configs.GoogleAdsConfig.CustomerID, configs.GoogleAdsConfig.DeveloperToken)

	return listadaccounts.NewListAdAccountsTool(service)
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
