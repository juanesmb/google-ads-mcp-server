package searchads

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

func (s *Service) SearchAds(ctx context.Context, filters Filters) (Result, error) {
	// Build endpoint URL with customer ID from filters
	endpoint, err := s.buildEndpoint(filters.CustomerID)
	if err != nil {
		return Result{}, fmt.Errorf("searchads: invalid customer ID: %w", err)
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
		return Result{}, fmt.Errorf("searchads: failed to get access token: %w", err)
	}

	headers := map[string]string{
		"Content-Type":      "application/json",
		"Authorization":     "Bearer " + accessToken,
		"developer-token":   s.developerToken,
		"login-customer-id":  s.loginCustomerID,
	}

	protoRequest := ProtoJSONRequest{Message: request}
	response, err := s.client.Post(ctx, endpoint, protoRequest, headers)
	if err != nil {
		return Result{}, fmt.Errorf("searchads: executing request: %w", err)
	}

	if response.StatusCode >= 400 {
		return Result{}, fmt.Errorf("searchads: api error status %d body %s", response.StatusCode, string(response.Body))
	}

	var protoResp services.SearchGoogleAdsResponse
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(response.Body, &protoResp); err != nil {
		return Result{}, fmt.Errorf("searchads: unmarshal response: %w", err)
	}

	ads := make([]Ad, 0, len(protoResp.Results))
	for _, row := range protoResp.Results {
		ad := s.mapRowToAd(row)
		if ad != nil {
			ads = append(ads, *ad)
		}
	}

	s.logger.Info(ctx, "google ads search", map[string]string{
		"request_id": getHeaderValue(response.Headers, "request-id"),
	})

	return Result{
		Ads:              ads,
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
	// Build SELECT clause with comprehensive ad fields
	fields := []string{
		// Core ad_group_ad fields
		"ad_group_ad.ad.id",
		"ad_group_ad.ad.resource_name",
		"ad_group_ad.ad.name",
		"ad_group_ad.ad.type",
		"ad_group_ad.ad.final_urls",
		"ad_group_ad.resource_name",
		"ad_group_ad.status",
		// Campaign context
		"campaign.id",
		"campaign.name",
		"campaign.resource_name",
		// Ad group context
		"ad_group.id",
		"ad_group.name",
		"ad_group.resource_name",
		// Policy
		"ad_group_ad.policy_summary.approval_status",
		// Expanded text ad fields
		"ad_group_ad.ad.expanded_text_ad.headline_part1",
		"ad_group_ad.ad.expanded_text_ad.headline_part2",
		"ad_group_ad.ad.expanded_text_ad.headline_part3",
		"ad_group_ad.ad.expanded_text_ad.description",
		"ad_group_ad.ad.expanded_text_ad.description2",
		"ad_group_ad.ad.expanded_text_ad.path1",
		"ad_group_ad.ad.expanded_text_ad.path2",
		// Responsive search ad fields
		"ad_group_ad.ad.responsive_search_ad.headlines",
		"ad_group_ad.ad.responsive_search_ad.descriptions",
		"ad_group_ad.ad.responsive_search_ad.path1",
		"ad_group_ad.ad.responsive_search_ad.path2",
		// Call-only ad fields (removed - only valid for specific campaign types)
		// Note: These fields cause UNRECOGNIZED_FIELD errors for Search campaigns
		// "ad_group_ad.ad.call_only_ad.headline1",
		// "ad_group_ad.ad.call_only_ad.headline2",
		// "ad_group_ad.ad.call_only_ad.description1",
		// "ad_group_ad.ad.call_only_ad.description2",
		// "ad_group_ad.ad.call_only_ad.phone_number",
		// "ad_group_ad.ad.call_only_ad.call_tracked",
		// "ad_group_ad.ad.call_only_ad.disable_call_conversion",
		// Comprehensive metrics
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
		"metrics.cost_per_all_conversions",
		"metrics.interactions",
		"metrics.engagement_rate",
		"metrics.search_impression_share",
		"metrics.search_rank_lost_impression_share",
		// Video metrics (removed - only valid for Video campaigns)
		// Note: These fields cause UNRECOGNIZED_FIELD errors for Search campaigns
		// "metrics.video_views",
		// "metrics.video_view_rate",
		// "metrics.average_cpv",
	}

	qb := gaql.NewQueryBuilder("ad_group_ad").Select(fields...)

	// Default to LAST_7_DAYS if no date range is provided (for metrics)
	hasDateRange := filters.DateRangeStart != "" || filters.DateRangeEnd != ""
	if !hasDateRange {
		qb.Where("segments.date DURING LAST_7_DAYS")
	} else {
		if err := qb.WhereDateRange(filters.DateRangeStart, filters.DateRangeEnd); err != nil {
			return "", fmt.Errorf("searchads: building query: %w", err)
		}
	}

	// Exclude REMOVED ads by default unless explicitly requested
	hasRemoved := false
	for _, status := range filters.Statuses {
		if strings.ToUpper(strings.TrimSpace(status)) == "REMOVED" {
			hasRemoved = true
			break
		}
	}
	if !hasRemoved {
		qb.Where("ad_group_ad.status != 'REMOVED'")
	}

	// Add campaign ID filters if present
	if len(filters.CampaignIDs) > 0 {
		if err := qb.WhereCampaignIDs(filters.CampaignIDs); err != nil {
			return "", fmt.Errorf("searchads: building query: %w", err)
		}
	}

	// Add campaign name filters if present
	if len(filters.CampaignNames) > 0 {
		qb.WhereCampaignNames(filters.CampaignNames)
	}

	// Add ad group ID filters if present
	if len(filters.AdGroupIDs) > 0 {
		if err := qb.WhereAdGroupIDs(filters.AdGroupIDs); err != nil {
			return "", fmt.Errorf("searchads: building query: %w", err)
		}
	}

	// Add ad group name filters if present
	if len(filters.AdGroupNames) > 0 {
		qb.WhereAdGroupNames(filters.AdGroupNames)
	}

	// Add status filters if present
	if len(filters.Statuses) > 0 {
		if err := qb.WhereAdGroupAdStatus(filters.Statuses); err != nil {
			return "", fmt.Errorf("searchads: building query: %w", err)
		}
	}

	// Add ad type filters if present
	if len(filters.AdTypes) > 0 {
		if err := qb.WhereAdTypes(filters.AdTypes); err != nil {
			return "", fmt.Errorf("searchads: building query: %w", err)
		}
	}

	return qb.Build(), nil
}

func (s *Service) mapRowToAd(row *services.GoogleAdsRow) *Ad {
	adGroupAdResource := row.GetAdGroupAd()
	if adGroupAdResource == nil {
		return nil
	}

	adResource := adGroupAdResource.GetAd()
	if adResource == nil {
		return nil
	}

	campaignResource := row.GetCampaign()
	adGroupResource := row.GetAdGroup()
	metricsResource := row.GetMetrics()
	policySummaryResource := adGroupAdResource.GetPolicySummary()

	var campaignID, campaignName, campaignResourceName string
	if campaignResource != nil {
		campaignID = fmt.Sprintf("%d", campaignResource.GetId())
		campaignName = campaignResource.GetName()
		campaignResourceName = campaignResource.GetResourceName()
	}

	var adGroupID, adGroupName, adGroupResourceName string
	if adGroupResource != nil {
		adGroupID = fmt.Sprintf("%d", adGroupResource.GetId())
		adGroupName = adGroupResource.GetName()
		adGroupResourceName = adGroupResource.GetResourceName()
	}

	var approvalStatus string
	if policySummaryResource != nil {
		approvalStatus = strings.ToLower(strings.TrimPrefix(policySummaryResource.GetApprovalStatus().String(), "POLICY_APPROVAL_STATUS_"))
	}

	// Extract ad type
	adType := strings.ToLower(strings.TrimPrefix(adResource.GetType().String(), "AD_TYPE_"))

	// Extract final URLs
	var finalURLs []string
	if adResource.GetFinalUrls() != nil {
		finalURLs = adResource.GetFinalUrls()
	}

	// Map metrics
	var clicks, impressions, interactions int64
	var ctr, averageCPC float64
	var costMicros, averageCPCInMicros int64
	var conversions, conversionsValue, costPerConversion float64
	var allConversions, allConversionsValue, allConversionsFromInteractionsRate float64
	var costPerAllConversions, engagementRate float64
	var searchImpressionShare, searchRankLostImpressionShare float64
	// Note: Video metrics (videoViews, videoViewRate, averageCPV) removed - not queried
	// as they cause UNRECOGNIZED_FIELD errors for Search campaigns

	if metricsResource != nil {
		clicks = metricsResource.GetClicks()
		impressions = metricsResource.GetImpressions()
		ctr = metricsResource.GetCtr()
		averageCPC = metricsResource.GetAverageCpc()
		costMicros = metricsResource.GetCostMicros()
		averageCPCInMicros = int64(averageCPC * 1000000)
		conversions = metricsResource.GetConversions()
		conversionsValue = metricsResource.GetConversionsValue()
		costPerConversion = metricsResource.GetCostPerConversion()
		allConversions = metricsResource.GetAllConversions()
		allConversionsValue = metricsResource.GetAllConversionsValue()
		allConversionsFromInteractionsRate = metricsResource.GetAllConversionsFromInteractionsRate()
		costPerAllConversions = metricsResource.GetCostPerAllConversions()
		interactions = metricsResource.GetInteractions()
		engagementRate = metricsResource.GetEngagementRate()
		searchImpressionShare = metricsResource.GetSearchImpressionShare()
		searchRankLostImpressionShare = metricsResource.GetSearchRankLostImpressionShare()
	}

	// Calculate conversion rate
	var conversionRate float64
	if clicks > 0 {
		conversionRate = (conversions / float64(clicks)) * 100.0
	}

	ad := &Ad{
		ID:                  fmt.Sprintf("%d", adResource.GetId()),
		ResourceName:        adGroupAdResource.GetResourceName(),
		Name:                adResource.GetName(),
		Type:                adType,
		Status:              strings.ToLower(strings.TrimPrefix(adGroupAdResource.GetStatus().String(), "AD_GROUP_AD_STATUS_")),
		FinalURLs:           finalURLs,
		ApprovalStatus:      approvalStatus,
		CampaignID:          campaignID,
		CampaignName:        campaignName,
		CampaignResourceName: campaignResourceName,
		AdGroupID:           adGroupID,
		AdGroupName:         adGroupName,
		AdGroupResourceName: adGroupResourceName,
		Metrics: AdMetrics{
			Clicks:                             clicks,
			Impressions:                        impressions,
			CTR:                                ctr,
			AverageCPC:                         averageCPCInMicros,
			CostMicros:                         costMicros,
			Conversions:                        conversions,
			ConversionsValue:                   conversionsValue,
			CostPerConversion:                  costPerConversion,
			ConversionRate:                     conversionRate,
			AllConversions:                     allConversions,
			AllConversionsValue:                allConversionsValue,
			AllConversionsFromInteractionsRate: allConversionsFromInteractionsRate,
			CostPerAllConversions:              costPerAllConversions,
			Interactions:                       interactions,
			EngagementRate:                     engagementRate,
			SearchImpressionShare:              searchImpressionShare,
			SearchRankLostImpressionShare:      searchRankLostImpressionShare,
			// VideoViews, VideoViewRate, AverageCPV set to 0 - not queried for compatibility
			// with Search campaigns that don't support video metrics
			VideoViews:                         0,
			VideoViewRate:                      0,
			AverageCPV:                         0,
		},
	}

	// Extract ad type specific fields dynamically
	switch adType {
	case "expanded_text_ad":
		expandedTextAd := adResource.GetExpandedTextAd()
		if expandedTextAd != nil {
			ad.ExpandedTextAd = &ExpandedTextAd{
				HeadlinePart1: expandedTextAd.GetHeadlinePart1(),
				HeadlinePart2: expandedTextAd.GetHeadlinePart2(),
				HeadlinePart3: expandedTextAd.GetHeadlinePart3(),
				Description:   expandedTextAd.GetDescription(),
				Description2:  expandedTextAd.GetDescription2(),
				Path1:         expandedTextAd.GetPath1(),
				Path2:         expandedTextAd.GetPath2(),
			}
		}
	case "responsive_search_ad":
		responsiveSearchAd := adResource.GetResponsiveSearchAd()
		if responsiveSearchAd != nil {
			var headlines, descriptions []string
			if responsiveSearchAd.GetHeadlines() != nil {
				for _, asset := range responsiveSearchAd.GetHeadlines() {
					text := asset.GetText()
					if text != "" {
						headlines = append(headlines, text)
					}
				}
			}
			if responsiveSearchAd.GetDescriptions() != nil {
				for _, asset := range responsiveSearchAd.GetDescriptions() {
					text := asset.GetText()
					if text != "" {
						descriptions = append(descriptions, text)
					}
				}
			}
			ad.ResponsiveSearchAd = &ResponsiveSearchAd{
				Headlines:    headlines,
				Descriptions: descriptions,
				Path1:        responsiveSearchAd.GetPath1(),
				Path2:        responsiveSearchAd.GetPath2(),
			}
		}
	case "call_only_ad":
		callAd := adResource.GetCallAd()
		if callAd != nil {
			ad.CallOnlyAd = &CallOnlyAd{
				Headline1:             callAd.GetHeadline1(),
				Headline2:             callAd.GetHeadline2(),
				Description1:          callAd.GetDescription1(),
				Description2:          callAd.GetDescription2(),
				PhoneNumber:           callAd.GetPhoneNumber(),
				CallTracked:           callAd.GetCallTracked(),
				DisableCallConversion: callAd.GetDisableCallConversion(),
			}
		}
	}

	return ad
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

