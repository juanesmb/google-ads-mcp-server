package searchcampaigns

import (
	"context"
	"encoding/json"
	"fmt"

	"google-ads-mcp/internal/infrastructure/api/searchcampaigns"

	"github.com/go-playground/validator/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validate = validator.New()

type Tool struct {
	service *searchcampaigns.Service
}

func NewSearchCampaignsTool(service *searchcampaigns.Service) *Tool {
	return &Tool{
		service: service,
	}
}

func (t *Tool) SearchCampaigns(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
	if req.Params.Arguments == nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchcampaigns: arguments payload is required")
	}

	// Validate input
	if err := validate.Struct(input); err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchcampaigns: validation error: %w", err)
	}

	filters := mapInputToFilters(input)

	result, err := t.service.SearchCampaigns(ctx, filters)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, err
	}

	output := ToolOutput{
		Campaigns:     mapCampaigns(result.Campaigns),
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalResultsCount,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchcampaigns: marshal response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, output, nil
}

func mapInputToFilters(input ToolInput) searchcampaigns.Filters {
	return searchcampaigns.Filters{
		CustomerID:     input.CustomerID,
		CampaignIDs:    input.CampaignIDs,
		CampaignNames:  input.CampaignNames,
		Statuses:       input.Statuses,
		DateRangeStart: input.DateRangeStart,
		DateRangeEnd:   input.DateRangeEnd,
	}
}

func mapCampaigns(campaigns []searchcampaigns.Campaign) []CampaignOutput {
	normalized := make([]CampaignOutput, 0, len(campaigns))
	for _, camp := range campaigns {
		normalized = append(normalized, CampaignOutput{
			ID:                     camp.ID,
			ResourceName:           camp.ResourceName,
			Name:                   camp.Name,
			Status:                 camp.Status,
			AdvertisingChannelType: camp.AdvertisingChannelType,
			BiddingStrategyType:    camp.BiddingStrategyType,
			BudgetAmountMicros:     camp.BudgetAmountMicros,
			OptimizationScore:      camp.OptimizationScore,
			Metrics: CampaignMetrics{
				Clicks:                             camp.Metrics.Clicks,
				Impressions:                        camp.Metrics.Impressions,
				CTR:                                camp.Metrics.CTR,
				AverageCPC:                         camp.Metrics.AverageCPC,
				CostMicros:                         camp.Metrics.CostMicros,
				Conversions:                        camp.Metrics.Conversions,
				ConversionsValue:                   camp.Metrics.ConversionsValue,
				CostPerConversion:                  camp.Metrics.CostPerConversion,
				ConversionRate:                     camp.Metrics.ConversionRate,
				AllConversions:                     camp.Metrics.AllConversions,
				AllConversionsValue:                camp.Metrics.AllConversionsValue,
				AllConversionsFromInteractionsRate: camp.Metrics.AllConversionsFromInteractionsRate,
				CostPerAllConversions:              camp.Metrics.CostPerAllConversions,
				Interactions:                       camp.Metrics.Interactions,
				EngagementRate:                     camp.Metrics.EngagementRate,
				SearchImpressionShare:              camp.Metrics.SearchImpressionShare,
				SearchRankLostImpressionShare:      camp.Metrics.SearchRankLostImpressionShare,
			},
		})
	}

	return normalized
}
