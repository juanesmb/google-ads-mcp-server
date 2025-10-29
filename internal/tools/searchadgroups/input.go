package searchadgroups

// ToolInput defines the parameters accepted by the MCP tool.
type ToolInput struct {
	CustomerID      string   `json:"customer_id" validate:"required"`
	AdGroupIDs      []string `json:"ad_group_ids,omitempty"`
	AdGroupNames    []string `json:"ad_group_names,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	CampaignIDs     []string `json:"campaign_ids,omitempty"`
	CampaignNames   []string `json:"campaign_names,omitempty"`
	DateRangeStart  string   `json:"date_range_start,omitempty"`
	DateRangeEnd    string   `json:"date_range_end,omitempty"`
}

