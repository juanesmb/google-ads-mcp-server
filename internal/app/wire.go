package app

import (
	"google-ads-mcp/internal/app/configs"
	repo "google-ads-mcp/internal/infrastructure/api/listadaccounts"
	searchcampaignsrepo "google-ads-mcp/internal/infrastructure/api/searchcampaigns"
	"google-ads-mcp/internal/infrastructure/auth"
	"google-ads-mcp/internal/infrastructure/http"
	"google-ads-mcp/internal/infrastructure/log/local"
	"google-ads-mcp/internal/tools/listadaccounts"
	"google-ads-mcp/internal/tools/searchcampaigns"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpServerInstructions string = "This is a Google Ads MCP server."

func initServer(configs configs.Configs) *mcp.Server {
	implementation := initImplementation()
	options := getMCPOptions()

	server := mcp.NewServer(implementation, options)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_ad_accounts",
		Description: "List Google Ads accounts",
	}, initListAdAccountsTool(configs).ListAdAccounts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_campaigns",
		Description: "Search Google Ads campaigns",
	}, initSearchCampaignsTool(configs).SearchCampaigns)

	return server
}

func initListAdAccountsTool(configs configs.Configs) *listadaccounts.Tool {
	httpClient := http.NewClient(nil)
	logger := local.NewLogger()

	// Use the service account JSON from Google Secret Manager
	tokenManager, err := auth.NewTokenManagerFromServiceAccount([]byte(configs.GoogleAdsConfig.ServiceAccountJSON), auth.GoogleAdsScope)
	if err != nil {
		panic("failed to initialize token manager: " + err.Error())
	}

	service := repo.NewService(httpClient, logger, tokenManager, configs.GoogleAdsConfig.CustomerID, configs.GoogleAdsConfig.DeveloperToken)

	return listadaccounts.NewListAdAccountsTool(service)
}

func initSearchCampaignsTool(configs configs.Configs) *searchcampaigns.Tool {
	httpClient := http.NewClient(nil)
	logger := local.NewLogger()

	// Use the service account JSON from Google Secret Manager
	tokenManager, err := auth.NewTokenManagerFromServiceAccount([]byte(configs.GoogleAdsConfig.ServiceAccountJSON), auth.GoogleAdsScope)
	if err != nil {
		panic("failed to initialize token manager: " + err.Error())
	}

	// loginCustomerID is the manager account ID from config (used in login-customer-id header)
	service := searchcampaignsrepo.NewService(httpClient, logger, tokenManager, configs.GoogleAdsConfig.CustomerID, configs.GoogleAdsConfig.DeveloperToken)

	return searchcampaigns.NewSearchCampaignsTool(service)
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
