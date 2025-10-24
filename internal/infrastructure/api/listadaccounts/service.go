package listadaccounts

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	endpoint       string
	developerToken string
	customerID     string
}

func NewService(client *infrahttp.Client, logger log.Logger, customerID, developerToken string) *Service {
	baseURL, _ := url.Parse(defaultBaseURL)
	version := strings.TrimPrefix(strings.TrimSpace(defaultAPIVersion), "/")
	path := strings.TrimSuffix(baseURL.Path, "/")
	path = fmt.Sprintf("%s/%s/customers/%s/googleAds:search", path, version, customerID)
	baseURL.Path = path

	return &Service{
		client:         client,
		logger:         logger,
		endpoint:       baseURL.String(),
		developerToken: developerToken,
		customerID:     customerID,
	}
}

func (s *Service) ListAccounts(ctx context.Context, filters Filters) (Result, error) {
	query, err := s.buildQuery(filters)
	if err != nil {
		return Result{}, err
	}

	request := &services.SearchGoogleAdsRequest{
		Query: query,
	}

	headers := map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer " + "accessToken",
		"developer-token": s.developerToken,
	}

	protoRequest := ProtoJSONRequest{Message: request}
	response, err := s.client.Post(ctx, s.endpoint, protoRequest, headers)
	if err != nil {
		return Result{}, fmt.Errorf("listadaccounts: executing request: %w", err)
	}

	if response.StatusCode >= 400 {
		return Result{}, fmt.Errorf("listadaccounts: api error status %d body %s", response.StatusCode, string(response.Body))
	}

	var protoResp services.SearchGoogleAdsResponse
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(response.Body, &protoResp); err != nil {
		return Result{}, fmt.Errorf("listadaccounts: unmarshal response: %w", err)
	}

	accounts := make([]Account, 0, len(protoResp.Results))
	for _, row := range protoResp.Results {
		custClient := row.GetCustomerClient()
		if custClient == nil {
			continue
		}

		accounts = append(accounts, Account{
			CustomerID:   custClient.GetClientCustomer(),
			CustomerName: custClient.GetDescriptiveName(),
			CurrencyCode: custClient.GetCurrencyCode(),
			TimeZone:     custClient.GetTimeZone(),
			Status:       custClient.GetStatus().String(),
			ResourceName: custClient.GetResourceName(),
		})
	}

	s.logger.Info(ctx, "google ads search", map[string]string{
		"request_id": getHeaderValue(response.Headers, "request-id"),
	})

	return Result{
		Accounts:          accounts,
		NextPageToken:     protoResp.GetNextPageToken(),
		TotalResultsCount: protoResp.GetTotalResultsCount(),
	}, nil
}

func (s *Service) buildQuery(filters Filters) (string, error) {
	qb := NewQueryBuilder("customer_client").
		Select(
			"customer_client.client_customer",
			"customer_client.descriptive_name",
			"customer_client.currency_code",
			"customer_client.time_zone",
			"customer_client.level",
			"customer_client.manager",
			"customer_client.status",
		).
		Where("customer_client.status = ENABLED").
		Where("customer_client.manager = false")

	// Add account ID filters if present
	if len(filters.AccountIDs) > 0 {
		if err := qb.WhereAccountIDs(filters.AccountIDs); err != nil {
			return "", fmt.Errorf("listadaccounts: building query: %w", err)
		}
	}

	// Add account name filters if present
	if len(filters.AccountNames) > 0 {
		qb.WhereAccountNames(filters.AccountNames)
	}

	return qb.Build(), nil
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
