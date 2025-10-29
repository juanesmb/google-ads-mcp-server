package searchcampaigns

// ToolInput defines the parameters accepted by the MCP tool.
type ToolInput struct {
	CustomerID      string   `json:"customer_id" validate:"required"`
	CampaignIDs     []string `json:"campaign_ids,omitempty"`
	CampaignNames   []string `json:"campaign_names,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	DateRangeStart  string   `json:"date_range_start,omitempty"`
	DateRangeEnd    string   `json:"date_range_end,omitempty"`
}

