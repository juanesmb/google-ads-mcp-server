package searchadgroups

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
	client         *infrahttp.Client
	logger         log.Logger
	tokenManager   auth.TokenProvider
	developerToken string
	loginCustomerID string
}

func NewService(client *infrahttp.Client, logger log.Logger, tokenManager auth.TokenProvider, loginCustomerID, developerToken string) *Service {
	return &Service{
		client:         client,
		logger:         logger,
		tokenManager:   tokenManager,
		developerToken: developerToken,
		loginCustomerID: loginCustomerID,
	}
}

func (s *Service) SearchAdGroups(ctx context.Context, filters Filters) (Result, error) {
	// Build endpoint URL with customer ID from filters
	endpoint, err := s.buildEndpoint(filters.CustomerID)
	if err != nil {
		return Result{}, fmt.Errorf("searchadgroups: invalid customer ID: %w", err)
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
		return Result{}, fmt.Errorf("searchadgroups: failed to get access token: %w", err)
	}

	headers := map[string]string{
		"Content-Type":      "application/json",
		"Authorization":     "Bearer " + accessToken,
		"developer-token":    s.developerToken,
		"login-customer-id":  s.loginCustomerID,
	}

	protoRequest := ProtoJSONRequest{Message: request}
	response, err := s.client.Post(ctx, endpoint, protoRequest, headers)
	if err != nil {
		return Result{}, fmt.Errorf("searchadgroups: executing request: %w", err)
	}

	if response.StatusCode >= 400 {
		return Result{}, fmt.Errorf("searchadgroups: api error status %d body %s", response.StatusCode, string(response.Body))
	}

	var protoResp services.SearchGoogleAdsResponse
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(response.Body, &protoResp); err != nil {
		return Result{}, fmt.Errorf("searchadgroups: unmarshal response: %w", err)
	}

	adGroups := make([]AdGroup, 0, len(protoResp.Results))
	for _, row := range protoResp.Results {
		adGroup := s.mapRowToAdGroup(row)
		if adGroup != nil {
			adGroups = append(adGroups, *adGroup)
		}
	}

	s.logger.Info(ctx, "google ads ad group search", map[string]string{
		"request_id": getHeaderValue(response.Headers, "request-id"),
	})

	return Result{
		AdGroups:          adGroups,
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
	// Build SELECT clause with ad group fields, campaign fields, and metrics
	fields := []string{
		"ad_group.id",
		"ad_group.resource_name",
		"ad_group.name",
		"ad_group.status",
		"ad_group.type",
		"campaign.id",
		"campaign.name",
		"campaign.resource_name",
		"metrics.clicks",
		"metrics.impressions",
		"metrics.ctr",
		"metrics.average_cpc",
		"metrics.cost_micros",
	}

	qb := gaql.NewQueryBuilder("ad_group").Select(fields...)

	// Default to LAST_7_DAYS if no date range is provided
	hasDateRange := filters.DateRangeStart != "" || filters.DateRangeEnd != ""
	if !hasDateRange {
		qb.Where("segments.date DURING LAST_7_DAYS")
	} else {
		if err := qb.WhereDateRange(filters.DateRangeStart, filters.DateRangeEnd); err != nil {
			return "", fmt.Errorf("searchadgroups: building query: %w", err)
		}
	}

	// Exclude REMOVED ad groups by default unless explicitly requested
	hasRemoved := false
	for _, status := range filters.Statuses {
		if strings.ToUpper(strings.TrimSpace(status)) == "REMOVED" {
			hasRemoved = true
			break
		}
	}
	if !hasRemoved {
		qb.Where("ad_group.status != REMOVED")
	}

	// Add ad group ID filters if present
	if len(filters.AdGroupIDs) > 0 {
		if err := qb.WhereAdGroupIDs(filters.AdGroupIDs); err != nil {
			return "", fmt.Errorf("searchadgroups: building query: %w", err)
		}
	}

	// Add ad group name filters if present
	if len(filters.AdGroupNames) > 0 {
		qb.WhereAdGroupNames(filters.AdGroupNames)
	}

	// Add status filters if present
	if len(filters.Statuses) > 0 {
		if err := qb.WhereAdGroupStatus(filters.Statuses); err != nil {
			return "", fmt.Errorf("searchadgroups: building query: %w", err)
		}
	}

	// Add campaign ID filters if present
	if len(filters.CampaignIDs) > 0 {
		if err := qb.WhereCampaignIDs(filters.CampaignIDs); err != nil {
			return "", fmt.Errorf("searchadgroups: building query: %w", err)
		}
	}

	// Add campaign name filters if present
	if len(filters.CampaignNames) > 0 {
		qb.WhereCampaignNames(filters.CampaignNames)
	}

	return qb.Build(), nil
}

func (s *Service) mapRowToAdGroup(row *services.GoogleAdsRow) *AdGroup {
	adGroupResource := row.GetAdGroup()
	if adGroupResource == nil {
		return nil
	}

	campaignResource := row.GetCampaign()
	metricsResource := row.GetMetrics()

	var campaignID, campaignName, campaignResourceName string
	if campaignResource != nil {
		campaignID = fmt.Sprintf("%d", campaignResource.GetId())
		campaignName = campaignResource.GetName()
		campaignResourceName = campaignResource.GetResourceName()
	}

	var clicks, impressions int64
	var ctr, averageCPC float64
	var costMicros int64

	if metricsResource != nil {
		clicks = metricsResource.GetClicks()
		impressions = metricsResource.GetImpressions()
		ctr = metricsResource.GetCtr()
		averageCPC = metricsResource.GetAverageCpc()
		costMicros = metricsResource.GetCostMicros()
	}

	return &AdGroup{
		ID:                   fmt.Sprintf("%d", adGroupResource.GetId()),
		ResourceName:         adGroupResource.GetResourceName(),
		Name:                 adGroupResource.GetName(),
		Status:               strings.ToLower(strings.TrimPrefix(adGroupResource.GetStatus().String(), "AD_GROUP_STATUS_")),
		Type:                 strings.ToLower(strings.TrimPrefix(adGroupResource.GetType().String(), "AD_GROUP_TYPE_")),
		CampaignID:           campaignID,
		CampaignName:         campaignName,
		CampaignResourceName: campaignResourceName,
		Metrics: AdGroupMetrics{
			Clicks:      clicks,
			Impressions: impressions,
			CTR:         ctr,
			AverageCPC:  int64(averageCPC * 1000000), // Convert to micros
			CostMicros:  costMicros,
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

