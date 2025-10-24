package listadaccounts

// ToolOutput captures the structured response returned to the MCP client.
type ToolOutput struct {
	Accounts      []AccountOutput `json:"accounts"`
	NextPageToken string          `json:"next_page_token,omitempty"`
	TotalCount    int64           `json:"total_count"`
}

// AccountOutput mirrors the normalized account representation returned to clients.
type AccountOutput struct {
	CustomerID   string `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CurrencyCode string `json:"currency_code"`
	TimeZone     string `json:"time_zone"`
	Status       string `json:"status"`
	ResourceName string `json:"resource_name"`
}
