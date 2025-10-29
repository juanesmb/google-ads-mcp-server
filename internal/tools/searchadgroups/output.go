package searchadgroups

// ToolOutput captures the structured response returned to the MCP client.
type ToolOutput struct {
	AdGroups      []AdGroupOutput `json:"ad_groups"`
	NextPageToken string          `json:"next_page_token,omitempty"`
	TotalCount    int64           `json:"total_count"`
}

// AdGroupOutput mirrors the normalized ad group representation returned to clients.
type AdGroupOutput struct {
	ID                   string          `json:"id"`
	ResourceName         string          `json:"resource_name"`
	Name                 string          `json:"name"`
	Status               string          `json:"status"`
	Type                 string          `json:"type"`
	CampaignID           string          `json:"campaign_id"`
	CampaignName         string          `json:"campaign_name"`
	CampaignResourceName string          `json:"campaign_resource_name"`
	Metrics              AdGroupMetrics  `json:"metrics"`
}

// AdGroupMetrics represents metrics for an ad group.
type AdGroupMetrics struct {
	Clicks      int64   `json:"clicks"`
	Impressions int64   `json:"impressions"`
	CTR         float64 `json:"ctr"`
	AverageCPC  int64   `json:"average_cpc_micros"` // in micros
	CostMicros  int64   `json:"cost_micros"`
}

