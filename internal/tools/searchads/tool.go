package searchads

import (
	"context"
	"encoding/json"
	"fmt"

	"google-ads-mcp/internal/infrastructure/api/searchads"

	"github.com/go-playground/validator/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validate = validator.New()

type Tool struct {
	service *searchads.Service
}

func NewSearchAdsTool(service *searchads.Service) *Tool {
	return &Tool{
		service: service,
	}
}

func (t *Tool) SearchAds(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
	if req.Params.Arguments == nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchads: arguments payload is required")
	}

	// Validate input
	if err := validate.Struct(input); err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchads: validation error: %w", err)
	}

	filters := mapInputToFilters(input)

	result, err := t.service.SearchAds(ctx, filters)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, err
	}

	output := ToolOutput{
		Ads:           mapAds(result.Ads),
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalResultsCount,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("searchads: marshal response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, output, nil
}

func mapInputToFilters(input ToolInput) searchads.Filters {
	return searchads.Filters{
		CustomerID:      input.CustomerID,
		CampaignIDs:    input.CampaignIDs,
		CampaignNames:  input.CampaignNames,
		AdGroupIDs:     input.AdGroupIDs,
		AdGroupNames:   input.AdGroupNames,
		Statuses:       input.Statuses,
		AdTypes:        input.AdTypes,
		DateRangeStart: input.DateRangeStart,
		DateRangeEnd:   input.DateRangeEnd,
	}
}

func mapAds(ads []searchads.Ad) []AdOutput {
	normalized := make([]AdOutput, 0, len(ads))
	for _, ad := range ads {
		output := AdOutput{
			ID:                  ad.ID,
			ResourceName:        ad.ResourceName,
			Name:                ad.Name,
			Type:                ad.Type,
			Status:              ad.Status,
			FinalURLs:           ad.FinalURLs,
			ApprovalStatus:      ad.ApprovalStatus,
			CampaignID:          ad.CampaignID,
			CampaignName:        ad.CampaignName,
			CampaignResourceName: ad.CampaignResourceName,
			AdGroupID:           ad.AdGroupID,
			AdGroupName:         ad.AdGroupName,
			AdGroupResourceName: ad.AdGroupResourceName,
			Metrics: AdMetrics{
				Clicks:                             ad.Metrics.Clicks,
				Impressions:                        ad.Metrics.Impressions,
				CTR:                                ad.Metrics.CTR,
				AverageCPC:                         ad.Metrics.AverageCPC,
				CostMicros:                         ad.Metrics.CostMicros,
				Conversions:                        ad.Metrics.Conversions,
				ConversionsValue:                   ad.Metrics.ConversionsValue,
				CostPerConversion:                  ad.Metrics.CostPerConversion,
				ConversionRate:                     ad.Metrics.ConversionRate,
				AllConversions:                     ad.Metrics.AllConversions,
				AllConversionsValue:                ad.Metrics.AllConversionsValue,
				AllConversionsFromInteractionsRate: ad.Metrics.AllConversionsFromInteractionsRate,
				CostPerAllConversions:              ad.Metrics.CostPerAllConversions,
				Interactions:                       ad.Metrics.Interactions,
				EngagementRate:                     ad.Metrics.EngagementRate,
				SearchImpressionShare:              ad.Metrics.SearchImpressionShare,
				SearchRankLostImpressionShare:      ad.Metrics.SearchRankLostImpressionShare,
				VideoViews:                         ad.Metrics.VideoViews,
				VideoViewRate:                      ad.Metrics.VideoViewRate,
				AverageCPV:                        ad.Metrics.AverageCPV,
			},
		}

		if ad.ExpandedTextAd != nil {
			output.ExpandedTextAd = &ExpandedTextAd{
				HeadlinePart1: ad.ExpandedTextAd.HeadlinePart1,
				HeadlinePart2: ad.ExpandedTextAd.HeadlinePart2,
				HeadlinePart3: ad.ExpandedTextAd.HeadlinePart3,
				Description:   ad.ExpandedTextAd.Description,
				Description2:  ad.ExpandedTextAd.Description2,
				Path1:          ad.ExpandedTextAd.Path1,
				Path2:          ad.ExpandedTextAd.Path2,
			}
		}

		if ad.ResponsiveSearchAd != nil {
			output.ResponsiveSearchAd = &ResponsiveSearchAd{
				Headlines:    ad.ResponsiveSearchAd.Headlines,
				Descriptions: ad.ResponsiveSearchAd.Descriptions,
				Path1:         ad.ResponsiveSearchAd.Path1,
				Path2:         ad.ResponsiveSearchAd.Path2,
			}
		}

		if ad.CallOnlyAd != nil {
			output.CallOnlyAd = &CallOnlyAd{
				Headline1:            ad.CallOnlyAd.Headline1,
				Headline2:            ad.CallOnlyAd.Headline2,
				Description1:         ad.CallOnlyAd.Description1,
				Description2:         ad.CallOnlyAd.Description2,
				PhoneNumber:          ad.CallOnlyAd.PhoneNumber,
				CallTracked:          ad.CallOnlyAd.CallTracked,
				DisableCallConversion: ad.CallOnlyAd.DisableCallConversion,
			}
		}

		normalized = append(normalized, output)
	}

	return normalized
}

