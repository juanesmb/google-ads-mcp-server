package listadaccounts

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"google-ads-mcp/internal/infrastructure/api/listadaccounts"

	"github.com/go-playground/validator/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validate = validator.New()

type Tool struct {
	service *listadaccounts.Service
}

func NewListAdAccountsTool(service *listadaccounts.Service) *Tool {
	return &Tool{
		service: service,
	}
}

func (t *Tool) ListAdAccounts(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
	if req.Params.Arguments == nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("listadaccounts: arguments payload is required")
	}

	filters := mapInputToFilters(input)

	result, err := t.service.ListAccounts(ctx, filters)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, err
	}

	output := ToolOutput{
		Accounts:      mapAccounts(result.Accounts),
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalResultsCount,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return &mcp.CallToolResult{}, ToolOutput{}, fmt.Errorf("listadaccounts: marshal response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, output, nil
}

func mapInputToFilters(input ToolInput) listadaccounts.Filters {
	return listadaccounts.Filters{
		AccountIDs:   input.AccountIDs,
		AccountNames: input.AccountNames,
	}
}

func mapAccounts(accounts []listadaccounts.Account) []AccountOutput {
	normalized := make([]AccountOutput, 0, len(accounts))
	for _, acc := range accounts {
		normalized = append(normalized, AccountOutput{
			CustomerID:   acc.CustomerID,
			CustomerName: acc.CustomerName,
			CurrencyCode: acc.CurrencyCode,
			TimeZone:     acc.TimeZone,
			Status:       strings.ToLower(acc.Status),
			ResourceName: acc.ResourceName,
		})
	}

	return normalized
}

func compactStrings(values []string) []string {
	values = slices.Clone(values)
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	values = slices.DeleteFunc(values, func(s string) bool {
		return s == ""
	})

	slices.Sort(values)
	values = slices.Compact(values)
	return values
}
