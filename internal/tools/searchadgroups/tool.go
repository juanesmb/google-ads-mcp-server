package searchadgroups

import (
	"context"
	"encoding/json"
	"fmt"

	"google-ads-mcp/internal/infrastructure/api/searchadgroups"

	"github.com/go-playground/validator/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validate = validator.New()

type Tool struct {
	service *searchadgroups.Service
}

func NewSearchAdGroupsTool(service *searchadgroups.Service) *Tool {
	return &Tool{
		service: service,
	}
}

func (t *Tool) SearchAdGroups(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
	if req.Params.Arguments == nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchadgroups: arguments payload is required")
	}

	// Validate input
	if err := validate.Struct(input); err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchadgroups: validation error: %w", err)
	}

	filters := mapInputToFilters(input)

	result, err := t.service.SearchAdGroups(ctx, filters)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, err
	}

	output := ToolOutput{
		AdGroups:      mapAdGroups(result.AdGroups),
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalResultsCount,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchadgroups: marshal response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, output, nil
}

func mapInputToFilters(input ToolInput) searchadgroups.Filters {
	return searchadgroups.Filters{
		CustomerID:     input.CustomerID,
		AdGroupIDs:     input.AdGroupIDs,
		AdGroupNames:   input.AdGroupNames,
		Statuses:       input.Statuses,
		CampaignIDs:    input.CampaignIDs,
		CampaignNames:  input.CampaignNames,
		DateRangeStart: input.DateRangeStart,
		DateRangeEnd:   input.DateRangeEnd,
	}
}

func mapAdGroups(adGroups []searchadgroups.AdGroup) []AdGroupOutput {
	normalized := make([]AdGroupOutput, 0, len(adGroups))
	for _, ag := range adGroups {
		normalized = append(normalized, AdGroupOutput{
			ID:                   ag.ID,
			ResourceName:         ag.ResourceName,
			Name:                 ag.Name,
			Status:               ag.Status,
			Type:                 ag.Type,
			CampaignID:           ag.CampaignID,
			CampaignName:         ag.CampaignName,
			CampaignResourceName: ag.CampaignResourceName,
			Metrics: AdGroupMetrics{
				Clicks:      ag.Metrics.Clicks,
				Impressions: ag.Metrics.Impressions,
				CTR:         ag.Metrics.CTR,
				AverageCPC:  ag.Metrics.AverageCPC,
				CostMicros:  ag.Metrics.CostMicros,
			},
		})
	}

	return normalized
}

