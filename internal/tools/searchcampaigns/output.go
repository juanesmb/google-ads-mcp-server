package searchcampaigns

// ToolOutput captures the structured response returned to the MCP client.
type ToolOutput struct {
	Campaigns     []CampaignOutput `json:"campaigns"`
	NextPageToken string           `json:"next_page_token,omitempty"`
	TotalCount    int64            `json:"total_count"`
}

// CampaignOutput mirrors the normalized campaign representation returned to clients.
type CampaignOutput struct {
	ID                     string          `json:"id"`
	ResourceName           string          `json:"resource_name"`
	Name                   string          `json:"name"`
	Status                 string          `json:"status"`
	AdvertisingChannelType string          `json:"advertising_channel_type"`
	BiddingStrategyType    string          `json:"bidding_strategy_type"`
	BudgetAmountMicros     int64           `json:"budget_amount_micros"`
	OptimizationScore      float64         `json:"optimization_score"`
	Metrics                CampaignMetrics `json:"metrics"`
}

// CampaignMetrics represents metrics for a campaign.
type CampaignMetrics struct {
	Clicks                             int64   `json:"clicks"`
	Impressions                        int64   `json:"impressions"`
	CTR                                float64 `json:"ctr"`
	AverageCPC                         int64   `json:"average_cpc_micros"` // in micros
	CostMicros                         int64   `json:"cost_micros"`
	Conversions                        float64 `json:"conversions"`
	ConversionsValue                   float64 `json:"conversions_value"`
	CostPerConversion                  float64 `json:"cost_per_conversion"`
	ConversionRate                     float64 `json:"conversion_rate"`
	AllConversions                     float64 `json:"all_conversions"`
	AllConversionsValue                float64 `json:"all_conversions_value"`
	AllConversionsFromInteractionsRate float64 `json:"all_conversions_from_interactions_rate"`
	AllConversionsValuePerCost         float64 `json:"all_conversions_value_per_cost"`
	CostPerAllConversions              float64 `json:"cost_per_all_conversions"`
	Interactions                       int64   `json:"interactions"`
	EngagementRate                     float64 `json:"engagement_rate"`
	SearchImpressionShare              float64 `json:"search_impression_share"`
	SearchRankLostImpressionShare      float64 `json:"search_rank_lost_impression_share"`
}
