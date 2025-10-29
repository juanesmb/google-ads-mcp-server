package searchcampaigns

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google-ads-mcp/internal/infrastructure/api/gaql"
	"google-ads-mcp/internal/infrastructure/auth"
	infrahttp "google-ads-mcp/internal/infrastructure/http"
	"google-ads-mcp/internal/infrastructure/log"

	"github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	defaultBaseURL    = "https://googleads.googleapis.com"
	defaultAPIVersion = "v22"
)

type Service struct {
	client          *infrahttp.Client
	logger          log.Logger
	tokenManager    auth.TokenProvider
	developerToken  string
	loginCustomerID string
}

func NewService(client *infrahttp.Client, logger log.Logger, tokenManager auth.TokenProvider, loginCustomerID, developerToken string) *Service {
	return &Service{
		client:          client,
		logger:          logger,
		tokenManager:    tokenManager,
		developerToken:  developerToken,
		loginCustomerID: loginCustomerID,
	}
}

func (s *Service) SearchCampaigns(ctx context.Context, filters Filters) (Result, error) {
	// Build endpoint URL with customer ID from filters
	endpoint, err := s.buildEndpoint(filters.CustomerID)
	if err != nil {
		return Result{}, fmt.Errorf("searchcampaigns: invalid customer ID: %w", err)
	}

	query, err := s.buildQuery(filters)
	if err != nil {
		return Result{}, err
	}

	request := &services.SearchGoogleAdsRequest{
		Query: query,
	}

	accessToken, err := s.tokenManager.GetAccessToken(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("searchcampaigns: failed to get access token: %w", err)
	}

	headers := map[string]string{
		"Content-Type":      "application/json",
		"Authorization":     "Bearer " + accessToken,
		"developer-token":   s.developerToken,
		"login-customer-id": s.loginCustomerID,
	}

	protoRequest := ProtoJSONRequest{Message: request}
	response, err := s.client.Post(ctx, endpoint, protoRequest, headers)
	if err != nil {
		return Result{}, fmt.Errorf("searchcampaigns: executing request: %w", err)
	}

	if response.StatusCode >= 400 {
		return Result{}, fmt.Errorf("searchcampaigns: api error status %d body %s", response.StatusCode, string(response.Body))
	}

	var protoResp services.SearchGoogleAdsResponse
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(response.Body, &protoResp); err != nil {
		return Result{}, fmt.Errorf("searchcampaigns: unmarshal response: %w", err)
	}

	campaigns := make([]Campaign, 0, len(protoResp.Results))
	for _, row := range protoResp.Results {
		campaign := s.mapRowToCampaign(row)
		if campaign != nil {
			campaigns = append(campaigns, *campaign)
		}
	}

	s.logger.Info(ctx, "google ads campaign search", map[string]string{
		"request_id": getHeaderValue(response.Headers, "request-id"),
	})

	return Result{
		Campaigns:         campaigns,
		NextPageToken:     protoResp.GetNextPageToken(),
		TotalResultsCount: protoResp.GetTotalResultsCount(),
	}, nil
}

func (s *Service) buildEndpoint(customerID string) (string, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return "", err
	}

	// Validate customer ID format (should be numeric)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return "", fmt.Errorf("customer ID is required")
	}

	// Remove any "customers/" prefix if present
	customerID = strings.TrimPrefix(customerID, "customers/")

	version := strings.TrimPrefix(strings.TrimSpace(defaultAPIVersion), "/")
	path := strings.TrimSuffix(baseURL.Path, "/")
	path = fmt.Sprintf("%s/%s/customers/%s/googleAds:search", path, version, customerID)
	baseURL.Path = path

	return baseURL.String(), nil
}

func (s *Service) buildQuery(filters Filters) (string, error) {
	// Build SELECT clause with campaign fields and metrics
	fields := []string{
		"campaign.id",
		"campaign.resource_name",
		"campaign.name",
		"campaign.status",
		"campaign.advertising_channel_type",
		"campaign.bidding_strategy_type",
		"campaign_budget.amount_micros",
		"campaign.optimization_score",
		"metrics.clicks",
		"metrics.impressions",
		"metrics.ctr",
		"metrics.average_cpc",
		"metrics.cost_micros",
		"metrics.conversions",
		"metrics.conversions_value",
		"metrics.cost_per_conversion",
		"metrics.all_conversions",
		"metrics.all_conversions_value",
		"metrics.all_conversions_from_interactions_rate",
		"metrics.all_conversions_value_per_cost",
		"metrics.cost_per_all_conversions",
		"metrics.interactions",
		"metrics.engagement_rate",
		"metrics.search_impression_share",
		"metrics.search_rank_lost_impression_share",
	}

	qb := gaql.NewQueryBuilder("campaign").Select(fields...)

	// Exclude REMOVED campaigns by default unless explicitly requested
	hasRemoved := false
	for _, status := range filters.Statuses {
		if strings.ToUpper(strings.TrimSpace(status)) == "REMOVED" {
			hasRemoved = true
			break
		}
	}
	if !hasRemoved {
		qb.Where("campaign.status != REMOVED")
	}

	// Add campaign ID filters if present
	if len(filters.CampaignIDs) > 0 {
		if err := qb.WhereCampaignIDs(filters.CampaignIDs); err != nil {
			return "", fmt.Errorf("searchcampaigns: building query: %w", err)
		}
	}

	// Add campaign name filters if present
	if len(filters.CampaignNames) > 0 {
		qb.WhereCampaignNames(filters.CampaignNames)
	}

	// Add status filters if present
	if len(filters.Statuses) > 0 {
		if err := qb.WhereStatus(filters.Statuses); err != nil {
			return "", fmt.Errorf("searchcampaigns: building query: %w", err)
		}
	}

	// Add date range filter if provided
	if filters.DateRangeStart != "" || filters.DateRangeEnd != "" {
		if err := qb.WhereDateRange(filters.DateRangeStart, filters.DateRangeEnd); err != nil {
			return "", fmt.Errorf("searchcampaigns: building query: %w", err)
		}
	}

	return qb.Build(), nil
}

func (s *Service) mapRowToCampaign(row *services.GoogleAdsRow) *Campaign {
	campaignResource := row.GetCampaign()
	if campaignResource == nil {
		return nil
	}

	metricsResource := row.GetMetrics()
	campaignBudgetResource := row.GetCampaignBudget()

	var budgetAmountMicros int64
	if campaignBudgetResource != nil {
		budgetAmountMicros = campaignBudgetResource.GetAmountMicros()
	}

	var optimizationScore float64
	optScore := campaignResource.GetOptimizationScore()
	if optScore != 0 {
		optimizationScore = optScore
	}

	var clicks, impressions, interactions int64
	var ctr, averageCPC float64
	var costMicros int64
	var conversions, conversionsValue, costPerConversion float64
	var allConversions, allConversionsValue, allConversionsFromInteractionsRate float64
	var allConversionsValuePerCost, costPerAllConversions float64
	var engagementRate, searchImpressionShare, searchRankLostImpressionShare float64

	if metricsResource != nil {
		clicks = metricsResource.GetClicks()
		impressions = metricsResource.GetImpressions()
		ctr = metricsResource.GetCtr()
		averageCPC = metricsResource.GetAverageCpc()
		costMicros = metricsResource.GetCostMicros()
		conversions = metricsResource.GetConversions()
		conversionsValue = metricsResource.GetConversionsValue()
		costPerConversion = metricsResource.GetCostPerConversion()
		// ConversionRate is calculated: conversions / clicks * 100 (if clicks > 0)
		allConversions = metricsResource.GetAllConversions()
		allConversionsValue = metricsResource.GetAllConversionsValue()
		allConversionsFromInteractionsRate = metricsResource.GetAllConversionsFromInteractionsRate()
		allConversionsValuePerCost = metricsResource.GetAllConversionsValuePerCost()
		costPerAllConversions = metricsResource.GetCostPerAllConversions()
		interactions = metricsResource.GetInteractions()
		engagementRate = metricsResource.GetEngagementRate()
		searchImpressionShare = metricsResource.GetSearchImpressionShare()
		searchRankLostImpressionShare = metricsResource.GetSearchRankLostImpressionShare()
	}

	// Calculate conversion rate from conversions and clicks
	var conversionRate float64
	if clicks > 0 {
		conversionRate = (conversions / float64(clicks)) * 100.0
	}

	return &Campaign{
		ID:                     fmt.Sprintf("%d", campaignResource.GetId()),
		ResourceName:           campaignResource.GetResourceName(),
		Name:                   campaignResource.GetName(),
		Status:                 strings.ToLower(strings.TrimPrefix(campaignResource.GetStatus().String(), "CAMPAIGN_STATUS_")),
		AdvertisingChannelType: strings.ToLower(strings.TrimPrefix(campaignResource.GetAdvertisingChannelType().String(), "ADVERTISING_CHANNEL_TYPE_")),
		BiddingStrategyType:    strings.ToLower(strings.TrimPrefix(campaignResource.GetBiddingStrategyType().String(), "BIDDING_STRATEGY_TYPE_")),
		BudgetAmountMicros:     budgetAmountMicros,
		OptimizationScore:      optimizationScore,
		Metrics: CampaignMetrics{
			Clicks:                             clicks,
			Impressions:                        impressions,
			CTR:                                ctr,
			AverageCPC:                         int64(averageCPC * 1000000), // Convert to micros
			CostMicros:                         costMicros,
			Conversions:                        conversions,
			ConversionsValue:                   conversionsValue,
			CostPerConversion:                  costPerConversion,
			ConversionRate:                     conversionRate,
			AllConversions:                     allConversions,
			AllConversionsValue:                allConversionsValue,
			AllConversionsFromInteractionsRate: allConversionsFromInteractionsRate,
			AllConversionsValuePerCost:         allConversionsValuePerCost,
			CostPerAllConversions:              costPerAllConversions,
			Interactions:                       interactions,
			EngagementRate:                     engagementRate,
			SearchImpressionShare:              searchImpressionShare,
			SearchRankLostImpressionShare:      searchRankLostImpressionShare,
		},
	}
}

// ProtoJSONRequest wraps a protobuf message to provide custom JSON marshaling
type ProtoJSONRequest struct {
	Message proto.Message
}

// MarshalJSON implements json.Marshaler interface to use protobuf JSON marshaling
func (p ProtoJSONRequest) MarshalJSON() ([]byte, error) {
	return protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(p.Message)
}

func getHeaderValue(headers map[string][]string, key string) string {
	if values, exists := headers[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}
